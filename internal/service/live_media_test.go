package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKickSRSDisconnectsPublisher(t *testing.T) {
	t.Parallel()
	var deleted []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/clients":
			_, _ = io.WriteString(w, `{"clients":[
				{"id":"pub1","stream":"stream1","url":"/live/stream1","type":"fmle-publish","publish":true},
				{"id":"play1","stream":"stream1","url":"/live/stream1","type":"flv-play","publish":false},
				{"id":"other","stream":"stream2","url":"/live/stream2","type":"fmle-publish","publish":true}
			]}`)
		case r.Method == http.MethodDelete && len(r.URL.Path) > len("/api/v1/clients/"):
			deleted = append(deleted, r.URL.Path[len("/api/v1/clients/"):])
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	cfg := readySelfAuth()
	cfg.SRSAPIBase = srv.URL
	err := KickLiveStream(context.Background(), srv.Client(), cfg, "stream1")
	require.NoError(t, err)
	assert.Equal(t, []string{"pub1"}, deleted)
}

func TestKickSRSFindsPublisherPastFirstPage(t *testing.T) {
	t.Parallel()
	var deleted []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/clients":
			start, _ := strconv.Atoi(r.URL.Query().Get("start"))
			if start == 0 {
				var b strings.Builder
				b.WriteString(`{"code":0,"clients":[`)
				for i := 0; i < srsClientPageSize; i++ {
					if i > 0 {
						b.WriteByte(',')
					}
					fmt.Fprintf(&b, `{"id":"play%d","stream":"stream1","url":"/live/stream1","type":"flv-play","publish":false}`, i)
				}
				b.WriteString(`]}`)
				_, _ = io.WriteString(w, b.String())
				return
			}
			_, _ = io.WriteString(w, `{"code":0,"clients":[{"id":"pub1","stream":"stream1","url":"/live/stream1","type":"fmle-publish","publish":true}]}`)
		case r.Method == http.MethodDelete && len(r.URL.Path) > len("/api/v1/clients/"):
			deleted = append(deleted, r.URL.Path[len("/api/v1/clients/"):])
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	cfg := readySelfAuth()
	cfg.SRSAPIBase = srv.URL
	require.NoError(t, KickLiveStream(context.Background(), srv.Client(), cfg, "stream1"))
	assert.Equal(t, []string{"pub1"}, deleted)
}

func TestKickSkippedWhenSelfAuthOff(t *testing.T) {
	t.Parallel()
	cfg := LiveStreamConfig{Provider: constant.LiveProviderSelf, PushServer: "rtmp://x/live", PlayHLSTemplate: "https://x/{stream_key}.m3u8"}
	err := KickLiveStream(context.Background(), nil, cfg, "stream1")
	require.NoError(t, err)
}

func TestAliyunForbidIsPersistentAndResumeTreatsNotForbidden(t *testing.T) {
	t.Parallel()
	var actions []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		actions = append(actions, q.Get("Action"))
		if q.Get("Action") == "ForbidLiveStream" {
			assert.Equal(t, "no", q.Get("Oneshot"))
			assert.Equal(t, "publisher", q.Get("LiveStreamType"))
			assert.Equal(t, "stream1", q.Get("StreamName"))
			assert.NotEmpty(t, q.Get("Signature"))
			_, _ = io.WriteString(w, `{"RequestId":"r1"}`)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"Code":"InvalidOperation","Message":"The stream is not in the forbidden list."}`)
	}))
	t.Cleanup(srv.Close)

	cfg := LiveStreamConfig{
		Provider: constant.LiveProviderAliyun,
		Aliyun: LiveAliyunConfig{
			PushDomain: "push.example.com", PlayDomain: "play.example.com", AppName: "live",
			Region: "cn-shanghai", PushAuthKey: "p", PlayAuthKey: "l", AccessKeyID: "id", AccessKeySecret: "secret",
		},
	}
	client := rewriteHostClient(srv)
	require.NoError(t, KickLiveStream(context.Background(), client, cfg, "stream1"))
	require.NoError(t, ResumeLiveStream(context.Background(), client, cfg, "stream1"))
	assert.Equal(t, []string{"ForbidLiveStream", "ResumeLiveStream"}, actions)
}

func TestAliyunResumeRejectsRAMForbidden(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"Code":"Forbidden.RAM","Message":"You are not authorized to resume this stream."}`)
	}))
	t.Cleanup(srv.Close)

	cfg := LiveStreamConfig{
		Provider: constant.LiveProviderAliyun,
		Aliyun: LiveAliyunConfig{
			PushDomain: "push.example.com", PlayDomain: "play.example.com", AppName: "live",
			Region: "cn-shanghai", PushAuthKey: "p", PlayAuthKey: "l", AccessKeyID: "id", AccessKeySecret: "secret",
		},
	}
	err := ResumeLiveStream(context.Background(), rewriteHostClient(srv), cfg, "stream1")
	require.Error(t, err)
	assert.False(t, aliyunNotForbidden("Forbidden.RAM", "You are not authorized to resume this stream."))
}

func TestResumeForStartReforbidsWhenMarkFails(t *testing.T) {
	t.Parallel()
	var actions []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actions = append(actions, r.URL.Query().Get("Action"))
		_, _ = io.WriteString(w, `{"RequestId":"r1"}`)
	}))
	t.Cleanup(srv.Close)

	cfg := LiveStreamConfig{
		Provider: constant.LiveProviderAliyun,
		Aliyun: LiveAliyunConfig{
			PushDomain: "push.example.com", PlayDomain: "play.example.com", AppName: "live",
			Region: "cn-shanghai", PushAuthKey: "p", PlayAuthKey: "l", AccessKeyID: "id", AccessKeySecret: "secret",
		},
	}
	client := rewriteHostClient(srv)
	markErr := errors.New("db down")
	err := ResumeForStart(context.Background(), client, cfg, "stream1", func() error { return markErr })
	require.ErrorIs(t, err, markErr)
	assert.Equal(t, []string{"ResumeLiveStream", "ForbidLiveStream"}, actions)

	actions = nil
	err = ResumeForStart(context.Background(), client, cfg, "stream1", func() error { return errcode.LiveRoomLiving })
	require.ErrorIs(t, err, errcode.LiveRoomLiving)
	assert.Equal(t, []string{"ResumeLiveStream"}, actions)
}

func TestQiniuDisableUsesURLSafeBase64(t *testing.T) {
	t.Parallel()
	var gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		assert.Contains(t, r.Header.Get("Authorization"), "Qiniu ")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	cfg := LiveStreamConfig{
		Provider: constant.LiveProviderQiniu,
		Qiniu: LiveQiniuConfig{
			Hub: "hub", PushDomain: "push.example.com", PlayDomain: "play.example.com",
			AccessKey: "ak", SecretKey: "sk",
		},
	}
	client := rewriteHostClient(srv)
	require.NoError(t, KickLiveStream(context.Background(), client, cfg, "stream/name"))
	encoded := base64.URLEncoding.EncodeToString([]byte("stream/name"))
	assert.Contains(t, gotPath, "/v2/hubs/hub/streams/"+encoded+"/disabled")
	assert.JSONEq(t, `{"disabledTill":-1}`, gotBody)

	require.NoError(t, ResumeLiveStream(context.Background(), client, cfg, "stream/name"))
	assert.JSONEq(t, `{"disabledTill":0}`, gotBody)
}

func rewriteHostClient(srv *httptest.Server) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			clone := req.Clone(req.Context())
			clone.URL.Scheme = "http"
			clone.URL.Host = srv.Listener.Addr().String()
			clone.Host = clone.URL.Host
			return http.DefaultTransport.RoundTrip(clone)
		}),
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
