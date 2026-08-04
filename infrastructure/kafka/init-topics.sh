#!/usr/bin/env bash
set -euo pipefail

BOOTSTRAP="${KAFKA_BOOTSTRAP_SERVERS:-kafka:9092}"
PARTITIONS="${KAFKA_PARTITIONS:-16}"
REPLICATION="${KAFKA_REPLICATION_FACTOR:-1}"
KAFKA_CLI_TIMEOUT="${KAFKA_CLI_TIMEOUT_SECONDS:-30}"

TOPICS=(
  "transactions.raw"
  "transactions.enriched"
  "fraud.blacklist"
  "fraud.rules"
  "fraud.ml"
  "fraud.final"
  "transactions.raw.DLT"
  "transactions.enriched.DLT"
  "fraud.blacklist.DLT"
  "fraud.rules.DLT"
  "fraud.ml.DLT"
  "fraud.final.DLT"
)

run_kafka_topics() {
  timeout "$KAFKA_CLI_TIMEOUT" kafka-topics --bootstrap-server "$BOOTSTRAP" "$@"
}

for _ in $(seq 1 60); do
  if run_kafka_topics --list >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

run_kafka_topics --list >/dev/null

for topic in "${TOPICS[@]}"; do
  run_kafka_topics --create --if-not-exists \
    --topic "$topic" --partitions "$PARTITIONS" --replication-factor "$REPLICATION" >/dev/null

  current_partitions="$(
    run_kafka_topics --describe --topic "$topic" \
      | awk -F'PartitionCount: ' 'NF > 1 { split($2, fields, /[[:space:]]+/); print fields[1]; exit }'
  )"

  if [[ -n "$current_partitions" && "$current_partitions" -lt "$PARTITIONS" ]]; then
    run_kafka_topics --alter --topic "$topic" --partitions "$PARTITIONS" >/dev/null
  elif [[ -n "$current_partitions" && "$current_partitions" -gt "$PARTITIONS" ]]; then
    echo "Topic $topic already has $current_partitions partitions, requested $PARTITIONS; leaving unchanged" >&2
  fi
done

echo "Kafka topics ensured partitions=$PARTITIONS"
