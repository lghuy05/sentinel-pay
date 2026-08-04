# SentinelPay Java to Go Refactor Tasks

## Ground Rules

- Do not delete Java code during the migration.
- Do not remove Maven files, Java Dockerfiles, Java entrypoints, or Java startup scripts during the migration.
- Add Go services beside the Java services first.
- Switch runtime traffic through Compose overrides, feature flags, or image tags.
- Remove Java only after the entire Go stack fully replaces the entire Java stack and the user explicitly approves cleanup.

## Task Status Legend

- `[ ]` Not started
- `[~]` In progress
- `[x]` Done
- `[!]` Blocked

## Phase 0: Baseline Current Java Behavior

- [ ] Start the current Java/Python/JS stack with the existing Compose setup.
- [ ] Verify Postgres, Redis, Kafka, and ML service containers are healthy.
- [ ] Verify Java service health endpoints:
  - [ ] `GET /health/account-service`
  - [ ] `GET /health/transaction-ingestor`
  - [ ] `GET /health/feature-extractor`
  - [ ] `GET /health/blacklist-service`
  - [ ] `GET /health/rule-engine`
  - [ ] `GET /health/fraud-orchestrator`
  - [ ] `GET /health/alert-service`
- [ ] Submit baseline transactions through `transaction-ingestor`.
- [ ] Capture HTTP request/response fixtures for account, transaction, decision, feedback, ML, system, and alert endpoints.
- [ ] Capture Kafka message fixtures for:
  - [ ] `transactions.raw`
  - [ ] `transactions.enriched`
  - [ ] `fraud.blacklist`
  - [ ] `fraud.rules`
  - [ ] `fraud.ml`
  - [ ] `fraud.final`
- [ ] Capture database schema snapshots after Java startup creates/updates tables.
- [ ] Capture expected final decisions for known `ALLOW`, `HOLD`, and `BLOCK` examples.
- [ ] Add a baseline smoke test script under `scripts/` that runs one full transaction through the current system.
- [ ] Add a Kafka verification script that checks topic existence, message production, message consumption, and DLT behavior.

## Phase 1: Go Workspace and Shared Platform

- [ ] Add root `go.work`.
- [ ] Add shared Go module/package layout under `internal/platform`.
- [ ] Add shared contract package under `internal/contracts`.
- [ ] Implement typed environment config loading.
- [ ] Implement HTTP server bootstrap with graceful shutdown.
- [ ] Implement service health helpers that preserve existing health route paths.
- [ ] Implement PostgreSQL connection helpers with `pgx`.
- [ ] Implement Redis connection helpers.
- [ ] Implement Kafka producer wrapper.
- [ ] Implement Kafka consumer wrapper with manual commit after successful processing.
- [ ] Implement DLT publishing behavior compatible with existing `*.DLT` topic convention.
- [ ] Implement JSON logging with `slog`.
- [ ] Implement Prometheus metrics helpers.
- [ ] Add unit tests for config, HTTP helpers, Kafka wrapper, Redis helper, and Postgres helper.
- [ ] Add fixture encode/decode tests for all shared event and HTTP DTO structs.
- [ ] Add `go test ./...` to local development docs.

## Phase 2: CI Foundation

- [ ] Update `.github/workflows/ci.yml` to keep existing Java build/test jobs during migration.
- [ ] Add Go setup and Go dependency cache.
- [ ] Add `go test ./...`.
- [ ] Add `go vet ./...`.
- [ ] Add optional formatting check with `gofmt`.
- [ ] Add SQL migration validation job.
- [ ] Add contract fixture test job.
- [ ] Add Docker image build checks for Go service Dockerfiles.
- [ ] Keep existing Python test detection and ML service checks.
- [ ] Add Compose smoke test job once Go services can start in CI.
- [ ] Avoid Docker push on pull requests unless explicitly intended.

## Phase 3: Docker and Compose Strategy

- [ ] Keep all current Java Dockerfiles unchanged.
- [ ] Add Go Dockerfiles beside Go service directories.
- [ ] Create `docker-compose.go.yml` as an override for running Go replacements.
- [ ] Keep the base `docker-compose.yml` stable until each Go service is proven.
- [ ] Add per-service image names for Go variants.
- [ ] Preserve service names, ports, env vars, health URLs, and dependencies.
- [ ] Add Compose profiles or override service definitions for Java-vs-Go switching.
- [ ] Update `scripts/start-services.sh` to support Java mode and Go mode without deleting Java behavior.
- [ ] Update `scripts/stop-services.sh` only if needed for Go local processes.
- [ ] Add `scripts/smoke-go-stack.sh`.
- [ ] Add `scripts/smoke-kafka-pipeline.sh`.
- [ ] Add `scripts/smoke-service-communication.sh`.

## Phase 4: Port `api-gateway`

- [ ] Implement Go `api-gateway` service.
- [ ] Preserve port `8080`.
- [ ] Add route table to downstream service URLs.
- [ ] Add unit tests for routing.
- [ ] Add integration tests against mock downstream services.
- [ ] Add Go Dockerfile.
- [ ] Add Compose override entry.
- [ ] Start gateway through Compose and verify health/routing.
- [ ] Keep Java `api-gateway` code untouched.

## Phase 5: Port `account-service`

- [ ] Implement Go HTTP routes matching Java account routes.
- [ ] Add SQL migrations for account tables.
- [ ] Add `sqlc` queries or explicit repository layer.
- [ ] Implement transactional top-up.
- [ ] Implement transactional debit with concurrency safety.
- [ ] Preserve account status and KYC enum JSON values.
- [ ] Add unit tests for validation and service logic.
- [ ] Add database integration tests.
- [ ] Add HTTP contract tests comparing Java fixtures to Go responses.
- [ ] Add Go Dockerfile.
- [ ] Add Compose override entry.
- [ ] Start Go `account-service` with Java stack and verify feature extractor can call it.
- [ ] Keep Java `account-service` code untouched.

## Phase 6: Port `transaction-ingestor`

- [ ] Implement Go transaction APIs.
- [ ] Add SQL migrations for transaction and outbox tables.
- [ ] Implement duplicate transaction handling.
- [ ] Implement high-value validation behavior.
- [ ] Implement outbox write in same DB transaction as transaction ingest.
- [ ] Implement outbox relay worker.
- [ ] Produce `transactions.raw` with Kafka key `transactionId`.
- [ ] Add unit tests for validation and mapping.
- [ ] Add DB integration tests for idempotency and outbox state transitions.
- [ ] Add Kafka integration test proving `transactions.raw` is emitted.
- [ ] Add Go Dockerfile.
- [ ] Add Compose override entry.
- [ ] Start Go ingestor with Java downstream pipeline and verify end-to-end transaction still flows.
- [ ] Keep Java `transaction-ingestor` code untouched.

## Phase 7: Port `feature-extractor`

- [ ] Implement `transactions.raw` consumer.
- [ ] Implement account HTTP client with timeout and retries.
- [ ] Port feature calculation logic.
- [ ] Port Redis velocity/device/account context behavior.
- [ ] Produce `transactions.enriched` with Kafka key `transactionId`.
- [ ] Add unit tests for feature calculation.
- [ ] Add Redis integration tests.
- [ ] Add Kafka input/output fixture tests.
- [ ] Add service communication test against Go or Java `account-service`.
- [ ] Add Go Dockerfile.
- [ ] Add Compose override entry.
- [ ] Start Go feature extractor with mixed stack and verify `transactions.enriched`.
- [ ] Keep Java `feature-extractor` code untouched.

## Phase 8: Port `blacklist-service`

- [ ] Implement `transactions.enriched` consumer.
- [ ] Add SQL migrations for blacklist tables.
- [ ] Implement active blacklist lookup.
- [ ] Preserve blacklist type enum values and reason strings.
- [ ] Preserve Redis caching behavior or document a compatible replacement.
- [ ] Produce `fraud.blacklist` with Kafka key `transactionId`.
- [ ] Add unit tests for blacklist matching.
- [ ] Add DB integration tests.
- [ ] Add Kafka input/output fixture tests.
- [ ] Add Go Dockerfile.
- [ ] Add Compose override entry.
- [ ] Start Go blacklist service and verify downstream Java rule engine and orchestrator consume its events.
- [ ] Keep Java `blacklist-service` code untouched.

## Phase 9: Port `rule-engine`

- [ ] Implement `fraud.blacklist` consumer.
- [ ] Add SQL migrations for fraud rules.
- [ ] Port default rule seeding.
- [ ] Port rule schema migration behavior into explicit migrations.
- [ ] Port risk scoring logic.
- [ ] Preserve rule names, score outputs, and explanation strings.
- [ ] Produce `fraud.rules` with Kafka key `transactionId`.
- [ ] Add unit tests for every rule.
- [ ] Add DB integration tests for rule loading.
- [ ] Add Kafka input/output fixture tests.
- [ ] Add Go Dockerfile.
- [ ] Add Compose override entry.
- [ ] Start Go rule engine and verify Java orchestrator consumes its events.
- [ ] Keep Java `rule-engine` code untouched.

## Phase 10: Port `fraud-orchestrator`

- [ ] Implement consumers for `fraud.blacklist`, `fraud.rules`, and `fraud.ml`.
- [ ] Port Redis signal aggregation.
- [ ] Preserve signal TTL behavior.
- [ ] Port final decision formula and thresholds.
- [ ] Add SQL migrations for decision records.
- [ ] Persist final decisions idempotently.
- [ ] Produce `fraud.final` with Kafka key `transactionId`.
- [ ] Implement decision, feedback, ML, system, Redis, Kafka, service health endpoints.
- [ ] Add unit tests for decision formula.
- [ ] Add Redis integration tests for signal aggregation.
- [ ] Add DB integration tests for decision persistence.
- [ ] Add Kafka multi-topic integration tests.
- [ ] Add service communication tests against ML service and downstream alert service.
- [ ] Add Go Dockerfile.
- [ ] Add Compose override entry.
- [ ] Start Go orchestrator and verify Java alert service consumes `fraud.final`.
- [ ] Keep Java `fraud-orchestrator` code untouched.

## Phase 11: Port `alert-service`

- [ ] Implement `fraud.final` consumer.
- [ ] Add SQL migrations for alerts, decisions, and transfers.
- [ ] Persist final decisions idempotently.
- [ ] Preserve alert query endpoints.
- [ ] Port transfer side effect behavior.
- [ ] Port retry worker behavior.
- [ ] Implement account HTTP client with timeout and retry policy.
- [ ] Add unit tests for decision handling and retry decisions.
- [ ] Add DB integration tests.
- [ ] Add Kafka input fixture tests.
- [ ] Add service communication tests against Go or Java `account-service`.
- [ ] Add Go Dockerfile.
- [ ] Add Compose override entry.
- [ ] Start Go alert service and verify dashboard alert reads.
- [ ] Keep Java `alert-service` code untouched.

## Phase 12: Full Go Stack Verification

- [ ] Start infrastructure: Postgres, Redis, Kafka, Zookeeper, ML service.
- [ ] Start every Go replacement through Compose override.
- [ ] Verify all Go health endpoints.
- [ ] Verify Kafka topics exist.
- [ ] Submit an `ALLOW` transaction and verify final dashboard state.
- [ ] Submit a `HOLD` transaction and verify final dashboard state.
- [ ] Submit a `BLOCK` transaction and verify final dashboard state.
- [ ] Verify `transactions.raw` message is produced.
- [ ] Verify `transactions.enriched` message is produced.
- [ ] Verify `fraud.blacklist` message is produced.
- [ ] Verify `fraud.rules` message is produced.
- [ ] Verify `fraud.ml` is consumed by orchestrator.
- [ ] Verify `fraud.final` message is produced.
- [ ] Verify alert service consumes `fraud.final`.
- [ ] Verify account service communication from feature extractor and alert service.
- [ ] Verify ML service communication from orchestrator.
- [ ] Verify Redis communication from feature extractor, blacklist service, rule engine, and orchestrator.
- [ ] Verify Postgres communication from all stateful services.
- [ ] Verify DLT behavior with malformed Kafka messages.
- [ ] Verify graceful shutdown and restart without losing outbox or Kafka state.
- [ ] Verify UI works against the Go stack without frontend route changes.

## Phase 13: Final CI Gate

- [ ] Java build/test still passes while Java remains in repo.
- [ ] Go test/vet/fmt checks pass.
- [ ] Python ML checks pass.
- [ ] Contract tests pass.
- [ ] SQL migration checks pass.
- [ ] Docker image builds pass for all Go services.
- [ ] Compose Go-stack smoke test passes.
- [ ] Kafka pipeline smoke test passes.
- [ ] Service communication smoke test passes.
- [ ] CI is green on pull request.

## Phase 14: Java Cleanup Planning Only

- [ ] Confirm every Java service has a fully verified Go replacement.
- [ ] Confirm base Compose can run the Go stack without Java service containers.
- [ ] Confirm CI no longer depends on Java behavior except retention checks.
- [ ] Ask for explicit approval before deleting Java code.
- [ ] Only after approval, create a separate cleanup task to remove Java source, Maven files, Java Dockerfiles, and Java-only scripts.
