# Java Cleanup Task

Only execute these removals after the Go stack remains green in CI and the retained Java fixtures are no longer needed.

1. Remove Java service source trees under:
   - `microservices/account-service`
   - `microservices/alert-service`
   - `microservices/api-gateway`
   - `microservices/blacklist-service`
   - `microservices/feature-extractor`
   - `microservices/fraud-orchestrator`
   - `microservices/rule-engine`
   - `microservices/transaction-ingestor`
2. Remove Maven wrapper files and Java-only build metadata from the repo root and service directories.
3. Remove Java Dockerfiles and Java Compose override once Go-only deployment is the only supported path.
4. Remove Java build/test stages from `.github/workflows/ci.yml`.
5. Remove Java mode branches from `scripts/start-services.sh` and `scripts/stop-services.sh`.
6. Keep contract fixtures, smoke scripts, and Go stack verification in place after cleanup.
