#!/usr/bin/env bash
# Initialize OpenBao, unseal, and enable the KV v2 engine.
# Run once after the first `docker compose up -d`.
#
# Usage: ./init.sh
set -euo pipefail

DC="docker compose exec openbao"
INIT_OUTPUT=$($DC bao operator init 2>&1)

echo "$INIT_OUTPUT"
echo ""
echo "========================================="
echo "  SAVE THE KEYS AND TOKEN ABOVE!"
echo "========================================="
echo ""

# Extract unseal keys and root token
UNSEAL_KEYS=$(echo "$INIT_OUTPUT" | grep "Unseal Key" | awk '{print $NF}')
ROOT_TOKEN=$(echo "$INIT_OUTPUT" | grep "Initial Root Token" | awk '{print $NF}')

# Unseal with the first 3 keys
COUNT=0
for KEY in $UNSEAL_KEYS; do
    if [ $COUNT -ge 3 ]; then break; fi
    $DC bao operator unseal "$KEY" > /dev/null
    COUNT=$((COUNT + 1))
done
echo "Unsealed with $COUNT keys."

# Enable KV v2 at the path edgectl expects
$DC env BAO_TOKEN="$ROOT_TOKEN" bao secrets enable -path=kv -version=2 kv
echo "KV v2 engine enabled at kv/."

echo ""
echo "OpenBao is ready. Set these on your host:"
echo ""
echo "  export VAULT_ADDR=\"http://127.0.0.1:8200\""
echo "  export BAO_TOKEN=\"$ROOT_TOKEN\""
