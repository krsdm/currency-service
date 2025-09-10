package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vctrl/currency-service/auth/internal/config"
	"go.uber.org/zap"
)

var (
	testLogger = zap.NewNop()
	testConfig = config.JWTConfig{
		JWTSecret:          "test-secret",
		JWTLifetimeSeconds: 60,
	}
)

func TestGenerateToken_RightConfig(t *testing.T) {
	jwtService := NewJWTService(testConfig, testLogger)
	userID := "user-123"

	tokenString, err := jwtService.GenerateToken(userID)
	require.NoError(t, err)
	assert.NotZero(t, tokenString)

	token, err := parseToken(tokenString)
	require.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)
	require.NotEmpty(t, claims)

	sub, ok := claims["sub"].(string)
	require.True(t, ok)
	assert.Equal(t, userID, sub)

	exp, ok := claims["exp"]
	require.True(t, ok)

	unixExp, err := asUnix(exp)
	require.NoError(t, err)
	assert.True(t, unixExp > time.Now().Unix())
	assert.True(t, time.Now().Unix()-unixExp < 60)
}

func TestValidateToken_Success(t *testing.T) {
	jwtService := NewJWTService(testConfig, testLogger)
	tokenString, err := jwtService.GenerateToken("user-123")
	require.NoError(t, err)
	assert.NoError(t, jwtService.ValidateToken(tokenString))
}

func TestValidateToken_Expired(t *testing.T) {
	cfg := config.JWTConfig{
		JWTSecret:          "test-secret",
		JWTLifetimeSeconds: -1, // already expired
	}
	jwtService := NewJWTService(cfg, testLogger)
	tokenString, err := jwtService.GenerateToken("user-123")
	err = jwtService.ValidateToken(tokenString)
	assert.ErrorIs(t, err, jwt.ErrTokenExpired)
}

func TestValidateToken_BadSignature(t *testing.T) {
	cfgA := config.JWTConfig{JWTSecret: "secret-A", JWTLifetimeSeconds: 60}
	jwtServiceA := NewJWTService(cfgA, testLogger)
	cfgB := config.JWTConfig{JWTSecret: "secret-B", JWTLifetimeSeconds: 60}
	jwtServiceB := NewJWTService(cfgB, testLogger)

	tokenString, err := jwtServiceA.GenerateToken("user-123")
	require.NoError(t, err)

	err = jwtServiceB.ValidateToken(tokenString)
	assert.ErrorIs(t, err, jwt.ErrTokenSignatureInvalid)
}

func asUnix(date interface{}) (int64, error) {
	switch v := date.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, fmt.Errorf("unknown date type")
	}
}

func parseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(testConfig.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
}
