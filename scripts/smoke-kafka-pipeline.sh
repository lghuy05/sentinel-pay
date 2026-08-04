#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/docker-compose.yml}"
KAFKA_SERVICE="${KAFKA_SERVICE:-kafka}"
BOOTSTRAP_IN_CONTAINER="${KAFKA_BOOTSTRAP_IN_CONTAINER:-localhost:9092}"
TIMEOUT_MS="${KAFKA_CONSUME_TIMEOUT_MS:-5000}"
KAFKA_PARTITIONS="${KAFKA_PARTITIONS:-16}"
KAFKA_REPLICATION_FACTOR="${KAFKA_REPLICATION_FACTOR:-1}"

TOPICS=(
  "transactions.raw"
  "transactions.enriched"
  "fraud.blacklist"
  "fraud.rules"
  "fraud.ml"
  "fraud.final"
)

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

kafka_exec() {
  docker compose -f "$COMPOSE_FILE" exec -T "$KAFKA_SERVICE" "$@"
}

compose_has_service() {
  local service="$1"
  docker compose -f "$COMPOSE_FILE" config --services | grep -qx "$service"
}

ensure_topic_direct() {
  local topic="$1"
  local current_partitions

  if kafka_exec kafka-topics --bootstrap-server "$BOOTSTRAP_IN_CONTAINER" --describe --topic "$topic" >/tmp/sentinelpay-topic-describe.txt 2>/dev/null; then
    current_partitions="$(grep -m1 -Eo 'PartitionCount: [0-9]+' /tmp/sentinelpay-topic-describe.txt | awk '{print $2}')"
    current_partitions="${current_partitions:-0}"

    if (( current_partitions < KAFKA_PARTITIONS )); then
      kafka_exec kafka-topics \
        --bootstrap-server "$BOOTSTRAP_IN_CONTAINER" \
        --alter \
        --topic "$topic" \
        --partitions "$KAFKA_PARTITIONS"
    fi
    return 0
  fi

  kafka_exec kafka-topics \
    --bootstrap-server "$BOOTSTRAP_IN_CONTAINER" \
    --create \
    --if-not-exists \
    --topic "$topic" \
    --partitions "$KAFKA_PARTITIONS" \
    --replication-factor "$KAFKA_REPLICATION_FACTOR"
}

require_cmd docker

echo "Checking Kafka service..."
docker compose -f "$COMPOSE_FILE" ps "$KAFKA_SERVICE"

if compose_has_service kafka-init; then
  echo "Ensuring Kafka topics through kafka-init..."
  docker compose -f "$COMPOSE_FILE" run --rm kafka-init
else
  echo "Ensuring Kafka topics directly in $KAFKA_SERVICE..."
  for topic in "${TOPICS[@]}"; do
    ensure_topic_direct "$topic"
  done
fi

echo "Verifying required topics..."
topic_list="$(kafka_exec kafka-topics --bootstrap-server "$BOOTSTRAP_IN_CONTAINER" --list)"
for topic in "${TOPICS[@]}"; do
  if ! grep -qx "$topic" <<<"$topic_list"; then
    echo "Missing Kafka topic: $topic" >&2
    exit 1
  fi
  kafka_exec kafka-topics --bootstrap-server "$BOOTSTRAP_IN_CONTAINER" --describe --topic "$topic"
done

if [[ "${KAFKA_PEEK:-false}" == "true" ]]; then
  echo "Peeking latest messages from each topic. Empty topics are allowed."
  for topic in "${TOPICS[@]}"; do
    echo "Topic: $topic"
    kafka_exec kafka-console-consumer \
      --bootstrap-server "$BOOTSTRAP_IN_CONTAINER" \
      --topic "$topic" \
      --from-beginning \
      --timeout-ms "$TIMEOUT_MS" \
      --max-messages 1 || true
  done
fi

echo "Kafka topic verification passed."
