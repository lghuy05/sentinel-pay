#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Running Go replacement stack smoke..."
"$ROOT_DIR/scripts/smoke-compose-network.sh"
