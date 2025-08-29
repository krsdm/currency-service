package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/vctrl/currency-service/currency/internal/clients/currency"
	"github.com/vctrl/currency-service/currency/internal/config"
	"github.com/vctrl/currency-service/currency/internal/db"
	"github.com/vctrl/currency-service/currency/internal/repository"
	"github.com/vctrl/currency-service/currency/internal/service"
	"github.com/vctrl/currency-service/currency/internal/worker"

	"go.uber.org/zap"

	"github.com/robfig/cron/v3"
)

// TODO:
// - Добавить run() error по аналогии с migrator
// - Вместо логов - возвращать ошибки

func main() {
	configPath := flag.String("config", "./config", "path to the config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	cfg.API.BaseURL = fmt.Sprintf(cfg.API.BaseURL, "latest")

	db, _, err := db.NewDatabaseConnection(cfg.Database)
	if err != nil {
		log.Fatalf("error init database connection: %v", err)
	}

	repo, err := repository.NewCurrency(db)
	if err != nil {
		log.Fatalf("error creating repository: %v", err)
	}

	// Logger инициировать как можно раньше.
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("init logger: %w", err)
	}

	client, err := currency.New(cfg.API, logger)
	if err != nil {
		log.Fatalf("error creating currency client: %v", err)
	}

	svc := service.NewCurrency(repo, client, logger)

	c := cron.New()

	currencyWorker := worker.NewCurrency(cfg.Worker, svc, c, logger)

	if err != nil {
		log.Fatalf("error adding cron job: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := currencyWorker.StartFetchingCurrencyRates(); err != nil {
		log.Fatalf("error start fetching currency rates: %v", err)
	}

	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	currencyWorker.Stop()
}
