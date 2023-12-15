#!/usr/bin/env bash 

echo "Starting node..."

# pull latest container images
echo "Updating..."
docker compose -f "$DATA_DIR/compose.yaml" pull

# start containers
docker compose -f "$DATA_DIR/compose.yaml" up -d

echo
echo "{::} Node Started! {::}"
echo