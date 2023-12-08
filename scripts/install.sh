#!/usr/bin/env bash 

echo "{::} Welcome to the NodeSet node installer for StakeWise {::}"
network_question()
{
    echo 
    echo "Which network do you want to use?"
    echo "1) holesky"
    echo "2) mainnet"
    echo
   
    read network
    if [ "$network" != "holesky" ] && [ "$network" != "mainnet"]
        network_question
}
network_question

usagemsg="Usage: install.sh [--data-directory|-d=DATA_DIRECTORY] [--mnemonic|-m=MNEMONIC] [--vault|-v=VAULT]\nSupported vaults: holesky, gravita\nExample: sudo bash install.sh -d "~/data" -m \"correct horse battery staple...\" -v=holesky"
data_dir="/home/${ whoami }/node-data"

while getopts "hd:m:-:" option; do
    case $option in
        -)
            case "${OPTARG}" in
                mnemonic=*)
                    mnemonic=${OPTARG#*=}
                    ;;
                mnemonic)
                    mnemonic="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                data-directory=*)
                    data_dir=${OPTARG#*=}
                    ;;
                data-directory)
                    data_dir="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
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
        d)
            data_dir=${OPTARG}
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

# create necessary directories and set permissions

if [ -d "$data_dir" ]; then
    if [ -n "$(ls -A "$data_dir")" ]; then  
        echo "Error: data directory exists is not empty."
        echo "Given directory: $data_dir"
        exit
    fi
else
    mkdir $data_dir
fi

mkdir $data_dir/nimbus-data
mkdir $data_dir/stakewise-data
chown $(logname) ./nimbus-data
chmod 700 ./nimbus-data
# v3-operator user is "nobody" for safety since keys are stored there
# you will need to use root to access this directory
chown nobody ./stakewise-data \

# generate jwtsecret

if [ ! -e ./tmp/jwtsecret ]; then
    echo "No prior jwtsecret found, creating..."
    echo "Generating jwtsecret..."
    # initialize EC, then wait a few seconds for it to create the jwtsecret
    docker compose run -d geth --authrpc.jwtsecret /tmp/jwtsecret
    sleep 3
    chown $(logname) ./tmp/jwtsecret
else
    echo "Prior jwtsecret found. Skipping creation."
fi

# setup stakewise operator

echo "Pulling latest StakeWise operator binary..."
docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master

display_funding_message()
{
    echo "Please send some ETH to the wallet address above (on the $NETWORK network), then type 'wallet is funded' to continue."
    read answer

    if [ "$answer" != "wallet is funded" ]; then 
        display_funding_message
    fi
}

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
printf "Each validator takes approximately 0.01 ETH to create when gas is 30 gwei. We recommend depositing AT LEAST 0.1 ETH.\nYou can withdraw this ETH at any time. For more information, see: http://nodeset.io/docs/stakewise\n"
display_funding_message
