package service

import (
	"net/url"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
)

// VerifyLiveHookSecret checks the shared SRS/Nginx callback secret before any room lookup.
func VerifyLiveHookSecret(cfg LiveStreamConfig, hookSecret string) error {
	if !secretMatch(cfg.HookSecret, hookSecret) {
		return ErrLiveTokenInvalid
	}
	return nil
}

// DecideSRSHook allows or rejects an SRS http_hook.
// on_unpublish is logged by the caller and never ends a session.
// A nil error means SRS should accept the event.
func DecideSRSHook(cfg LiveStreamConfig, hookSecret, action, app, stream, param string, room *model.LiveRooms, now time.Time) error {
	if err := VerifyLiveHookSecret(cfg, hookSecret); err != nil {
		return err
	}
	switch action {
	case "on_unpublish":
		return nil
	case "on_publish", "on_play":
	default:
		return ErrLiveTokenInvalid
	}
	if !cfg.SelfAuthOn() {
		return ErrLiveTokenInvalid
	}
	expectApp, err := pushApp(cfg.PushServer)
	if err != nil || expectApp == "" || app != expectApp {
		return ErrLiveTokenInvalid
	}
	if room == nil || room.StreamKey != stream || room.Status != constant.LiveRoomLiving {
		return ErrLiveTokenInvalid
	}
	kind := liveTokenKindPlay
	if action == "on_publish" {
		kind = liveTokenKindPush
	}
	token := tokenFromQuery(param, "token")
	return verifySelfToken(cfg.SignSecret, kind, stream, token, now)
}

// DecidePlayAuth checks an Nginx auth_request for an m3u8 or flv URL.
// Slice requests (.ts) are rejected here; they must not be sent to this API.
func DecidePlayAuth(cfg LiveStreamConfig, hookSecret, originalURI string, lookup func(stream string) *model.LiveRooms, now time.Time) error {
	if err := VerifyLiveHookSecret(cfg, hookSecret); err != nil {
		return err
	}
	if !cfg.SelfAuthOn() {
		return ErrLiveTokenInvalid
	}
	u, err := parseOriginalURI(originalURI)
	if err != nil {
		return ErrLiveTokenInvalid
	}
	app, stream, ext, err := splitPlayPath(u.Path)
	if err != nil {
		return err
	}
	if ext == ".ts" {
		return ErrLiveTokenInvalid
	}
	expectApp, err := pushApp(cfg.PushServer)
	if err != nil || expectApp == "" || app != expectApp {
		return ErrLiveTokenInvalid
	}
	room := lookup(stream)
	if room == nil || room.StreamKey != stream || room.Status != constant.LiveRoomLiving {
		return ErrLiveTokenInvalid
	}
	token := u.Query().Get("token")
	return verifySelfToken(cfg.SignSecret, liveTokenKindPlay, stream, token, now)
}

func secretMatch(stored, given string) bool {
	if stored == "" || given == "" || len(stored) != len(given) {
		return false
	}
	return hmacEqual(stored, given)
}

func parseOriginalURI(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrLiveTokenInvalid
	}
	u, err := url.Parse(raw)
	if err != nil || u.Path == "" {
		return nil, ErrLiveTokenInvalid
	}
	return u, nil
}

func splitPlayPath(path string) (app, stream, ext string, err error) {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", "", ErrLiveTokenInvalid
	}
	file := parts[len(parts)-1]
	lower := strings.ToLower(file)
	switch {
	case strings.HasSuffix(lower, ".m3u8"):
		ext = ".m3u8"
		file = file[:len(file)-len(".m3u8")]
	case strings.HasSuffix(lower, ".flv"):
		ext = ".flv"
		file = file[:len(file)-len(".flv")]
	case strings.HasSuffix(lower, ".ts"):
		return "", "", ".ts", ErrLiveTokenInvalid
	default:
		return "", "", "", ErrLiveTokenInvalid
	}
	if file == "" {
		return "", "", "", ErrLiveTokenInvalid
	}
	app = parts[0]
	if len(parts) == 2 {
		stream = file
		return app, stream, ext, nil
	}
	mid := strings.Join(parts[1:len(parts)-1], "/")
	stream = mid + "/" + file
	return app, stream, ext, nil
}
