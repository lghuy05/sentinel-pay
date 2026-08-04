# SentinelPay Go Platform Plan

## Goal

Run SentinelPay as a Go-based microservice platform while preserving the existing service boundaries, HTTP contracts, Kafka event flow, database ownership, Redis usage, Docker Compose wiring, observability, and UI behavior.

This project remains a microservice system. It is not a rewrite into a single binary.

## Current Runtime

Platform services are implemented in Go:

| Service | Port | Primary role | Data stores | Event contracts |
| --- | ---: | --- | --- | --- |
| `account-service` | 8087 | Account CRUD, balance top-up/debit, account lookup for enrichment | PostgreSQL `sentinelpay` | HTTP only |
| `transaction-ingestor` | 8081 | Accept transactions, persist raw transaction, publish raw event through outbox relay | PostgreSQL `sentinelpay`, Kafka | Produces `transactions.raw` |
| `feature-extractor` | 8082 | Consume raw transaction, enrich with account/velocity/device/context features | Redis, Kafka, account HTTP | Consumes `transactions.raw`, produces `transactions.enriched` |
| `blacklist-service` | 8084 | Check enriched transactions against active blacklist entries | PostgreSQL `sentinelpay`, Redis, Kafka | Consumes `transactions.enriched`, produces `fraud.blacklist` |
| `rule-engine` | 8083 | Evaluate fraud rules against pipeline inputs | PostgreSQL `sentinelpay`, Kafka | Consumes `fraud.blacklist`, produces `fraud.rules` |
| `fraud-orchestrator` | 8085 | Combine blacklist, rule, and ML signals into final fraud decisions | PostgreSQL `sentinelpay`, Redis, Kafka, ML HTTP | Consumes `fraud.blacklist`, `fraud.rules`, `fraud.ml`; produces `fraud.final` |
| `alert-service` | 8086 | Persist final decisions, expose alerts, apply transfer side effects/retry | PostgreSQL `sentinelpay`, Kafka, account HTTP | Consumes `fraud.final` |
| `api-gateway` | 8080 | HTTP gateway shell | None | HTTP routing target |

Other retained components:

| Service | Stack | Role |
| --- | --- | --- |
| `ml-service` | Python | Fraud ML scoring over Kafka and HTTP health/status |
| `sentinelpay-ui` | JavaScript/Vite | Demo and operations UI |

## Go Service Layout

```text
cmd/
  account-service/
  transaction-ingestor/
  feature-extractor/
  blacklist-service/
  rule-engine/
  fraud-orchestrator/
  alert-service/
  api-gateway/
internal/
  account/
  alert/
  blacklist/
  contracts/
  feature/
  gateway/
  ingest/
  orchestrator/
  platform/
  rules/
microservices/
  account-service-go/
  transaction-ingestor-go/
  feature-extractor-go/
  blacklist-service-go/
  rule-engine-go/
  fraud-orchestrator-go/
  alert-service-go/
  api-gateway-go/
  ml-services/
```

## Operating Principles

1. Preserve service boundaries and service names.
2. Preserve wire contracts before changing internals.
3. Keep Kafka topics, routing, and health routes stable.
4. Keep database ownership scoped to each service’s domain behavior.
5. Use explicit migrations and deterministic startup behavior.
6. Keep smoke tests, Kafka verification, and inter-service checks runnable locally and in CI.
7. Prefer one runtime path: the root Go Compose stack.

## Contract Surface

### HTTP routes to preserve

| Service | Routes |
| --- | --- |
| `account-service` | `POST /api/v1/accounts`, `GET /api/v1/accounts/{userId}`, `GET /api/v1/accounts`, `POST /api/v1/accounts/{userId}/topup`, `POST /api/v1/accounts/{userId}/debit`, `DELETE /api/v1/accounts/{userId}`, `GET /health/account-service` |
| `transaction-ingestor` | `POST /api/v1/transactions`, `GET /api/v1/transactions`, `GET /health/transaction-ingestor` |
| `feature-extractor` | `GET /health/feature-extractor` |
| `blacklist-service` | `GET /health/blacklist-service` |
| `rule-engine` | `GET /health/rule-engine` |
| `fraud-orchestrator` | `GET /decisions`, `GET /decisions/{transactionId}`, `POST /feedback`, `GET /ml/status`, `POST /ml/retrain`, `POST /ml/reload`, `GET /system/kafka`, `GET /system/redis`, `GET /system/services`, `GET /health/fraud-orchestrator` |
| `alert-service` | `GET /alerts`, `GET /alerts/{transactionId}`, `GET /health/alert-service` |

### Kafka topics to preserve

| Topic | Producer | Consumer |
| --- | --- | --- |
| `transactions.raw` | `transaction-ingestor` | `feature-extractor` |
| `transactions.enriched` | `feature-extractor` | `blacklist-service` |
| `fraud.blacklist` | `blacklist-service` | `rule-engine`, `fraud-orchestrator` |
| `fraud.rules` | `rule-engine` | `fraud-orchestrator` |
| `fraud.ml` | `ml-service` | `fraud-orchestrator` |
| `fraud.final` | `fraud-orchestrator` | `alert-service` |
| `*.DLT` | error handling | debugging and recovery |

## Current Verification Baseline

The current Go stack should continue to satisfy:

- `go test ./...`
- `go vet ./...`
- `./scripts/smoke-service-communication.sh`
- `./scripts/smoke-kafka-pipeline.sh`
- `./scripts/smoke-go-stack.sh`

The Compose stack should:

- wait for Postgres and Redis readiness before starting dependent services
- initialize Kafka topics without host bind-mount assumptions
- emit a final fraud decision for the end-to-end smoke transaction

## Near-Term Work

1. Keep CI aligned with the Go-only runtime path.
2. Keep Docker Compose and smoke scripts aligned with the root stack.
3. Maintain parity between API behavior, Kafka payloads, and UI expectations.
4. Tighten observability and startup diagnostics where needed.
5. Keep regression coverage around event ordering in the fraud pipeline.

## Acceptance Criteria

The platform is in a good state when:

- all Go services build and test cleanly
- the root Compose stack starts without manual intervention
- Kafka topics are created reliably
- a smoke transaction reaches `fraud.final`
- `alert-service` and `fraud-orchestrator` expose the final decision over HTTP
- the UI can run against the stack without route changes

## Notes

- Historical migration details should live in git history, not in the active operating plan.
- The repo is now Go-first for platform services; stale references to the retired migration path should be removed when encountered.
