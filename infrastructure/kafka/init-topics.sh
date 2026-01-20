#!/usr/bin/env bash
set -euo pipefail

BOOTSTRAP="${KAFKA_BOOTSTRAP_SERVERS:-kafka:9092}"
PARTITIONS="${KAFKA_PARTITIONS:-16}"
REPLICATION="${KAFKA_REPLICATION_FACTOR:-1}"

TOPICS=(
  "transactions.raw"
  "transactions.enriched"
  "fraud.blacklist"
  "fraud.rules"
  "fraud.ml"
  "fraud.final"
)

for topic in "${TOPICS[@]}"; do
  if kafka-topics --bootstrap-server "$BOOTSTRAP" --list | grep -q "^${topic}$"; then
    kafka-topics --bootstrap-server "$BOOTSTRAP" --alter --topic "$topic" --partitions "$PARTITIONS" >/dev/null
  else
    kafka-topics --bootstrap-server "$BOOTSTRAP" --create --if-not-exists \
      --topic "$topic" --partitions "$PARTITIONS" --replication-factor "$REPLICATION" >/dev/null
  fi
done

echo "Kafka topics ensured partitions=$PARTITIONS"
