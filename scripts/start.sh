#!/usr/bin/env bash 

echo "Starting node..."

# pull latest container images
echo "Updating..."
if [ $INTERNALCLIENTS ]; then
    docker compose -f "$DATA_DIR/compose.yaml" -f "$DATA_DIR/compose.internal.yaml" pull
    docker compose -f "$DATA_DIR/compose.yaml" -f "$DATA_DIR/compose.internal.yaml" up -d
else
    docker compose -f "$DATA_DIR/compose.yaml" pull
    docker compose -f "$DATA_DIR/compose.yaml" up -d
fi

echo
echo "{::} Node Started! {::}"
echo