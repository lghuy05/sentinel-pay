package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/account"
	"github.com/lghuy05/sentinel-pay/internal/platform/config"
	"github.com/lghuy05/sentinel-pay/internal/platform/httpserver"
	"github.com/lghuy05/sentinel-pay/internal/platform/logging"
	"github.com/lghuy05/sentinel-pay/internal/platform/metrics"
)

func main() {
	logger := logging.New(account.ServiceName, config.NewEnv().String("LOG_LEVEL", "INFO"))
	if err := run(context.Background(), logger); err != nil {
		logger.Error("account service failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	env := config.NewEnv()
	port, err := env.Int("SERVER_PORT", account.DefaultPort)
	if err != nil {
		return err
	}
	dsn := env.String("DATABASE_URL", "")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			env.String("DB_USER", "sentinel"),
			env.String("DB_PASSWORD", "sentinel123"),
			env.String("DB_HOST", "localhost"),
			env.String("DB_PORT", "15432"),
			env.String("DB_NAME", "sentinelpay"),
		)
	}

	db, err := account.OpenPostgres(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()

	store := account.NewPostgresStore(db)
	if err := store.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate postgres: %w", err)
	}

	mux := http.NewServeMux()
	account.NewHandler(account.NewService(store)).Register(mux)
	registry := metrics.NewRegistry(account.ServiceName)
	mux.Handle("/metrics", registry.Handler())

	server := httpserver.New(fmt.Sprintf(":%d", port), registry.Middleware(mux), 10*time.Second)
	errs := make(chan error, 1)
	go func() {
		logger.Info("account service listening", slog.Int("port", port))
		errs <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stop:
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errs:
		return err
	}
}
