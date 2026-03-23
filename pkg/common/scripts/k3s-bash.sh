#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/k3s-bash.sh) `

# Source shared functions
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# Function: setup_k3s_node_bash_env
# Description: Configures the shell environment for K3s administration on a server or agent node.
# K3s installs kubectl to /usr/local/bin by default, so PATH setup is typically not needed.
# This function ensures KUBECONFIG is set for the root user.
setup_k3s_node_bash_env() {
  local profile_file="/etc/profile.d/k3s.sh"

  # Ensure the file exists
  sudo touch "$profile_file"

  # Add KUBECONFIG if not already present (K3s writes kubeconfig to /etc/rancher/k3s/k3s.yaml)
  grep -q 'export KUBECONFIG=/etc/rancher/k3s/k3s.yaml' "$profile_file" || echo "export KUBECONFIG=/etc/rancher/k3s/k3s.yaml" | sudo tee -a "$profile_file" > /dev/null

  # Source the profile file to apply changes immediately
  # shellcheck source=/dev/null
  source "$profile_file"
}

# setup_kubectl_bash_env is provided by common.sh

# Required or `RunBashFunction` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi
