package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/vctrl/currency-service/gateway/internal/clients/auth"
	"github.com/vctrl/currency-service/gateway/internal/dto"
	"github.com/vctrl/currency-service/gateway/internal/password"
	"github.com/vctrl/currency-service/gateway/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type authClientInterface interface {
	GenerateToken(ctx context.Context, login string) (string, error)
	ValidateToken(ctx context.Context, token string) error
}

type AuthService struct {
	authClient      authClientInterface
	userRepo        repository.UserRepository // todo interface
	passwordManager password.ProtectedPasswordManager
}

func NewAuth(authClient authClientInterface, userRepo repository.UserRepository, passwordManager password.ProtectedPasswordManager) AuthService {
	return AuthService{
		authClient:      authClient,
		userRepo:        userRepo,
		passwordManager: passwordManager,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) error {
	protectedPassword, err := s.passwordManager.CreatePassword(req.Password)
	if err != nil {
		return err
	}

	user := repository.User{Login: req.Username, Password: protectedPassword}
	if err := s.userRepo.AddUser(ctx, user); err != nil {
		return fmt.Errorf("userRepo.AddUser: %w", err)
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepo.GetUser(ctx, login)
	if err != nil {
		return "", fmt.Errorf("userRepo.GetUser: %w", err)
	}

	if err := s.passwordManager.VerifyPassword(password, user.Password); err != nil {
		return "", auth.ErrInvalidCredentials
	}

	res, err := s.authClient.GenerateToken(ctx, login)
	if err != nil {
		return "", fmt.Errorf("authClient.GenerateToken: %w", err)
	}

	return res, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) error {
	err := s.authClient.ValidateToken(ctx, token)
	if err != nil {
		return fmt.Errorf("authClient.ValidateToken: %w", err)
	}

	return nil
}

func (s *AuthService) Logout(token string) error {
	return errors.New("logout is not implemented")
}
