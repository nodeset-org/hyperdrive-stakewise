set -a 
source ./.env
set +a
rm -rd ./data/db
docker compose run nimbus trustedNodeSync -d=data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false