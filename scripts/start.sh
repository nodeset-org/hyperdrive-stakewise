#!/usr/bin/env bash 

echo "Starting node..."

if [ $ECNAME != "external" ]; then
    composeFile=(-f "$DATA_DIR/compose.yaml" -f "$DATA_DIR/compose.internal.yaml")
else
    composeFile=(-f "$DATA_DIR/compose.yaml")
fi

# pull latest container images
echo "Updating..."
docker compose ${composeFile[@]} pull
docker compose ${composeFile[@]} up -d

echo
echo "{::} Node Started! {::}"
echo