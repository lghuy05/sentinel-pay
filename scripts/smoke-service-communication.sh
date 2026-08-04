#!/usr/bin/env bash
set -euo pipefail

ACCOUNT_URL="${ACCOUNT_URL:-http://localhost:18080}"
INGESTOR_URL="${INGESTOR_URL:-http://localhost:18087}"
FEATURE_URL="${FEATURE_URL:-http://localhost:18084}"
RULE_URL="${RULE_URL:-http://localhost:18086}"
BLACKLIST_URL="${BLACKLIST_URL:-http://localhost:18083}"
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-http://localhost:18085}"
ALERT_URL="${ALERT_URL:-http://localhost:18081}"
ML_URL="${ML_URL:-http://localhost:5000}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

check_http() {
  local name="$1"
  local url="$2"
  local status

  status="$(curl -sS -o /tmp/sentinelpay-comm-response.txt -w "%{http_code}" "$url" || true)"
  if [[ "${status:0:1}" != "2" ]]; then
    echo "$name failed: GET $url returned $status" >&2
    cat /tmp/sentinelpay-comm-response.txt >&2 || true
    exit 1
  fi
  echo "$name OK"
}

require_cmd curl

echo "Checking direct service health communication..."
check_http "account-service" "$ACCOUNT_URL/health/account-service"
check_http "transaction-ingestor" "$INGESTOR_URL/health/transaction-ingestor"
check_http "feature-extractor" "$FEATURE_URL/health/feature-extractor"
check_http "blacklist-service" "$BLACKLIST_URL/health/blacklist-service"
check_http "rule-engine" "$RULE_URL/health/rule-engine"
check_http "fraud-orchestrator" "$ORCHESTRATOR_URL/health/fraud-orchestrator"
check_http "alert-service" "$ALERT_URL/health/alert-service"

if curl -fsS "$ML_URL/health/ml-service" >/dev/null 2>&1; then
  echo "ml-service OK"
else
  echo "ml-service skipped or unavailable at $ML_URL"
fi

echo "Checking orchestrator dependency endpoints when available..."
check_http "orchestrator kafka dependency view" "$ORCHESTRATOR_URL/system/kafka"
check_http "orchestrator redis dependency view" "$ORCHESTRATOR_URL/system/redis"
check_http "orchestrator service dependency view" "$ORCHESTRATOR_URL/system/services"

echo "Service communication verification passed."
