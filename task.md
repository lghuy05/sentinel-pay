# SentinelPay Java to Go Refactor Tasks

## Rules

- Do not delete Java code for a service until its Go replacement is implemented, wired into runtime, tested, and verified against the old behavior.
- Do not remove Maven files, Java Dockerfiles, Java entrypoints, or Java startup scripts for a service until that service's Go replacement is done.
- Add Go services beside the Java services first.
- Switch runtime traffic through Compose overrides, profiles, feature flags, or image tags.
- Remove Java service-by-service only after the matching Go service is complete and verified.

## Phase 0: Capture Current Java Behavior

- [x] Start the current Java/Python/JS stack.
- [x] Verify Postgres, Redis, Kafka, Zookeeper, and ML service are healthy.
- [x] Verify all Java service health endpoints.
- [x] Submit baseline transactions through `transaction-ingestor`.
- [x] Capture HTTP request/response fixtures for account, transaction, decision, feedback, ML, system, and alert endpoints.
- [x] Capture Kafka fixtures for `transactions.raw`, `transactions.enriched`, `fraud.blacklist`, `fraud.rules`, `fraud.ml`, and `fraud.final`.
- [x] Capture database schema snapshots after Java startup.
- [x] Capture expected `ALLOW`, `HOLD`, and `BLOCK` decisions.
- [x] Add `scripts/smoke-java-stack.sh`.
- [x] Add `scripts/smoke-account-service.sh`.
- [x] Add `scripts/smoke-kafka-pipeline.sh`.
- [x] Add `scripts/smoke-service-communication.sh`.
- [x] Add `scripts/smoke-compose-network.sh`.

## Phase 1: Go Workspace

- [x] Add root `go.work`.
- [x] Add shared Go packages under `internal/platform`.
- [x] Add shared contract packages under `internal/contracts`.
- [x] Implement typed environment config loading.
- [x] Implement HTTP server bootstrap and graceful shutdown.
- [x] Implement health route helpers preserving current paths.
- [x] Implement PostgreSQL helpers with `pgx`.
- [x] Implement Redis helpers.
- [x] Implement Kafka producer and consumer wrappers.
- [x] Implement manual Kafka offset commit after successful processing.
- [x] Implement DLT publishing compatible with `*.DLT`.
- [x] Implement JSON logging with `slog`.
- [x] Implement Prometheus metrics helpers.
- [x] Add unit tests for shared platform code.
- [x] Add JSON fixture encode/decode tests.

## Phase 2: CI

- [x] Keep existing Java Maven build/test jobs during migration.
- [x] Add Go setup and dependency cache to `.github/workflows/ci.yml`.
- [x] Add `go test ./...`.
- [x] Add `go vet ./...`.
- [x] Add `gofmt` check.
- [x] Add SQL migration validation.
- [x] Add contract fixture tests.
- [x] Add Go Docker image build checks.
- [x] Keep Python ML dependency/test checks.
- [x] Add Compose smoke tests once Go services can start in CI.
- [x] Avoid Docker push on pull requests unless explicitly intended.

## Phase 3: Docker and Compose

- [x] Keep current Java Dockerfiles unchanged.
- [x] Add Go Dockerfiles beside Go services.
- [x] Create `docker-compose.go.yml` for Go replacements.
- [x] Keep base `docker-compose.yml` stable until Go services are proven.
- [x] Preserve service names, ports, env vars, dependencies, and health routes for `account-service`.
- [x] Add Compose profiles or overrides for Java-vs-Go switching.
- [x] Update `scripts/start-services.sh` to support Java mode and Go mode for `account-service`.
- [x] Update `scripts/stop-services.sh` for Go account-service local process cleanup.
- [x] Add `scripts/smoke-go-stack.sh`.
- [x] Add Kafka pipeline smoke test.
- [x] Add service communication smoke test.
- [x] Add Compose-network end-to-end smoke test.

## Phase 4: Port `api-gateway`

- [x] Implement Go `api-gateway`.
- [x] Preserve port `8080`.
- [x] Add downstream route table.
- [x] Add routing unit tests.
- [x] Add integration tests with mock downstream services.
- [x] Add Go Dockerfile.
- [x] Wire root Compose `api-gateway` to the Go image.
- [x] Start through Compose and verify routing.
- [x] Keep Java `api-gateway` untouched.

## Phase 5: Port `account-service`

- [x] Implement Go account HTTP routes.
- [x] Add SQL migrations for account tables.
- [x] Implement account repository/query layer.
- [x] Implement transactional top-up.
- [x] Implement transactional debit with concurrency safety.
- [x] Preserve account status and KYC enum JSON values.
- [x] Add validation and service unit tests.
- [x] Add database integration tests.
- [x] Add HTTP contract tests against captured Java fixtures.
- [x] Add Go Dockerfile.
- [x] Wire root Compose `account-service` to the Go image.
- [x] Run Go account service with Java stack.
- [x] Verify feature extractor can call Go account service.
- [x] Keep Java `account-service` source untouched until DB integration and contract fixture checks are added.

## Phase 6: Port `transaction-ingestor`

- [x] Implement Go transaction APIs.
- [x] Add SQL migrations for transaction and outbox tables.
- [x] Implement duplicate transaction handling.
- [x] Implement high-value validation.
- [x] Persist transaction and outbox record in one DB transaction.
- [x] Implement outbox relay worker.
- [x] Produce `transactions.raw` with Kafka key `transactionId`.
- [x] Add validation and mapping unit tests.
- [x] Add DB integration tests for idempotency and outbox state transitions.
- [x] Add Kafka smoke proving `transactions.raw` is emitted and consumed downstream.
- [x] Add Go Dockerfile.
- [x] Wire root Compose `transaction-ingestor` to the Go image.
- [x] Run Go ingestor with Java downstream pipeline.
- [x] Keep Java `transaction-ingestor` source untouched until DB integration and contract fixture checks are added.

## Phase 7: Port `feature-extractor`

- [x] Implement `transactions.raw` consumer.
- [x] Implement account HTTP client with timeout.
- [x] Port feature calculation logic.
- [x] Port Redis velocity/device/account context behavior.
- [x] Produce `transactions.enriched` with Kafka key `transactionId`.
- [x] Add feature calculation unit tests.
- [x] Add Redis integration tests.
- [x] Add Kafka smoke proving `transactions.enriched` is produced and consumed downstream.
- [x] Add service communication tests against account service through E2E smoke.
- [x] Add Go Dockerfile.
- [x] Wire root Compose `feature-extractor` to the Go image.
- [x] Verify `transactions.enriched` is produced.
- [x] Keep Java `feature-extractor` source untouched until Redis integration and Kafka fixture checks are added.

## Phase 8: Port `blacklist-service`

- [x] Implement `transactions.enriched` consumer.
- [x] Add SQL migrations for blacklist tables.
- [x] Implement active blacklist lookup.
- [x] Preserve blacklist type enum values and reason strings.
- [x] Preserve Redis caching behavior or document compatible replacement.
- [x] Produce `fraud.blacklist` with Kafka key `transactionId`.
- [x] Add blacklist matching unit tests.
- [x] Add DB integration tests.
- [x] Add Kafka smoke proving `fraud.blacklist` is consumed downstream.
- [x] Add Go Dockerfile.
- [x] Wire root Compose `blacklist-service` to the Go image.
- [x] Verify Java rule engine and orchestrator consume Go blacklist events.
- [x] Keep Java `blacklist-service` source untouched until DB integration and Kafka fixture checks are added.

## Phase 9: Port `rule-engine`

- [x] Implement `fraud.blacklist` consumer.
- [x] Add SQL migrations for fraud rules.
- [x] Port default rule seeding.
- [x] Port risk scoring logic.
- [x] Preserve rule names, scores, and explanations.
- [x] Produce `fraud.rules` with Kafka key `transactionId`.
- [x] Add focused unit tests for hard rules, score stacking, score caps, and fixture smoke path.
- [x] Add DB integration tests.
- [x] Add Kafka input/output fixture tests.
- [x] Add Go Dockerfile.
- [x] Wire root Compose `rule-engine` to the Go image.
- [x] Verify Java orchestrator consumes Go rule events.
- [x] Keep Java `rule-engine` untouched.

## Phase 10: Port `fraud-orchestrator`

- [x] Implement consumers for `fraud.blacklist`, `fraud.rules`, and `fraud.ml`.
- [x] Port Redis signal aggregation and TTL behavior.
- [x] Port final decision formula and thresholds.
- [x] Add SQL migrations for decision records.
- [x] Persist final decisions idempotently.
- [x] Produce `fraud.final` with Kafka key `transactionId`.
- [x] Implement feedback and ML proxy endpoints.
- [x] Implement decision, system, Redis, Kafka, and health endpoints.
- [x] Add focused decision formula unit tests.
- [x] Add Redis integration tests.
- [x] Add DB integration tests.
- [x] Add Kafka multi-topic integration tests.
- [x] Add ML and alert service communication tests.
- [x] Add Go Dockerfile.
- [x] Wire root Compose `fraud-orchestrator` to the Go image.
- [x] Verify Java alert service consumes Go final decisions.
- [x] Keep Java `fraud-orchestrator` untouched.

## Phase 11: Port `alert-service`

- [x] Implement `fraud.final` consumer.
- [x] Add SQL migrations for alerts and transfers.
- [x] Persist final decisions idempotently.
- [x] Preserve alert query endpoints.
- [x] Port transfer side effects.
- [x] Port retry worker behavior.
- [x] Implement account HTTP client with timeout.
- [x] Add focused decision handling unit tests.
- [x] Add DB integration tests.
- [x] Add Kafka input fixture tests.
- [x] Add account service communication smoke coverage.
- [x] Add Go Dockerfile.
- [x] Wire root Compose `alert-service` to the Go image.
- [x] Verify alert query reads.
- [x] Keep Java `alert-service` untouched.

## Phase 12: Full Go Stack Verification

- [x] Start infrastructure: Postgres, Redis, Kafka, Zookeeper, ML service.
- [x] Start every Go replacement through Compose override.
- [x] Verify all Go health endpoints.
- [x] Verify Kafka topics exist.
- [x] Submit `ALLOW`, `HOLD`, and `BLOCK` transactions.
- [x] Verify `transactions.raw` is produced.
- [x] Verify `transactions.enriched` is produced.
- [x] Verify `fraud.blacklist` is produced.
- [x] Verify `fraud.rules` is produced.
- [x] Verify `fraud.ml` is consumed.
- [x] Verify `fraud.final` is produced.
- [x] Verify alert service consumes `fraud.final`.
- [x] Verify account service communication from feature extractor and alert service.
- [x] Verify ML service communication from orchestrator.
- [x] Verify Redis communication from all Redis-backed services.
- [x] Verify Postgres communication from all stateful services.
- [x] Verify DLT behavior with malformed Kafka messages.
- [x] Verify graceful shutdown and restart.
- [x] Verify UI works against the Go stack.

## Phase 13: Final CI Gate

- [x] Java build/test still passes while Java remains in repo.
- [x] Go test/vet/fmt checks pass.
- [x] Python ML checks pass.
- [x] Contract tests pass.
- [x] SQL migration checks pass.
- [x] Docker image builds pass for all Go services.
- [x] Compose Go-stack smoke test passes.
- [x] Kafka pipeline smoke test passes.
- [x] Service communication smoke test passes.
- [ ] CI is green on pull request.

## Phase 14: Java Cleanup Planning Only

- [x] Confirm every Java service has a verified Go replacement.
- [x] Confirm base Compose can run Go services without Java service containers.
- [x] Confirm CI no longer depends on Java behavior except retention checks.
- [x] Ask for explicit approval before deleting Java code.
- [x] After approval, create a separate cleanup task for Java source, Maven files, Java Dockerfiles, and Java-only scripts.
