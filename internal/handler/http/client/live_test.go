package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	clientdto "github.com/ilaziness/orange-tv/internal/dto/client"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
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

type stubLiveService struct{}

func (s *stubLiveService) List(ctx context.Context, req *clientdto.LiveListRequest) ([]clientdto.LiveChannelItem, int, error) {
	return nil, 0, nil
}

func (s *stubLiveService) GetStreamURL(ctx context.Context, id uint32) (string, error) {
	return "", nil
}

func (s *stubLiveService) AllowedStreamDomains(ctx context.Context) (map[string]struct{}, error) {
	return nil, nil
}

type stubLiveProxyService struct {
	proxyErr error
}

func (s *stubLiveProxyService) Proxy(c *gin.Context, channelID uint32, segURL string) error {
	return s.proxyErr
}
