#!/usr/bin/env bash
set -euo pipefail

wait_for_host_port() {
  local name="$1"
  local host="$2"
  local port="$3"
  local timeout="$4"

  echo "Waiting for ${name} at ${host}:${port} for up to ${timeout}s..."

  local deadline=$((SECONDS + timeout))
  while ! bash -c ">/dev/tcp/${host}/${port}" 2>/dev/null; do
    if (( SECONDS >= deadline )); then
      echo "${name} did not become reachable at ${host}:${port} within ${timeout}s" >&2
      exit 1
    fi
    sleep 2
  done

  echo "${name} is reachable."
}

if [[ "${DB_WAIT_ENABLED:-false}" == "true" ]]; then
  wait_for_host_port \
    "PostgreSQL" \
    "${DB_HOST:-localhost}" \
    "${DB_PORT:-15432}" \
    "${DB_WAIT_TIMEOUT:-60}"
fi

if [[ "${REDIS_WAIT_ENABLED:-false}" == "true" ]]; then
  wait_for_host_port \
    "Redis" \
    "${REDIS_HOST:-localhost}" \
    "${REDIS_PORT:-16379}" \
    "${REDIS_WAIT_TIMEOUT:-60}"
fi

if [[ "${KAFKA_WAIT_ENABLED:-false}" == "true" ]]; then
  kafka_bootstrap="${KAFKA_BOOTSTRAP_SERVERS:-localhost:19092}"
  kafka_host="${kafka_bootstrap%%,*}"
  kafka_host="${kafka_host%%:*}"
  kafka_port="${kafka_bootstrap%%,*}"
  kafka_port="${kafka_port##*:}"

  wait_for_host_port \
    "Kafka" \
    "${kafka_host}" \
    "${kafka_port}" \
    "${KAFKA_WAIT_TIMEOUT:-60}"
fi

echo "Starting ${SERVICE_NAME:-java-service}."
exec java ${JAVA_OPTS:-} -jar /app/app.jar
