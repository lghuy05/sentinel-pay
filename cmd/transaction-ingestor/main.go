package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/account"
	"github.com/lghuy05/sentinel-pay/internal/ingest"
	"github.com/lghuy05/sentinel-pay/internal/platform/config"
	"github.com/lghuy05/sentinel-pay/internal/platform/httpserver"
	"github.com/lghuy05/sentinel-pay/internal/platform/kafkautil"
	"github.com/lghuy05/sentinel-pay/internal/platform/logging"
	"github.com/lghuy05/sentinel-pay/internal/platform/metrics"
)

func main() {
	logger := logging.New(ingest.ServiceName, config.NewEnv().String("LOG_LEVEL", "INFO"))
	if err := run(context.Background(), logger); err != nil {
		logger.Error("transaction ingestor failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	env := config.NewEnv()
	port, err := env.Int("SERVER_PORT", ingest.DefaultPort)
	if err != nil {
		return err
	}
	relayEnabled := !strings.EqualFold(env.String("EVENT_PUBLISHER", "kafka"), "noop")
	relayDelay, err := env.Duration("OUTBOX_RELAY_DELAY", time.Second)
	if err != nil {
		return err
	}
	highValueLimitEnabled := !strings.EqualFold(env.String("INGEST_HIGH_VALUE_LIMIT_ENABLED", "true"), "false")

	dsn := env.String("DATABASE_URL", "")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			env.String("DB_USER", "sentinel"),
			env.String("DB_PASSWORD", "sentinel123"),
			env.String("DB_HOST", "localhost"),
			env.String("DB_PORT", "15432"),
			env.String("DB_NAME", "transaction_ingestor_db"),
		)
	}

	db, err := account.OpenPostgres(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()

	store := ingest.NewPostgresStore(db)
	if err := store.Migrate(ctx); err != nil {
		return fmt.Errorf("migrate postgres: %w", err)
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var relay *ingest.Relay
	if relayEnabled {
		brokers := kafkautil.ParseBrokers(env.String("KAFKA_BOOTSTRAP_SERVERS", ""), "localhost:19092")
		relay = ingest.NewRelay(store, brokers, logger)
		defer relay.Close()
		go relay.Run(runCtx, relayDelay)
	}

	mux := http.NewServeMux()
	ingest.NewHandler(ingest.NewService(store, highValueLimitEnabled)).Register(mux)
	registry := metrics.NewRegistry(ingest.ServiceName)
	mux.Handle("/metrics", registry.Handler())

	server := httpserver.New(fmt.Sprintf(":%d", port), registry.Middleware(mux), 10*time.Second)
	errs := make(chan error, 1)
	go func() {
		logger.Info("transaction ingestor listening", slog.Int("port", port), slog.Bool("relayEnabled", relayEnabled))
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
