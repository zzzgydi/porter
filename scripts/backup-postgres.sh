#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Read POSTGRES_USER / POSTGRES_DB from .env when available.
POSTGRES_USER="${POSTGRES_USER:-registry}"
POSTGRES_DB="${POSTGRES_DB:-registry_console}"
if [ -f "${SCRIPT_DIR}/../.env" ]; then
  set -a
  # shellcheck disable=SC1091
  . "${SCRIPT_DIR}/../.env"
  set +a
fi

BACKUP_DIR="${BACKUP_DIR:-./backups}"
mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date +%F_%H%M%S)
OUTPUT="${BACKUP_DIR}/${POSTGRES_DB}-${TIMESTAMP}.sql"

docker exec registry-postgres pg_dump \
  -U "$POSTGRES_USER" \
  "$POSTGRES_DB" > "$OUTPUT"

echo "Backed up to $OUTPUT"
