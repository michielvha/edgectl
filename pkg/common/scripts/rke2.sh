#!/bin/bash
# RKE2 module for RKE2 installation and configuration
# purpose: bootstrap RKE2 nodes.
# usage: quickly source this module with the following command:
# ` source <(curl -fsSL https://raw.githubusercontent.com/michielvha/edgectl/main/pkg/common/scripts/rke2.sh) `
# ------------------------------------------------------------------------------------------------------------------------------------------------

# TODO: add logic if already installed, skip installation and proceed with configuration. or provide some kind of update functionality. We could check for the existance of these folders /etc/rancher /var/lib/kubelet /var/lib/etcd
# TODO: Look into harding the RKE2 installation with CIS benchmarks. SEL linux etc etc. Verify with [kube-bench](https://github.com/aquasecurity/kube-bench)
# Hardening Guide created in edge cloud repo: edge-cloud/docs/setup/software/kubernetes/rke2/hardening/readme.md. For ubuntu we'll have to manually create the profiles.
# code snippets added but currently failing, check what's going wrong.
# TODO: Add support for Fedora based systems.
# TODO: Refactor tailscale management plane into GO CLI so i can be passed to the script.
# bootstrap a RKE2 server node
install_rke2_server() {
  # usage: install_rke2_server [-l <loadbalancer-hostname>]

  # Pre checks
  if systemctl list-unit-files | grep -q "^rke2-server.service"; then
    echo "❌ RKE2 Server service already exists. Exiting."
    return 1
  fi
  # TODO: Check for ``/etc/rancher/rke2`` and ``/var/lib/kubelet`` and ``/var/lib/etcd`` folders to see if RKE2 is already installed. If so recommend to run rke2_status or purge_rke2.

  echo "🚀 Configuring RKE2 Server Node..."

  # Default parameter values
  local LB_HOSTNAME="loadbalancer.example.com"

  # Parse options using getopts
  while getopts "l:" opt; do
    case "$opt" in
      l) LB_HOSTNAME="$OPTARG" ;;  # -l <loadbalancer-hostname>
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
  local TS="$HOST.tail6948f.ts.net" # get tailscale domain for internal management interface, will be needed to add to SAN.

  # perform default bootstrap configurations required on each RKE2 node.
  configure_rke2_host

 # Install RKE2
  echo "⬇️  Downloading and installing RKE2..."
  curl -sfL https://get.rke2.io | sudo sh - || { echo "❌ Failed to download RKE2. Exiting."; return 1; }

  # Ensure the config directory exists
  sudo mkdir -p /etc/rancher/rke2

  # Write configuration to /etc/rancher/rke2/config.yaml
  # https://docs.rke2.io/reference/server_config
  cat <<EOF | sudo tee /etc/rancher/rke2/config.yaml
write-kubeconfig-mode: "0644"
profile: "cis"
node-label:
  - "environment=production"
  - "arch=${ARCH}"
  - "purpose=system"

cni: cilium
disable-kube-proxy: true    # Disable kube-proxy (since eBPF replaces it)
disable-cloud-controller: true # disable cloud controller since we are onprem.

tls-san: ["$FQDN", "$LB_HOSTNAME", "$TS"]  

# reference: https://docs.rke2.io/reference/server_config#listener

# node-ip: 192.168.1.241 # we should not have to hardcode this, change tailscale from hostname and use internal dns.

EOF

  # Add token and server IP to config if they are set as environment variables - for secondary server installations.
  if [ -n "$RKE2_TOKEN" ]; then
    echo "token: \"$RKE2_TOKEN\"" | sudo tee -a /etc/rancher/rke2/config.yaml
    echo "🔑 Added token to config"
  fi

  if [ -n "$RKE2_SERVER_IP" ]; then
    echo "server: \"https://$RKE2_SERVER_IP:9345\"" | sudo tee -a /etc/rancher/rke2/config.yaml
    echo "🌐 Added server URL to config: $RKE2_SERVER_IP"
  fi

  # Cilium debug
  # check if bpf is enabled
  # bpftool feature  | zgrep CONFIG_BPF /proc/config.gz if available.


  sudo mkdir -p /var/lib/rancher/rke2/server/manifests/
  cat <<EOF | sudo tee /var/lib/rancher/rke2/server/manifests/rke2-cilium-config.yaml
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

  # Hardening RKE2 with CIS benchmarks
  configure_rke2_cis

  # Configure UFW for RKE2 server
  configure_ufw_rke2_server

  # Enable and start RKE2 server
  echo "⚙️  Starting RKE2 server..."
  sudo systemctl enable --now rke2-server || { echo "❌ RKE2 Server node bootstrap failed."; return 1; }
  echo "✅ RKE2 Server node bootstrapped."
}

# bootstrap a RKE2 agent node
install_rke2_agent() {
  # usage: install_rke2_agent [-l <loadbalancer-hostname>]
  # Pre checks
  if systemctl list-unit-files | grep -q "^rke2-agent.service"; then
    echo "❌ RKE2 Server service already exists. Exiting."
    return 1
  fi

  # Check if token is available via environment variable
  if [ -n "$RKE2_TOKEN" ]; then
    echo "🔑 Using RKE2_TOKEN from environment variable"
  else
    echo "❌ RKE2_TOKEN environment variable not set. Token is required."
    return 1
  fi

  echo "🚀 Configuring RKE2 Agent Node..."

  # Default parameter values
  local LB_HOSTNAME="loadbalancer.example.com"

  # Parse options using getopts
  while getopts "l:" opt; do
    case "$opt" in
      l) LB_HOSTNAME="$OPTARG" ;;  # -l <loadbalancer-hostname>
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
  local TS="$HOST.tail6948f.ts.net" # get tailscale domain for internal management interface, will be needed to add to SAN.
  # Default purpose for agent nodes if not set
  local PURPOSE=${PURPOSE:-"worker"}

  # perform default bootstrap configurations required on each RKE2 node.
  configure_rke2_host

  # Install RKE2
  echo "⬇️  Downloading and installing RKE2..."
  curl -sfL https://get.rke2.io | sudo sh - || { echo "❌ Failed to download RKE2. Exiting."; return 1; }

  # Ensure the config directory exists
  sudo mkdir -p /etc/rancher/rke2

  # Write configuration to /etc/rancher/rke2/config.yaml
  # https://docs.rke2.io/reference/linux_agent_config
  cat <<EOF | sudo tee /etc/rancher/rke2/config.yaml
server: "https://$LB_HOSTNAME:9345"
token: $RKE2_TOKEN
profile: "cis"
node-label:
  - "environment=production"
  - "arch=${ARCH}"
  - "purpose=$PURPOSE"
tls-san: ["$FQDN", "$LB_HOSTNAME", "$TS"]
EOF

  # Hardening RKE2 with CIS benchmarks
  configure_rke2_cis

  # Enable and start RKE2 agent
  echo "⚙️  Starting RKE2 agent..."
  sudo systemctl enable --now rke2-agent || { echo "❌ RKE2 Agent node bootstrap failed."; return 1; }

  configure_ufw_rke2_agent

  echo "✅ RKE2 Agent node bootstrapped."
}



# perform default bootstrap configurations required on each RKE2 node.
configure_rke2_host() {
  # TODO: maybe write to a file after config to check if already configured. When running another command that calls this command.
  echo "🚀 Default RKE2 Node Config..."

  # Disable swap if not already disabled
  if free | awk '/^Swap:/ {exit !$2}'; then
    echo "⚙️  Disabling swap..."
    sudo swapoff -a
    sudo sed -i '/swapfile/s/^/#/' /etc/fstab
  else
    echo "✅ Swap is already disabled."
  fi

  # TODO: add check if cilium ebpf is enabled, this config is only needed in kube-proxy mode.
  local sysctl_file="/etc/sysctl.d/k8s.conf"

  # Load br_netfilter kernel module
  echo "🛠️  Loading br_netfilter kernel module..."
  sudo modprobe br_netfilter || { echo "❌ Failed to load br_netfilter kernel module. Exiting."; return 1; }
  lsmod | grep br_netfilter
  # make the config persistent
  echo "br_netfilter" | sudo tee /etc/modules-load.d/br_netfilter.conf

  echo "🛠️  Applying sysctl settings for Kubernetes (kube-proxy) networking..."
  cat <<EOF | sudo tee "$sysctl_file" > /dev/null
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward = 1
EOF

  sudo sysctl --system > /dev/null && echo "✅ Sysctl settings applied successfully."

  sysctl net.bridge.bridge-nf-call-iptables net.bridge.bridge-nf-call-ip6tables net.ipv4.ip_forward
}

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
  getent group etcd >/dev/null || sudo groupadd --system etcd
  id -u etcd >/dev/null 2>&1 || sudo useradd --system --no-create-home --shell /sbin/nologin --gid etcd etcd
}

# configure the firewall for a RKE2 server node
configure_ufw_rke2_server() {
  # Allow ssh access (22) for administration
  sudo ufw allow proto tcp from any to any port 22 comment "SSH server access"

  # Allow Kubernetes API (6443) from agent nodes
  sudo ufw allow proto tcp from any to any port 6443 comment "RKE2 API Server"

  # Allow RKE2 supervisor API (9345) from agent nodes
  sudo ufw allow proto tcp from any to any port 9345 comment "RKE2 Supervisor API"

  # Allow kubelet metrics (10250) from all nodes
  sudo ufw allow proto tcp from any to any port 10250 comment "kubelet metrics"

  # Allow etcd client port (2379) between RKE2 server nodes
  sudo ufw allow proto tcp from any to any port 2379 comment "etcd client port"

  # Allow etcd peer port (2380) between RKE2 server nodes
  sudo ufw allow proto tcp from any to any port 2380 comment "etcd peer port"

  # Allow etcd metrics port (2381) between RKE2 server nodes
  sudo ufw allow proto tcp from any to any port 2381 comment "etcd metrics port"

  # Allow NodePort range (30000-32767) between all nodes
  sudo ufw allow proto tcp from any to any port 30000:32767 comment "Kubernetes NodePort range"

  echo "✅ UFW rules configured for RKE2 Server Node."
  # TODO: enable ufw with ``sudo ufw enable`` wait until config is refined and add port 22.
 }

# configure the firewall for a RKE2 agent node
configure_ufw_rke2_agent() {
  # Allow ssh access (22) for administration
  sudo ufw allow proto tcp from any to any port 22 comment "SSH server access"

  # Allow kubelet metrics (10250) from all nodes
  sudo ufw allow proto tcp from any to any port 10250 comment "kubelet metrics"

  # Allow NodePort range (30000-32767) between all nodes
  sudo ufw allow proto tcp from any to any port 30000:32767 comment "Kubernetes NodePort range"

  echo "✅ UFW rules configured for RKE2 Agent Node."
}

# Dispatcher: allows calling the function by name
if declare -f "$1" > /dev/null; then
  "$@"
else
  echo "❌ Unknown function: $1"
  exit 1
fi