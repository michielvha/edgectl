#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/rke2-purge.sh) ` 

# Function: purge_rke2
# Description: 🗑️ Purge RKE2 install from the current system
purge_rke2() {
  echo "🛑 Stopping and disabling RKE2..."

  if systemctl is-active --quiet rke2-server; then
    echo "🧹 Running official RKE2 server uninstall script..."
    if [ -f "/usr/local/bin/rke2-uninstall.sh" ]; then
      sudo /usr/local/bin/rke2-uninstall.sh
    else
      echo "❌ Server uninstall script not found!"
    fi
  elif systemctl is-active --quiet rke2-agent; then
    echo "🧹 Running official RKE2 agent uninstall script..."
    if [ -f "/usr/local/bin/rke2-agent-uninstall.sh" ]; then
      sudo /usr/local/bin/rke2-agent-uninstall.sh
    else
      echo "❌ Agent uninstall script not found!"
    fi
  else
    echo "ℹ️ Neither rke2-server nor rke2-agent are currently active."
  fi

  echo "🗑️ Cleaning up leftover systemd service files..."
  sudo rm -f /usr/local/lib/systemd/system/rke2-server.service
  sudo rm -f /usr/local/lib/systemd/system/rke2-agent.service

  echo "🔁 Rexecuting systemd daemon..."
  if ! sudo systemctl daemon-reexec; then
    echo "❌ Failed to Rexecute systemd daemon."
    return 1
  fi

  echo "🔁 Reloading systemd daemon..."
  if ! sudo systemctl daemon-reload; then
    echo "❌ Failed to reload systemd daemon."
    return 1
  fi

  echo "🔄 Resetting failed systemd services..."
  if ! sudo systemctl reset-failed; then
    echo "❌ Failed to reset failed systemd services."
    return 1
  fi

  echo "✅ RKE2 completely purged from this system."
}

if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi