# SentinelPay Java to Go Refactor Plan

## Goal

Replace every Java/Spring service in SentinelPay with Go services while preserving the current microservice architecture, external API behavior, Kafka event flow, database ownership, Redis usage, Docker Compose wiring, and UI/demo behavior.

This is not a rewrite into one Go binary. SentinelPay is a microservice project, so each service should remain independently buildable, deployable, observable, and replaceable.

## Java Retention Rule

Do not remove, delete, move, or gut any Java service code during the Go migration.

The Java implementation is the reference system and fallback until every Java microservice has a production-equivalent Go replacement, the full Go Compose stack is running, contract tests pass, Kafka pipeline tests pass, inter-service communication tests pass, and CI is green. Only after that full replacement milestone should Java removal be planned as a separate cleanup task.

## Current Java Service Inventory

| Service | Current port | Primary role | Data stores | Event contracts |
| --- | ---: | --- | --- | --- |
| `account-service` | 8087 | Account CRUD, balance top-up/debit, account lookup for enrichment | PostgreSQL `sentinelpay` | HTTP only |
| `transaction-ingestor` | 8081 | Accept transactions, persist raw transaction, publish raw event through outbox relay | PostgreSQL `transaction_ingestor_db`, Kafka | Produces `transactions.raw` |
| `feature-extractor` | 8082 | Consume raw transaction, enrich with account/velocity/device/context features | Redis, PostgreSQL config/schema placeholder, Kafka, account HTTP | Consumes `transactions.raw`, produces `transactions.enriched` |
| `blacklist-service` | 8084 | Check enriched transactions against active blacklist entries | PostgreSQL `blacklist_db`, Redis, Kafka | Consumes `transactions.enriched`, produces `fraud.blacklist` |
| `rule-engine` | 8083 | Evaluate blacklist/enriched transaction against fraud rules | PostgreSQL `rule_engine_db`, Redis, Kafka | Consumes `fraud.blacklist`, produces `fraud.rules` |
| `fraud-orchestrator` | 8085 | Combine blacklist, rule, and ML signals into final fraud decision | PostgreSQL `orchestrator_db`, Redis, Kafka, ML HTTP | Consumes `fraud.blacklist`, `fraud.rules`, `fraud.ml`; produces `fraud.final` |
| `alert-service` | 8086 | Persist final decisions, expose alerts, apply transfer side effects/retry | PostgreSQL `alert_db`, Kafka, account HTTP | Consumes `fraud.final` |
| `api-gateway` | 8080 | Gateway shell | None visible in code | HTTP routing target |

Non-Java services stay in place unless separately scoped:

| Service | Stack | Keep as-is? |
| --- | --- | --- |
| `ml-service` | Python | Yes. Go services should continue calling it over HTTP and consuming its Kafka output. |
| `sentinelpay-ui` | JavaScript/Vite | Yes. Preserve API paths and ports so the UI does not need a broad rewrite. |

## Migration Principles

1. Preserve service boundaries. Each Java service gets a corresponding Go service with the same container/service name, port, health route, API routes, Kafka topics, consumer group intent, database ownership, and environment variables.
2. Preserve wire contracts before improving internals. JSON field names, enum strings, timestamp formats, HTTP status codes, pagination shape, and Kafka keys must match Spring behavior first.
3. Migrate one service at a time behind the existing Docker Compose names. The rest of the system should not know whether a service is Java or Go.
4. Avoid shared database ownership expansion. Go services may read/write only the tables owned by the Java service they replace.
5. Replace Hibernate auto-DDL with explicit SQL migrations. Go services should not create schemas implicitly at startup.
6. Keep Kafka event handling durable. Maintain retries, dead-letter topics, idempotency, and offset commit behavior explicitly in Go.
7. Use a small, consistent Go stack across services to reduce operational drift.
8. Keep Java source, Maven files, Java Dockerfiles, and Java startup scripts intact until the complete Go stack has replaced the complete Java stack.

## Target Go Stack

Use one repo-level Go workspace:

```text
go.work
microservices/
  account-service-go/
  transaction-ingestor-go/
  feature-extractor-go/
  blacklist-service-go/
  rule-engine-go/
  fraud-orchestrator-go/
  alert-service-go/
  api-gateway-go/
internal/
  platform/
    config/
    httpserver/
    kafka/
    postgres/
    redis/
    observability/
    validation/
  contracts/
    events/
    httpdto/
```

Recommended libraries:

| Concern | Go choice |
| --- | --- |
| HTTP router | `chi` or `gin`; prefer `chi` for small, explicit services |
| PostgreSQL | `pgx` |
| SQL queries | `sqlc` for typed queries |
| Migrations | `goose` or `atlas`; prefer `goose` for simple per-service migrations |
| Kafka | `segmentio/kafka-go` or `confluent-kafka-go`; prefer `segmentio/kafka-go` unless Confluent-specific features are required |
| Redis | `go-redis` |
| Config | environment variables parsed into typed structs |
| Logging | `slog` with JSON output |
| Metrics | Prometheus client |
| Validation | `go-playground/validator` or explicit validation functions |
| Tests | standard `testing`, `httptest`, `testcontainers-go` for integration |

## Contract Freeze

Before writing Go replacements, freeze the current contracts in tests and fixtures.

### HTTP contracts to capture

| Service | Routes to preserve |
| --- | --- |
| `account-service` | `POST /api/v1/accounts`, `GET /api/v1/accounts/{userId}`, `GET /api/v1/accounts`, `POST /api/v1/accounts/{userId}/topup`, `POST /api/v1/accounts/{userId}/debit`, `DELETE /api/v1/accounts/{userId}`, `GET /health/account-service` |
| `transaction-ingestor` | `POST /api/v1/transactions`, `GET /api/v1/transactions`, `GET /health/transaction-ingestor` |
| `feature-extractor` | `GET /health/feature-extractor` |
| `blacklist-service` | `GET /health/blacklist-service` |
| `rule-engine` | `GET /health/rule-engine` |
| `fraud-orchestrator` | `GET /decisions`, `GET /decisions/{transactionId}`, `POST /feedback`, `GET /ml/status`, `POST /ml/retrain`, `POST /ml/reload`, `GET /system/kafka`, `GET /system/redis`, `GET /system/services`, `GET /health/fraud-orchestrator` |
| `alert-service` | `GET /alerts`, `GET /alerts/{transactionId}`, `GET /health/alert-service` |

### Kafka contracts to capture

| Topic | Producer | Consumer |
| --- | --- | --- |
| `transactions.raw` | `transaction-ingestor` | `feature-extractor` |
| `transactions.enriched` | `feature-extractor` | `blacklist-service` |
| `fraud.blacklist` | `blacklist-service` | `rule-engine`, `fraud-orchestrator` |
| `fraud.rules` | `rule-engine` | `fraud-orchestrator` |
| `fraud.ml` | `ml-service` | `fraud-orchestrator` |
| `fraud.final` | `fraud-orchestrator` | `alert-service` |
| `*.DLT` | Kafka error handlers | Operational debugging |

Create fixture files under `contracts/fixtures/`:

```text
contracts/fixtures/http/
contracts/fixtures/kafka/
contracts/fixtures/db/
```

Each fixture should include at least one happy path, validation failure, duplicate/idempotency path, and missing dependency path where applicable.

## Phased Migration

### Phase 0: Baseline and Safety Net

Deliverables:

- Add contract tests against the existing Java services.
- Add JSON fixtures for all Kafka event structs.
- Add database schema snapshots for each Java-owned database.
- Add a local smoke test that runs the full current Compose stack and submits one transaction through the full fraud pipeline.
- Document current expected decisions for `ALLOW`, `HOLD`, and `BLOCK`.

Acceptance criteria:

- Current Java stack passes smoke tests.
- Test suite can detect an incompatible JSON, API, topic, or DB behavior change.
- No Go migration starts until the baseline is reproducible.

### Phase 1: Shared Go Platform

Deliverables:

- Create `go.work` and shared `internal/platform` packages.
- Implement config loading from the same env vars used by Compose: `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `REDIS_HOST`, `REDIS_PORT`, `KAFKA_BOOTSTRAP_SERVERS`, service URL env vars.
- Implement common HTTP server boot/shutdown, `/health/...` helpers, Kafka producer/consumer wrappers, JSON logging, Prometheus metrics, PostgreSQL and Redis clients.
- Implement event structs in `internal/contracts/events` with JSON tags matching Spring payloads exactly.
- Implement HTTP DTOs in `internal/contracts/httpdto` with JSON tags matching Spring payloads exactly.

Acceptance criteria:

- `go test ./...` passes.
- Contract fixture encode/decode tests pass.
- Go services can be built into small Docker images.

### Phase 2: Low-Risk Edge Services

Migrate services with limited dependencies first.

1. `api-gateway`
   - Replace Spring shell with a Go reverse proxy or simple gateway.
   - Preserve port `8080`.
   - Add explicit route table to downstream services.

2. `account-service`
   - Replace account CRUD and balance mutation APIs.
   - Convert JPA repository behavior into explicit SQL queries.
   - Add database migrations for `accounts`.
   - Pay special attention to balance updates and concurrent debit safety.

Acceptance criteria:

- UI and feature extractor can call Go `account-service` without changes.
- Account balance operations are transactional and race-tested.
- Compose can run Java and Go variants side by side using alternate image tags during testing.

### Phase 3: Transaction Ingestion

Migrate `transaction-ingestor`.

Deliverables:

- Implement `POST /api/v1/transactions` and transaction list endpoint.
- Implement explicit transaction table and outbox table migrations.
- Preserve `event.publisher=kafka` and a no-op mode if still useful.
- Rebuild outbox relay in Go with retry count, next attempt time, status transitions, and Kafka key `transactionId`.
- Preserve high-value validation behavior.

Acceptance criteria:

- Duplicate transaction handling matches Java behavior.
- A persisted transaction always results in an eventual `transactions.raw` event unless the outbox is exhausted and marked failed.
- Outbox relay can restart without losing or duplicating non-idempotent state.

### Phase 4: Pipeline Consumers

Migrate event processors from left to right.

1. `feature-extractor`
   - Consume `transactions.raw`.
   - Call `account-service`.
   - Use Redis for velocity and feature cache behavior.
   - Produce `transactions.enriched`.

2. `blacklist-service`
   - Consume `transactions.enriched`.
   - Query active blacklist records.
   - Preserve Redis optimization.
   - Produce `fraud.blacklist`.

3. `rule-engine`
   - Consume `fraud.blacklist`.
   - Port `DefaultRuleSet`, schema migration logic, and risk scoring exactly first.
   - Produce `fraud.rules`.

Acceptance criteria:

- For each service, feed captured input fixtures and compare output events byte-compatible where possible and semantically equivalent where ordering/timestamps differ.
- Consumer offsets commit only after successful processing and publish.
- Failure path publishes to matching `.DLT` topic or has an explicitly documented equivalent.

### Phase 5: Decision and Alerting Services

1. `fraud-orchestrator`
   - Consume `fraud.blacklist`, `fraud.rules`, and `fraud.ml`.
   - Preserve Redis signal aggregation keys and TTL behavior.
   - Preserve final decision formula and thresholds.
   - Persist decisions to `orchestrator_db`.
   - Produce `fraud.final`.
   - Preserve dashboard/system endpoints.

2. `alert-service`
   - Consume `fraud.final`.
   - Persist alert and transfer records.
   - Preserve retry behavior for retryable transfer failures.
   - Preserve account balance side effects through `account-service`.

Acceptance criteria:

- Full pipeline produces the same final decision and reason fields for baseline fixtures.
- Dashboard alert and decision views work without frontend changes.
- Retry workers are idempotent and restart-safe.

### Phase 6: Compose, CI, and Java Retention

Deliverables:

- Add Go multi-stage Dockerfiles beside the Go services without deleting the Java Dockerfiles.
- Update Compose through explicit Go override files first, while preserving service names and ports.
- Keep root Maven `pom.xml`, service `pom.xml`, `mvnw`, `mvnw.cmd`, Java Dockerfiles, and `src/main/java` trees in the repo during migration.
- Plan Java deletion only after all services are fully replaced in Go and the full Go-only stack passes CI and smoke tests.
- Add CI jobs:
  - `go test ./...`
  - SQL migration validation
  - contract fixture tests
  - Docker image builds
  - Compose smoke test

Acceptance criteria:

- Java and Go variants can coexist during development.
- Compose starts the full Go stack through an override file.
- Compose starts all services.
- End-to-end transaction demo works.
- UI still displays account, transaction, decision, alert, and health data.

Java cleanup is a later milestone, not part of service-by-service porting. The cleanup acceptance criterion will be:

- `rg -n "\.java$|pom.xml|mvnw|SpringApplication|org.springframework" microservices pom.xml` returns no Java application code, after the user explicitly approves Java removal.

## Per-Service Porting Notes

### `account-service`

- Replace JPA entity `Account` and repository methods with SQL.
- Use row-level locking or atomic SQL for debit/top-up to avoid lost updates.
- Preserve account status and KYC enum strings.
- Return the same response shape as `AccountResponse`.

### `transaction-ingestor`

- Preserve the outbox pattern. Do not publish to Kafka before the transaction record commits.
- Use explicit unique constraints for `transactionId`.
- Use context timeouts around database writes and Kafka publish.
- Port validation and error response shape from `GlobalExceptionHandler`.

### `feature-extractor`

- Keep feature calculation deterministic.
- Put account HTTP calls behind an interface and timeout.
- Redis key naming and TTLs must be carried over or intentionally versioned.
- Kafka message key remains `transactionId`.

### `blacklist-service`

- Cache active blacklist entries carefully. Invalidation can be simple at first because the current code mostly reads active entries.
- Preserve blacklist hit reason strings because they feed final explanations.
- Keep `BlacklistType` enum values stable.

### `rule-engine`

- Port default rule seeding before changing schema.
- Keep rule score outputs stable.
- Preserve rule names/reasons, because UI and final explanations depend on them.

### `fraud-orchestrator`

- This is the highest coordination risk service.
- Port after upstream event producers are stable in Go.
- Signal aggregation must be idempotent because Kafka can redeliver.
- Final decision should be written and published once per transaction, or the duplicate behavior must match the current Java code.

### `alert-service`

- Make transfer side effects idempotent.
- Store consumed final decisions before side effects so failed account calls can be retried.
- Preserve retry delay configuration.

### `api-gateway`

- If it is currently only a Spring Boot placeholder, replace it with a thin Go gateway.
- If UI already calls services directly by exposed ports, either keep gateway minimal or move UI routing behind it as a separate explicit task.

## Database Migration Strategy

1. Generate schema snapshots from the current Java-created tables.
2. Write explicit `goose` migrations per service under:

```text
microservices/<service>-go/migrations/
```

3. Use one database per existing service database. Do not collapse databases during this migration.
4. Add unique constraints for event idempotency:
   - `transactions.transaction_id`
   - `outbox_events.event_id` or equivalent
   - `fraud_decisions.transaction_id`
   - `alerts.transaction_id`
5. Use migration checks in CI before service startup tests.

## Kafka Semantics

Go consumers must explicitly define:

- topic name
- consumer group
- partitioning key
- JSON schema
- retry count
- dead-letter behavior
- commit timing
- concurrency
- shutdown behavior

Consumer rule:

```text
decode -> validate -> process -> publish/store side effects -> commit offset
```

Producer rule:

```text
set key = transactionId when event belongs to a transaction
disable Java/Spring type headers
use JSON field names from fixtures
log topic, partition, offset, transactionId
```

## Observability

Each Go service should expose:

- service-specific health endpoint matching the current route
- structured JSON logs
- request count, latency, and error metrics
- Kafka consumed/produced counters
- Kafka processing latency from event timestamp where available
- PostgreSQL and Redis dependency health checks where meaningful

## Migration Order

Recommended order:

1. Contract tests and fixtures
2. Shared Go platform
3. `api-gateway`
4. `account-service`
5. `transaction-ingestor`
6. `feature-extractor`
7. `blacklist-service`
8. `rule-engine`
9. `fraud-orchestrator`
10. `alert-service`
11. Compose/CI cleanup
12. Full Go stack sign-off
13. Separate Java removal task only after explicit approval

Reasoning:

- `account-service` is needed by `feature-extractor` and `alert-service`.
- `transaction-ingestor` starts the event pipeline and validates outbox behavior early.
- `fraud-orchestrator` should wait until all upstream event shapes are proven stable.
- `alert-service` should be last because it depends on final decision correctness and account side effects.

## Definition of Done

The migration is complete when:

- Every previous Java service has a Go replacement with the same service name, port, health route, and public contract.
- The full Kafka fraud pipeline works from transaction ingestion to final alert.
- The Python ML service and UI continue to work unchanged.
- All service images build from Go Dockerfiles.
- Contract, unit, integration, and Compose smoke tests pass in CI.
- Operational docs describe how to run, test, and troubleshoot the Go stack.
- Java code remains available as a reference until a separately approved cleanup removes it.

## Immediate Next Actions

1. Add contract fixture capture tests for the existing Java services.
2. Create the Go workspace and shared platform packages.
3. Port `account-service` first as the first real Java removal.
4. Run both Java and Go `account-service` against the same fixtures until responses match.
5. Replace the Compose image for `account-service` only after the matching tests pass.
