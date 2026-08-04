#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_FILE="${ROOT_DIR}/logs/service-pids.txt"

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

if [[ -f "${PID_FILE}" ]]; then
  while read -r name pid; do
    if [[ -n "${pid}" ]] && kill -0 "${pid}" 2>/dev/null; then
      echo "Stopping ${name} (pid ${pid})..."
      kill "${pid}" || true
    fi
  done < "${PID_FILE}"
  rm -f "${PID_FILE}"
fi

for service in "${services[@]}"; do
  if [[ "$service" == "api-gateway" ]]; then
    pids=$(pgrep -f "cmd/api-gateway" || true)
  elif [[ "$service" == "account-service" ]]; then
    pids=$(pgrep -f "cmd/account-service" || true)
  elif [[ "$service" == "transaction-ingestor" ]]; then
    pids=$(pgrep -f "cmd/transaction-ingestor" || true)
  elif [[ "$service" == "feature-extractor" ]]; then
    pids=$(pgrep -f "cmd/feature-extractor" || true)
  elif [[ "$service" == "blacklist-service" ]]; then
    pids=$(pgrep -f "cmd/blacklist-service" || true)
  elif [[ "$service" == "rule-engine" ]]; then
    pids=$(pgrep -f "cmd/rule-engine" || true)
  elif [[ "$service" == "fraud-orchestrator" ]]; then
    pids=$(pgrep -f "cmd/fraud-orchestrator" || true)
  elif [[ "$service" == "alert-service" ]]; then
    pids=$(pgrep -f "cmd/alert-service" || true)
  fi
  if [[ -n "${pids}" ]]; then
    echo "Stopping ${service} (pids ${pids})..."
    for pid in ${pids}; do
      kill "${pid}" || true
    done
  fi
done

echo "Stopping infrastructure (including fraud-ml-service)..."
docker compose -f "${ROOT_DIR}/infrastructure/docker-compose.yml" down

echo "Done."
