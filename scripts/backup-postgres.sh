#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-./backups}"
mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date +%F_%H%M%S)
OUTPUT="${BACKUP_DIR}/registry_console-${TIMESTAMP}.sql"

docker exec registry-postgres pg_dump \
  -U registry \
  registry_console > "$OUTPUT"

echo "Backed up to $OUTPUT"
