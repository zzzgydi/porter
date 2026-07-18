#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Ensure auth certs exist (never overwrites existing certs)
bash "${SCRIPT_DIR}/generate-auth-cert.sh"

# Ensure .env exists
if [ ! -f "${SCRIPT_DIR}/../.env" ]; then
    cp "${SCRIPT_DIR}/../.env.example" "${SCRIPT_DIR}/../.env"
    echo "Created .env from .env.example. Please review values before first run."
fi

cd "${SCRIPT_DIR}/.."

docker compose -f docker-compose.dev.yml up -d --build

echo ""
echo "Dev stack starting. Tail logs:"
echo "  docker compose -f docker-compose.dev.yml logs -f"
