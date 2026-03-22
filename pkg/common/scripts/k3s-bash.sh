#!/bin/bash
# Usage: ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/k3s-bash.sh) `

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
