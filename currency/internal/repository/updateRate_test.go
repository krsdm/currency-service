package repository

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vctrl/currency-service/currency/internal/dto"
)

type UpdateRateSuit struct {
	suite.Suite
	pgContainer *postgres.PostgresContainer
	repo        CurrencyRepository
}

func TestUpdateRateSuit(t *testing.T) {
	suite.Run(t, new(UpdateRateSuit))
}

func (s *UpdateRateSuit) SetupSuite() {
	migrationPath, _ := filepath.Abs("../migrations/000001_create_exchange_rates_table.up.sql")
	pgContainer, err := postgres.Run(
		context.Background(), "postgres:16-alpine",
		postgres.WithDatabase("currency_db"),
		postgres.WithUsername("admin"),
		postgres.WithPassword("password"),
		postgres.WithInitScripts(migrationPath),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	s.Require().NoError(err)
	dsn, err := pgContainer.ConnectionString(context.Background(), "sslmode=disable")
	s.Require().NoError(err)
	s.pgContainer = pgContainer
	db, err := sql.Open("postgres", dsn)
	s.Require().NoError(err)
	err = db.Ping()
	s.Require().NoError(err, "database container is running")
	s.repo, err = NewCurrency(db)
	s.Require().NoError(err)
}

func (s *UpdateRateSuit) TearDownSuite() {
	err := s.pgContainer.Terminate(context.Background())
	s.Assert().NoError(err)
}

func (s *UpdateRateSuit) SetupTest() {
	rates := map[string]float64{"usd": 10.10, "eur": 9.10}
	err := s.repo.Save(context.Background(), time.Now(), "RUB", rates)
	s.Require().NoError(err)
}

func (s *UpdateRateSuit) TestUpdateRate() {
	var newRate float32 = 11.11
	today := time.Now()
	updateDTO := &dto.UpdateCurrencyRequestDTO{
		BaseCurrency:   "RUB",
		TargetCurrency: "usd",
		RateRecord:     dto.RateRecordDTO{Date: today, Rate: newRate},
	}
	affectedRowCount, err := s.repo.UpdateRate(context.Background(), updateDTO)
	s.Require().NoError(err)
	s.Assert().EqualValues(1, affectedRowCount)

	rates, err := s.repo.FindInInterval(context.Background(), &dto.CurrencyRequestDTO{
		BaseCurrency:   "RUB",
		TargetCurrency: "usd",
		DateFrom:       today,
		DateTo:         today,
	})
	s.Require().NoError(err)
	s.Require().Len(rates, 1)
	s.EqualValues(newRate, rates[0].Rate)
}
