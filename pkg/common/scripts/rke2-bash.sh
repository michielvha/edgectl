#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/rke2-bash.sh) `

# Function: setup_rke2_node_bash_env
# Description: Configures the shell environment for RKE2 administration on a server or agent node. This config is only available to the root account.
setup_rke2_node_bash_env() {
  # configure the shell for administration on an RKE2 server/agent node
  local profile_file="/etc/profile.d/rke2.sh"

  # Ensure the file exists
  sudo touch "$profile_file"

  # Add RKE2 to the PATH if not already present
  grep -q 'export PATH=.*:/var/lib/rancher/rke2/bin' "$profile_file" || echo "export PATH=\$PATH:/var/lib/rancher/rke2/bin" | sudo tee -a "$profile_file" > /dev/null
  # Add KUBECONFIG if not already present
  grep -q 'export KUBECONFIG=/etc/rancher/rke2/rke2.yaml' "$profile_file" || echo "export KUBECONFIG=/etc/rancher/rke2/rke2.yaml" | sudo tee -a "$profile_file" > /dev/null

  # Source the profile file to apply changes immediately
  # shellcheck source=/dev/null
  source "$profile_file"
}

# Function: setup_kubectl_bash_env
# Description: Configures the shell environment on any remote administrator's machine.
setup_kubectl_bash_env() {
  local profile_file="$HOME/.bashrc"
  local user_home="$HOME/.kube/config"

  # Ensure required files and directories exist
  sudo touch "$profile_file"
  mkdir -p ~/.kube

  # Add KUBECONFIG if not already present
  grep -q "export KUBECONFIG=$user_home" "$profile_file" || echo "export KUBECONFIG=$user_home" | sudo tee -a "$profile_file" > /dev/null
  # enable bash completion for kubectl
  grep -q 'source <(kubectl completion bash)' "$profile_file" || echo "source <(kubectl completion bash)" | sudo tee -a "$profile_file" > /dev/null
  # add kubectl alias
  grep -q 'alias k=kubectl' "$profile_file" || echo "alias k=kubectl" | sudo tee -a "$profile_file" > /dev/null

  # Source the profile file to apply changes immediately
  # shellcheck source=/dev/null
  source "$HOME/.bashrc" && echo "🔧 User-specific Kubernetes configuration set up for $(whoami)"
}

# Required or `RunBashFunction` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi