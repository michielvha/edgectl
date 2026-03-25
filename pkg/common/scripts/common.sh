#!/bin/bash
# common.sh - Shared functions for all Kubernetes distributions
# This script is sourced by distro-specific scripts (k3s.sh, rke2.sh, etc.)
# It provides OS detection, firewall abstraction, host configuration, and other shared utilities.
# ------------------------------------------------------------------------------------------------------------------------------------------------

# ============================================================
# OS & Environment Detection
# ============================================================

# Detect Linux distribution ID (e.g., ubuntu, debian, fedora, rocky, rhel)
detect_os() {
  if [ -f /etc/os-release ]; then
    # shellcheck source=/dev/null
    . /etc/os-release
    echo "$ID"
  else
    echo "unknown"
  fi
}

# Detect OS family (e.g., debian, rhel)
detect_os_family() {
  if [ -f /etc/os-release ]; then
    # shellcheck source=/dev/null
    . /etc/os-release
    echo "${ID_LIKE:-$ID}"
  else
    echo "unknown"
  fi
}

# ============================================================
# Firewall Abstraction
# ============================================================

# Detect available firewall tool
detect_firewall() {
  if command -v ufw &>/dev/null; then
    echo "ufw"
  elif command -v firewall-cmd &>/dev/null; then
    echo "firewalld"
  elif command -v iptables &>/dev/null; then
    echo "iptables"
  else
    echo "none"
  fi
}

# Allow a single TCP port with a comment
# Usage: firewall_allow_port <port> <comment>
firewall_allow_port() {
  local port="$1"
  local comment="$2"
  local fw
  fw=$(detect_firewall)

  case "$fw" in
    ufw)
      sudo ufw allow proto tcp from any to any port "$port" comment "$comment" \
        || { echo "❌ Failed to allow port $port via UFW"; return 1; }
      ;;
    firewalld)
      sudo firewall-cmd --permanent --add-port="${port}/tcp" \
        || { echo "❌ Failed to allow port $port via firewalld"; return 1; }
      ;;
    iptables)
      # Handle port ranges for iptables (uses : instead of - for range)
      local iptables_port="${port/-/:}"
      if [[ "$iptables_port" == *":"* ]]; then
        sudo iptables -A INPUT -p tcp --dport "$iptables_port" -m multiport -j ACCEPT \
          || { echo "❌ Failed to allow port $port via iptables"; return 1; }
      else
        sudo iptables -A INPUT -p tcp --dport "$iptables_port" -j ACCEPT \
          || { echo "❌ Failed to allow port $port via iptables"; return 1; }
      fi
      ;;
    none)
      echo "⚠️  No supported firewall detected. Skipping port $port ($comment)."
      ;;
  esac
}

# Allow multiple ports from an array of "port comment" strings
# Usage: firewall_allow_ports "${ports_array[@]}"
firewall_allow_ports() {
  local ports=("$@")
  for port_info in "${ports[@]}"; do
    local port="${port_info%% *}"
    local comment="${port_info#* }"
    firewall_allow_port "$port" "$comment" || return 1
  done
}

# Enable/reload the firewall
firewall_enable() {
  local fw
  fw=$(detect_firewall)

  case "$fw" in
    ufw)
      sudo ufw --force enable || { echo "❌ Failed to enable UFW."; return 1; }
      ;;
    firewalld)
      sudo firewall-cmd --reload || { echo "❌ Failed to reload firewalld."; return 1; }
      ;;
    iptables)
      echo "ℹ️  iptables rules applied (no enable step needed)."
      ;;
    none)
      echo "⚠️  No firewall to enable."
      ;;
  esac
}

# Configure firewall for a Kubernetes agent node (same ports for all distros)
# Usage: firewall_configure_agent <distro_name>
firewall_configure_agent() {
  local distro="$1"
  local agent_ports=(
    "22 SSH server access"
    "10250 kubelet metrics"
    "30000:32767 Kubernetes NodePort range"
  )
  firewall_allow_ports "${agent_ports[@]}" || return 1
  firewall_enable || return 1
  echo "✅ Firewall rules configured for $distro Agent Node."
}

# ============================================================
# Host Configuration (shared across all K8s distros)
# ============================================================

# Configure host for Kubernetes: disable swap, load kernel modules, apply sysctl
configure_host() {
  echo "🔧 Configuring host for Kubernetes..."

  # Disable swap if not already disabled
  if free | awk '/^Swap:/ {exit !$2}'; then
    echo "⚙️  Disabling swap..."
    sudo swapoff -a
    sudo sed -i '/swapfile/s/^/#/' /etc/fstab
  else
    echo "✅ Swap is already disabled."
  fi

  # Load br_netfilter kernel module
  echo "🛠️  Loading br_netfilter kernel module..."
  sudo modprobe br_netfilter || { echo "❌ Failed to load br_netfilter kernel module. Exiting."; return 1; }
  echo "br_netfilter" | sudo tee /etc/modules-load.d/br_netfilter.conf > /dev/null

  # Apply sysctl settings for Kubernetes networking
  local sysctl_file="/etc/sysctl.d/k8s.conf"
  echo "🛠️  Applying sysctl settings for Kubernetes networking..."
  cat <<EOF | sudo tee "$sysctl_file" > /dev/null
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward = 1
EOF

  sudo sysctl --system > /dev/null && echo "✅ Sysctl settings applied successfully."
  sysctl net.bridge.bridge-nf-call-iptables net.bridge.bridge-nf-call-ip6tables net.ipv4.ip_forward
}

# ============================================================
# Addon Deployment
# ============================================================

# Deploy Stakater Reloader addon via HelmChart manifest
# Usage: enable_addon_reloader <manifest_dir>
enable_addon_reloader() {
  local manifest_dir="$1"

  echo "📦 Enabling Addon: Stakater's Reloader"
  sudo mkdir -p "$manifest_dir"
  cat <<EOF | sudo tee "${manifest_dir}/reloader.yaml"
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

# ============================================================
# Kubectl Bash Environment
# ============================================================

# Configure kubectl bash environment for remote admin machines
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

# ============================================================
# Systemd Cleanup (used by purge scripts)
# ============================================================

# Clean up systemd state after uninstalling a Kubernetes distribution
# Usage: systemd_cleanup <service_file_path1> [service_file_path2] ...
systemd_cleanup() {
  local paths=("$@")

  echo "🗑️ Cleaning up leftover systemd service files..."
  for path in "${paths[@]}"; do
    sudo rm -f "$path"
  done

  echo "🔁 Re-executing systemd daemon..."
  if ! sudo systemctl daemon-reexec; then
    echo "❌ Failed to re-execute systemd daemon."
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
}

# ============================================================
# Status Check
# ============================================================

# Check Kubernetes service status
# Usage: check_status <server_service> <agent_service>
check_status() {
  local server_svc="$1"
  local agent_svc="$2"

  # Output cluster-id
  [ -s /etc/edgectl/cluster-id ] && echo "🔑 Cluster ID: $(cat /etc/edgectl/cluster-id)" || echo "❌ Cluster ID not found. Please check if the cluster is initialized."

  # Check the status of Kubernetes services
  if systemctl is-active --quiet "$server_svc"; then
    sudo systemctl status "$server_svc"
  elif systemctl is-active --quiet "$agent_svc"; then
    sudo systemctl status "$agent_svc"
  else
    echo "Neither $server_svc nor $agent_svc are running."
  fi
}
