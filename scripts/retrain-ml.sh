#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/infrastructure/docker-compose.yml"

echo "Running ML retrain inside fraud-ml-service container..."
docker compose -f "${COMPOSE_FILE}" exec fraud-ml-service \
  python /app/training/train.py

echo "Retrain complete. Consider calling /ml/reload if needed."
