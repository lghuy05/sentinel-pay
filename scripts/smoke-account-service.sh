#!/usr/bin/env bash
set -euo pipefail

ACCOUNT_URL="${ACCOUNT_URL:-http://localhost:8087}"

if command -v uuidgen >/dev/null 2>&1; then
  RUN_ID="$(uuidgen | tr '[:upper:]' '[:lower:]')"
else
  RUN_ID="$(date +%s)-$$"
fi

USER_ID="${SMOKE_USER_ID:-$((700000 + (RANDOM % 99999)))}"
NOW="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
RESPONSE_FILE="/tmp/sentinelpay-account-smoke-response.txt"

http_request() {
  local method="$1"
  local url="$2"
  local payload="${3:-}"
  local expected="${4:-2}"
  local status

  if [[ -n "$payload" ]]; then
    status="$(curl -sS -o "$RESPONSE_FILE" -w "%{http_code}" \
      -H "Content-Type: application/json" \
      -X "$method" \
      --data "$payload" \
      "$url" || true)"
  else
    status="$(curl -sS -o "$RESPONSE_FILE" -w "%{http_code}" -X "$method" "$url" || true)"
  fi

  if [[ "${status:0:1}" != "$expected" ]]; then
    echo "$method $url expected ${expected}xx got $status" >&2
    cat "$RESPONSE_FILE" >&2 || true
    exit 1
  fi
}

require_field() {
  local field="$1"
  if ! grep -q "$field" "$RESPONSE_FILE"; then
    echo "Expected response to contain $field" >&2
    cat "$RESPONSE_FILE" >&2 || true
    exit 1
  fi
}

echo "Running account-service smoke for user $USER_ID..."

http_request GET "$ACCOUNT_URL/health/account-service"
require_field '"status":"UP"'

http_request POST "$ACCOUNT_URL/api/v1/accounts" "{
  \"userId\": $USER_ID,
  \"accountCountry\": \"US\",
  \"homeCurrency\": \"USD\",
  \"createdAt\": \"$NOW\",
  \"kycLevel\": \"FULL\",
  \"status\": \"ACTIVE\",
  \"initialBalance\": 1000
}"
require_field '"balanceMinor":1000'

http_request GET "$ACCOUNT_URL/api/v1/accounts/$USER_ID"
require_field "\"userId\":$USER_ID"

http_request PATCH "$ACCOUNT_URL/api/v1/accounts/$USER_ID" '{"status":"LOCKED"}'
require_field '"status":"LOCKED"'

http_request POST "$ACCOUNT_URL/api/v1/accounts/$USER_ID/topup" '{"amount":250,"currency":"USD"}'
require_field '"balanceMinor":1250'

http_request POST "$ACCOUNT_URL/api/v1/accounts/$USER_ID/debit" '{"amount":100,"currency":"USD"}'
require_field '"balanceMinor":1150'

http_request GET "$ACCOUNT_URL/api/v1/accounts?limit=10&offset=0"
require_field "\"userId\":$USER_ID"

http_request DELETE "$ACCOUNT_URL/api/v1/accounts/$USER_ID" "" 2

http_request GET "$ACCOUNT_URL/api/v1/accounts/$USER_ID" "" 4

echo "Account-service smoke passed. run=$RUN_ID"
