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
	"github.com/lghuy05/sentinel-pay/internal/blacklist"
	"github.com/lghuy05/sentinel-pay/internal/platform/config"
	"github.com/lghuy05/sentinel-pay/internal/platform/httpserver"
	"github.com/lghuy05/sentinel-pay/internal/platform/kafkautil"
	"github.com/lghuy05/sentinel-pay/internal/platform/logging"
	"github.com/lghuy05/sentinel-pay/internal/platform/metrics"
	"github.com/lghuy05/sentinel-pay/internal/platform/redisutil"
)

func main() {
	logger := logging.New(blacklist.ServiceName, config.NewEnv().String("LOG_LEVEL", "INFO"))
	if err := run(context.Background(), logger); err != nil {
		logger.Error("blacklist service failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	env := config.NewEnv()
	port, err := env.Int("SERVER_PORT", blacklist.DefaultPort)
	if err != nil {
		return err
	}
	redisPort, err := env.Int("REDIS_PORT", 16379)
	if err != nil {
		return err
	}
	redisClient, err := redisutil.NewClient(ctx, redisutil.Config{
		Host: env.String("REDIS_HOST", "localhost"),
		Port: redisPort,
	})
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer redisClient.Close()

	dsn := env.String("DATABASE_URL", "")
	if dsn == "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			env.String("DB_USER", "sentinel"),
			env.String("DB_PASSWORD", "sentinel123"),
			env.String("DB_HOST", "localhost"),
			env.String("DB_PORT", "15432"),
			env.String("DB_NAME", "blacklist_db"),
		)
	}
	db, err := account.OpenPostgres(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()
	store := blacklist.NewPostgresStore(db)
	if err := store.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate postgres: %w", err)
	}

	service := blacklist.NewService(store, redisClient)
	if err := service.WarmCache(ctx); err != nil {
		return fmt.Errorf("warm blacklist cache: %w", err)
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	worker := blacklist.NewWorker(service, kafkautil.ParseBrokers(env.String("KAFKA_BOOTSTRAP_SERVERS", ""), "localhost:19092"), logger)
	defer worker.Close()
	go worker.Run(runCtx)

	mux := http.NewServeMux()
	httpserver.RegisterHealth(mux, blacklist.ServiceName)
	registry := metrics.NewRegistry(blacklist.ServiceName)
	mux.Handle("/metrics", registry.Handler())
	server := httpserver.New(fmt.Sprintf(":%d", port), registry.Middleware(mux), 10*time.Second)
	errs := make(chan error, 1)
	go func() {
		logger.Info("blacklist service listening", slog.Int("port", port))
		errs <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stop:
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
		defer shutdownCancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errs:
		cancel()
		return err
	}
}
