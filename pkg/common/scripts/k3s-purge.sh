#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/k3s-purge.sh) `

# Function: k3s_purge
# Description: 🗑️ Purge K3s install from the current system
k3s_purge() {
  echo "🛑 Stopping and disabling K3s..."

  # check if k3s uninstall script exists
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

  echo "🗑️ Cleaning up leftover systemd service files..."
  sudo rm -f /etc/systemd/system/k3s.service
  sudo rm -f /etc/systemd/system/k3s-agent.service

  echo "🔁 Rexecuting systemd daemon..."
  if ! sudo systemctl daemon-reexec; then
    echo "❌ Failed to Rexecute systemd daemon."
    return 1
  else
    echo "✅ Systemd daemon re-executed successfully."
  fi

  echo "🔁 Reloading systemd daemon..."
  if ! sudo systemctl daemon-reload; then
    echo "❌ Failed to reload systemd daemon."
    return 1
  else
    echo "✅ Systemd daemon reloaded successfully."
  fi

  echo "🔄 Resetting failed systemd services..."
  if ! sudo systemctl reset-failed; then
    echo "❌ Failed to reset failed systemd services."
    return 1
  else
    echo "✅ Failed systemd services reset successfully."
  fi

  echo "✅ K3s completely purged from this system."
}

# Required or `RunBashFunction` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi
