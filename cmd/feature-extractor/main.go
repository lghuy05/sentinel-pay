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

	"github.com/lghuy05/sentinel-pay/internal/feature"
	"github.com/lghuy05/sentinel-pay/internal/platform/config"
	"github.com/lghuy05/sentinel-pay/internal/platform/httpserver"
	"github.com/lghuy05/sentinel-pay/internal/platform/kafkautil"
	"github.com/lghuy05/sentinel-pay/internal/platform/logging"
	"github.com/lghuy05/sentinel-pay/internal/platform/metrics"
	"github.com/lghuy05/sentinel-pay/internal/platform/redisutil"
)

func main() {
	logger := logging.New(feature.ServiceName, config.NewEnv().String("LOG_LEVEL", "INFO"))
	if err := run(context.Background(), logger); err != nil {
		logger.Error("feature extractor failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	env := config.NewEnv()
	port, err := env.Int("SERVER_PORT", feature.DefaultPort)
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

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	service := feature.NewService(
		feature.NewAccountClient(env.String("ACCOUNT_SERVICE_URL", "http://localhost:8087")),
		feature.NewRedisFeatures(redisClient),
	)
	worker := feature.NewWorker(service, kafkautil.ParseBrokers(env.String("KAFKA_BOOTSTRAP_SERVERS", ""), "localhost:19092"), logger)
	defer worker.Close()
	go worker.Run(runCtx)

	mux := http.NewServeMux()
	httpserver.RegisterHealth(mux, feature.ServiceName)
	registry := metrics.NewRegistry(feature.ServiceName)
	mux.Handle("/metrics", registry.Handler())
	server := httpserver.New(fmt.Sprintf(":%d", port), registry.Middleware(mux), 10*time.Second)

	errs := make(chan error, 1)
	go func() {
		logger.Info("feature extractor listening", slog.Int("port", port))
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
