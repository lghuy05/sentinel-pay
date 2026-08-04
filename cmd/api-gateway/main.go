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

	"github.com/lghuy05/sentinel-pay/internal/gateway"
	"github.com/lghuy05/sentinel-pay/internal/platform/config"
	"github.com/lghuy05/sentinel-pay/internal/platform/httpserver"
	"github.com/lghuy05/sentinel-pay/internal/platform/logging"
	"github.com/lghuy05/sentinel-pay/internal/platform/metrics"
)

func main() {
	logger := logging.New(gateway.ServiceName, config.NewEnv().String("LOG_LEVEL", "INFO"))
	if err := run(context.Background(), logger); err != nil {
		logger.Error("api gateway failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	env := config.NewEnv()
	port, err := env.Int("SERVER_PORT", gateway.DefaultPort)
	if err != nil {
		return err
	}
	proxy, err := gateway.NewProxy([]gateway.Route{
		{Prefix: "/api/v1/accounts", Target: env.String("ACCOUNT_SERVICE_URL", "http://localhost:8087")},
		{Prefix: "/api/v1/transactions", Target: env.String("TRANSACTION_INGESTOR_URL", "http://localhost:8081")},
		{Prefix: "/api/decisions", Target: env.String("FRAUD_ORCHESTRATOR_URL", "http://localhost:8085"), StripAPIPrefix: true},
		{Prefix: "/api/feedback", Target: env.String("FRAUD_ORCHESTRATOR_URL", "http://localhost:8085"), StripAPIPrefix: true},
		{Prefix: "/decisions", Target: env.String("FRAUD_ORCHESTRATOR_URL", "http://localhost:8085")},
		{Prefix: "/alerts", Target: env.String("ALERT_SERVICE_URL", "http://localhost:8086")},
		{Prefix: "/ml", Target: env.String("FRAUD_ORCHESTRATOR_URL", "http://localhost:8085")},
		{Prefix: "/system", Target: env.String("FRAUD_ORCHESTRATOR_URL", "http://localhost:8085")},
		{Prefix: "/health/account-service", Target: env.String("ACCOUNT_SERVICE_URL", "http://localhost:8087")},
		{Prefix: "/health/transaction-ingestor", Target: env.String("TRANSACTION_INGESTOR_URL", "http://localhost:8081")},
		{Prefix: "/health/feature-extractor", Target: env.String("FEATURE_EXTRACTOR_URL", "http://localhost:8082")},
		{Prefix: "/health/blacklist-service", Target: env.String("BLACKLIST_SERVICE_URL", "http://localhost:8084")},
		{Prefix: "/health/rule-engine", Target: env.String("RULE_ENGINE_URL", "http://localhost:8083")},
		{Prefix: "/health/fraud-orchestrator", Target: env.String("FRAUD_ORCHESTRATOR_URL", "http://localhost:8085")},
		{Prefix: "/health/alert-service", Target: env.String("ALERT_SERVICE_URL", "http://localhost:8086")},
		{Prefix: "/health/ml-service", Target: env.String("ML_SERVICE_URL", "http://localhost:5000")},
	})
	if err != nil {
		return err
	}
	registry := metrics.NewRegistry(gateway.ServiceName)
	mux := http.NewServeMux()
	mux.Handle("/metrics", registry.Handler())
	mux.Handle("/", proxy)
	server := httpserver.New(fmt.Sprintf(":%d", port), registry.Middleware(mux), 10*time.Second)
	errs := make(chan error, 1)
	go func() {
		logger.Info("api gateway listening", slog.Int("port", port))
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
