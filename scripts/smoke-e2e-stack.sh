#!/usr/bin/env bash
set -euo pipefail

if [[ -n "${BASH_SOURCE[0]:-}" ]]; then
  ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
else
  ROOT_DIR="$(pwd)"
fi

ACCOUNT_URL="${ACCOUNT_URL:-http://localhost:8087}"
INGESTOR_URL="${INGESTOR_URL:-http://localhost:8081}"
FEATURE_URL="${FEATURE_URL:-http://localhost:8082}"
RULE_URL="${RULE_URL:-http://localhost:8083}"
BLACKLIST_URL="${BLACKLIST_URL:-http://localhost:8084}"
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-http://localhost:8085}"
ALERT_URL="${ALERT_URL:-http://localhost:8086}"
ML_URL="${ML_URL:-http://localhost:18091}"

if command -v uuidgen >/dev/null 2>&1; then
  RUN_ID="$(uuidgen | tr '[:upper:]' '[:lower:]')"
else
  RUN_ID="$(date +%s)-$$"
fi

USER_ID="${SMOKE_USER_ID:-$((900000 + (RANDOM % 99999)))}"
RECEIVER_ID="${SMOKE_RECEIVER_ID:-$((800000 + (RANDOM % 99999)))}"
TRANSACTION_ID="${SMOKE_TRANSACTION_ID:-smoke-${RUN_ID}}"
NOW="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

http_get() {
  local url="$1"
  local expected="${2:-200}"
  local status

  status="$(curl -sS -o /tmp/sentinelpay-smoke-response.txt -w "%{http_code}" "$url" || true)"
  if [[ "$status" != "$expected" ]]; then
    echo "GET $url expected $expected got $status" >&2
    cat /tmp/sentinelpay-smoke-response.txt >&2 || true
    exit 1
  fi
}

http_post_json() {
  local url="$1"
  local payload="$2"
  local expected_prefix="${3:-2}"
  local status

  status="$(curl -sS -o /tmp/sentinelpay-smoke-response.txt -w "%{http_code}" \
    -H "Content-Type: application/json" \
    -X POST \
    --data "$payload" \
    "$url" || true)"

  if [[ "${status:0:1}" != "$expected_prefix" ]]; then
    echo "POST $url expected ${expected_prefix}xx got $status" >&2
    cat /tmp/sentinelpay-smoke-response.txt >&2 || true
    exit 1
  fi
}

wait_for_http() {
  local url="$1"
  local attempts="${2:-60}"
  local delay="${3:-2}"

  for _ in $(seq 1 "$attempts"); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$delay"
  done

  echo "Timed out waiting for $url" >&2
  exit 1
}

require_cmd curl

echo "Checking service health endpoints..."
wait_for_http "$ACCOUNT_URL/health/account-service"
wait_for_http "$INGESTOR_URL/health/transaction-ingestor"
wait_for_http "$FEATURE_URL/health/feature-extractor"
wait_for_http "$BLACKLIST_URL/health/blacklist-service"
wait_for_http "$RULE_URL/health/rule-engine"
wait_for_http "$ORCHESTRATOR_URL/health/fraud-orchestrator"
wait_for_http "$ALERT_URL/health/alert-service"

if curl -fsS "$ML_URL/health/ml-service" >/dev/null 2>&1; then
  echo "ML service health OK at $ML_URL"
else
  echo "ML service health skipped or unavailable at $ML_URL"
fi

echo "Creating sender and receiver accounts..."
http_post_json "$ACCOUNT_URL/api/v1/accounts" "{
  \"userId\": $USER_ID,
  \"accountCountry\": \"US\",
  \"homeCurrency\": \"USD\",
  \"createdAt\": \"$NOW\",
  \"kycLevel\": \"FULL\",
  \"status\": \"ACTIVE\",
  \"initialBalance\": 100000
}" "2"

http_post_json "$ACCOUNT_URL/api/v1/accounts" "{
  \"userId\": $RECEIVER_ID,
  \"accountCountry\": \"US\",
  \"homeCurrency\": \"USD\",
  \"createdAt\": \"$NOW\",
  \"kycLevel\": \"FULL\",
  \"status\": \"ACTIVE\",
  \"initialBalance\": 1000
}" "2"

http_get "$ACCOUNT_URL/api/v1/accounts/$USER_ID"
http_get "$ACCOUNT_URL/api/v1/accounts/$RECEIVER_ID"

echo "Submitting transaction $TRANSACTION_ID..."
http_post_json "$INGESTOR_URL/api/v1/transactions" "{
  \"transactionId\": \"$TRANSACTION_ID\",
  \"type\": \"P2P_TRANSFER\",
  \"senderUserId\": $USER_ID,
  \"receiverUserId\": $RECEIVER_ID,
  \"merchantId\": null,
  \"amount\": 125,
  \"currency\": \"USD\",
  \"deviceId\": \"smoke-device-$RUN_ID\",
  \"timestamp\": \"$NOW\"
}" "2"

echo "Waiting for final decision or alert..."
for _ in $(seq 1 60); do
  if curl -fsS "$ORCHESTRATOR_URL/decisions/$TRANSACTION_ID" >/tmp/sentinelpay-smoke-decision.txt 2>/dev/null; then
    echo "Decision available:"
    cat /tmp/sentinelpay-smoke-decision.txt
    echo
    exit 0
  fi

  if curl -fsS "$ALERT_URL/alerts/$TRANSACTION_ID" >/tmp/sentinelpay-smoke-alert.txt 2>/dev/null; then
    echo "Alert available:"
    cat /tmp/sentinelpay-smoke-alert.txt
    echo
    exit 0
  fi

  sleep 2
done

echo "Transaction was accepted, but no final decision or alert appeared before timeout." >&2
echo "Check service logs under $ROOT_DIR/logs/ and Kafka topics." >&2
exit 1
