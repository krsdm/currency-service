package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	currencyClient "github.com/vctrl/currency-service/currency/internal/clients/currency"
	"github.com/vctrl/currency-service/currency/internal/config"
	"github.com/vctrl/currency-service/currency/internal/db"
	"github.com/vctrl/currency-service/currency/internal/handler"
	"github.com/vctrl/currency-service/currency/internal/repository"
	"github.com/vctrl/currency-service/currency/internal/service"
	"github.com/vctrl/currency-service/pkg/currency"

	"flag"
	"fmt"
	"log"
	"net"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

// TODO:
// - Добавить run() error по аналогии с migrator
// - Вместо логов - возвращать ошибки

var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "currency_requests_total",
			Help: "Total number of requests handled by the currency service",
		},
		[]string{"method"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "currency_request_duration_seconds",
			Help:    "Histogram of response times for requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	appUptime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "currency_service_uptime_seconds",
			Help: "Time since service start in seconds",
		},
	)
)

func init() {
	// Регистрируем метрики
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(appUptime)
}

func main() {
	configPath := flag.String("config", "./config", "path to the config file")

	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}

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
		log.Fatalf("error init exchange rate repository: %v", err)
	}

	client, err := currencyClient.New(cfg.API, logger)
	if err != nil {
		log.Fatalf("error creating currency client: %v", err)
	}

	svc := service.NewCurrency(repo, client, logger)

	// todo
	//metrics := initMetrics()
	//
	//middleware := initMiddleware(metrics)

	// todo apply middleware

	currencyServer := handler.NewCurrencyServer(&svc,
		logger,
		requestCount,
		requestDuration,
		appUptime,
		/*metrics*/)

	go func() {
		if err := startGRPCServer(cfg, currencyServer); err != nil {
			log.Fatalf("Error starting GRPC server: %s", err)
		}
	}()

	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			_, err := fmt.Fprintf(w, "health is ok")
			if err != nil {
				log.Fatalf("Failed health check: %s", err)
			} else {
				log.Println("health check successful")
			}
		})
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Prometheus metrics server running on :8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Fatalf("Error starting Prometheus metrics server: %s", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		startTime := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				appUptime.Set(time.Since(startTime).Seconds())
				time.Sleep(5 * time.Second)
			}
		}
	}()

	select {} // Блокируем main() чтобы горутины работали // todo graceful shutdown
}

func startGRPCServer(cfg config.AppConfig, srv handler.CurrencyServer) error {
	lis, err := net.Listen("tcp", ":"+cfg.Service.ServerPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	currency.RegisterCurrencyServiceServer(s, srv)

	log.Printf("gRPC server is listening on :%s", cfg.Service.ServerPort)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
