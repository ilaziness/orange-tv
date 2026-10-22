package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	liveTokenKindPush = "push"
	liveTokenKindPlay = "play"
)

var (
	// ErrLiveTokenInvalid is a bad signature, secret, or token shape.
	ErrLiveTokenInvalid = errors.New("live token invalid")
	// ErrLiveTokenExpired is a token past its exp timestamp.
	ErrLiveTokenExpired = errors.New("live token expired")
)

// signSelfToken builds token=exp-rand-sign. The MAC covers kind, stream name, exp, and rand.
func signSelfToken(secret, kind, streamName, randToken string, exp int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = fmt.Fprintf(mac, "%s\n%s\n%d\n%s", kind, streamName, exp, randToken)
	sign := hex.EncodeToString(mac.Sum(nil))
	return strconv.FormatInt(exp, 10) + "-" + randToken + "-" + sign
}

func verifySelfToken(secret, kind, streamName, token string, now time.Time) error {
	parts := strings.Split(token, "-")
	if len(parts) != 3 || secret == "" || streamName == "" {
		return ErrLiveTokenInvalid
	}
	exp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || exp <= 0 || parts[1] == "" || parts[2] == "" {
		return ErrLiveTokenInvalid
	}
	if now.Unix() > exp {
		return ErrLiveTokenExpired
	}
	expect := signSelfToken(secret, kind, streamName, parts[1], exp)
	if !hmac.Equal([]byte(expect), []byte(parts[0]+"-"+parts[1]+"-"+parts[2])) {
		return ErrLiveTokenInvalid
	}
	return nil
}

func issueSelf(cfg LiveStreamConfig, streamName string, now time.Time) (LiveMediaURLs, error) {
	randToken, err := randomHex(8)
	if err != nil {
		return LiveMediaURLs{}, err
	}
	pushExp := now.Add(time.Duration(cfg.pushTTL()) * time.Second).Unix()
	playExp := now.Add(time.Duration(cfg.playTTL()) * time.Second).Unix()
	pushToken := signSelfToken(cfg.SignSecret, liveTokenKindPush, streamName, randToken, pushExp)
	playRand, err := randomHex(8)
	if err != nil {
		return LiveMediaURLs{}, err
	}
	playToken := signSelfToken(cfg.SignSecret, liveTokenKindPlay, streamName, playRand, playExp)
	hls := appendToken(ApplyStreamKeyTemplate(cfg.PlayHLSTemplate, streamName), "token", playToken)
	flv := appendToken(ApplyStreamKeyTemplate(cfg.PlayFLVTemplate, streamName), "token", playToken)
	return LiveMediaURLs{
		PushServer:    cfg.PushServer,
		StreamKey:     streamName + "?token=" + url.QueryEscape(pushToken),
		PlayHLS:       hls,
		PlayFLV:       flv,
		PushExpiresAt: time.Unix(pushExp, 0),
		PlayExpiresAt: time.Unix(playExp, 0),
	}, nil
}

func appendToken(rawURL, key, value string) string {
	if rawURL == "" {
		return ""
	}
	escaped := url.QueryEscape(value)
	if strings.Contains(rawURL, "?") {
		return rawURL + "&" + key + "=" + escaped
	}
	return rawURL + "?" + key + "=" + escaped
}

func tokenFromQuery(raw, key string) string {
	raw = strings.TrimPrefix(strings.TrimSpace(raw), "?")
	if raw == "" {
		return ""
	}
	q, err := url.ParseQuery(raw)
	if err != nil {
		return ""
	}
	return q.Get(key)
}
