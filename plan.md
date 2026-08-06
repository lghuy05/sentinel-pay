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

## Dataset Direction

SentinelPay should move its fraud model training away from repo-generated demo CSVs and toward the public PaySim dataset:

- Dataset: https://www.kaggle.com/datasets/ealaxi/paysim1?resource=download
- See: [docs/paysim-dataset-note.md](/home/huy/workspace/projects/sentinel-pay/docs/paysim-dataset-note.md)

Important nuance:

- PaySim is public and much stronger than the current self-created dataset
- PaySim is still simulator-generated, not a raw production transaction dump
- we should call it a public, real-world-inspired fraud dataset

## ML Objective

The ML layer should aim to catch as much fraud as possible, not maximize plain accuracy.

Primary evaluation targets:

1. high recall on fraud
2. strong PR-AUC
3. acceptable precision at deployment thresholds
4. stable score calibration
5. manageable false positives for downstream review and alerting

Accuracy should not be used as the main headline metric because the fraud target is heavily imbalanced.

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
6. Replace self-generated ML training data with a PaySim-backed benchmark training pipeline.
7. Keep repo-generated demo data only for quick local development and smoke usage.
8. Compare simple ML baselines against stronger tabular models before adopting deep learning.

## ML Refactor Plan

### Phase 1: Dataset Adoption

- add PaySim dataset notes and usage guidance to the repo
- add a dataset ingestion script for Kaggle-exported CSV input
- add schema validation for expected PaySim columns
- keep raw dataset files out of git
- store only code, metadata, and reproducible transforms in the repo

### Phase 2: Feature Pipeline

- map PaySim fields into a reproducible training feature pipeline
- explicitly prevent leakage-heavy balance fields from becoming default training inputs
- derive behavioral, temporal, and transaction-type features
- add train/validation/test splits that respect time ordering
- add a separate raw PaySim schema and canonical training schema instead of forcing PaySim columns into live service contracts
- keep `transactions.enriched` as the online inference contract, with a feature builder that maps it into the same model feature space

### Phase 3: Baselines

- keep logistic regression as the first benchmark
- add gradient boosting benchmarks
- record recall, PR-AUC, precision, threshold behavior, and calibration
- make the best non-deep baseline the minimum bar for any future deep model

### Phase 4: Advanced Modeling

- evaluate deep learning only after strong tabular baselines are stable
- candidates may include MLPs for tabular features or temporal/sequence models if event-history context is added
- do not ship deep learning just because it is more complex; it must improve fraud capture materially

### Phase 5: Productionization

- version model artifacts with dataset source and training metadata
- expose model version, dataset source, and evaluation summary through ML status endpoints
- keep inference feature expectations aligned with `transactions.enriched`
- add retraining notes and threshold-management guidance for operators

## Acceptance Criteria

The platform is in a good state when:

- all Go services build and test cleanly
- the root Compose stack starts without manual intervention
- Kafka topics are created reliably
- a smoke transaction reaches `fraud.final`
- `alert-service` and `fraud-orchestrator` expose the final decision over HTTP
- the UI can run against the stack without route changes
- the ML service has a documented path from local demo data to PaySim-backed benchmark training
- future model upgrades are evaluated on fraud recall and PR-AUC, not accuracy alone

## Notes

- Historical migration details should live in git history, not in the active operating plan.
- The repo is now Go-first for platform services; stale references to the retired migration path should be removed when encountered.
