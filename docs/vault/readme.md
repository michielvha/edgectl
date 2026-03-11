# OpenBao integration

As a secret store backend for the **EdgeCloud** we use [OpenBao](https://openbao.org/), the Linux Foundation community fork of HashiCorp Vault.
OpenBao is wire-compatible with the Vault KV v2 API.

## Features

- [x] Save secrets
- [x] Fetch secrets

## Install the CLI

OpenBao replaces the `vault` binary with `bao`:

```shell
# See https://openbao.org/docs/install/ for platform-specific instructions
```

## Docker (dev mode)

```bash
docker run --cap-add=IPC_LOCK \
  -e BAO_DEV_ROOT_TOKEN_ID="root" \
  -p 8200:8200 \
  quay.io/openbao/openbao:latest server -dev
```

## Set env vars

- **linux:**
    ```shell
    export BAO_ADDR="https://edgevault.duckdns.org"
    export BAO_TOKEN=""
    ```

## RKE2 integration
