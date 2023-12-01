# ensure root access
if [ "$(id -u)" -ne 0 ]
  then echo "Please run this script as root (or with sudo)"
  exit
fi

USAGEMSG="Usage: initialize.sh VAULT [--preserve-data|-p]
Supported vaults: holesky, gravita"

# check parameters make sense
if [ "$1" != "holesky" ] && [ "$1" != "gravita" ]
  then echo $USAGEMSG
  exit
fi

# set env based on vault provided
set -a 
source ./${1}.env
set +a

# check for preserve-data flag
if [ "$2" != "" ] && [ "$2" != "-r" ] && [ "$2" != "--reset" ]
    then echo $USAGEMSG
    exit
fi

if [ "$2" = "-r" ] || [ "$2" == "--reset" ]
    then clean
fi

generate_jwtsecret
checkpoint
setup_stakewise

echo Starting node...
docker compose up -d
echo {::} Node started successfully! After continuing, logs will be displayed. Please inspect them to ensure correct operation.
read -rsn1 -p "Press any key to continue..."; echo
sleep 10
docker compose logs -f

clean()
{
    echo Cleaning up previous configuration...
    # clear old data (if any)
    rm -rd ./nimbus-data
    rm -rd ./tmp
    rm -rd ./stakewise-data
    rm -rd ./geth-data

    mkdir ./nimbus-data
    chown $(logname) ./nimbus-data
}

generate_jwtsecret()
{
    echo Generating jwtsecret...
    # initialize geth
    docker compose run -d geth --authrpc.jwtsecret /tmp/jwtsecret
    sleep 3 # wait a few seconds for geth to boot up and create the jwtsecret
    chown $(logname) ./tmp/jwtsecret
}

checkpoint()
{
    echo Performing checkpoint sync...
    if [ "$NETWORK" != "mainnet" ]
        then docker compose run nimbus trustedNodeSync -d=/home/user/data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false
    fi
}

setup_stakewise()
{
    echo Pulling latest StakeWise operator binary...
    # pull latest stakewise operator image
    docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master
}

