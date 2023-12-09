#!/usr/bin/env bash 

echo "Shutting down containers..."

docker compose -f "$SCRIPT_DIR/compose.yaml" down --remove-orphans