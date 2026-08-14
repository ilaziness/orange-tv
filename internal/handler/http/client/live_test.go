package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/response"
	"github.com/stretchr/testify/assert"
)

func TestLiveHandler_Play_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/live/play/abc", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h := NewLiveHandler(&stubLiveService{}, &stubLiveProxyService{})
	h.Play(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLiveHandler_Play_NegativeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/live/play/-1", nil)
	c.Params = gin.Params{{Key: "id", Value: "-1"}}

	h := NewLiveHandler(&stubLiveService{}, &stubLiveProxyService{})
	h.Play(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLiveHandler_Play_ProxyError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/live/play/1?u=invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h := NewLiveHandler(&stubLiveService{}, &stubLiveProxyService{proxyErr: errcode.ParamError})
	h.Play(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLiveHandler_List_StreamURLJSONContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		withStream   bool
		expectStream bool
	}{
		{"web no stream_url field", false, false},
		{"app includes stream_url field", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/client/v1/live", nil)

			h := NewLiveHandler(&stubLiveService{withStream: tt.withStream}, &stubLiveProxyService{})
			h.List(c)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp struct {
				Data response.PageData `json:"data"`
			}
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

			listJSON, err := json.Marshal(resp.Data.List)
			assert.NoError(t, err)
			body := string(listJSON)

			if tt.expectStream {
				assert.Contains(t, body, `"stream_url"`, "app response should include stream_url")
				assert.Contains(t, body, "http://example.com/cctv1.m3u8")
			} else {
				assert.NotContains(t, body, "stream_url", "web response must NOT include stream_url")
			}
		})
	}
}

type stubLiveService struct {
	withStream bool
}

func (s *stubLiveService) List(ctx context.Context, req *clientdto.LiveListRequest) ([]clientdto.LiveChannelItem, int, error) {
	item := clientdto.LiveChannelItem{
		ID:          1,
		Name:        "CCTV1",
		Category:    "央视",
		Description: "测试频道",
		Format:      "hls",
	}
	if s.withStream {
		item.StreamURL = "http://example.com/cctv1.m3u8"
	}
	return []clientdto.LiveChannelItem{item}, 1, nil
}

func (s *stubLiveService) GetStreamURL(ctx context.Context, id uint32) (string, error) {
	return "", nil
}

type stubLiveProxyService struct {
	proxyErr error
}

func (s *stubLiveProxyService) Proxy(c *gin.Context, channelID uint32, segURL string) error {
	return s.proxyErr
}
