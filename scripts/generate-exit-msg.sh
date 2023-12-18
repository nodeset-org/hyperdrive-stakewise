#!/usr/bin/env bash 

echo "Generating exit messages..."
docker compose -f "$DATA_DIR/compose.yaml" run ethdo validator exit --validator=$DATA_DIR/stakewise-data/$VAULT/keystore-m_12381_3600_0_0_0-1702688242.json --offline