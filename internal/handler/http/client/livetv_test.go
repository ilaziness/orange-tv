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

func TestLiveTVHandler_Play_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/livetv/play/abc", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}

	h := NewLiveTVHandler(&stubLiveTVService{}, &stubLiveTVProxyService{})
	h.Play(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLiveTVHandler_Play_NegativeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/livetv/play/-1", nil)
	c.Params = gin.Params{{Key: "id", Value: "-1"}}

	h := NewLiveTVHandler(&stubLiveTVService{}, &stubLiveTVProxyService{})
	h.Play(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLiveTVHandler_Play_ProxyError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/livetv/play/1?u=invalid", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h := NewLiveTVHandler(&stubLiveTVService{}, &stubLiveTVProxyService{proxyErr: errcode.ParamError})
	h.Play(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLiveTVHandler_List_StreamURLJSONContract(t *testing.T) {
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
			c.Request = httptest.NewRequest(http.MethodGet, "/api/client/v1/livetv", nil)

			h := NewLiveTVHandler(&stubLiveTVService{withStream: tt.withStream}, &stubLiveTVProxyService{})
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

type stubLiveTVService struct {
	withStream bool
}

func (s *stubLiveTVService) List(ctx context.Context, req *clientdto.LiveTVListRequest) ([]clientdto.LiveTVChannelItem, int, error) {
	item := clientdto.LiveTVChannelItem{
		ID:          1,
		Name:        "CCTV1",
		Category:    "央视",
		Description: "测试频道",
		Format:      "hls",
	}
	if s.withStream {
		item.StreamURL = "http://example.com/cctv1.m3u8"
	}
	return []clientdto.LiveTVChannelItem{item}, 1, nil
}

func (s *stubLiveTVService) GetStreamURL(ctx context.Context, id uint32) (string, error) {
	return "", nil
}

type stubLiveTVProxyService struct {
	proxyErr error
}

func (s *stubLiveTVProxyService) Proxy(c *gin.Context, channelID uint32, segURL string) error {
	return s.proxyErr
}
