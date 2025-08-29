package migrations

import (
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/vctrl/currency-service/currency/internal/config"
	"github.com/vctrl/currency-service/currency/internal/db"
)

//go:embed *.sql
var fs embed.FS

func RunPgMigrations(cfg config.DatabaseConfig) error {
	d, err := iofs.New(fs, ".")
	if err != nil {
		return err
	}

	database, _, err := db.NewDatabaseConnection(cfg)
	if err != nil {
		return err
	}

	databaseDriver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", d, "postgres", databaseDriver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
