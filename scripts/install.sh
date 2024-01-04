#!/usr/bin/env bash 

# check if bash
if [ "$BASH_VERSION" = '' ]; then
    printf "Please execute this with a bash-compatible shell.\nExample: sudo bash install-node.sh"
    exit
fi

# check for docker
if ! docker -v > /dev/null || ! docker compose version > /dev/null; then
    echo "Error: Dependencies are missing! Please ensure you have docker and docker compose installed."
    exit 2
fi

export SCRIPT_DIR=$( dirname -- "$( readlink -f -- "${BASH_SOURCE[0]}"; )"; )
APP_DIR="$SCRIPT_DIR/.."
LOCAL_DIR="$APP_DIR/local"
CLIENT_DIR="$APP_DIR/local/clients"
VAULT_DIR="$APP_DIR/local/vaults"
version=$(< "$SCRIPT_DIR/version.txt")
help=$(< "$SCRIPT_DIR/install-help.txt")
usagemsg="\n"${help/VERSION/"v"$version}"\n\n"
export DATA_DIR=""
eth1client=""
eth2client=""
mnemonic=""
checkpoint=true
vault=""
remove=false

while getopts "hre:c:v:d:m:-:" option; do
    case $option in
        -)
            case "${OPTARG}" in
                eth1client=*)
                    eth1client=${OPTARG#*=}
                    ;;
                eth1client)
                    eth1client="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                eth2client=*)
                    eth2client=${OPTARG#*=}
                    ;;
                eth2client)
                    eth2client="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                data-directory=*)
                    export DATA_DIR=${OPTARG#*=}
                    ;;
                data-directory)
                    export DATA_DIR="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                help)
                    printf "$usagemsg\n"
                    exit 0
                    ;;
                mnemonic=*)
                    mnemonic=${OPTARG#*=}
                    ;;
                mnemonic)
                    mnemonic="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                no-checkpoint)
                    checkpoint=false
                    ;;
                remove)
                    remove=true
                    ;;
                vault=*)
                    vault=${OPTARG#*=}
                    ;;
                vault)
                    vault="${!OPTIND}"; OPTIND=$(( $OPTIND + 1 ))
                    ;;
                \?)
                    printf "$usagemsg\n"
                    exit 1
                    ;;
                :)
                    printf "Error: Option -$option requires an argument\n\n"
                    printf "$usagemsg\n"
                    exit 1
                    ;;
                *) 
                    echo "Unknown arg -$option"
                    exit 1
                    ;;
            esac
            ;;
        c)
            eth2client=${OPTARG}
            ;;
        d)
            export DATA_DIR=${OPTARG}
            ;;
        e)
            eth1client=${OPTARG}
            ;;
        h)
            printf "$usagemsg\n"
            exit 0
            ;;
        m)
            mnemonic=${OPTARG}
            ;;
        v)
            vault=${OPTARG}
            ;;
        r)
            remove=true
            ;;
        \?)
            printf "$usagemsg\n"
            exit 1
            ;;
        :)
            printf "Error: Option -$option requires an argument\n\n"
            printf "$usagemsg\n"
            exit 1
            ;;
        *) 
            echo "Unknown arg"
            exit 1
            ;;
    esac
done

### ensure root access
if [ "$(id -u)" -ne 0 ]; then
  echo "Please run as root (or with sudo)"
  exit
fi

### find calling user
if [ $SUDO_USER ]; then 
    callinguser=$SUDO_USER; 
else 
    callinguser=`whoami`
fi

### set default data dir if not passed in
if [ "$DATA_DIR" = "" ]; then
    export DATA_DIR="/home/$callinguser/.node-data"
fi

echo "{::} Welcome to the NodeSet node installer for StakeWise {::}"
echo

if [ "$remove" = true ]; then
    "$SCRIPT_DIR/nodeset.sh" "-d" "$DATA_DIR" "remove" 
    c=$?
    if [ $c -ne 0 ]; then
        exit $c
    fi
fi

### create data directory
if [ -d "$DATA_DIR" ]; then
    if [ -n "$(ls -A "$DATA_DIR")" ]; then  
        echo "Error: Data directory exists and is not empty."
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
    echo "1) NodeSet Test Vault (holesky)"
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
elif [ "$vault" != "holesky" ] && [ "$vault" != "holesky-dev" ] && [ "$vault" != "gravita" ]; then
    echo "Error: incorrect vault name provided."
    printf $usagemsg
    exit 1
fi

### determine and install client config
get_eth1()
{
    echo 
    echo "Which execution (eth1) client do you want to use?"
    echo "1) Nethermind (recommended)"
    echo "2) Geth"
    echo
    read choice
    if [ "$choice" = "1" ] || [ "$choice" = "nethermind" ]; then
        eth1client="nethermind"
    fi
    if [ "$choice" = "2" ] || [ "$choice" = "geth" ]; then
        eth1client="geth"
    fi
    if [ "$eth1client" != "nethermind" ] && [ "$eth1client" != "geth" ]; then
        get_eth1
    fi
}
if [ "$eth1client" = "" ]; then
    get_eth1
elif [ "$eth1client" != "geth" ] && [ "$eth1client" != "nethermind" ]; then
    echo "Error: incorrect eth1 client name provided."
    printf $usagemsg
    exit 1
fi

get_eth2()
{
    echo 
    echo "Which consensus (eth2) client do you want to use?"
    echo "1) Nimbus (recommended)"
    echo "2) Teku"
    echo
    read choice
    if [ "$choice" = "1" ] || [ "$choice" = "nimbus" ]; then
        eth2client="nimbus"
    fi
    if [ "$choice" = "2" ] || [ "$choice" = "teku" ]; then
        eth2client="teku"
    fi
    if [ "$eth2client" != "nimbus" ] && [ "$eth2client" != "teku" ]; then
        get_eth2
    fi
}
if [ "$eth2client" = "" ]; then
    get_eth2
elif [ "$eth2client" != "nimbus" ] && [ "$eth2client" != "teku" ]; then
    echo "Error: incorrect eth2 client name provided."
    printf $usagemsg
    exit 1
fi

# install default vault config
cp "$VAULT_DIR/$vault.env" "$DATA_DIR/nodeset.env"

# replace default client names in installed configuration
sed -i -e "s/ECNAME=.*/ECNAME=$eth1client/g" "$DATA_DIR/nodeset.env"
sed -i -e "s/CCNAME=.*/CCNAME=$eth2client/g" "$DATA_DIR/nodeset.env"

### set local env
set -a 
source "$DATA_DIR/nodeset.env"
set +a

### prep data directory
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
if [ ! -e ./jwtsecret/jwtsecret ]; then
    echo "Generating jwtsecret..."
    # initialize EC, then wait a few seconds for it to create the jwtsecret
    docker compose -f "$DATA_DIR/compose.yaml" up -d $ECNAME
    i=6
    until [ -f "$DATA_DIR/jwtsecret/jwtsecret" ] || [ $i = 0 ]; do
        echo "Waiting for jwtsecret..."
        sleep 5
        i=$((i-1))
    done
    if [ ! -f "$DATA_DIR/jwtsecret/jwtsecret" ]; then
        echo "Error: Could not generate jwtsecret before timeout!"
        exit 3
    fi

    chown $callinguser $DATA_DIR/jwtsecret/jwtsecret || exit 3
fi

### checkpoint sync
if [ $checkpoint = true ] && [ "$NETWORK" != "mainnet" ]; then
    case $CCNAME in
        nimbus) 
            echo "Performing checkpoint sync..."
            docker compose -f "$DATA_DIR/compose.yaml" run nimbus trustedNodeSync -d=/home/user/data --network=$NETWORK --trusted-node-url=https://checkpoint-sync.holesky.ethpandaops.io --backfill=false
            ;;
    esac
fi

### set bashrc
nsalias="\n\n# NodeSet\nalias nodeset='sudo bash \"$SCRIPT_DIR/nodeset.sh\" \"-d\" \"$DATA_DIR\"'\n"
if [ "$( tail -n 4 /etc/bash.bashrc )" != nsalias ]; then
    printf "$nsalias" >> /etc/bash.bashrc
    echo "Added NodeSet bashrc entry"
else
    echo "NodeSet bashrc entry already set"
fi

### install systemd service
if [[ -d /run/systemd/system ]]; then
    if [ -f "/etc/systemd/system/nodeset.service" ]; then 
        # service is already installed, so stop and disable before overwriting just in case
        systemctl stop nodeset.service
        systemctl disable nodeset.service
    fi 
    cp "$LOCAL_DIR/nodeset.service" "/etc/systemd/system/nodeset.service"
    systemctl enable nodeset.service   
else
    echo "WARNING: you are not using systemd! You will need to create your own boot and shutdown automation."
    read -p "Press enter to acknowledge and continue..."
fi

### setup stakewise operator
echo "Pulling latest StakeWise operator binary..."
docker pull europe-west4-docker.pkg.dev/stakewiselabs/public/v3-operator:master

if [ "$mnemonic" != "" ]; then
    echo "supplying a mnemonic is not yet supported, please check back later!"
    exit

    echo "Recreating StakeWise configuration using existing mnemonic..."
    # todo: recover setup using deposit data downloaded from NodeSet API
    #docker compose run stakewise src/main.py get-validators-root --deposit-data-file=<DEPOSIT DATA FILE>
    docker compose -f "$DATA_DIR/compose.yaml" run stakewise src/main.py recover --network="$NETWORK" --vault="$VAULT" --consensus-endpoints="http://$CCNAME:$CCAPIPORT" --execution-endpoints="http://$ECNAME:$ECAPIPORT" --mnemonic="$mnemonic"
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
    echo "Please send some ETH to the wallet address above (on the $NETWORK network), then type 'wallet is funded' to continue..."
    read answer

    if [ "$answer" != "wallet is funded" ]; then 
        display_funding_message
    fi
}

echo
echo "Please note that you must have enough Ether in this node wallet to register validators."
printf "Each validator takes approximately 0.01 ETH to create when gas is 30 gwei. We recommend depositing AT LEAST 0.1 ETH.\nYou can withdraw this ETH at any time. For more information, see: http://nodeset.io/docs/stakewise\n"
display_funding_message

### start node
echo "Starting node..."
sudo bash $SCRIPT_DIR/nodeset.sh -d "$DATA_DIR" start

### complete
echo 
echo "{::} Installation Complete! {::}"
echo
echo "Your new node is started!"
echo
echo "We recommend that you check two things from here:"
echo "1. Verify that your node is syncing correctly and watch its progress with \"nodeset logs\""
echo "2. Verify the configuration file in your installation directory looks correct:"
echo $DATA_DIR/nodeset.env
echo
echo "Note that you must reload your environment (exit and log in again) to enable the \"nodeset\" commands."