#!/bin/bash
set -o pipefail
# K3s module for K3s installation and configuration
# purpose: bootstrap K3s nodes.
# usage: quickly source this module with the following command:
# ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/k3s.sh) `
# ------------------------------------------------------------------------------------------------------------------------------------------------

# Source shared functions
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# bootstrap a K3s server node
install_k3s_server() {
  # usage: install_k3s_server [-l <loadbalancer-hostname>]

  # Pre checks
  systemctl list-unit-files | grep -q "^k3s.service" && {
    echo "❌ K3s Server service already exists. Use 'edgectl k3s system purge' Exiting."
    return 1
  }

  echo "📦 Configuring K3s Server Node..."

  # Default parameter values
  local LB_HOSTNAME="loadbalancer.example.com"

  # Parse options using getopts
  while getopts "l:" opt; do
    case "$opt" in
      l) LB_HOSTNAME="$OPTARG" ;;
      \?)
        echo "❌ Invalid option: -$OPTARG"
        echo "Usage: install_k3s_server [-l <loadbalancer-hostname>]"
        return 1
        ;;
    esac
  done

  # environment
  local ARCH
  ARCH=$(uname -m | cut -c1-3)
  local FQDN
  FQDN=$(hostname -f)
  local HOST
  HOST=$(hostname -s)
  local PURPOSE=${PURPOSE:-"server"}

  configure_host   # shared host configuration from common.sh

  # Install K3s
  echo "⬇️  Downloading and installing K3s..."
  curl -sfL https://get.k3s.io | K3S_TOKEN="$K3S_TOKEN" K3S_URL="$K3S_URL" sudo -E sh -s - server \
    --write-kubeconfig-mode "0644" \
    --node-label "environment=production" \
    --node-label "arch=$ARCH" \
    --node-label "purpose=$PURPOSE" \
    --tls-san "$FQDN" \
    --tls-san "$LB_HOSTNAME" \
    --flannel-backend=none \
    --disable-kube-proxy \
    --disable-network-policy \
    --disable=traefik \
    || { echo "❌ Failed to install K3s. Exiting."; return 1; }

  # If K3S_TOKEN is set, this is a secondary server joining an existing cluster
  if [ -n "$K3S_TOKEN" ]; then
    echo "🔑 Token detected, joining existing cluster"
  fi

  # If K3S_URL is set, join as additional server
  if [ -n "$K3S_URL" ]; then
    echo "🌐 Server URL detected: $K3S_URL"
  fi

  # Write Cilium HelmChart manifest for automatic deployment
  echo "🛠️  Writing Cilium Helm Chart Config..."
  sudo mkdir -p /var/lib/rancher/k3s/server/manifests/ && cat <<EOF | sudo tee /var/lib/rancher/k3s/server/manifests/cilium.yaml
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: cilium
  namespace: kube-system
spec:
  repo: https://helm.cilium.io/
  chart: cilium
  targetNamespace: kube-system
  valuesContent: |-
    kubeProxyReplacement: true
    k8sServiceHost: "localhost"
    k8sServicePort: "6443"
    hubble:
      enabled: true
      relay:
        enabled: true
      ui:
        enabled: true
    operator:
      replicas: 1
EOF

  enable_addon_reloader "/var/lib/rancher/k3s/server/manifests/"   # shared from common.sh

  configure_firewall_k3s_server     # K3s-specific server firewall rules

  echo "✅ K3s Server node bootstrapped."
}

# bootstrap a K3s agent node
install_k3s_agent() {
  # usage: install_k3s_agent [-l <loadbalancer-hostname>]
  # Pre checks
  systemctl list-unit-files | grep -q "^k3s-agent.service" && {
    echo "❌ K3s Agent service already exists. Exiting."
    return 1
  }

  # Check for token
  [ -z "$K3S_TOKEN" ] && {
    echo "❌ K3S_TOKEN environment variable not set. Token is required."
    return 1
  } || echo "🔑 Using K3S_TOKEN from environment variable"

  echo "📦 Configuring K3s Agent Node..."

  # Default parameter values
  local LB_HOSTNAME="loadbalancer.example.com"

  # Parse options using getopts
  while getopts "l:" opt; do
    case "$opt" in
      l) LB_HOSTNAME="$OPTARG" ;;
      \?)
        echo "❌ Invalid option: -$OPTARG"
        echo "Usage: install_k3s_agent [-l <loadbalancer-hostname>]"
        return 1
        ;;
    esac
  done

  # environment
  local ARCH
  ARCH=$(uname -m | cut -c1-3)
  local FQDN
  FQDN=$(hostname -f)
  local PURPOSE=${PURPOSE:-"worker"}

  configure_host         # shared host configuration from common.sh

  # Install K3s agent
  echo "⬇️  Downloading and installing K3s agent..."
  curl -sfL https://get.k3s.io | K3S_URL="https://$LB_HOSTNAME:6443" K3S_TOKEN="$K3S_TOKEN" sudo -E sh -s - agent \
    --node-label "environment=production" \
    --node-label "arch=$ARCH" \
    --node-label "purpose=$PURPOSE" \
    || { echo "❌ Failed to install K3s agent. Exiting."; return 1; }

  firewall_configure_agent "K3s"    # shared agent firewall from common.sh

  echo "✅ K3s Agent node bootstrapped."
}

# configure the firewall for a K3s server node (K3s-specific port list)
configure_firewall_k3s_server() {
  local server_ports=(
    "22 SSH server access"
    "6443 K3s API Server"
    "10250 kubelet metrics"
    "2379 etcd client port"
    "2380 etcd peer port"
    "30000:32767 Kubernetes NodePort range"
  )
  firewall_allow_ports "${server_ports[@]}" || return 1
  firewall_enable || return 1
  echo "✅ Firewall rules configured for K3s Server Node."
}

# Required or `RunBashFunction` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi
