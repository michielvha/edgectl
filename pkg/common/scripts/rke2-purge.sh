#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/rke2-purge.sh) ` 

# Function: rke2_purge
# Description: 🗑️ Purge RKE2 install from the current system
rke2_purge() {
  echo "🛑 Stopping and disabling RKE2..."

  # check if rke2 uninstall script exists
  [ -f "/usr/local/bin/rke2-uninstall.sh" ] || { echo "❌ RKE2 uninstall script not found!"; return 1; }

  # Check if any RKE2 service exists and remove it
  if systemctl list-unit-files | grep -q "^rke2-server.service" || systemctl list-unit-files | grep -q "^rke2-agent.service"; then
    echo "🧹 Running official RKE2 uninstall script..."
    sudo /usr/local/bin/rke2-uninstall.sh
  else
    echo "ℹ️ Neither rke2-server nor rke2-agent service exists."
  fi

  echo "🗑️ Cleaning up leftover systemd service files..."
  sudo rm -f /usr/local/lib/systemd/system/rke2-server.service
  sudo rm -f /usr/local/lib/systemd/system/rke2-agent.service

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

  echo "✅ RKE2 completely purged from this system."
}

# Required or `CommonGoHelper` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi