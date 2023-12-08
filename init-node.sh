#!/usr/bin/env bash 





display_funding_message()
{
    echo "Please send some ETH to the wallet address above (on the $NETWORK network), then type 'wallet is funded' to continue."
    read answer

    if [ "$answer" != "wallet is funded" ]; then 
        display_funding_message
    fi
}







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