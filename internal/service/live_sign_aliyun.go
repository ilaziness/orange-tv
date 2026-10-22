package service

import (
	"context"
	"crypto/hmac"
	"crypto/md5"  // #nosec G501 -- Aliyun URL auth type A requires MD5
	"crypto/sha1" // #nosec G505 -- Aliyun OpenAPI signs requests with HMAC-SHA1
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
)

// aliyunAuthKey is Aliyun URL auth type A: auth_key=timestamp-rand-uid-md5(URI-timestamp-rand-uid-PrivateKey).
func aliyunAuthKey(uriPath, privateKey, randToken, uid string, exp int64) string {
	if !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	ts := strconv.FormatInt(exp, 10)
	plain := uriPath + "-" + ts + "-" + randToken + "-" + uid + "-" + privateKey
	sum := md5.Sum([]byte(plain)) //nolint:gosec // G401: Aliyun URL auth type A specifies MD5
	return ts + "-" + randToken + "-" + uid + "-" + hex.EncodeToString(sum[:])
}

func verifyAliyunAuthKey(uriPath, privateKey, authKey string, now time.Time) error {
	parts := strings.Split(authKey, "-")
	if len(parts) != 4 || privateKey == "" {
		return ErrLiveTokenInvalid
	}
	exp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || exp <= 0 {
		return ErrLiveTokenInvalid
	}
	if now.Unix() > exp {
		return ErrLiveTokenExpired
	}
	expect := aliyunAuthKey(uriPath, privateKey, parts[1], parts[2], exp)
	if !hmac.Equal([]byte(expect), []byte(authKey)) {
		return ErrLiveTokenInvalid
	}
	return nil
}

func issueAliyun(cfg LiveStreamConfig, streamName string, now time.Time) (LiveMediaURLs, error) {
	randToken, err := randomHex(8)
	if err != nil {
		return LiveMediaURLs{}, err
	}
	playRand, err := randomHex(8)
	if err != nil {
		return LiveMediaURLs{}, err
	}
	app := cfg.Aliyun.AppName
	pushExp := now.Add(time.Duration(cfg.pushTTL()) * time.Second).Unix()
	playExp := now.Add(time.Duration(cfg.playTTL()) * time.Second).Unix()
	pushPath := "/" + app + "/" + streamName
	hlsPath := "/" + app + "/" + streamName + ".m3u8"
	flvPath := "/" + app + "/" + streamName + ".flv"
	pushKey := aliyunAuthKey(pushPath, cfg.Aliyun.PushAuthKey, randToken, "0", pushExp)
	hlsKey := aliyunAuthKey(hlsPath, cfg.Aliyun.PlayAuthKey, playRand, "0", playExp)
	flvRand, err := randomHex(8)
	if err != nil {
		return LiveMediaURLs{}, err
	}
	flvKey := aliyunAuthKey(flvPath, cfg.Aliyun.PlayAuthKey, flvRand, "0", playExp)
	return LiveMediaURLs{
		PushServer:    "rtmp://" + cfg.Aliyun.PushDomain + "/" + app,
		StreamKey:     streamName + "?auth_key=" + pushKey,
		PlayHLS:       "https://" + cfg.Aliyun.PlayDomain + hlsPath + "?auth_key=" + hlsKey,
		PlayFLV:       "https://" + cfg.Aliyun.PlayDomain + flvPath + "?auth_key=" + flvKey,
		PushExpiresAt: time.Unix(pushExp, 0),
		PlayExpiresAt: time.Unix(playExp, 0),
	}, nil
}

type aliyunRPCBody struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

func aliyunCall(ctx context.Context, client *http.Client, cfg LiveAliyunConfig, action string, extra map[string]string) error {
	if client == nil {
		return fmt.Errorf("aliyun %s: http client is nil", action)
	}
	region := cfg.Region
	if region == "" {
		region = constant.LiveAliyunRegionDefault
	}
	nonce, err := randomHex(16)
	if err != nil {
		return err
	}
	params := map[string]string{
		"Format":           "JSON",
		"Version":          "2016-11-01",
		"AccessKeyId":      cfg.AccessKeyID,
		"SignatureMethod":  "HMAC-SHA1",
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureVersion": "1.0",
		"SignatureNonce":   nonce,
		"Action":           action,
		"RegionId":         region,
		"DomainName":       cfg.PushDomain,
		"AppName":          cfg.AppName,
		"StreamName":       extra["StreamName"],
		"LiveStreamType":   "publisher",
	}
	for k, v := range extra {
		params[k] = v
	}
	canonical := aliCanonical(params)
	stringToSign := "GET&" + aliPercentEncode("/") + "&" + aliPercentEncode(canonical)
	mac := hmac.New(sha1.New, []byte(cfg.AccessKeySecret+"&"))
	_, _ = mac.Write([]byte(stringToSign))
	params["Signature"] = base64.StdEncoding.EncodeToString(mac.Sum(nil))

	endpoint := "https://live." + region + ".aliyuncs.com/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+aliCanonical(params), nil)
	if err != nil {
		return fmt.Errorf("aliyun %s: %w", action, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("aliyun %s: %w", action, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	var parsed aliyunRPCBody
	_ = json.Unmarshal(body, &parsed)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 && parsed.Code == "" {
		return nil
	}
	if action == "ResumeLiveStream" && aliyunNotForbidden(parsed.Code, parsed.Message) {
		return nil
	}
	return fmt.Errorf("aliyun %s: status %d code %s message %s body %s", action, resp.StatusCode, parsed.Code, parsed.Message, string(body))
}

func aliyunNotForbidden(code, message string) bool {
	blob := strings.ToLower(code + " " + message)
	// Match Aliyun's "stream is not forbidden" result only.
	// A RAM Forbidden error also contains "not" and "forbid" and must stay a failure.
	return strings.Contains(blob, "not forbidden") ||
		strings.Contains(blob, "not in the forbidden") ||
		strings.Contains(blob, "notforbid")
}

func aliCanonical(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, aliPercentEncode(k)+"="+aliPercentEncode(params[k]))
	}
	return strings.Join(parts, "&")
}

func aliPercentEncode(s string) string {
	e := url.QueryEscape(s)
	e = strings.ReplaceAll(e, "+", "%20")
	e = strings.ReplaceAll(e, "*", "%2A")
	e = strings.ReplaceAll(e, "%7E", "~")
	return e
}
