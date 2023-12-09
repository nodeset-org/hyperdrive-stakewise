#!/usr/bin/env bash 

# always pull latest stakewise operator image in case it's been updated
echo "Pulling latest StakeWise operator binary..."
docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master

# todo: only run checkpoint sync if no db exists
# nimbus checks this already, but we shouldn't even do it if the db exists
if [ "$NETWORK" != "mainnet" ]; then
    echo "Performing checkpoint sync..."
    docker compose run nimbus trustedNodeSync -d=/home/user/data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false
fi


echo Starting node...
docker compose up -d
echo "{::} Node started successfully!"
echo
echo "After continuing, logs will be displayed. Please inspect them to verify everything is working as expected. Once you're satisfied, you may exit safely with ^C and the containers will continue running."
read -rsn1 -p "Press any key to continue..."; echo
docker compose logs -f