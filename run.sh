#!/usr/bin/env bash 

# todo: only run checkpoint sync if no db exists
# nimbus checks this already, but we shouldn't even do it if the db exists
if [ "$NETWORK" != "mainnet" ]; then
    echo "Performing checkpoint sync..."
    docker compose run nimbus trustedNodeSync -d=/home/user/data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false
fi