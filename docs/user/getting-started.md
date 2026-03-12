# Getting Started

## Prerequisites

- Linux host (Debian/Ubuntu-based)
- Root access (required for RKE2 installation)
- Go 1.24+ (for installing from source)
- An [OpenBao](https://openbao.org/) instance for secret management

## Install edgectl

```bash
go install github.com/michielvha/edgectl@latest
edgectl version
```

Or download a pre-built binary from the [releases page](https://github.com/michielvha/edgectl/releases/latest).

### Homebrew (macOS/Linux)

```bash
brew install michielvha/tap/edgectl
```

## Configure the secret store

edgectl uses OpenBao to store and retrieve cluster secrets (join tokens, kubeconfigs, etc.).

Set the following environment variables:

```bash
export BAO_ADDR="https://your-openbao-instance:8200"
export BAO_TOKEN="your-token"
```

For a local dev setup, you can run OpenBao in Docker:

```bash
docker run --cap-add=IPC_LOCK \
  -e BAO_DEV_ROOT_TOKEN_ID="root" \
  -p 8200:8200 \
  quay.io/openbao/openbao:latest server -dev
```

Then:
```bash
export BAO_ADDR="http://127.0.0.1:8200"
export BAO_TOKEN="root"
```

## Create your first cluster

### 1. Bootstrap the control plane

```bash
sudo edgectl rke2 server install --vip 172.16.12.232
```

This will:
- Install RKE2 in server mode
- Generate a unique cluster ID (e.g., `rke2-abc12345`)
- Store the join token in OpenBao

### 2. Join additional server nodes

```bash
sudo edgectl rke2 server install --cluster-id rke2-abc12345
```

The join token is automatically retrieved from the secret store.

### 3. Join agent (worker) nodes

```bash
sudo edgectl rke2 agent install --cluster-id rke2-abc12345
```

### 4. Fetch kubeconfig

```bash
edgectl rke2 system kubeconfig --cluster-id rke2-abc12345
```

## Next steps

- [RKE2 Cluster Management](rke2.md) — full command reference and architecture
- [Load Balancer Setup](loadbalancer.md) — HA load balancing with HAProxy + Keepalived
- [Secret Management](secret-management.md) — OpenBao integration details
