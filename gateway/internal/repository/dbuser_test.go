package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vctrl/currency-service/gateway/internal/config"
	"github.com/vctrl/currency-service/gateway/internal/password"
)

type UserRepositorySuite struct {
	suite.Suite
	pgContainer     *postgres.PostgresContainer
	userRepository  UserRepository
	passwordManager password.ProtectedPasswordManager
	ctx             context.Context
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositorySuite))
}

func (s *UserRepositorySuite) SetupSuite() {
	s.ctx = context.Background()
	migrationPath, _ := filepath.Abs("../migrations/000001_create_users_table.up.sql")
	pgContainer, err := postgres.Run(
		s.ctx, "postgres:16-alpine",
		postgres.WithDatabase("gateway_db"),
		postgres.WithUsername("admin"),
		postgres.WithPassword("password"),
		postgres.WithInitScripts(migrationPath),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	s.Require().NoError(err)
	dsn, err := pgContainer.ConnectionString(s.ctx, "sslmode=disable")
	s.Require().NoError(err)
	s.pgContainer = pgContainer
	db, err := sql.Open("postgres", dsn)
	s.Require().NoError(err)
	err = db.Ping()
	s.Require().NoError(err, "database container is running")
	s.userRepository = NewUserRepository(db)
	s.passwordManager, err = password.NewUserPasswordManager(config.PasswordPolicy{})
	s.Require().NoError(err)
}

func (s *UserRepositorySuite) TearDownSuite() {
	err := s.pgContainer.Terminate(s.ctx)
	s.Assert().NoError(err)
}

func (s *UserRepositorySuite) TestCreateNewUser() {
	protectedPassword, err := s.passwordManager.CreatePassword("123")
	s.Require().NoError(err)
	user := User{Login: "petrov", Password: protectedPassword}

	err = s.userRepository.AddUser(context.Background(), user)
	s.Assert().NoError(err)
}

func (s *UserRepositorySuite) TestCreateExistingUser() {
	protectedPassword, err := s.passwordManager.CreatePassword("123")
	s.Require().NoError(err)
	user := User{Login: "ivanov", Password: protectedPassword}

	err = s.userRepository.AddUser(context.Background(), user)
	s.Require().NoError(err)

	err = s.userRepository.AddUser(context.Background(), user)
	s.Assert().ErrorIs(ErrUserAlreadyExist, err)
}

func (s *UserRepositorySuite) TestGetExistingUser() {
	protectedPassword, err := s.passwordManager.CreatePassword("123")
	s.Require().NoError(err)
	user := User{Login: "sidorov", Password: protectedPassword}

	err = s.userRepository.AddUser(context.Background(), user)
	s.Require().NoError(err)

	_, err = s.userRepository.GetUser(s.ctx, user.Login)
	s.Assert().NoError(err)
}

func (s *UserRepositorySuite) TestGetUserWhenUserNoExist() {
	_, err := s.userRepository.GetUser(s.ctx, "any user")
	s.Assert().ErrorIs(ErrUserNotFound, err)
}
