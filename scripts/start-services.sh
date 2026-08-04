#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_FILE="${ROOT_DIR}/logs/service-pids.txt"

usage() {
  cat <<'EOF'
Usage: start-services.sh [-f COMPOSE_FILE] [--verify]

Options:
  -f COMPOSE_FILE  Path to docker-compose file (default: infrastructure/docker-compose.yml)
  --verify         Run smoke checks after startup while local services are alive
EOF
}

COMPOSE_FILE="${ROOT_DIR}/infrastructure/docker-compose.yml"
VERIFY=false

args=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --verify)
      VERIFY=true
      shift
      ;;
    *)
      args+=("$1")
      shift
      ;;
  esac
done
set -- "${args[@]}"

while getopts ":f:h" opt; do
  case "${opt}" in
    f)
      COMPOSE_FILE="${OPTARG}"
      ;;
    h)
      usage
      exit 0
      ;;
    \?)
      echo "Unknown option: -${OPTARG}" >&2
      usage >&2
      exit 2
      ;;
    :)
      echo "Missing argument for -${OPTARG}" >&2
      usage >&2
      exit 2
      ;;
  esac
done
shift $((OPTIND - 1))

services=(
  "api-gateway"
  "account-service"
  "transaction-ingestor"
  "feature-extractor"
  "rule-engine"
  "blacklist-service"
  "fraud-orchestrator"
  "alert-service"
)

declare -A health_paths=(
  ["api-gateway"]="/actuator/health"
  ["account-service"]="/health/account-service"
  ["transaction-ingestor"]="/health/transaction-ingestor"
  ["feature-extractor"]="/health/feature-extractor"
  ["rule-engine"]="/health/rule-engine"
  ["blacklist-service"]="/health/blacklist-service"
  ["fraud-orchestrator"]="/health/fraud-orchestrator"
  ["alert-service"]="/health/alert-service"
)

declare -A ports=(
  ["api-gateway"]="8080"
  ["account-service"]="8087"
  ["transaction-ingestor"]="8081"
  ["feature-extractor"]="8082"
  ["rule-engine"]="8083"
  ["blacklist-service"]="8084"
  ["fraud-orchestrator"]="8085"
  ["alert-service"]="8086"
)

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    return 1
  fi
}

missing_runtime=false
required_commands=(docker go)

for cmd in "${required_commands[@]}"; do
  if ! require_cmd "$cmd"; then
    missing_runtime=true
  fi
done

if [[ "$missing_runtime" == "true" ]]; then
  cat >&2 <<'EOF'
Cannot start local services.

This script runs the Go services on the host, so the host needs:
- Docker
- Go

Install the missing runtime or run the containerized Compose stack instead.
EOF
  exit 1
fi

wait_for_service() {
  local service="$1"
  local url="http://localhost:${ports[$service]}${health_paths[$service]}"

  for _ in $(seq 1 90); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi

    local pid
    pid="$(awk -v service="$service" '$1 == service { print $2 }' "$PID_FILE" | tail -1)"
    if [[ -n "$pid" ]] && ! kill -0 "$pid" 2>/dev/null; then
      echo "${service} exited before health check passed. Last log lines:" >&2
      tail -80 "${ROOT_DIR}/logs/${service}.log" >&2 || true
      exit 1
    fi

    sleep 2
  done

  echo "${service} did not become healthy at $url. Last log lines:" >&2
  tail -80 "${ROOT_DIR}/logs/${service}.log" >&2 || true
  exit 1
}

existing_pids=$(pgrep -f "(cmd/api-gateway|cmd/account-service|cmd/transaction-ingestor|cmd/feature-extractor|cmd/blacklist-service|cmd/rule-engine|cmd/fraud-orchestrator|cmd/alert-service)" || true)
if [[ -n "${existing_pids}" ]]; then
  echo "Existing service processes detected; stopping them first..."
  "${ROOT_DIR}/scripts/stop-services.sh"
fi

mkdir -p "${ROOT_DIR}/logs"
> "${PID_FILE}"

echo "Starting infrastructure (Kafka/Redis/Postgres/ML)..."
docker compose -f "${COMPOSE_FILE}" up -d --build

echo "Ensuring Kafka topics/partitions..."
docker compose -f "${COMPOSE_FILE}" run --rm kafka-init

export KAFKA_BOOTSTRAP_SERVERS="127.0.0.1:19092"
export REDIS_HOST="localhost"
export REDIS_PORT="16379"
export ML_SERVICE_URL="${ML_SERVICE_URL:-http://localhost:18091}"

echo "Ensuring fraud-ml-service is running..."
docker compose -f "${COMPOSE_FILE}" up -d --build fraud-ml-service

for service in "${services[@]}"; do
  echo "Starting ${service}..."
  if [[ "$service" == "api-gateway" ]]; then
    (
      cd "${ROOT_DIR}"
      nohup env \
        SERVER_PORT="${ports[$service]}" \
        ACCOUNT_SERVICE_URL="${ACCOUNT_SERVICE_URL:-http://localhost:8087}" \
        TRANSACTION_INGESTOR_URL="${TRANSACTION_INGESTOR_URL:-http://localhost:8081}" \
        FEATURE_EXTRACTOR_URL="${FEATURE_EXTRACTOR_URL:-http://localhost:8082}" \
        BLACKLIST_SERVICE_URL="${BLACKLIST_SERVICE_URL:-http://localhost:8084}" \
        RULE_ENGINE_URL="${RULE_ENGINE_URL:-http://localhost:8083}" \
        FRAUD_ORCHESTRATOR_URL="${FRAUD_ORCHESTRATOR_URL:-http://localhost:8085}" \
        ALERT_SERVICE_URL="${ALERT_SERVICE_URL:-http://localhost:8086}" \
        ML_SERVICE_URL="${ML_SERVICE_URL}" \
        go run ./cmd/api-gateway > "${ROOT_DIR}/logs/${service}.log" 2>&1 &
      echo "${service} $!" >> "${PID_FILE}"
    )
  elif [[ "$service" == "account-service" ]]; then
    (
      cd "${ROOT_DIR}"
      nohup env \
        SERVER_PORT="${ports[$service]}" \
        DB_HOST="${DB_HOST:-localhost}" \
        DB_PORT="${DB_PORT:-15432}" \
        DB_NAME="${DB_NAME:-sentinelpay}" \
        DB_USER="${DB_USER:-sentinel}" \
        DB_PASSWORD="${DB_PASSWORD:-sentinel123}" \
        go run ./cmd/account-service > "${ROOT_DIR}/logs/${service}.log" 2>&1 &
      echo "${service} $!" >> "${PID_FILE}"
    )
  elif [[ "$service" == "transaction-ingestor" ]]; then
    (
      cd "${ROOT_DIR}"
      nohup env \
        SERVER_PORT="${ports[$service]}" \
        DB_HOST="${DB_HOST:-localhost}" \
        DB_PORT="${DB_PORT:-15432}" \
        DB_NAME="${DB_NAME:-transaction_ingestor_db}" \
        DB_USER="${DB_USER:-sentinel}" \
        DB_PASSWORD="${DB_PASSWORD:-sentinel123}" \
        KAFKA_BOOTSTRAP_SERVERS="${KAFKA_BOOTSTRAP_SERVERS}" \
        go run ./cmd/transaction-ingestor > "${ROOT_DIR}/logs/${service}.log" 2>&1 &
      echo "${service} $!" >> "${PID_FILE}"
    )
  elif [[ "$service" == "feature-extractor" ]]; then
    (
      cd "${ROOT_DIR}"
      nohup env \
        SERVER_PORT="${ports[$service]}" \
        REDIS_HOST="${REDIS_HOST}" \
        REDIS_PORT="${REDIS_PORT}" \
        ACCOUNT_SERVICE_URL="${ACCOUNT_SERVICE_URL:-http://localhost:8087}" \
        KAFKA_BOOTSTRAP_SERVERS="${KAFKA_BOOTSTRAP_SERVERS}" \
        go run ./cmd/feature-extractor > "${ROOT_DIR}/logs/${service}.log" 2>&1 &
      echo "${service} $!" >> "${PID_FILE}"
    )
  elif [[ "$service" == "blacklist-service" ]]; then
    (
      cd "${ROOT_DIR}"
      nohup env \
        SERVER_PORT="${ports[$service]}" \
        DB_HOST="${DB_HOST:-localhost}" \
        DB_PORT="${DB_PORT:-15432}" \
        DB_NAME="${DB_NAME:-blacklist_db}" \
        DB_USER="${DB_USER:-sentinel}" \
        DB_PASSWORD="${DB_PASSWORD:-sentinel123}" \
        REDIS_HOST="${REDIS_HOST}" \
        REDIS_PORT="${REDIS_PORT}" \
        KAFKA_BOOTSTRAP_SERVERS="${KAFKA_BOOTSTRAP_SERVERS}" \
        go run ./cmd/blacklist-service > "${ROOT_DIR}/logs/${service}.log" 2>&1 &
      echo "${service} $!" >> "${PID_FILE}"
    )
  elif [[ "$service" == "rule-engine" ]]; then
    (
      cd "${ROOT_DIR}"
      nohup env \
        SERVER_PORT="${ports[$service]}" \
        DB_HOST="${DB_HOST:-localhost}" \
        DB_PORT="${DB_PORT:-15432}" \
        DB_NAME="${DB_NAME:-rule_engine_db}" \
        DB_USER="${DB_USER:-sentinel}" \
        DB_PASSWORD="${DB_PASSWORD:-sentinel123}" \
        KAFKA_BOOTSTRAP_SERVERS="${KAFKA_BOOTSTRAP_SERVERS}" \
        go run ./cmd/rule-engine > "${ROOT_DIR}/logs/${service}.log" 2>&1 &
      echo "${service} $!" >> "${PID_FILE}"
    )
  elif [[ "$service" == "fraud-orchestrator" ]]; then
    (
      cd "${ROOT_DIR}"
      nohup env \
        SERVER_PORT="${ports[$service]}" \
        DB_HOST="${DB_HOST:-localhost}" \
        DB_PORT="${DB_PORT:-15432}" \
        DB_NAME="${DB_NAME:-fraud_orchestrator_db}" \
        DB_USER="${DB_USER:-sentinel}" \
        DB_PASSWORD="${DB_PASSWORD:-sentinel123}" \
        REDIS_HOST="${REDIS_HOST}" \
        REDIS_PORT="${REDIS_PORT}" \
        KAFKA_BOOTSTRAP_SERVERS="${KAFKA_BOOTSTRAP_SERVERS}" \
        ML_SERVICE_URL="${ML_SERVICE_URL}" \
        SERVICES_INGESTOR_URL="${SERVICES_INGESTOR_URL:-http://localhost:8081/health/transaction-ingestor}" \
        SERVICES_EXTRACTOR_URL="${SERVICES_EXTRACTOR_URL:-http://localhost:8082/health/feature-extractor}" \
        SERVICES_BLACKLIST_URL="${SERVICES_BLACKLIST_URL:-http://localhost:8084/health/blacklist-service}" \
        SERVICES_RULE_URL="${SERVICES_RULE_URL:-http://localhost:8083/health/rule-engine}" \
        SERVICES_ML_URL="${SERVICES_ML_URL:-http://localhost:18091/health/ml-service}" \
        SERVICES_ORCHESTRATOR_URL="${SERVICES_ORCHESTRATOR_URL:-http://localhost:8085/health/fraud-orchestrator}" \
        go run ./cmd/fraud-orchestrator > "${ROOT_DIR}/logs/${service}.log" 2>&1 &
      echo "${service} $!" >> "${PID_FILE}"
    )
  elif [[ "$service" == "alert-service" ]]; then
    (
      cd "${ROOT_DIR}"
      nohup env \
        SERVER_PORT="${ports[$service]}" \
        DB_HOST="${DB_HOST:-localhost}" \
        DB_PORT="${DB_PORT:-15432}" \
        DB_NAME="${DB_NAME:-alert_db}" \
        DB_USER="${DB_USER:-sentinel}" \
        DB_PASSWORD="${DB_PASSWORD:-sentinel123}" \
        ACCOUNT_SERVICE_URL="${ACCOUNT_SERVICE_URL:-http://localhost:8087}" \
        KAFKA_BOOTSTRAP_SERVERS="${KAFKA_BOOTSTRAP_SERVERS}" \
        go run ./cmd/alert-service > "${ROOT_DIR}/logs/${service}.log" 2>&1 &
      echo "${service} $!" >> "${PID_FILE}"
    )
  fi
  wait_for_service "$service"
done

echo "All services started. Logs in ${ROOT_DIR}/logs/"
echo "Use scripts/stop-services.sh to stop them."

if [[ "$VERIFY" == "true" ]]; then
  echo "Running account-service smoke..."
  "${ROOT_DIR}/scripts/smoke-account-service.sh"

  echo "Running service communication smoke..."
  "${ROOT_DIR}/scripts/smoke-service-communication.sh"

  echo "Running end-to-end stack smoke..."
  "${ROOT_DIR}/scripts/smoke-compose-network.sh"

  echo "Running Kafka pipeline smoke..."
  "${ROOT_DIR}/scripts/smoke-kafka-pipeline.sh"

  echo "Startup verification passed."
fi
