package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-must-be-at-least-32bytes"

func newTestManager() *JWTManager {
	return NewJWTManager(testSecret, 7200, 604800)
}

func TestGenerateAccessToken(t *testing.T) {
	mgr := newTestManager()

	token, err := mgr.GenerateAccessToken(42)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := mgr.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint32(42), claims.UserID)
	assert.Equal(t, TokenTypeAccess, claims.TokenType)
	assert.Equal(t, "orange-tv", claims.Issuer)
	assert.True(t, claims.ExpiresAt.After(time.Now().Add(time.Hour)))
}

func TestGenerateRefreshToken(t *testing.T) {
	mgr := newTestManager()

	token, err := mgr.GenerateRefreshToken(99)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := mgr.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint32(99), claims.UserID)
	assert.Equal(t, TokenTypeRefresh, claims.TokenType)
	assert.True(t, claims.ExpiresAt.After(time.Now().Add(24*time.Hour)))
}

func TestParseToken_Valid(t *testing.T) {
	mgr := newTestManager()
	token, err := mgr.GenerateAccessToken(1)
	require.NoError(t, err)

	claims, err := mgr.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint32(1), claims.UserID)
}

func TestParseToken_Expired(t *testing.T) {
	mgr := NewJWTManager(testSecret, -1, 604800) // already expired

	token, err := mgr.GenerateAccessToken(1)
	require.NoError(t, err)

	_, err = mgr.ParseToken(token)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
}

func TestParseToken_InvalidSignature(t *testing.T) {
	mgr := newTestManager()
	token, err := mgr.GenerateAccessToken(1)
	require.NoError(t, err)

	// tamper with the signature
	tampered := token[:len(token)-4] + "xxxx"
	_, err = mgr.ParseToken(tampered)
	assert.Error(t, err)
}

func TestParseToken_Malformed(t *testing.T) {
	mgr := newTestManager()

	_, err := mgr.ParseToken("not.a.valid.jwt.string")
	assert.Error(t, err)
}
