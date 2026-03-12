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

After the first start, you need to initialize, unseal, and enable the KV engine. An init script is provided that does all three in one go:

```bash
cd deploy/openbao
docker compose up -d
./init.sh
```

The script will:
1. Initialize OpenBao and print **5 unseal keys** + **1 root token**
2. Auto-unseal using 3 of the 5 keys
3. Enable the KV v2 engine at the `kv/` path that edgectl expects

> **Save the unseal keys and root token.** The unseal keys are the only way to unlock OpenBao after a restart. If you lose them, you must wipe the volume and reinitialize (`docker compose down -v && docker compose up -d && ./init.sh`).

> **After every container restart** you must unseal again (3 of 5 keys). For unattended operation, configure [auto-unseal](https://openbao.org/docs/configuration/seal/) via Transit, AWS KMS, Azure Key Vault, or GCP Cloud KMS.

<details>
<summary>Manual steps (if you prefer not to use the script)</summary>

```bash
# 1. Initialize — prints unseal keys + root token
docker compose exec openbao bao operator init

# 2. Unseal (repeat 3 times with different unseal keys)
docker compose exec openbao bao operator unseal    # paste key 1
docker compose exec openbao bao operator unseal    # paste key 2
docker compose exec openbao bao operator unseal    # paste key 3

# 3. Enable KV v2
export BAO_TOKEN="<root-token-from-init>"
docker compose exec openbao env BAO_TOKEN="$BAO_TOKEN" bao secrets enable -path=kv -version=2 kv
# docker exec -e BAO_TOKEN=your-token openbao bao secrets enable -path=kv -version=2 kv

```

</details>

### Option 2: Dev mode (quick testing, no persistence)

```bash
docker run --cap-add=IPC_LOCK \
  -e BAO_DEV_ROOT_TOKEN_ID="root" \
  -p 8200:8200 \
  openbao/openbao:2.3.1 server -dev
```

Dev mode is auto-unsealed, runs fully in-memory, and mounts a KV v1 engine at `secret/` by default (just like the prod). You still need to enable the `kv/` path that edgectl uses:

```bash
export VAULT_ADDR="http://127.0.0.1:8200"
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
export VAULT_ADDR="https://your-openbao-instance:8200"
export BAO_TOKEN="<your-token>"
```

For local development:

```bash
export VAULT_ADDR="http://127.0.0.1:8200"
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

If `bao` is installed on your host, you can run commands directly (with `VAULT_ADDR` and `BAO_TOKEN` set). Otherwise, exec into the container:

```bash
# Check connectivity
docker compose exec openbao bao status

# Write a test secret
docker compose exec openbao bao kv put kv/test foo=bar

# Read it back
docker compose exec openbao bao kv get kv/test

# Clean up
docker compose exec openbao bao kv metadata delete kv/test
```
