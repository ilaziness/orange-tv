// Package auth provides JWT token generation and parsing.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ilaziness/orange-tv/internal/config"
)

// TokenType represents the type of JWT token.
type TokenType string

const (
	// TokenTypeAccess is the access token type.
	TokenTypeAccess TokenType = "access"
	// TokenTypeRefresh is the refresh token type.
	TokenTypeRefresh TokenType = "refresh"
)

// Subject identifies the token owner kind: admin or user.
type Subject string

const (
	// SubjectAdmin marks tokens issued to admin accounts.
	SubjectAdmin Subject = "admin"
	// SubjectUser marks tokens issued to regular user accounts.
	SubjectUser Subject = "user"
)

// Claims represents the custom JWT claims.
type Claims struct {
	UserID    int64     `json:"user_id"`
	Subject   Subject   `json:"subject"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTManager manages JWT token generation and parsing using HS256.
type JWTManager struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

// NewJWTManagerFromConfig creates a JWTManager from configuration.
// Returns nil if JWT secret is not configured.
func NewJWTManagerFromConfig(cfg *config.Config) *JWTManager {
	if cfg.JWT.Secret == "" {
		return nil
	}
	return NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
}

// NewJWTManager creates a new JWTManager with HS256 algorithm.
// accessTokenTTL and refreshTokenTTL are in seconds.
func NewJWTManager(secret string, accessTokenTTL, refreshTokenTTL int) *JWTManager {
	return &JWTManager{
		secret:          []byte(secret),
		accessTokenTTL:  time.Duration(accessTokenTTL) * time.Second,
		refreshTokenTTL: time.Duration(refreshTokenTTL) * time.Second,
		issuer:          "orange-tv",
	}
}

// GenerateAccessToken generates a new access token for the given user ID.
func (m *JWTManager) GenerateAccessToken(userID int64) (string, error) {
	return m.generateToken(userID, SubjectAdmin, TokenTypeAccess, m.accessTokenTTL)
}

// GenerateRefreshToken generates a new refresh token for the given user ID.
func (m *JWTManager) GenerateRefreshToken(userID int64) (string, error) {
	return m.generateToken(userID, SubjectAdmin, TokenTypeRefresh, m.refreshTokenTTL)
}

// GenerateAccessTokenFor generates an access token for a specific subject (admin/user).
func (m *JWTManager) GenerateAccessTokenFor(userID int64, sub Subject) (string, error) {
	return m.generateToken(userID, sub, TokenTypeAccess, m.accessTokenTTL)
}

// GenerateRefreshTokenFor generates a refresh token for a specific subject (admin/user).
func (m *JWTManager) GenerateRefreshTokenFor(userID int64, sub Subject) (string, error) {
	return m.generateToken(userID, sub, TokenTypeRefresh, m.refreshTokenTTL)
}

// generateToken creates a signed JWT token with the given parameters.
func (m *JWTManager) generateToken(userID int64, sub Subject, tokenType TokenType, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Subject:   sub,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ParseToken parses and validates a JWT token string, returning the claims.
func (m *JWTManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// Validate signing method is HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method, expected HS256")
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}
