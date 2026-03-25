#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/k3s-purge.sh) `

# Source shared functions
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Function: k3s_purge
# Description: 🗑️ Purge K3s install from the current system
k3s_purge() {
  echo "🛑 Stopping and disabling K3s..."

  # Check for k3s uninstall script (server or agent)
  if [ -f "/usr/local/bin/k3s-uninstall.sh" ]; then
    echo "🧹 Running official K3s server uninstall script..."
    sudo /usr/local/bin/k3s-uninstall.sh
  elif [ -f "/usr/local/bin/k3s-agent-uninstall.sh" ]; then
    echo "🧹 Running official K3s agent uninstall script..."
    sudo /usr/local/bin/k3s-agent-uninstall.sh
  else
    echo "❌ K3s uninstall script not found!"
    return 1
  fi

  systemd_cleanup \
    "/etc/systemd/system/k3s.service" \
    "/etc/systemd/system/k3s-agent.service"

  echo "✅ K3s completely purged from this system."
}

# Required or `RunBashFunction` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi
