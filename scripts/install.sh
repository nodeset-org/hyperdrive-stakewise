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

export SCRIPT_DIR=$( dirname -- "$( readlink -f -- "${BASH_SOURCE[0]}"; )"; )
APP_DIR="$SCRIPT_DIR/.."
LOCAL_DIR="$APP_DIR/local"
CLIENT_DIR="$APP_DIR/local/clients"
VAULT_DIR="$APP_DIR/local/vaults"

usagemsg="Usage: install-node.sh [--data-directory|-d=DATA_DIRECTORY] [--mnemonic|-m=MNEMONIC] [--vault|-v=VAULT] [--remove|-r]\nSupported vaults: holesky, gravita\nExample: sudo bash install-node.sh -d "~/data" -m \"correct horse battery staple...\" -v=holesky"
export DATA_DIR=""
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
                    export DATA_DIR=${OPTARG#*=}
                    ;;
                data-directory)
                    export DATA_DIR="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
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
            export DATA_DIR=${OPTARG}
            ;;
        d=*)
            export DATA_DIR=${OPTARG#*=}
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

### set default data dir if not passed in
if [ "$DATA_DIR" = "" ]; then
    export DATA_DIR="/home/$callinguser/node-data"
fi

echo "{::} Welcome to the NodeSet node installer for StakeWise {::}"
echo

if [ "$remove" = true ]; then
    "$SCRIPT_DIR/nodeset.sh" "-d" "$DATA_DIR" "remove" 
    if [ $? -ne 0 ]; then
        exit 2
    fi
fi

### create data directory
if [ -d "$DATA_DIR" ]; then
    if [ -n "$(ls -A "$DATA_DIR")" ]; then  
        echo "Data directory exists and is not empty."
        echo "To remove your existing installation, use the -r or --remove option"
        echo
        echo "Given data directory: $DATA_DIR"
        exit 1
    fi
else
    mkdir "$DATA_DIR" || exit 1
    chown $callinguser "$DATA_DIR"
fi

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

cp "$VAULT_DIR/$vault.env" "$DATA_DIR/nodeset.env"

### set local env
set -a 
source "$DATA_DIR/nodeset.env"
set +a

### install compose configs
# todo: first ask which clients to install here
mkdir $DATA_DIR/$CCNAME-data
mkdir $DATA_DIR/stakewise-data
chown $callinguser $DATA_DIR/$CCNAME-data
chmod 700 $DATA_DIR/$CCNAME-data
# v3-operator user is "nobody" for safety since keys are stored there
# you will need to use root to access this directory
chown nobody $DATA_DIR/stakewise-data
cp "$LOCAL_DIR/compose.yaml" "$DATA_DIR/compose.yaml"
cp "$CLIENT_DIR/$ECNAME.yaml" "$DATA_DIR/$ECNAME.yaml"
cp "$CLIENT_DIR/$CCNAME.yaml" "$DATA_DIR/$CCNAME.yaml"


### generate jwtsecret
if [ ! -e ./tmp/jwtsecret ]; then
    echo "Generating jwtsecret..."
    # initialize EC, then wait a few seconds for it to create the jwtsecret
    docker compose -f "$DATA_DIR/compose.yaml" run -d $ECNAME
    sleep 3
    chown $callinguser $DATA_DIR/tmp/jwtsecret
fi


### setup stakewise operator
echo "Pulling latest StakeWise operator binary..."
docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master

if [ "$mnemonic" != "" ]; then
    echo "Recreating StakeWise configuration using existing mnemonic..."
    docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py recover --network="$NETWORK" --vault="$VAULT" --execution-endpoints="http://$ECNAME:$ECAPIPORT" --consensus-endpoints="http://$CCNAME:$CCAPIPORT" --mnemonic="$mnemonic"
    docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py create-wallet --vault="$VAULT" --mnemonic="$mnemonic"
else
    echo "Initializing new StakeWise configuration..."
    docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py init --network="$NETWORK" --vault="$VAULT" --language=english
    docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py create-keys --vault="$VAULT" --count="$NUMKEYS"
    docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py create-wallet --vault="$VAULT"
fi

display_funding_message()
{
    echo
    echo "Please send some ETH to the wallet address above (on the $NETWORK network), then type 'wallet is funded' to continue."
    read answer

    if [ "$answer" != "wallet is funded" ]; then 
        display_funding_message
    fi
}

echo
echo "Please note that you must have enough Ether in this node wallet to register validators."
printf "Each validator takes approximately 0.01 ETH to create when gas is 30 gwei. We recommend depositing AT LEAST 0.1 ETH.\nYou can withdraw this ETH at any time. For more information, see: http://nodeset.io/docs/stakewise\n"
display_funding_message

### checkpoint sync
if [ "$NETWORK" != "mainnet" ]; then
    if [ "$CCNAME" = "nimbus" ]; then 
        echo "Performing checkpoint sync..."
        docker compose -f "$DATA_DIR/compose.yaml" run nimbus trustedNodeSync -d=/home/user/data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false
    fi
fi

### set bashrc
nsalias="\n\n# NodeSet\nalias nodeset='bash \"$SCRIPT_DIR/nodeset.sh\" \"-d\" \"$DATA_DIR\"'\n"
if [ "$( tail -n 4 /etc/bash.bashrc )" != nsalias ]; then
    printf $nsalias >> /etc/bash.bashrc
    echo "Added NodeSet bashrc entry"
else
    echo "NodeSet bashrc entry already set"
fi

### install systemd service
if [ -f "/etc/systemd/system/nodeset.service" ]; then
    systemctl stop nodeset.service
    systemctl disable nodeset.service
fi
cp "$LOCAL_DIR/nodeset.service" "/etc/systemd/system/nodeset.service"
systemctl enable nodeset.service

### complete
echo 
echo "{::} Installation Complete! {::}"
echo 
echo "We recommend that you verify the configuration file in your installation directory looks correct:"
echo $DATA_DIR/nodeset.env
echo
echo "Please log out, then log back in to reload your environment, then start the node with:"
echo "nodeset start"