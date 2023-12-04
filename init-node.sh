#!/bin/bash

clean()
{
    echo "Cleaning up previous configuration..."
    
    docker compose down --remove-orphans

    # clear old data (if any)
    rm -rd ./nimbus-data
    rm -rd ./tmp
    rm -rd ./stakewise-data
    rm -rd ./geth-data

    mkdir ./nimbus-data
    mkdir ./stakewise-data
    chown $(logname) ./nimbus-data
    chmod 700 ./nimbus-data
    # v3-operator user is "nobody" for safety since keys are stored there
    # you will need to use root to access it
    chown nobody ./stakewise-data 
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
    docker compose run nimbus trustedNodeSync -d=/home/user/data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false
}

display_funding_message()
{
    echo "Please send some ETH to the wallet address above (on the $NETWORK network), then type 'wallet is funded' to continue."
    read answer

    if [ "$answer" != "wallet is funded" ]; then 
        display_funding_message
    fi
}

setup_stakewise()
{
    if [ "$mnemonic" != "" ]; then
        echo "Recreating StakeWise configuration using existing mnemonic..."

        docker compose run stakewise src/main.py recover --network="$NETWORK" --vault="$VAULT" --execution-endpoints="http://$ECNAME:$ECAPIPORT" --consensus-endpoints="http://$CCNAME:$CCAPIPORT" --mnemonic="$mnemonic"
        docker compose run stakewise src/main.py create-wallet --vault="$VAULT" --mnemonic="$mnemonic"
    else
        echo "Initializing new StakeWise configuration..."
        docker compose run stakewise src/main.py init --network="$NETWORK" --vault="$VAULT" --language=english
        docker compose run stakewise src/main.py create-keys --vault="$VAULT" --count="$NUMKEYS"
        docker compose run stakewise src/main.py create-wallet --vault="$VAULT"
    fi
    
    echo "Please note that you must have enough Ether in this node wallet to register validators."
    printf "Each validator takes approximately 0.01 ETH to create. We recommend depositing AT LEAST 0.1 ETH.\nYou can withdraw this ETH at any time. For more information, see: http://nodeset.io/docs/stakewise"
    display_funding_message
}

## -- Start of Script -- ##

# ensure root access
if [ "$(id -u)" -ne 0 ]; then
  echo "Please run this script as root (or with sudo)"
  exit
fi

usagemsg="Usage: init-node.sh [--reset|-r] [--mnemonic|-m=mnemonic] VAULT\nSupported vaults: holesky, gravita\nExample: sudo sh init-node.sh -m \"correct horse battery staple...\" holesky"
reset=false
shutdown=false

while getopts "rhsm:-:" option; do
    case $option in
        -)
            case "${OPTARG}" in
                reset)
                    reset=true
                    ;;
                mnemonic=*)
                    mnemonic=${OPTARG#*=}
                    ;;
                mnemonic)
                    mnemonic="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                shutdown)
                    shutdown=true
                    ;;
                help)
                    printf "$usagemsg\n"
                    exit 0
                    ;;
                \?)
                    printf "$usagemsg\n"
                    exit 1
                    ;;
                :)
                    printf "ERROR: Option -$option requires an argument\n\n"
                    printf "$usagemsg\n"
                    exit 1
                    ;;
                *) 
                    echo "Unknown arg -$option"
                    exit 1
                    ;;
            esac
            ;;
        r)
            reset=true
            ;;
        s)
            shutdown=true
            ;;
        h)
            printf "$usagemsg\n"
            exit 0
            ;;
        m)
            mnemonic=${OPTARG}
            ;;
        m=*)
            mnemonic=${OPTARG#*=}
            ;;
        \?)
            printf "$usagemsg\n"
            exit 1
            ;;
        :)
            printf "ERROR: Option -$option requires an argument\n\n"
            printf "$usagemsg\n"
            exit 1
            ;;
        *) 
            echo "Unknown arg"
            exit 1
            ;;
    esac
done
shift $(( OPTIND - 1 ))

# check vault name makes sense
if [ "$1" != "holesky" ] && [ "$1" != "gravita" ]; then
  printf "Error: you must provide a valid vault name\n\n"
  printf "$usagemsg\n"
  exit
fi

# set env based on vault provided
set -a 
source ./${1}.env
set +a

if [ $shutdown ]; then
    docker compose down --remove-orphans
    exit
fi

if [ $reset ]; then
     if [ "$1" != "holesky" ]; then
        
        # todo: check if there are any active validators before giving this warning
        # i.e. docker compose up geth "check validators request"
        echo "DANGER: You are attempting to reset your configuration for a mainnet vault!"
        echo "This will require you to resync the chain completely before you can begin validating again, which may take several days."
        echo "Remember, if you're offline for too long, you may be kicked out of NodeSet!"
        echo "Are you sure you want to continue? You must type 'I UNDERSTAND' to continue."
        read answer

        if [ "$answer" != "I UNDERSTAND" ]; then 
            echo Cancelled
            exit
        fi

        echo "THIS IS YOUR FINAL WARNING! Are you absolutely sure that you want to delete all of your data for this mainnet configuration?"
        echo "You must type 'DELETE EVERYTHING' to continue."
        read answer2

        if [ "$answer2" != "DELETE EVERYTHING" ]; then 
            echo Cancelled
            exit
        fi
    fi
    clean
fi

if [ ! -e ./tmp/jwtsecret ]; then
    echo "No prior jwtsecret found, creating..."
    generate_jwtsecret
else
    echo "Prior jwtsecret found. Skipping creation."
fi

# todo: only run checkpoint sync if no db exists
# if [ "$NETWORK" != "mainnet" ]; then
#     checkpoint
# fi

# always pull latest stakewise operator image in case it's been updated
echo "Pulling latest StakeWise operator binary..."
docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master

if [ ! -d "./stakewise-data/$VAULT/keystores" ]; then
    echo "No prior StakeWise setup found. Initializing..."
    setup_stakewise
else
    echo "Prior StakeWise setup found. Skipping initialization."
fi

echo Starting node...
docker compose up -d
echo "{::} Node started successfully!"
echo
echo "After continuing, logs will be displayed. Please inspect them to verify everything is working as expected. Once you're satisfied, you may exit safely with ^C and the containers will continue running."
read -rsn1 -p "Press any key to continue..."; echo
docker compose logs -f