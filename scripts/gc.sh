#!/usr/bin/env bash
set -euo pipefail

echo "=== Registry Maintenance GC ==="
echo "1. Set registry to readonly"

# This script assumes you temporarily update registry config and restart before running GC.
# Example manual steps:
#   docker compose stop registry
#   edit registry/config.yml -> storage.maintenance.readonly.enabled = true
#   docker compose up -d registry
#   sleep 2
#   docker exec -it registry-core registry garbage-collect /etc/distribution/config.yml
#   edit back to readonly.enabled = false
#   docker compose up -d registry

echo "This is a manual maintenance script. Please edit registry/config.yml readonly=true, restart, then run:"
echo "  docker exec -it registry-core registry garbage-collect /etc/distribution/config.yml"
echo "After GC, revert readonly to false and restart."
