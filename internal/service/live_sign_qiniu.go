package service

import (
	"bytes"
	"context"
	"crypto/md5" // #nosec G501 -- Qiniu timestamp anti-leech requires MD5
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/pili"
)

const qiniuPiliHost = "pili.qiniuapi.com"

// qiniuTimestampURL is Qiniu's current timestamp anti-leech:
// sign = md5(key + escapedPath + hex(expire)), query sign & t.
func qiniuTimestampURL(rawURL, encryptKey string, expireUnix int64) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("qiniu play url: %w", err)
	}
	t := strconv.FormatInt(expireUnix, 16)
	plain := encryptKey + u.EscapedPath() + t
	sum := md5.Sum([]byte(plain)) //nolint:gosec // G401: Qiniu timestamp anti-leech specifies MD5
	sign := hex.EncodeToString(sum[:])
	q := url.Values{}
	q.Set("sign", sign)
	q.Set("t", t)
	if u.RawQuery == "" {
		return u.String() + "?" + q.Encode(), nil
	}
	return u.String() + "&" + q.Encode(), nil
}

func verifyQiniuTimestamp(rawURL, encryptKey string, now time.Time) error {
	u, err := url.Parse(rawURL)
	if err != nil || encryptKey == "" {
		return ErrLiveTokenInvalid
	}
	sign := u.Query().Get("sign")
	t := u.Query().Get("t")
	if sign == "" || t == "" {
		return ErrLiveTokenInvalid
	}
	exp, err := strconv.ParseInt(t, 16, 64)
	if err != nil || exp <= 0 {
		return ErrLiveTokenInvalid
	}
	if now.Unix() > exp {
		return ErrLiveTokenExpired
	}
	u.RawQuery = ""
	expect, err := qiniuTimestampURL(u.String(), encryptKey, exp)
	if err != nil {
		return ErrLiveTokenInvalid
	}
	eu, err := url.Parse(expect)
	if err != nil || !hmacEqual(eu.Query().Get("sign"), sign) {
		return ErrLiveTokenInvalid
	}
	return nil
}

func hmacEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func issueQiniu(cfg LiveStreamConfig, streamName string, now time.Time) (LiveMediaURLs, error) {
	pushExp := now.Add(time.Duration(cfg.pushTTL()) * time.Second).Unix()
	playExp := now.Add(time.Duration(cfg.playTTL()) * time.Second).Unix()
	publish := pili.RTMPPublishURL(cfg.Qiniu.Hub, cfg.Qiniu.PushDomain, streamName)
	signedPush, err := pili.SignPublishURL(publish, pili.SignPublishURLArgs{
		SecurityType: pili.SecurityTypeExpirySK,
		ExpireAt:     pushExp,
		AccessKey:    cfg.Qiniu.AccessKey,
		SecretKey:    cfg.Qiniu.SecretKey,
	})
	if err != nil {
		return LiveMediaURLs{}, fmt.Errorf("qiniu push sign: %w", err)
	}
	server, streamKey, err := splitQiniuPush(signedPush, cfg.Qiniu.Hub)
	if err != nil {
		return LiveMediaURLs{}, err
	}
	hlsRaw := pili.HLSPlayURL(cfg.Qiniu.Hub, cfg.Qiniu.PlayDomain, streamName)
	flvRaw := pili.HDLPlayURL(cfg.Qiniu.Hub, cfg.Qiniu.PlayDomain, streamName)
	// Playback anti-leech key is the account SecretKey; the console key must match it.
	hls, err := qiniuTimestampURL(hlsRaw, cfg.Qiniu.SecretKey, playExp)
	if err != nil {
		return LiveMediaURLs{}, err
	}
	flv, err := qiniuTimestampURL(flvRaw, cfg.Qiniu.SecretKey, playExp)
	if err != nil {
		return LiveMediaURLs{}, err
	}
	return LiveMediaURLs{
		PushServer:    server,
		StreamKey:     streamKey,
		PlayHLS:       hls,
		PlayFLV:       flv,
		PushExpiresAt: time.Unix(pushExp, 0),
		PlayExpiresAt: time.Unix(playExp, 0),
	}, nil
}

func splitQiniuPush(signed, hub string) (server, streamKey string, err error) {
	u, err := url.Parse(signed)
	if err != nil {
		return "", "", fmt.Errorf("qiniu push url: %w", err)
	}
	prefix := "/" + hub + "/"
	if !strings.HasPrefix(u.Path, prefix) {
		return "", "", fmt.Errorf("qiniu push url path %s", u.Path)
	}
	name := strings.TrimPrefix(u.Path, prefix)
	server = u.Scheme + "://" + u.Host + "/" + hub
	if u.RawQuery == "" {
		return server, name, nil
	}
	return server, name + "?" + u.RawQuery, nil
}

func qiniuStreamDisabled(ctx context.Context, client *http.Client, cfg LiveQiniuConfig, streamName string, disabledTill int) error {
	if client == nil {
		return fmt.Errorf("qiniu disable: http client is nil")
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(streamName))
	rawURL := fmt.Sprintf("https://%s/v2/hubs/%s/streams/%s/disabled", qiniuPiliHost, cfg.Hub, encoded)
	payload := fmt.Appendf(nil, `{"disabledTill":%d}`, disabledTill)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("qiniu disable: %w", err)
	}
	req.Host = req.URL.Host
	req.Header.Set("Content-Type", "application/json")
	mac := auth.New(cfg.AccessKey, cfg.SecretKey)
	token, err := mac.SignRequestV2(req)
	if err != nil {
		return fmt.Errorf("qiniu disable sign: %w", err)
	}
	req.Header.Set("Authorization", "Qiniu "+token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("qiniu disable: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	blob := strings.ToLower(string(body))
	missing := resp.StatusCode == http.StatusNotFound || resp.StatusCode == 612 ||
		strings.Contains(blob, "not found") || strings.Contains(blob, "no such") || strings.Contains(blob, "not exist")
	if missing {
		return nil
	}
	if disabledTill == 0 && (strings.Contains(blob, "not disabled") || strings.Contains(blob, "not forbidden")) {
		return nil
	}
	return fmt.Errorf("qiniu disable: status %d body %s", resp.StatusCode, string(body))
}
