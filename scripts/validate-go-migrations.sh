#!/usr/bin/env bash

set -euo pipefail

compose_file="${COMPOSE_FILE:-docker-compose.yml}"
compose=(docker compose -f "$compose_file")
psql_cmd=("${compose[@]}" exec -T postgres psql -U sentinel -d sentinelpay -v ON_ERROR_STOP=1 -At)

run_sql() {
  local sql="$1"
  "${psql_cmd[@]}" -c "$sql"
}

expect_equals() {
  local label="$1"
  local sql="$2"
  local expected="$3"
  local actual
  actual="$(run_sql "$sql")"
  if [[ "$actual" != "$expected" ]]; then
    echo "validation failed: $label expected=$expected actual=$actual" >&2
    exit 1
  fi
}

expect_gte() {
  local label="$1"
  local sql="$2"
  local minimum="$3"
  local actual
  actual="$(run_sql "$sql")"
  if (( actual < minimum )); then
    echo "validation failed: $label expected>=$minimum actual=$actual" >&2
    exit 1
  fi
}

expect_equals "accounts table" "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='accounts';" "1"
expect_equals "transaction_records table" "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='transaction_records';" "1"
expect_equals "outbox_event table" "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='outbox_event';" "1"
expect_equals "fraud_rules table" "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='fraud_rules';" "1"
expect_equals "fraud_decisions table" "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='fraud_decisions';" "1"
expect_equals "fraud_alerts table" "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='fraud_alerts';" "1"
expect_equals "applied_transfer table" "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='applied_transfer';" "1"

expect_equals "accounts version column" "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='accounts' AND column_name='version';" "1"
expect_equals "outbox_event last_error column" "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='outbox_event' AND column_name='last_error';" "1"
expect_equals "fraud_decisions reviewed column" "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='fraud_decisions' AND column_name='reviewed';" "1"
expect_equals "fraud_alerts payload_json column" "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='fraud_alerts' AND column_name='payload_json';" "1"
expect_equals "applied_transfer payload_json column" "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='applied_transfer' AND column_name='payload_json';" "1"

expect_equals "transaction_records received_at index" "SELECT COUNT(*) FROM pg_indexes WHERE schemaname='public' AND tablename='transaction_records' AND indexname='idx_transaction_records_received_at';" "1"
expect_equals "transaction_records sender_time_amount_currency index" "SELECT COUNT(*) FROM pg_indexes WHERE schemaname='public' AND tablename='transaction_records' AND indexname='idx_transaction_records_sender_time_amount_currency';" "1"
expect_equals "outbox_event due index" "SELECT COUNT(*) FROM pg_indexes WHERE schemaname='public' AND tablename='outbox_event' AND indexname='idx_outbox_event_due';" "1"
expect_equals "fraud_rules unique name index" "SELECT COUNT(*) FROM pg_indexes WHERE schemaname='public' AND tablename='fraud_rules' AND indexname='fraud_rules_name_key';" "1"

expect_gte "seeded fraud rules" "SELECT COUNT(*) FROM fraud_rules;" "13"
expect_equals "absolute-high-amount rule" "SELECT COUNT(*) FROM fraud_rules WHERE name='absolute-high-amount' AND type='ABSOLUTE_HIGH_AMOUNT' AND enabled=true AND active=true;" "1"
expect_equals "cross-border-high-amount rule" "SELECT COUNT(*) FROM fraud_rules WHERE name='cross-border-high-amount' AND type='CROSS_BORDER_HIGH_AMOUNT' AND enabled=true AND active=true;" "1"

echo "Go migration validation passed."
