#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/rke2-status.sh) `

# Source shared functions
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Function: rke2_status
# Description: ℹ️ Get detailed information about rke2 installation
# TODO: expand this status check, possibly with some details about the cluster state in vault provided by the vault client.
rke2_status() {
  check_status "rke2-server" "rke2-agent"
}

# Required or `RunBashFunction` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi