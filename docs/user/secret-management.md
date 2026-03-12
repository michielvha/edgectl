# Secret Management (OpenBao)

As a secret store backend for the **EdgeCloud** we use [OpenBao](https://openbao.org/), the Linux Foundation community fork of HashiCorp Vault.
OpenBao is open-source (MPL-2.0) and wire-compatible with the Vault KV v2 API.

## Features

- [x] Save secrets
- [x] Fetch secrets
- [x] List keys
- [x] Delete secrets (soft + permanent)
- [x] Generic `get`/`set` CLI commands for ad-hoc use

---

## Deploy OpenBao

### Option 1: Docker Compose (recommended for self-hosting)

A ready-to-use compose file is provided in [`deploy/openbao/`](../../deploy/openbao/):

```bash
cd deploy/openbao
docker compose up -d
```

This runs OpenBao with **Raft integrated storage** (persistent, transactional, recommended over the file backend). Data is stored in a named Docker volume.

After the first start, you need to initialize and unseal:

```bash
# Initialize — save the unseal keys and root token!
docker compose exec openbao bao operator init

# Unseal (repeat 3 times with different unseal keys)
docker compose exec openbao bao operator unseal

# Login with the root token
export BAO_ADDR="http://127.0.0.1:8200"
export BAO_TOKEN="<root-token-from-init>"

# Enable KV v2 at the path edgectl expects
bao secrets enable -path=kv -version=2 kv
```

> **Note:** After every container restart you must unseal again (3 of 5 keys). For unattended operation, configure [auto-unseal](https://openbao.org/docs/configuration/seal/) via Transit, AWS KMS, Azure Key Vault, or GCP Cloud KMS.

### Option 2: Dev mode (quick testing, no persistence)

```bash
docker run --cap-add=IPC_LOCK \
  -e BAO_DEV_ROOT_TOKEN_ID="root" \
  -p 8200:8200 \
  openbao/openbao:2.3.1 server -dev
```

Dev mode is auto-unsealed, runs fully in-memory, and mounts a KV v2 engine at `secret/` by default. You still need to enable the `kv/` path that edgectl uses:

```bash
export BAO_ADDR="http://127.0.0.1:8200"
export BAO_TOKEN="root"
bao secrets enable -path=kv -version=2 kv
```

### Option 3: Install the CLI directly

```bash
# See https://openbao.org/docs/install/ for platform-specific instructions
```

OpenBao replaces the `vault` binary with `bao`. All CLI commands are identical to Vault, just swap `vault` with `bao`.

---

## Configure edgectl

Set these environment variables so edgectl can reach your OpenBao instance:

```bash
export BAO_ADDR="https://your-openbao-instance:8200"
export BAO_TOKEN="<your-token>"
```

For local development:

```bash
export BAO_ADDR="http://127.0.0.1:8200"
export BAO_TOKEN="root"
```

---

## How edgectl uses OpenBao

All cluster data is stored under the `kv/` KV v2 mount using the following path structure:

```
kv/data/rke2/<cluster-id>/token         # Join token
kv/data/rke2/<cluster-id>/kubeconfig    # Kubeconfig
kv/data/rke2/<cluster-id>/masters       # Master node list
kv/data/rke2/<cluster-id>/lb/<hostname> # Load balancer node info
```

The `kv/metadata/` prefix is used for permanent deletion (all versions) during cluster cleanup (`edgectl rke2 system purge --cluster-id`).

### CLI commands

```bash
# Generic secret operations
edgectl vault get --path kv/data/myapp/config --key api_url
edgectl vault set --path kv/data/myapp/config --key api_url --value https://example.com

# RKE2-specific (used internally by cluster commands)
edgectl vault upload --cluster-id <id>   # Upload join token
edgectl vault fetch --cluster-id <id>    # Fetch join token
```

---

## Verify your setup

```bash
# Check connectivity
bao status

# Write a test secret
bao kv put kv/test foo=bar

# Read it back
bao kv get kv/test

# Clean up
bao kv metadata delete kv/test
```
