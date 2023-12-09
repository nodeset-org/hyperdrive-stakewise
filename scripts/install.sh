#!/usr/bin/env bash 

# check if bash
if [ "$BASH_VERSION" = '' ]; then
    printf "Please execute this with a bash-compatible shell.\nExample: sudo bash install-node.sh"
    exit
fi

# ensure root access
if [ "$(id -u)" -ne 0 ]; then
  echo "Please run as root (or with sudo)"
  exit
fi

export SCRIPT_DIR=$( cd -- "$( dirname -- "$( realpath ${BASH_SOURCE[0]} )" )" &> /dev/null && pwd )
usagemsg="Usage: node-install.sh [--data-directory|-d=DATA_DIRECTORY] [--mnemonic|-m=MNEMONIC] [--vault|-v=VAULT] [--remove|-r]\nSupported vaults: holesky, gravita\nExample: sudo bash node-install.sh -d "~/data" -m \"correct horse battery staple...\" -v=holesky"
data_dir=""
mnemonic=""
vault=""
remove=false
if [ $SUDO_USER ]; then 
    callinguser=$SUDO_USER; 
else 
    callinguser=`whoami`
fi


while getopts "hrv:d:m:-:" option; do
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
                vault=*)
                    vault=${OPTARG#*=}
                    ;;
                vault)
                    vault="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                help)
                    printf "$usagemsg\n"
                    exit 0
                    ;;
                remove)
                    remove=true
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
        d=*)
            data_dir=${OPTARG#*=}
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
        v)
            vault=${OPTARG}
            ;;
        v=*)
            vault=${OPTARG#*=}
            ;;
        r)
            remove=true
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

echo "{::} Welcome to the NodeSet node installer for StakeWise {::}"
echo

if [ "$remove" = true ]; then
    "$SCRIPT_DIR/remove.sh"
fi

### set default data dir if not passed in
if [ "$data_dir" = "" ]; then
    data_dir="/home/$callinguser/node-data"
fi

### create necessary directories and set permissions
if [ -d "$data_dir" ]; then
    if [ -n "$(ls -A "$data_dir")" ]; then  
        echo "Error: data directory exists and is not empty."
        echo "Given directory: $data_dir"
        exit 1
    fi
else
    mkdir $data_dir
fi
mkdir $data_dir/nimbus-data
mkdir $data_dir/stakewise-data
chown $callinguser $data_dir/nimbus-data
chmod 700 $data_dir/nimbus-data
# v3-operator user is "nobody" for safety since keys are stored there
# you will need to use root to access this directory
chown nobody $data_dir/stakewise-data

### determine and install vault config
get_vault()
{
    echo 
    echo "Which vault do you want to use?"
    echo "1) NodeSet Holesky Test Vault"
    echo "2) Gravita (mainnet)"
    echo
    read vault
    if [ "$vault" = "1" ]; then
        vault="holesky"
    fi
    if [ "$vault" = "2" ]; then
        vault="gravita"
    fi
    if [ "$vault" != "holesky" ] && [ "$vault" != "gravita" ]; then
        get_vault
    fi
}

if [ "$vault" = "" ]; then
    get_vault
elif [ "$vault" != "holesky" ] && [ "$vault" != "gravita" ]; then
    echo "Error: incorrect vault name provided."
    printf $usagemsg
    exit 1
fi

cp "$SCRIPT_DIR/../vaults/$vault.env" "$data_dir/$vault.env"


### set env
set -a 
source "$data_dir/$vault.env"
set +a


### generate jwtsecret
if [ ! -e ./tmp/jwtsecret ]; then
    echo "Generating jwtsecret..."
    # initialize EC, then wait a few seconds for it to create the jwtsecret
    docker compose -f ./scripts/compose.yaml run -d  geth --authrpc.jwtsecret /tmp/jwtsecret
    sleep 3
    chown $callinguser $data_dir/tmp/jwtsecret
fi


### setup stakewise operator
echo "Pulling latest StakeWise operator binary..."
docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master

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

display_funding_message()
{
    echo "Please send some ETH to the wallet address above (on the $NETWORK network), then type 'wallet is funded' to continue."
    read answer

    if [ "$answer" != "wallet is funded" ]; then 
        display_funding_message
    fi
}

echo "Please note that you must have enough Ether in this node wallet to register validators."
printf "Each validator takes approximately 0.01 ETH to create when gas is 30 gwei. We recommend depositing AT LEAST 0.1 ETH.\nYou can withdraw this ETH at any time. For more information, see: http://nodeset.io/docs/stakewise\n"
display_funding_message

### set bashrc
# todo: move this to bashrc
sudo echo "alias nodeset='bash nodeset.sh'" >> /etc/bash.bashrc


### complete
echo 
echo "{::} Installation Complete! {::}"
echo 
echo "We recommend that you verify the configuration file in your installation directory looks correct:"
echo $data_dir/$vault.env
echo
echo "Please log out, then log back in to reload your environment, then start the node with:"
echo "nodeset run"