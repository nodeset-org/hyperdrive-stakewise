#!/usr/bin/env bash 

echo "Starting node..."

if [ $ECNAME != "external" ]; then
    composeFile=("$DATA_DIR/compose.yaml" "$DATA_DIR/compose.internal.yaml")
else
    composeFile=("$DATA_DIR/compose.yaml")
fi

# pull latest container images
echo "Updating..."
docker compose -f "$composeFile" pull
docker compose -f "$composeFile" up -d

echo
echo "{::} Node Started! {::}"
echo