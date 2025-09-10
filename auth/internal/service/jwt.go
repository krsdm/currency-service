package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vctrl/currency-service/auth/internal/config"
	"go.uber.org/zap"
)

const (
	DefaultLifetimeSeconds int = 300
)

type JWTService struct {
	cfg    config.JWTConfig
	logger *zap.Logger
}

func NewJWTService(cfg config.JWTConfig, logger *zap.Logger) *JWTService {
	return &JWTService{cfg: cfg, logger: logger}
}

func (s *JWTService) GenerateToken(userID string) (string, error) {
	var lifetime time.Duration
	if s.cfg.JWTLifetimeSeconds == 0 {
		lifetime = time.Second * time.Duration(DefaultLifetimeSeconds)
	} else {
		lifetime = time.Second * time.Duration(s.cfg.JWTLifetimeSeconds)
	}

	exp := time.Now().Add(lifetime).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": userID, "exp": exp})
	s.logger.Info("JWT token created. Payload: ", zap.Any("sub", userID), zap.Any("exp", exp))

	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *JWTService) ValidateToken(tokenString string) error {
	options := []jwt.ParserOption{
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	}
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	}, options...)

	if err != nil {
		return err
	}

	return nil
}
