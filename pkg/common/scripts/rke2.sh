#!/bin/bash
# RKE2 module for RKE2 installation and configuration
# purpose: bootstrap RKE2 nodes.
# usage: quickly source this module with the following command:
# ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/rke2.sh) `
# ------------------------------------------------------------------------------------------------------------------------------------------------

# Source shared functions
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# TODO: add logic if already installed, skip installation and proceed with configuration. or provide some kind of update functionality. We could check for the existance of these folders /etc/rancher /var/lib/kubelet /var/lib/etcd
# TODO: Look into harding the RKE2 installation with CIS benchmarks. SEL linux etc etc. Verify with [kube-bench](https://github.com/aquasecurity/kube-bench)
# Hardening Guide created in edge cloud repo: edge-cloud/docs/setup/software/kubernetes/rke2/hardening/readme.md. For ubuntu we'll have to manually create the profiles.
# TODO: Add support for Fedora based systems.
# TODO: Refactor tailscale management plane into GO CLI so it can be passed to the script. Issue #29
# TODO: we should write purpose (agent/server) env var to a file so we can check if the host is a worker or server node and based on that apply appropriate cis config.

# bootstrap a RKE2 server node
install_rke2_server() {
  # usage: install_rke2_server [-l <loadbalancer-hostname>]

  # Pre checks
  systemctl list-unit-files | grep -q "^rke2-server.service" && {
    echo "❌ RKE2 Server service already exists. Use 'edgectl rke2 system purge' Exiting."
    return 1
  }

  echo "📦 Configuring RKE2 Server Node..."

  # Default parameter values
  local LB_HOSTNAME="loadbalancer.example.com"

  # Parse options using getopts
  while getopts "l:" opt; do
    case "$opt" in
      l) LB_HOSTNAME="$OPTARG" ;;
      \?)
        echo "❌ Invalid option: -$OPTARG"
        echo "Usage: install_rke2_server [-l <loadbalancer-hostname>]"
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
  HOST=$(hostname -s) # hostname without domain
  local TS_DOMAIN="tail6948f.ts.net" # TODO: this should be set in the environment or passed as a parameter.
  local TS="$HOST.$TS_DOMAIN" # get tailscale domain for internal management interface, will be needed to add to SAN.
  local PURPOSE=${PURPOSE:-"server"}

  configure_host   # shared host configuration from common.sh

 # Install RKE2
  echo "⬇️  Downloading and installing RKE2..."
  curl -sfL https://get.rke2.io | sudo sh - || { echo "❌ Failed to download RKE2. Exiting."; return 1; }

  # Write configuration to /etc/rancher/rke2/config.yaml
  # https://docs.rke2.io/reference/server_config
  sudo mkdir -p /etc/rancher/rke2 && cat <<EOF | sudo tee /etc/rancher/rke2/config.yaml
write-kubeconfig-mode: "0644"
profile: "cis"
node-label:
  - "environment=production"
  - "arch=$ARCH"
  - "purpose=$PURPOSE"

cni: cilium
disable-kube-proxy: true    # Disable kube-proxy (since eBPF replaces it)
disable-cloud-controller: true # disable cloud controller since we are onprem.

tls-san: ["$FQDN", "$LB_HOSTNAME", "$TS"]
EOF

  # TODO: Decide to use long or shorthand syntax, check again if we cannot just add this above in the config.yaml, had some issues with it before but might not have been related to the way we create the config file.
  # Add token and server IP to config if they are set as environment variables - for secondary server installations.
  [ -n "$RKE2_TOKEN" ] && echo "token: \"$RKE2_TOKEN\"" | sudo tee -a /etc/rancher/rke2/config.yaml && echo "🔑 Added token to config"

  if [ -n "$RKE2_SERVER_IP" ]; then
    echo "server: \"https://$RKE2_SERVER_IP:9345\"" | sudo tee -a /etc/rancher/rke2/config.yaml
    echo "🌐 Added server URL to config: $RKE2_SERVER_IP"
  fi

  # TODO: we should make cilium the default but provide a fallback. and then use kube-proxy config else skip it probably wrap this in it's own function.
  echo "🛠️  Writing Cilium Helm Chart Config..."
  sudo mkdir -p /var/lib/rancher/rke2/server/manifests/ && cat <<EOF | sudo tee /var/lib/rancher/rke2/server/manifests/rke2-cilium-config.yaml
apiVersion: helm.cattle.io/v1
kind: HelmChartConfig
metadata:
  name: rke2-cilium
  namespace: kube-system
spec:
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

  enable_addon_reloader "/var/lib/rancher/rke2/server/manifests/"   # shared from common.sh
  
  configure_rke2_cis            # Hardening RKE2 with CIS benchmarks (RKE2-specific)

  configure_firewall_rke2_server     # RKE2-specific server firewall rules

  echo "⚙️  Enabling RKE2 server..."
  sudo systemctl enable --now rke2-server || { echo "❌ RKE2 Server node bootstrap failed."; return 1; }
  echo "✅ RKE2 Server node bootstrapped."
}

# bootstrap a RKE2 agent node
install_rke2_agent() {
  # usage: install_rke2_agent [-l <loadbalancer-hostname>]
  # Pre checks
  systemctl list-unit-files | grep -q "^rke2-agent.service" && {
    echo "❌ RKE2 Agent service already exists. Exiting."
    return 1
  }

  # Check for token, this check could be removed or moved up into the Go wrapper
  [ -z "$RKE2_TOKEN" ] && {
    echo "❌ RKE2_TOKEN environment variable not set. Token is required."
    return 1
  } || echo "🔑 Using RKE2_TOKEN from environment variable"

  echo "📦 Configuring RKE2 Agent Node..."

  # Default parameter values
  local LB_HOSTNAME="loadbalancer.example.com"

  # Parse options using getopts
  while getopts "l:" opt; do
    case "$opt" in
      l) LB_HOSTNAME="$OPTARG" ;;
      \?)
        echo "❌ Invalid option: -$OPTARG"
        echo "Usage: install_rke2_agent [-l <loadbalancer-hostname>]"
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
  HOST=$(hostname -s) # hostname without domain
  local TS="$HOST.$TS_DOMAIN" # get tailscale domain for internal management interface, will be needed to add to SAN.
  local PURPOSE=${PURPOSE:-"worker"}

  configure_host         # shared host configuration from common.sh

  # Install RKE2
  echo "⬇️  Downloading and installing RKE2..."
  curl -sfL https://get.rke2.io | sudo sh - || { echo "❌ Failed to download RKE2. Exiting."; return 1; }

  # Write configuration to /etc/rancher/rke2/config.yaml
  # https://docs.rke2.io/reference/linux_agent_config
  sudo mkdir -p /etc/rancher/rke2 && cat <<EOF | sudo tee /etc/rancher/rke2/config.yaml
server: "https://$LB_HOSTNAME:9345"
token: $RKE2_TOKEN
profile: "cis"
node-label:
  - "environment=production"
  - "arch=$ARCH"
  - "purpose=$PURPOSE"
tls-san: ["$FQDN", "$LB_HOSTNAME", "$TS"]
EOF

  configure_rke2_cis          # Hardening RKE2 with CIS benchmarks (RKE2-specific)

  firewall_configure_agent "RKE2"    # shared agent firewall from common.sh

  # Enable and start RKE2 agent
  echo "⚙️  Enabling RKE2 agent..."
  sudo systemctl enable --now rke2-agent || { echo "❌ RKE2 Agent node bootstrap failed."; return 1; }
  echo "✅ RKE2 Agent node bootstrapped."
}

# RKE2-specific: CIS hardening
configure_rke2_cis() {
  # https://docs.rke2.io/security/hardening_guide/#kernel-parameters
  local cis_sysctl="/usr/local/share/rke2/rke2-cis-sysctl.conf"
  if [ -f "$cis_sysctl" ]; then
    echo "🔐 Applying CIS sysctl settings..."
    sudo cp -f "$cis_sysctl" /etc/sysctl.d/60-rke2-cis.conf
    sudo systemctl restart systemd-sysctl
    echo "✅ CIS sysctl settings applied."
  else
    echo "⚠️  CIS sysctl config not found at $cis_sysctl. Skipping."
  fi

  # Check if etcd user and group exist, if not create them
  # TODO: This should only be done on server nodes, not agent nodes.
  getent group etcd >/dev/null || sudo groupadd --system etcd
  id -u etcd >/dev/null 2>&1 || sudo useradd --system --no-create-home --shell /sbin/nologin --gid etcd etcd
}

# RKE2-specific: server firewall rules (includes supervisor port 9345)
configure_firewall_rke2_server() {
  local server_ports=(
    "22 SSH server access"
    "6443 RKE2 API Server"
    "9345 RKE2 Supervisor API"
    "10250 kubelet metrics"
    "2379 etcd client port"
    "2380 etcd peer port"
    "2381 etcd metrics port"
    "30000:32767 Kubernetes NodePort range"
  )
  firewall_allow_ports "${server_ports[@]}" || return 1
  firewall_enable || return 1
  echo "✅ Firewall rules configured for RKE2 Server Node."
}

# Required or `RunBashFunction` will not be able to call the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi