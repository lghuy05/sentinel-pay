#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/docker-compose.yml}"
RUNNER_SERVICE="${SMOKE_RUNNER_SERVICE:-host}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

if [[ "$RUNNER_SERVICE" != "host" ]]; then
  require_cmd docker
fi

if [[ "$RUNNER_SERVICE" == "host" ]]; then
  echo "Running end-to-end smoke through Compose host ports..."
  ACCOUNT_URL="${ACCOUNT_URL:-http://localhost:18080}" \
  INGESTOR_URL="${INGESTOR_URL:-http://localhost:18087}" \
  FEATURE_URL="${FEATURE_URL:-http://localhost:18084}" \
  BLACKLIST_URL="${BLACKLIST_URL:-http://localhost:18083}" \
  RULE_URL="${RULE_URL:-http://localhost:18086}" \
  ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-http://localhost:18085}" \
  ALERT_URL="${ALERT_URL:-http://localhost:18081}" \
  ML_URL="${ML_URL:-http://localhost:5000}" \
  "$ROOT_DIR/scripts/smoke-e2e-stack.sh"
else
  echo "Running end-to-end smoke inside Compose network via $RUNNER_SERVICE..."
  docker compose -f "$COMPOSE_FILE" exec -T "$RUNNER_SERVICE" \
    env \
    ACCOUNT_URL="${ACCOUNT_URL:-http://account-service:8087}" \
    INGESTOR_URL="${INGESTOR_URL:-http://transaction-ingestor:8081}" \
    FEATURE_URL="${FEATURE_URL:-http://feature-extractor:8082}" \
    BLACKLIST_URL="${BLACKLIST_URL:-http://blacklist-service:8084}" \
    RULE_URL="${RULE_URL:-http://rule-engine:8083}" \
    ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-http://fraud-orchestrator:8085}" \
    ALERT_URL="${ALERT_URL:-http://alert-service:8086}" \
    ML_URL="${ML_URL:-http://ml-service:5000}" \
    bash -s < "$ROOT_DIR/scripts/smoke-e2e-stack.sh"
fi
