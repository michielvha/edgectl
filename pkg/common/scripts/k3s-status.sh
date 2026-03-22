#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/k3s-status.sh) `

# Function: k3s_status
# Description: ℹ️ Get detailed information about K3s installation
k3s_status() {
  # output cluster-id
  [ -s /etc/edgectl/cluster-id ] && echo "🔑 Cluster ID: $(cat /etc/edgectl/cluster-id)" || echo "❌ Cluster ID not found. Please check if the cluster is initialized."

  # Check the status of K3s services
  if systemctl is-active --quiet k3s; then
    sudo systemctl status k3s
  elif systemctl is-active --quiet k3s-agent; then
    sudo systemctl status k3s-agent
  else
    echo "Neither k3s nor k3s-agent are running."
  fi
}

# Required or `RunBashFunction` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi
