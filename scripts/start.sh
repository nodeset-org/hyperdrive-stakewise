#!/usr/bin/env bash 

echo "Starting node..."

# always pull latest stakewise operator image in case it's been updated
echo "Pulling latest StakeWise operator binary..."
docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master

# start containers
docker compose -f "$DATA_DIR/compose.yaml" up -d