#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_FILE="${ROOT_DIR}/logs/service-pids.txt"

services=(
  "transaction-ingestor"
  "feature-extractor"
  "rule-engine"
  "blacklist-service"
  "fraud-orchestrator"
  "alert-service"
)

mkdir -p "${ROOT_DIR}/logs"
> "${PID_FILE}"

echo "Starting infrastructure (Kafka/Redis/Postgres/ML)..."
docker compose -f "${ROOT_DIR}/infrastructure/docker-compose.yml" up -d --build

export KAFKA_BOOTSTRAP_SERVERS="localhost:19092"
export REDIS_HOST="localhost"
export REDIS_PORT="16379"

echo "Ensuring fraud-ml-service is running..."
docker compose -f "${ROOT_DIR}/infrastructure/docker-compose.yml" up -d --build fraud-ml-service

for service in "${services[@]}"; do
  echo "Starting ${service}..."
  (
    cd "${ROOT_DIR}/microservices/${service}"
    ./mvnw spring-boot:run > "${ROOT_DIR}/logs/${service}.log" 2>&1 &
    echo "${service} $!" >> "${PID_FILE}"
  )
done

echo "All services started. Logs in ${ROOT_DIR}/logs/"
echo "Use scripts/stop-services.sh to stop them."
