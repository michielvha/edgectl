# K3s Cluster Management

## Overview

EdgeCTL supports automated deployment of **K3s Kubernetes clusters** across bare-metal or hybrid environments. K3s is a lightweight, CNCF-certified Kubernetes distribution ideal for edge, IoT, and resource-constrained environments.

Like RKE2, K3s clusters use:

- **OpenBao** (secret store) to securely store and retrieve join tokens
- A persistent **cluster ID** for every control plane (e.g., `k3s-abc12345`)
- Embedded bash scripts for modular system-level execution

---

## Core Concepts

### Cluster ID

- A **unique identifier** generated at control plane bootstrap (e.g., `k3s-abc12345`)
- Stored persistently at `/etc/edgectl/cluster-id`
- Used as the secret store key for all token-related operations
- Ensures agents can connect to the right control plane without handling raw tokens

### Secret Store Integration

- All tokens are stored/retrieved using the **Cluster ID** path:
  ```
  kv/data/k3s/<cluster-id>
  ```
- Users never manually copy tokens
- Tokens are only exposed during bootstrap and handled programmatically afterward

---

## Differences from RKE2

| Feature | RKE2 | K3s |
|---------|------|-----|
| Supervisor port | 9345 (separate) | None (uses 6443) |
| CIS hardening | Built-in (`profile: cis`) | Not included |
| etcd metrics port | 2381 | Not exposed |
| Config method | `/etc/rancher/rke2/config.yaml` | CLI flags |
| CNI | Cilium (via HelmChartConfig) | Cilium (via HelmChart manifest) |
| Default components | Full | Traefik, kube-proxy, and network policy disabled by default |
| Manifest directory | `/var/lib/rancher/rke2/server/manifests/` | `/var/lib/rancher/k3s/server/manifests/` |

---

## Workflow

### 1. Bootstrap the control plane

```bash
sudo edgectl k3s server install --vip 172.16.12.232
```

This will:
- Configure the host (disable swap, load kernel modules, apply sysctl settings)
- Install K3s in server mode with Cilium CNI and Hubble observability
- Deploy the Stakater Reloader addon
- Configure firewall rules for server ports
- Generate a unique cluster ID (e.g., `k3s-abc12345`)
- Store the join token in OpenBao

### 2. Join additional server nodes

```bash
sudo edgectl k3s server install --cluster-id k3s-abc12345
```

The join token is automatically retrieved from the secret store.

### 3. Join agent (worker) nodes

```bash
sudo edgectl k3s agent install --cluster-id k3s-abc12345
```

### 4. Fetch kubeconfig

```bash
edgectl k3s system kubeconfig --cluster-id k3s-abc12345
```

---

## Command Reference

### Server & Agent

```bash
edgectl k3s server install [--cluster-id <id>] [--vip <ip>]
edgectl k3s agent install --cluster-id <id> [--vip <ip>]
```

### Load Balancer

```bash
edgectl k3s lb create --cluster-id <id> [--vip <ip>]
edgectl k3s lb status --cluster-id <id>
edgectl k3s lb cleanup --cluster-id <id>
```

### System

```bash
edgectl k3s system status
edgectl k3s system purge [--cluster-id <id>]
edgectl k3s system kubeconfig --cluster-id <id> [--output <path>]
edgectl k3s system bash
```

---

## Firewall Ports

EdgeCTL automatically configures firewall rules during installation. See [Firewall Configuration](firewall.md) for details.

**Server node:**

| Port | Protocol | Purpose |
|------|----------|---------|
| 22 | TCP | SSH access |
| 6443 | TCP | K3s API Server |
| 10250 | TCP | kubelet metrics |
| 2379 | TCP | etcd client |
| 2380 | TCP | etcd peer |
| 30000-32767 | TCP | Kubernetes NodePort range |

**Agent node:**

| Port | Protocol | Purpose |
|------|----------|---------|
| 22 | TCP | SSH access |
| 10250 | TCP | kubelet metrics |
| 30000-32767 | TCP | Kubernetes NodePort range |

---

## File Layout

| Path | Purpose |
|------|---------|
| `/etc/edgectl/cluster-id` | Stores generated Cluster ID |
| `kv/data/k3s/<cluster-id>` (OpenBao) | Join token + metadata for that cluster |
| `/var/lib/rancher/k3s/server/manifests/` | Auto-deployed Kubernetes manifests (Cilium, Reloader) |
| `/etc/rancher/k3s/` | K3s configuration directory |

---

## K3s Installation Defaults

EdgeCTL configures K3s with the following defaults:

- **Flannel disabled** (`--flannel-backend=none`) — replaced by Cilium
- **kube-proxy disabled** (`--disable-kube-proxy`) — replaced by Cilium eBPF
- **Network policy disabled** (`--disable-network-policy`) — handled by Cilium
- **Traefik disabled** (`--disable=traefik`) — bring your own ingress
- **Cilium CNI** with Hubble UI and relay enabled
- **Stakater Reloader** for automatic workload restarts on ConfigMap/Secret changes
