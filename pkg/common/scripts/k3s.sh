#!/bin/bash
# K3s module for K3s installation and configuration
# purpose: bootstrap K3s nodes.
# usage: quickly source this module with the following command:
# ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/k3s.sh) `
# ------------------------------------------------------------------------------------------------------------------------------------------------

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

  configure_k3s_host   # perform default bootstrap configurations required on each K3s node.

  # Install K3s
  echo "⬇️  Downloading and installing K3s..."
  curl -sfL https://get.k3s.io | sudo sh -s - server \
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

  enable_k3s_addons_reloader   # enabling the reloader addon by default

  configure_ufw_k3s_server     # Configure UFW for K3s server

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

  configure_k3s_host         # perform common bootstrap configurations.

  # Install K3s agent
  echo "⬇️  Downloading and installing K3s agent..."
  curl -sfL https://get.k3s.io | K3S_URL="https://$LB_HOSTNAME:6443" K3S_TOKEN="$K3S_TOKEN" sudo -E sh -s - agent \
    --node-label "environment=production" \
    --node-label "arch=$ARCH" \
    --node-label "purpose=$PURPOSE" \
    || { echo "❌ Failed to install K3s agent. Exiting."; return 1; }

  configure_ufw_k3s_agent    # Configure UFW for K3s agent

  echo "✅ K3s Agent node bootstrapped."
}

enable_k3s_addons_reloader(){
  echo "📦 Enabling K3s Addon: Stakater's Reloader"
  cat <<EOF | sudo tee /var/lib/rancher/k3s/server/manifests/reloader.yaml
# Reference: https://docs.k3s.io/helm
apiVersion: helm.cattle.io/v1
kind: HelmChart
metadata:
  name: reloader
  namespace: kube-system
spec:
  chart: reloader
  repo: https://stakater.github.io/stakater-charts
  targetNamespace: kube-system
  valuesContent: |-
    reloader:
      autoReloadAll: true
EOF
}

# perform default bootstrap configurations required on each K3s node.
configure_k3s_host() {
  echo "🔧 Default K3s Node Config..."

  # Disable swap if not already disabled
  if free | awk '/^Swap:/ {exit !$2}'; then
    echo "⚙️  Disabling swap..."
    sudo swapoff -a
    sudo sed -i '/swapfile/s/^/#/' /etc/fstab
  else
    echo "✅ Swap is already disabled."
  fi

  local sysctl_file="/etc/sysctl.d/k8s.conf"

  # Load br_netfilter kernel module
  echo "🛠️  Loading br_netfilter kernel module..."
  sudo modprobe br_netfilter || { echo "❌ Failed to load br_netfilter kernel module. Exiting."; return 1; }
  lsmod | grep br_netfilter
  echo "br_netfilter" | sudo tee /etc/modules-load.d/br_netfilter.conf

  echo "🛠️  Applying sysctl settings for Kubernetes networking..."
  cat <<EOF | sudo tee "$sysctl_file" > /dev/null
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward = 1
EOF

  sudo sysctl --system > /dev/null && echo "✅ Sysctl settings applied successfully."

  sysctl net.bridge.bridge-nf-call-iptables net.bridge.bridge-nf-call-ip6tables net.ipv4.ip_forward
}

# Function: ufw_allow_ports - Helper function to configure UFW rules
ufw_allow_ports() {
  local ports=("$@")
  for port_info in "${ports[@]}"; do
    local port="${port_info%% *}"
    local comment="${port_info#* }"
    sudo ufw allow proto tcp from any to any port "$port" comment "$comment" || { echo "❌ Failed to create rule for $port"; return 1;}
  done
}

# configure the firewall for a K3s server node
configure_ufw_k3s_server() {
  local server_ports=(
    "22 SSH server access"
    "6443 K3s API Server"
    "10250 kubelet metrics"
    "2379 etcd client port"
    "2380 etcd peer port"
    "30000:32767 Kubernetes NodePort range"
  )
  ufw_allow_ports "${server_ports[@]}"

  sudo ufw enable || { echo "❌ Failed to enable UFW."; return 1; }
  echo "✅ UFW rules configured for K3s Server Node."
}

# configure the firewall for a K3s agent node
configure_ufw_k3s_agent() {
  local agent_ports=(
    "22 SSH server access"
    "10250 kubelet metrics"
    "30000:32767 Kubernetes NodePort range"
  )
  ufw_allow_ports "${agent_ports[@]}"

  sudo ufw enable || { echo "❌ Failed to enable UFW."; return 1; }
  echo "✅ UFW rules configured for K3s Agent Node."
}

# Required or `RunBashFunction` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi
