#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_FILE="${ROOT_DIR}/logs/service-pids.txt"

if [[ -f "${PID_FILE}" ]]; then
  while read -r name pid; do
    if [[ -n "${pid}" ]] && kill -0 "${pid}" 2>/dev/null; then
      echo "Stopping ${name} (pid ${pid})..."
      kill "${pid}" || true
    fi
  done < "${PID_FILE}"
  rm -f "${PID_FILE}"
fi

echo "Stopping infrastructure..."
docker compose -f "${ROOT_DIR}/infrastructure/docker-compose.yml" down

echo "Done."
