#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_FILE="${ROOT_DIR}/logs/service-pids.txt"

usage() {
  cat <<'EOF'
Usage: start-services.sh [-f COMPOSE_FILE]

Options:
  -f COMPOSE_FILE  Path to docker-compose file (default: infrastructure/docker-compose.yml)
EOF
}

COMPOSE_FILE="${ROOT_DIR}/infrastructure/docker-compose.yml"
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
  "account-service"
  "transaction-ingestor"
  "feature-extractor"
  "rule-engine"
  "blacklist-service"
  "fraud-orchestrator"
  "alert-service"
)

existing_pids=$(pgrep -f "microservices/(account-service|transaction-ingestor|feature-extractor|rule-engine|blacklist-service|fraud-orchestrator|alert-service).*spring-boot:run" || true)
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
export ML_SERVICE_URL="http://localhost:8091"
export JAVA_TOOL_OPTIONS="${JAVA_TOOL_OPTIONS:-} -Djava.net.preferIPv4Stack=true"

echo "Ensuring fraud-ml-service is running..."
docker compose -f "${COMPOSE_FILE}" up -d --build fraud-ml-service

for service in "${services[@]}"; do
  echo "Starting ${service}..."
  (
    cd "${ROOT_DIR}/microservices/${service}"
    # Force a clean compile to avoid stale/invalid class artifacts causing runtime "Unresolved compilation problems".
    ./mvnw clean compile
    ./mvnw spring-boot:run > "${ROOT_DIR}/logs/${service}.log" 2>&1 &
    echo "${service} $!" >> "${PID_FILE}"
  )
done

echo "All services started. Logs in ${ROOT_DIR}/logs/"
echo "Use scripts/stop-services.sh to stop them."
