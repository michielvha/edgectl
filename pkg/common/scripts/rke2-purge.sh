#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/rke2-purge.sh) `

# Source shared functions
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Function: rke2_purge
# Description: 🗑️ Purge RKE2 install from the current system
rke2_purge() {
  echo "🛑 Stopping and disabling RKE2..."

  # Check if rke2 uninstall script exists
  [ -f "/usr/local/bin/rke2-uninstall.sh" ] || { echo "❌ RKE2 uninstall script not found!"; return 1; }

  # Check if any RKE2 service exists and remove it
  if systemctl list-unit-files | grep -q "^rke2-server.service" || systemctl list-unit-files | grep -q "^rke2-agent.service"; then
    echo "🧹 Running official RKE2 uninstall script..."
    sudo /usr/local/bin/rke2-uninstall.sh
  else
    echo "ℹ️ Neither rke2-server nor rke2-agent service exists."
  fi

  systemd_cleanup \
    "/usr/local/lib/systemd/system/rke2-server.service" \
    "/usr/local/lib/systemd/system/rke2-agent.service"

  echo "✅ RKE2 completely purged from this system."
}

# Required or `RunBashFunction` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi