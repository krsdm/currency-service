package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

type UserRepository interface {
	AddUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, login string) (User, error)
}

type DBUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *DBUserRepository {
	return &DBUserRepository{db}
}

func (r *DBUserRepository) AddUser(ctx context.Context, user User) error {
	existing, err := r.GetUser(ctx, user.Login)
	if err == nil && existing.Login == user.Login {
		return ErrUserAlreadyExist
	}
	query := `insert into users(login, password_hash) values ($1, $2)`
	_, err = r.db.ExecContext(ctx, query, user.Login, user.Password.String())
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

func (r *DBUserRepository) GetUser(ctx context.Context, login string) (User, error) {
	query := `SELECT login, password_hash FROM users WHERE login = $1`
	row := r.db.QueryRowContext(ctx, query, login)

	var user User
	err := row.Scan(&user.Login, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	} else if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, err
}
