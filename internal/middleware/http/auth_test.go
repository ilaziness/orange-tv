package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ilaziness/orange-tv/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const authTestSecret = "auth-test-secret-key-must-be-32bytes!!"

func newAuthTestManager() *auth.JWTManager {
	return auth.NewJWTManager(authTestSecret, 7200, 604800)
}

func setupAuthRouter(mgr *auth.JWTManager, skipPaths ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(mgr, skipPaths...))
	r.GET("/protected", func(c *gin.Context) {
		claims := GetClaims(c)
		c.JSON(http.StatusOK, gin.H{"user_id": claims.UserID})
	})
	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestJWTAuth_NoAuthorizationHeader(t *testing.T) {
	r := setupAuthRouter(newAuthTestManager())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(3000001), body["code"])
}

func TestJWTAuth_MissingBearerPrefix(t *testing.T) {
	r := setupAuthRouter(newAuthTestManager())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Token somevalue")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_ValidToken(t *testing.T) {
	mgr := newAuthTestManager()
	token, err := mgr.GenerateAccessToken(7)
	require.NoError(t, err)

	r := setupAuthRouter(mgr)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(7), body["user_id"])
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	mgr := auth.NewJWTManager(authTestSecret, -1, 604800)
	token, err := mgr.GenerateAccessToken(1)
	require.NoError(t, err)

	r := setupAuthRouter(mgr)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(3000002), body["code"])
}

func TestJWTAuth_InvalidSignature(t *testing.T) {
	mgr := newAuthTestManager()
	token, err := mgr.GenerateAccessToken(1)
	require.NoError(t, err)

	tampered := token[:len(token)-4] + "xxxx"
	r := setupAuthRouter(mgr)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_SkipPath(t *testing.T) {
	mgr := newAuthTestManager()
	r := setupAuthRouter(mgr, "/public")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/public", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuth_SkipPathPrefix(t *testing.T) {
	mgr := newAuthTestManager()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(mgr, "/api/client/*", "/api/admin/v1/auth/login", "/sitemap*", "/sitemaps/*"))
	r.GET("/api/client/v1/categories", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.POST("/api/admin/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/api/admin/v1/videos", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	r.GET("/sitemap.xml", func(c *gin.Context) {
		c.String(http.StatusOK, "sitemap")
	})
	r.GET("/sitemap-41.xml", func(c *gin.Context) {
		c.String(http.StatusOK, "shard")
	})
	r.GET("/sitemaps/videos-1.xml", func(c *gin.Context) {
		c.String(http.StatusOK, "videos")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/client/v1/categories", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/admin/v1/auth/login", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/admin/v1/videos", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	for _, path := range []string{"/sitemap.xml", "/sitemap-41.xml", "/sitemaps/videos-1.xml"} {
		w = httptest.NewRecorder()
		req, _ = http.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, path)
	}
}

func TestPathSkipped(t *testing.T) {
	assert.True(t, pathSkipped("/api/client/v1/videos", "/api/client/v1/videos", []string{"/api/client/*"}))
	assert.True(t, pathSkipped("/swagger/index.html", "/swagger/index.html", []string{"/swagger/*any"}))
	assert.True(t, pathSkipped("/health", "/health", []string{"/health"}))
	assert.True(t, pathSkipped("", "/sitemap-41.xml", []string{"/sitemap*"}))
	assert.False(t, pathSkipped("/api/admin/v1/videos", "/api/admin/v1/videos", []string{"/api/client/*"}))
}

func TestJWTMiddleware_ContextClaims(t *testing.T) {
	mgr := newAuthTestManager()
	token, err := mgr.GenerateAccessToken(123)
	require.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(mgr))

	var capturedUserID uint32
	r.GET("/check", func(c *gin.Context) {
		claims := GetClaims(c)
		require.NotNil(t, claims)
		capturedUserID = claims.UserID
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/check", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint32(123), capturedUserID)
}

func TestRequireAuth_NoClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequireAuth())
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_WithAccessTokenClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(ClaimsKey, &auth.Claims{UserID: 9, TokenType: auth.TokenTypeAccess})
		c.Next()
	})
	r.Use(RequireAuth())
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInternalServiceAuth_WithServiceKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(InternalServiceAuth("secret-key"))
	r.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/internal", nil)
	req.Header.Set("X-Internal-Service-Key", "secret-key")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/internal", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInternalServiceAuth_WithoutServiceKeyRequiresJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(InternalServiceAuth(""))
	r.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/internal", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	r2 := gin.New()
	r2.Use(func(c *gin.Context) {
		c.Set(ClaimsKey, &auth.Claims{UserID: 1, TokenType: auth.TokenTypeAccess})
		c.Next()
	})
	r2.Use(InternalServiceAuth(""))
	r2.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/internal", nil)
	r2.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
