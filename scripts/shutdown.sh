#!/usr/bin/env bash 

docker compose down --remove-orphans "$SCRIPT_DIR/compose.yaml"