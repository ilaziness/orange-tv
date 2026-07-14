// Package http provides HTTP middleware implementations.
package http

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ilaziness/orange-tv/internal/auth"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/response"
)

const (
	// ClaimsKey is the key used to store JWT claims in the Gin context.
	ClaimsKey = "claims"
)

// JWTAuth returns a middleware that validates JWT tokens from the Authorization header.
// On success, the parsed claims are stored in the Gin context under ClaimsKey.
// Paths in skipPaths are allowed through without authentication.
func JWTAuth(manager *auth.JWTManager, skipPaths ...string) gin.HandlerFunc {
	skipSet := make(map[string]struct{}, len(skipPaths))
	for _, p := range skipPaths {
		skipSet[p] = struct{}{}
	}

	return func(c *gin.Context) {
		if _, skip := skipSet[c.FullPath()]; skip {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errcode.WithMessage(errcode.AuthFailed, "缺少Authorization头"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, errcode.WithMessage(errcode.AuthFailed, "Authorization头格式错误，应为 Bearer <token>"))
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := manager.ParseToken(tokenString)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.Error(c, errcode.TokenExpired)
			} else {
				response.Error(c, errcode.InvalidToken)
			}
			c.Abort()
			return
		}

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

// GetClaims retrieves the JWT claims from the Gin context.
// Returns nil if no claims are stored.
func GetClaims(c *gin.Context) *auth.Claims {
	val, exists := c.Get(ClaimsKey)
	if !exists {
		return nil
	}
	claims, ok := val.(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}

// RequireAuth returns a middleware that requires valid JWT access token claims in context.
// Use on route groups that must not be accessed without authentication (e.g. admin API).
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.Error(c, errcode.WithMessage(errcode.AuthFailed, "未认证"))
			c.Abort()
			return
		}
		if claims.TokenType != auth.TokenTypeAccess {
			response.Error(c, errcode.InvalidToken)
			c.Abort()
			return
		}
		c.Next()
	}
}

const internalServiceKeyHeader = "X-Internal-Service-Key"

// InternalServiceAuth protects internal service-to-service routes.
// When serviceKey is non-empty, requests must provide a matching X-Internal-Service-Key header.
// When serviceKey is empty, falls back to RequireAuth (JWT access token required).
func InternalServiceAuth(serviceKey string) gin.HandlerFunc {
	requireAuth := RequireAuth()
	return func(c *gin.Context) {
		if serviceKey != "" {
			if c.GetHeader(internalServiceKeyHeader) != serviceKey {
				response.Error(c, errcode.WithMessage(errcode.AuthFailed, "无效的内部服务密钥"))
				c.Abort()
				return
			}
			c.Next()
			return
		}
		requireAuth(c)
	}
}
