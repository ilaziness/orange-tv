package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/clienttype"
	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectClientType(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		userAgent string
		want      string
	}{
		{"X-Client-Type app", constant.ClientTypeApp, "Dart/3.5 (dart:io)", constant.ClientTypeApp},
		{"X-Client-Type tv", constant.ClientTypeTV, "Dart/3.5 (dart:io)", constant.ClientTypeTV},
		{"X-Client-Type desktop", constant.ClientTypeDesktop, "Dart/3.5 (dart:io)", constant.ClientTypeDesktop},
		{"X-Client-Type web", constant.ClientTypeWeb, "Dart/3.5 (dart:io)", constant.ClientTypeWeb},
		{"X-Client-Type uppercase App", "App", "Dart/3.5 (dart:io)", constant.ClientTypeApp},
		{"X-Client-Type mixed TV", "Tv", "Dart/3.5 (dart:io)", constant.ClientTypeTV},
		{"invalid header falls back to UA mobile", "h4x", "Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36", constant.ClientTypeApp},
		{"Dart UA no header -> web safe default", "", "Dart/3.5 (dart:io)", constant.ClientTypeWeb},
		{"empty UA -> web safe default", "", "", constant.ClientTypeWeb},
		{"TV keyword UA", "", "Mozilla/5.0 (SMART-TV; Linux; Tizen 6.0) AppleWebKit/537.36", constant.ClientTypeTV},
		{"tizen UA", "", "Mozilla/5.0 (SMART-TV; Linux; Tizen 6.0) AppleWebKit/537.36", constant.ClientTypeTV},
		{"androidtv UA", "", "Mozilla/5.0 (Linux; AndroidTV 12) AppleWebKit/537.36", constant.ClientTypeTV},
		{"android tv UA with space", "", "Mozilla/5.0 (Linux; Android TV 12) AppleWebKit/537.36", constant.ClientTypeTV},
		{"webos TV UA", "", "Mozilla/5.0 (Web0S; Linux/SmartTV) AppleWebKit/537.36", constant.ClientTypeTV},
		{"nettv UA", "", "Mozilla/5.0 (NetTV; U; Linux) AppleWebKit/537.36", constant.ClientTypeTV},
		{"iptv UA must NOT be tv", "", "Mozilla/5.0 (Linux; IPTV-STB) AppleWebKit/537.36", constant.ClientTypeWeb},
		{"atv UA must NOT be tv", "", "Mozilla/5.0 (ATV3,1; CPU OS 10_0 like Mac OS X)", constant.ClientTypeWeb},
		{"android browser -> app", "", "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 Mobile", constant.ClientTypeApp},
		{"iphone safari -> app", "", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile", constant.ClientTypeApp},
		{"desktop chrome -> web", "", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/126.0", constant.ClientTypeWeb},
		{"desktop mac firefox -> web", "", "Mozilla/5.0 (Macintosh; Intel Mac OS X 14.0) Gecko/20100101 Firefox/127.0", constant.ClientTypeWeb},
		{"curl UA -> web safe default", "", "curl/8.6.0", constant.ClientTypeWeb},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectClientType(tt.header, tt.userAgent)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClientTypeMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ginRouter := gin.New()
	ginRouter.Use(ClientTypeMiddleware(testSystemPaths()...))
	ginRouter.GET("/echo", func(c *gin.Context) {
		t := clienttype.FromContext(c.Request.Context())
		c.JSON(http.StatusOK, gin.H{"client_type": t})
	})
	ginRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	tests := []struct {
		name       string
		path       string
		header     string
		userAgent  string
		wantBody   string
		wantHeader string
	}{
		{"header app takes priority over Dart UA", "/echo", constant.ClientTypeApp, "Dart/3.5 (dart:io)", `"client_type":"app"`, constant.ClientTypeApp},
		{"desktop UA no header -> web safe default", "/echo", "", "Dart/3.5 (dart:io)", `"client_type":"web"`, constant.ClientTypeWeb},
		{"android UA -> app", "/echo", "", "Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 Mobile", `"client_type":"app"`, constant.ClientTypeApp},
		{"uppercase header -> app", "/echo", "App", "Dart/3.5 (dart:io)", `"client_type":"app"`, constant.ClientTypeApp},
		{"system path skips header echo", "/health", "", "curl/8.6.0", `"status":"ok"`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.header != "" {
				req.Header.Set(constant.ClientTypeHeader, tt.header)
			}
			req.Header.Set("User-Agent", tt.userAgent)
			ginRouter.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
			assert.Equal(t, tt.wantHeader, w.Header().Get(constant.ClientTypeHeader))
		})
	}
}

// testSystemPaths 返回需要跳过的系统路径字面量。
// 避免 import internal/router 造成 middleware → router → middleware 循环依赖。
func testSystemPaths() []string {
	return []string{"/health", "/readiness", "/liveness", "/version"}
}
