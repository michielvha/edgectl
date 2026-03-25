# Script Modularization Plan

## Problem

The 8 bash scripts in `pkg/common/scripts/` contain ~40% duplication between K3s and RKE2 variants. Shared logic (host config, firewall, systemd cleanup, bash env, addon deployment, status checks) is copy-pasted with only distro-specific names changed. This makes maintenance harder and blocks future extensibility (new distros, new firewall backends).

## Current Architecture

### Script Inventory (8 files)

| Script | Functions | Purpose |
|--------|-----------|---------|
| `rke2.sh` | `install_rke2_server()`, `install_rke2_agent()`, `enable_rke2_addons_reloader()`, `configure_rke2_host()`, `configure_rke2_cis()`, `ufw_allow_ports()`, `configure_ufw_rke2_server()`, `configure_ufw_rke2_agent()` | Server/agent install + host config |
| `k3s.sh` | `install_k3s_server()`, `install_k3s_agent()`, `enable_k3s_addons_reloader()`, `configure_k3s_host()`, `ufw_allow_ports()`, `configure_ufw_k3s_server()`, `configure_ufw_k3s_agent()` | Server/agent install + host config |
| `rke2-bash.sh` | `setup_rke2_node_bash_env()`, `setup_kubectl_bash_env()` | Shell environment setup |
| `k3s-bash.sh` | `setup_k3s_node_bash_env()`, `setup_kubectl_bash_env()` | Shell environment setup |
| `rke2-purge.sh` | `rke2_purge()` | Uninstall + systemd cleanup |
| `k3s-purge.sh` | `k3s_purge()` | Uninstall + systemd cleanup |
| `rke2-status.sh` | `rke2_status()` | Service status check |
| `k3s-status.sh` | `k3s_status()` | Service status check |

### How Scripts are Delivered

Scripts are **compiled into the Go binary** via `//go:embed scripts/*.sh` in `pkg/common/embedded.go`. At runtime:

1. `RunBashFunction(scriptName, commandString)` is called from Go
2. `ExtractEmbeddedScript(scriptName)` writes the script to `/tmp/edgectl-{uid}/{scriptName}`
3. The script is executed as `bash /tmp/edgectl-{uid}/{scriptName} functionName [args...]`
4. Each script has a dispatch block at the bottom: `if declare -f "$1" > /dev/null; then "$@"; fi`

**Key constraint:** Only the requested script is extracted. For `common.sh` to work, it must also be extracted to the same temp directory so distro scripts can `source` it.

### Duplication Matrix

| Shared Logic | K3s Location | RKE2 Location | Identical? |
|-------------|--------------|----------------|------------|
| `ufw_allow_ports()` | `k3s.sh` | `rke2.sh` | 100% |
| `configure_*_host()` (swap, br_netfilter, sysctl) | `k3s.sh` | `rke2.sh` | 100% |
| `configure_ufw_*_agent()` (same 3 port groups) | `k3s.sh` | `rke2.sh` | 100% |
| `enable_*_addons_reloader()` | `k3s.sh` | `rke2.sh` | 98% (path only) |
| `setup_kubectl_bash_env()` | `k3s-bash.sh` | `rke2-bash.sh` | 100% |
| Systemd cleanup (daemon-reexec, reload, reset-failed) | `k3s-purge.sh` | `rke2-purge.sh` | 100% |
| Status check pattern (cluster-id + systemctl) | `k3s-status.sh` | `rke2-status.sh` | 95% |
| Script dispatch boilerplate | All 8 files | All 8 files | 100% |

## Proposed Architecture

### New File: `common.sh`

A shared library that all distro scripts source. Contains:

#### 1. OS/Distro Detection

```bash
# Detect Linux distribution family
detect_os() {
  if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo "$ID"       # ubuntu, debian, fedora, rocky, rhel, centos, etc.
  else
    echo "unknown"
  fi
}

# Detect OS family (debian, rhel, etc.)
detect_os_family() {
  if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo "${ID_LIKE:-$ID}"  # "debian" for ubuntu, "rhel fedora" for rocky, etc.
  else
    echo "unknown"
  fi
}
```

#### 2. Firewall Abstraction

```bash
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

# Allow a TCP port with a comment/description
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
      sudo iptables -A INPUT -p tcp --dport "$port" -j ACCEPT \
        || { echo "❌ Failed to allow port $port via iptables"; return 1; }
      ;;
    none)
      echo "⚠️  No supported firewall detected. Skipping port $port."
      ;;
  esac
}

# Allow multiple ports from an array
# Usage: firewall_allow_ports "${ports_array[@]}"
firewall_allow_ports() {
  local ports=("$@")
  for port_info in "${ports[@]}"; do
    local port="${port_info%% *}"
    local comment="${port_info#* }"
    firewall_allow_port "$port" "$comment"
  done
}

# Enable the firewall
firewall_enable() {
  local fw
  fw=$(detect_firewall)

  case "$fw" in
    ufw)
      sudo ufw --force enable || { echo "❌ Failed to enable UFW."; return 1; }
      ;;
    firewalld)
      sudo firewall-cmd --reload || { echo "❌ Failed to reload firewalld."; return 1; }
      sudo systemctl enable --now firewalld
      ;;
    iptables)
      echo "ℹ️  iptables rules applied (no enable needed)."
      ;;
    none)
      echo "⚠️  No firewall to enable."
      ;;
  esac
}

# Configure firewall for a Kubernetes agent node (same ports for all distros)
firewall_configure_agent() {
  local distro="$1"
  local agent_ports=(
    "22 SSH server access"
    "10250 kubelet metrics"
    "30000:32767 Kubernetes NodePort range"
  )
  firewall_allow_ports "${agent_ports[@]}"
  firewall_enable
  echo "✅ Firewall rules configured for $distro Agent Node."
}
```

#### 3. Host Configuration (shared across all K8s distros)

```bash
# Configure host for Kubernetes (swap, kernel modules, sysctl)
configure_host() {
  echo "🔧 Configuring host for Kubernetes..."

  # Disable swap
  if free | awk '/^Swap:/ {exit !$2}'; then
    echo "⚙️  Disabling swap..."
    sudo swapoff -a
    sudo sed -i '/swapfile/s/^/#/' /etc/fstab
  else
    echo "✅ Swap is already disabled."
  fi

  # Load br_netfilter kernel module
  echo "🛠️  Loading br_netfilter kernel module..."
  sudo modprobe br_netfilter || { echo "❌ Failed to load br_netfilter kernel module."; return 1; }
  echo "br_netfilter" | sudo tee /etc/modules-load.d/br_netfilter.conf > /dev/null

  # Apply sysctl settings
  local sysctl_file="/etc/sysctl.d/k8s.conf"
  echo "🛠️  Applying sysctl settings for Kubernetes networking..."
  cat <<EOF | sudo tee "$sysctl_file" > /dev/null
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward = 1
EOF

  sudo sysctl --system > /dev/null && echo "✅ Sysctl settings applied."
  sysctl net.bridge.bridge-nf-call-iptables net.bridge.bridge-nf-call-ip6tables net.ipv4.ip_forward
}
```

#### 4. Systemd Cleanup (shared purge logic)

```bash
# Clean up systemd after uninstall
# Usage: systemd_cleanup <service_path1> [service_path2] ...
systemd_cleanup() {
  local paths=("$@")

  echo "🗑️ Cleaning up leftover systemd service files..."
  for path in "${paths[@]}"; do
    sudo rm -f "$path"
  done

  echo "🔁 Re-executing systemd daemon..."
  sudo systemctl daemon-reexec || { echo "❌ Failed to re-execute systemd daemon."; return 1; }

  echo "🔁 Reloading systemd daemon..."
  sudo systemctl daemon-reload || { echo "❌ Failed to reload systemd daemon."; return 1; }

  echo "🔄 Resetting failed systemd services..."
  sudo systemctl reset-failed || { echo "❌ Failed to reset failed services."; return 1; }
}
```

#### 5. Addon Deployment (shared Helm chart logic)

```bash
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
```

#### 6. Kubectl Bash Environment (shared)

```bash
# Configure kubectl bash environment for remote admin machines
setup_kubectl_bash_env() {
  local profile_file="$HOME/.bashrc"
  local user_home="$HOME/.kube/config"

  sudo touch "$profile_file"
  mkdir -p ~/.kube

  grep -q "export KUBECONFIG=$user_home" "$profile_file" || \
    echo "export KUBECONFIG=$user_home" | sudo tee -a "$profile_file" > /dev/null
  grep -q 'source <(kubectl completion bash)' "$profile_file" || \
    echo "source <(kubectl completion bash)" | sudo tee -a "$profile_file" > /dev/null
  grep -q 'alias k=kubectl' "$profile_file" || \
    echo "alias k=kubectl" | sudo tee -a "$profile_file" > /dev/null

  source "$HOME/.bashrc" && echo "🔧 kubectl environment configured for $(whoami)"
}
```

#### 7. Status Check (shared)

```bash
# Check Kubernetes service status
# Usage: check_status <server_service> <agent_service>
check_status() {
  local server_svc="$1"
  local agent_svc="$2"

  [ -s /etc/edgectl/cluster-id ] && \
    echo "🔑 Cluster ID: $(cat /etc/edgectl/cluster-id)" || \
    echo "❌ Cluster ID not found."

  if systemctl is-active --quiet "$server_svc"; then
    sudo systemctl status "$server_svc"
  elif systemctl is-active --quiet "$agent_svc"; then
    sudo systemctl status "$agent_svc"
  else
    echo "Neither $server_svc nor $agent_svc are running."
  fi
}
```

### Refactored Distro Scripts

After extracting shared logic to `common.sh`, each distro script becomes a thin wrapper.

#### `k3s.sh` (after refactoring)

Only contains:
- `install_k3s_server()` — K3s-specific install flags (`--flannel-backend=none`, `--disable-kube-proxy`, etc.), Cilium HelmChart manifest, calls `configure_host`, `firewall_allow_ports`, `enable_addon_reloader` from `common.sh`
- `install_k3s_agent()` — K3s agent install with `K3S_TOKEN`/`K3S_URL`, calls `configure_host`, `firewall_configure_agent` from `common.sh`
- `configure_firewall_k3s_server()` — K3s-specific port list (no 9345), calls `firewall_allow_ports` + `firewall_enable` from `common.sh`

#### `rke2.sh` (after refactoring)

Only contains:
- `install_rke2_server()` — RKE2-specific config.yaml, systemctl enable, Cilium HelmChartConfig, calls shared functions from `common.sh`
- `install_rke2_agent()` — RKE2 agent install with config.yaml approach, calls shared functions
- `configure_rke2_cis()` — CIS hardening (RKE2-only, stays here)
- `configure_firewall_rke2_server()` — RKE2-specific port list (includes 9345, 2381), calls `firewall_allow_ports` + `firewall_enable`

#### `k3s-purge.sh` / `rke2-purge.sh` (after refactoring)

Each becomes ~10 lines: locate uninstall script, run it, call `systemd_cleanup` from `common.sh`.

#### `k3s-status.sh` / `rke2-status.sh` (after refactoring)

Each becomes a one-liner calling `check_status "k3s" "k3s-agent"` or `check_status "rke2-server" "rke2-agent"`.

#### `k3s-bash.sh` / `rke2-bash.sh` (after refactoring)

Each keeps only the node-specific function (`setup_k3s_node_bash_env` / `setup_rke2_node_bash_env`). The shared `setup_kubectl_bash_env()` moves to `common.sh`.

### Go Changes: `pkg/common/embedded.go`

`RunBashFunction` needs to ensure `common.sh` is always available in the temp directory:

```go
func RunBashFunction(scriptName, commandString string) {
    scriptPath := ExtractEmbeddedScript(scriptName)

    // Always extract common.sh alongside the target script
    // so distro scripts can source it
    if scriptName != "common.sh" {
        ExtractEmbeddedScript("common.sh")
    }

    // ... rest unchanged
}
```

Each distro script sources it at the top:

```bash
#!/bin/bash
# Source shared functions from the same temp directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"
```

This works because `ExtractEmbeddedScript` writes all scripts to the same `/tmp/edgectl-{uid}/` directory.

## File Changes Summary

| Action | File | Description |
|--------|------|-------------|
| **Create** | `pkg/common/scripts/common.sh` | Shared functions: host config, firewall, systemd, kubectl env, status, reloader |
| **Modify** | `pkg/common/embedded.go` | Extract `common.sh` alongside every script |
| **Modify** | `pkg/common/scripts/k3s.sh` | Remove duplicated functions, source `common.sh` |
| **Modify** | `pkg/common/scripts/rke2.sh` | Remove duplicated functions, source `common.sh` |
| **Modify** | `pkg/common/scripts/k3s-bash.sh` | Remove `setup_kubectl_bash_env()`, source `common.sh` |
| **Modify** | `pkg/common/scripts/rke2-bash.sh` | Remove `setup_kubectl_bash_env()`, source `common.sh` |
| **Modify** | `pkg/common/scripts/k3s-purge.sh` | Use `systemd_cleanup()` from `common.sh` |
| **Modify** | `pkg/common/scripts/rke2-purge.sh` | Use `systemd_cleanup()` from `common.sh` |
| **Modify** | `pkg/common/scripts/k3s-status.sh` | Use `check_status()` from `common.sh` |
| **Modify** | `pkg/common/scripts/rke2-status.sh` | Use `check_status()` from `common.sh` |

**Total: 1 new file, 9 modified files**

## What Stays Distro-Specific

| Logic | Why it can't be shared |
|-------|----------------------|
| Install script URL (`get.k3s.io` vs `get.rke2.io`) | Different installers |
| Install approach (CLI args vs config.yaml) | K3s uses flags, RKE2 uses config file |
| CIS hardening (`configure_rke2_cis`) | RKE2-only feature |
| Cilium config (HelmChart vs HelmChartConfig) | Different K8s API types |
| Tailscale TLS-SAN integration | Currently RKE2-only |
| Supervisor port 9345 in server firewall rules | RKE2-only port |
| `systemctl enable --now` | RKE2 needs explicit enable; K3s installer does it |
| Node bash env (profile path, PATH additions) | RKE2 needs PATH for `/var/lib/rancher/rke2/bin`; K3s doesn't |

## Estimated Line Reduction

| Script | Current Lines | After Refactor | Saved |
|--------|-------------|----------------|-------|
| `common.sh` | 0 (new) | ~150 | - |
| `k3s.sh` | ~260 | ~120 | ~140 |
| `rke2.sh` | ~420 | ~230 | ~190 |
| `k3s-bash.sh` | ~48 | ~25 | ~23 |
| `rke2-bash.sh` | ~50 | ~27 | ~23 |
| `k3s-purge.sh` | ~56 | ~20 | ~36 |
| `rke2-purge.sh` | ~56 | ~20 | ~36 |
| `k3s-status.sh` | ~25 | ~12 | ~13 |
| `rke2-status.sh` | ~28 | ~12 | ~16 |
| **Total** | **~943** | **~616** | **~327 (35%)** |

## Implementation Order

1. Create `common.sh` with all shared functions
2. Modify `embedded.go` to always extract `common.sh`
3. Refactor small scripts first (status, purge, bash) — easy wins, low risk
4. Refactor `k3s.sh` and `rke2.sh` — largest change, needs careful testing
5. Test all commands end-to-end

## Future Extensibility

This architecture makes it trivial to add:
- **New K8s distros** (kubeadm, microk8s) — create `kubeadm.sh`, source `common.sh`, only add distro-specific install logic
- **New firewall backends** — add a case to `firewall_allow_port()` and `firewall_enable()` in `common.sh`
- **New Linux distros** — `detect_os()` already reads `/etc/os-release`; package manager abstraction could be added to `common.sh` if needed (e.g., `apt` vs `dnf` vs `yum`)
- **New addons** — add more `enable_addon_*()` functions to `common.sh`
