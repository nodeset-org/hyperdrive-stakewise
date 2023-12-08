#!/usr/bin/env bash 

# create necessary directories and set permissions
mkdir $data_dir/nimbus-data
mkdir $data_dir/stakewise-data
chown $(logname) ./nimbus-data
chmod 700 ./nimbus-data
# v3-operator user is "nobody" for safety since keys are stored there
# you will need to use root to access this directory
chown nobody ./stakewise-data \

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
    printf "Each validator takes approximately 0.01 ETH to create when gas is 30 gwei. We recommend depositing AT LEAST 0.1 ETH.\nYou can withdraw this ETH at any time. For more information, see: http://nodeset.io/docs/stakewise\n"
    display_funding_message
}
