clean()
{
    echo "Cleaning up previous configuration..."
    
    docker compose down

    # clear old data (if any)
    rm -rd ./nimbus-data
    rm -rd ./tmp
    rm -rd ./stakewise-data
    rm -rd ./geth-data

    mkdir ./nimbus-data
    mkdir ./stakewise-data
    chown $(logname) ./nimbus-data
    chown $(logname) ./stakewise-data
}

generate_jwtsecret()
{
    echo "Generating jwtsecret..."
    # initialize geth
    docker compose run -d geth --authrpc.jwtsecret /tmp/jwtsecret
    sleep 3 # wait a few seconds for geth to boot up and create the jwtsecret
    chown $(logname) ./tmp/jwtsecret
}

checkpoint()
{
    echo "Performing checkpoint sync..."
    if [ "$NETWORK" != "mainnet" ]; then
        docker compose run nimbus trustedNodeSync -d=/home/user/data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false
    fi
}

display_funding_message()
{
    echo "Please note that you must have enough Ether in this node wallet to register validators."
    echo "Each validator takes approximately 0.01 ETH to create. We recommend depositing AT LEAST 0.1 ETH."
    echo "To continue, first send some ETH to this wallet, then type 'wallet is funded' to continue."
    read answer

    if [ "$answer" != "wallet is funded" ]; then 
        display_funding_message
    fi
}

setup_stakewise()
{
    echo "Pulling latest StakeWise operator binary..."
    # pull latest stakewise operator image
    docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master

    docker compose run stakewise src/main.py init --network=$NETWORK --vault=$VAULT --language=english
    docker compose run stakewise src/main.py create-keys --vault=$VAULT --language=english --count=$NUMKEYS
    docker compose run stakewise src/main.py create-wallet --vault=$VAULT --language=english --count=$NUMKEYS
    
    display_funding_message
}

## -- Start of Script -- ##

# ensure root access
if [ "$(id -u)" -ne 0 ]; then
  echo "Please run this script as root (or with sudo)"
  exit
fi

USAGEMSG="Usage: init-node.sh VAULT [--preserve-data|-p]
Supported vaults: holesky, gravita"

# check parameters make sense
if [ "$1" != "holesky" ] && [ "$1" != "gravita" ]; then
  echo $USAGEMSG
  exit
fi

# set env based on vault provided
set -a 
source ./${1}.env
set +a

# check for preserve-data flag
if [ "$2" != "" ] && [ "$2" != "-r" ] && [ "$2" != "--reset" ]; then
    echo $USAGEMSG
    exit
fi

if [ "$2" = "-r" ] || [ "$2" == "--reset" ]; then
     if [ "$1" != "holesky" ]; then
        
        # todo: check if there are any active validators before giving this warning
        echo "DANGER: You are attempting to reset your configuration for a mainnet vault!"
        echo "This will require you to resync the chain completely before you can begin validating again, which may take several days."
        echo "Remember, if you're offline for too long, you may be kicked out of NodeSet!"
        echo "Are you sure you want to continue? You must type YES to continue."
        read answer

        if [ "$answer" != "YES" ]; then 
            echo Cancelled
            exit
        fi

        echo "THIS IS YOUR FINAL WARNING! Are you absolutely sure that you want to delete all of your data for this mainnet configuration?"
        echo "You must type YES to continue."
        read answer2

        if [ "$answer2" != "YES" ]; then 
            echo Cancelled
            exit
        fi
    fi
    clean
fi

generate_jwtsecret
#checkpoint
setup_stakewise

echo Starting node...
docker compose up -d
echo "{::} Node started successfully!"
echo
echo "After continuing, logs will be displayed. Please inspect them to verify everything is working as expected. Once you're satisfied, you may exit safely and the containers will continue running."
read -rsn1 -p "Press any key to continue..."; echo
docker compose logs -f