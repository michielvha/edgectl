#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/rke2-status.sh) ` 

# Function: rke2_status
# Description: ℹ️ Get detailed information about rke2 installation
# TODO: expand this status check, possibly with some details about the cluster state in vault provided by the vault client.
rke2_status() {
  # Check the status of RKE2 services
  if systemctl is-active --quiet rke2-server; then
    sudo systemctl status rke2-server
  elif systemctl is-active --quiet rke2-agent; then
    sudo systemctl status rke2-agent
  else
    echo "Neither rke2-server nor rke2-agent are running."
  fi

  # TODO: output the cluster-id
}

# Required or `CommonGoHelper` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi