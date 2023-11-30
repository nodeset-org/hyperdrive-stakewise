# ensure root access
if [ "$(id -u)" -ne 0 ];
  then echo "Please run this script as root (or with sudo)"
  exit
fi

echo $1

# check parameters make sense
if [ "$1" != "holesky" ] && [ "$1" != "gravita" ];
  then echo "Usage: initialize.sh VAULT
Supported vaults: holesky, gravita"
  exit
fi

# set env based on vault provided
set -a 
source ./${1}.env
set +a

# clear old data (if any)
rm -rd ./nimbus-data
rm -rd ./tmp
rm -rd ./stakewise-data
rm -rd ./geth-data

mkdir ./nimbus-data
chown $USER ./nimbus-data

# initialize geth
docker compose run -d geth --authrpc.jwtsecret /tmp/jwtsecret
sleep 3 # wait a few seconds for geth to boot up and create the jwtsecret
echo "looking for jwtsecret"
chown $USER ./tmp/jwtsecret

if [ "$NETWORK" != "mainnet" ];
    then docker compose run nimbus trustedNodeSync -d=data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false
fi

# pull latest stakewise operator image
docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master

docker compose up

