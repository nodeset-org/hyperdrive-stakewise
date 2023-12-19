#!/usr/bin/env bash 

echo "Generating exit messages..."
docker compose -f "$DATA_DIR/compose.yaml" run ethdo validator exit --account=$DATA_DIR/stakewise-data/$VAULT/keystore-m_12381_3600_0_0_0-1702688242.json --passphrase=$DATA_DIR/stakewise-data/$VAULT/password.txt --offline