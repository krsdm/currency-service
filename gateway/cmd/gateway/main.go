package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/vctrl/currency-service/gateway/internal/clients/auth"
	"github.com/vctrl/currency-service/gateway/internal/config"
	"github.com/vctrl/currency-service/gateway/internal/handler"
	"github.com/vctrl/currency-service/gateway/internal/middleware"
	"github.com/vctrl/currency-service/gateway/internal/password"
	"github.com/vctrl/currency-service/gateway/internal/repository"
	"github.com/vctrl/currency-service/gateway/internal/service"
	"github.com/vctrl/currency-service/pkg/grpc_client"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}

func run() error {
	configPath := flag.String("config", "./config", "path to the config file")

	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	logger.Info("Server initializing with config", zap.Any("config", cfg))

	router := gin.New()
	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))

	// auth middleware
	authClient, err := auth.NewAuthClient(cfg.Auth)
	if err != nil {
		return fmt.Errorf("auth.NewAuthClient: %w", err)
	}

	resp, err := authClient.Ping()
	if err != nil {
		return fmt.Errorf("authClient.Ping: %w", err)
	}

	if resp != "pong" {
		return fmt.Errorf("auth client answered with invalid response: %w", err)
	}

	authMiddleware := middleware.NewAuthorization(authClient, shouldSkipAuthMiddleware, logger)
	router.Use(authMiddleware.Authorize())

	currencyClient, conn, err := grpc_client.NewCurrencyServiceClient(cfg.GRPC.CurrencyServiceURL)
	if err != nil {
		return fmt.Errorf("grpc_client.NewCurrencyServiceClient: %w", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			logger.Warn("Cannot close GRPC Client for auth service", zap.Error(err))
		}
	}()

	/*
		userRepo := user.NewRepository()
		authService := auuth.NewService(authClient, userRepo)
		currencyService := currency.NewService(currencyClient)
	*/

	userRepo := repository.NewUser()
	passwordManager := password.NewUserPasswordManager(cfg.PasswordPolicy)
	authService := service.NewAuth(authClient, userRepo, passwordManager)
	currencyService := service.NewCurrency(currencyClient)

	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: router,
	}

	handler.RegisterRoutes(authService, currencyService, router, logger)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("srv.ListenAndServe: %s\n", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	log.Println("Server exiting")
	return nil
}

func shouldSkipAuthMiddleware(c *gin.Context) bool {
	if strings.HasSuffix(c.Request.URL.Path, "/login") ||
		strings.HasSuffix(c.Request.URL.Path, "/register") ||
		strings.HasSuffix(c.Request.URL.Path, "/metrics") {
		return true
	}

	return false
}
