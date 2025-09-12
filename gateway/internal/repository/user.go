package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/vctrl/currency-service/gateway/internal/password"
)

var (
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrUserNotFound     = errors.New("user not found")
)

type User struct {
	Login    string
	Password password.ProtectedPassword
}

type MapUserRepository struct {
	users map[string]User
	mu    *sync.RWMutex
}

func NewUser() *MapUserRepository {
	return &MapUserRepository{
		users: make(map[string]User),
		mu:    &sync.RWMutex{},
	}
}

func (repo *MapUserRepository) AddUser(_ context.Context, user User) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.users[user.Login]; exists {
		return ErrUserAlreadyExist
	}
	repo.users[user.Login] = user

	return nil
}

func (repo *MapUserRepository) GetUser(_ context.Context, login string) (User, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	user, exists := repo.users[login]
	if !exists {
		return User{}, ErrUserNotFound
	}

	return user, nil
}
