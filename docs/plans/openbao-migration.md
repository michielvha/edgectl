# OpenBao Migration Plan

HashiCorp changed Vault to BSL (Business Source License) in 2023, making it incompatible with open-source redistribution.
OpenBao is the Linux Foundation community fork of Vault. It is a near-identical API with different import paths and env var names.

OpenBao SDK: `github.com/openbao/openbao/api/v2 @ v2.5.1` (as of 2026-02-04)

---

## Current State Assessment

### Maintenance status

Go 1.24.1 — current. Most dependencies are up to date.

| Dependency | Current | Available |
|------------|---------|-----------|
| `go-viper/mapstructure/v2` | v2.4.0 | v2.5.0 |
| All other direct deps | current | — |

### Open TODOs

| File | Line | Description |
|------|------|-------------|
| `pkg/vault/generic.go` | 47 | `InitVaultClient()` exists but not used consistently everywhere |
| `cmd/vault.go` | 13 | Vault CLI too RKE2-specific; planned generic abstraction never implemented |
| `cmd/rke2/system/commands.go` | 44 | `rke2 system purge` doesn't clean up Vault data |
| `cmd/rke2/lb/commands.go` | 58 | LB command logic not moved to handler package |
| `pkg/rke2/agent/install.go` | 38 | No VIP fallback when not found in Vault |
| `pkg/common/embedded.go` | 18 | Script permission bug for multi-user environments |

### Vault integration surface

All Vault code lives in `pkg/vault/` (5 files). The SDK is used for KV v2 operations only:
- `StoreSecret` / `RetrieveSecret` / `ListKeys` / `DeleteSecret` via `.Logical().Read/Write/List/Delete()`
- No auth method complexity — token auth only via env vars
- Paths: `kv/data/rke2/{clusterID}/{token,kubeconfig,masters,lb/hostname}`

---

## Migration Steps

### Step 1 — Replace the SDK

In `go.mod`, replace:
```
github.com/hashicorp/vault/api v1.22.0
```
with:
```
github.com/openbao/openbao/api/v2 v2.5.1
```

Then run:
```bash
go get github.com/openbao/openbao/api/v2
go mod tidy
```

`go mod tidy` will remove the BSL-licensed `hashicorp/vault/api` dependency. Note: some lower-level Apache-2.0 hashicorp libraries (e.g. `hcl`) may persist as transitive deps of OpenBao itself — this is fine as they are not BSL-licensed.

### Step 2 — Update imports (5 files)

All files in `pkg/vault/` import:
```go
vault "github.com/hashicorp/vault/api"
```

Change to:
```go
vault "github.com/openbao/openbao/api/v2"
```

**Zero API changes needed.** Same `*vault.Client`, same `.Logical().Read/Write/List/Delete()`, same `DefaultConfig()`, same `NewClient()`. The alias `vault` stays the same so nothing else in the files changes.

Files to update:
- `pkg/vault/generic.go`
- `pkg/vault/rke2_token.go`
- `pkg/vault/rke2_server.go`
- `pkg/vault/rke2_kubeconfig.go`
- `pkg/vault/rke2_loadbalancer.go`

### Step 3 — Update environment variables (breaking for users)

OpenBao dropped all `VAULT_*` env vars. They are now `BAO_*`.

| Old | New |
|-----|-----|
| `VAULT_ADDR` | `BAO_ADDR` |
| `VAULT_TOKEN` | `BAO_TOKEN` |
| `VAULT_CACERT` | `BAO_CACERT` |
| `VAULT_SKIP_VERIFY` | `BAO_SKIP_VERIFY` |
| `VAULT_NAMESPACE` | `BAO_NAMESPACE` |

**Code change required:**

`pkg/vault/generic.go` lines 37–39:
```go
// Before
token := os.Getenv("VAULT_TOKEN")
if token == "" {
    return nil, fmt.Errorf("VAULT_TOKEN not set")
}

// After
token := os.Getenv("BAO_TOKEN")
if token == "" {
    return nil, fmt.Errorf("BAO_TOKEN not set")
}
```

`DefaultConfig()` automatically picks up `BAO_ADDR` after the import swap — no manual change needed for the address.

### Step 4 — Update user-facing strings and comments

There are ~30+ references to "Vault" in user-facing output (fmt.Printf/Println) and comments across the codebase. These are spread across files that import the local `pkg/vault` package — they don't need import changes but DO need string/comment updates:

- `pkg/vault/generic.go` — package doc says "HashiCorp Vault", error messages say "Vault client"
- `pkg/vault/rke2_*.go` (4 files) — package docs reference "Vault"
- `pkg/rke2/server/install.go` — ~15 references: "VIP fetched from Vault", "stored in Vault", etc.
- `pkg/rke2/agent/install.go` — ~5 references: "Vault client", "VIP from Vault"
- `pkg/lb/handler.go` — ~12 references: "Connect to Vault", "Removing load balancer entry from Vault"
- `cmd/vault.go` — "Uploading token to Vault", "Fetching token from Vault"
- `cmd/rke2/system/commands.go` — "Fetch kubeconfig from Vault"
- `cmd/rke2/lb/commands.go` — "Connect to Vault", "Clean up LB and remove from Vault"
- `cmd/rke2.go` — help text says "Fetch kubeconfig from Vault"

Decision needed: replace with "OpenBao" everywhere, or use a generic term like "secret store" for user-facing strings. Using "OpenBao" is more accurate; using "secret store" would be more future-proof if an interface pattern is added later (see root readme.md discussion).

### Step 5 — Update docs

`docs/vault/readme.md` needs to be rewritten:
- Replace HashiCorp Vault install instructions with OpenBao equivalents
- Update env var names to `BAO_ADDR` / `BAO_TOKEN`
- Docker image: `hashicorp/vault` → `quay.io/openbao/openbao`
- Title/references: "HashiCorp Vault" → "OpenBao"

### Step 6 — Update root readme.md

The project root `readme.md` (not in `docs/`) has:
- Line 57–61: "Secret Management" section that mentions HashiCorp Vault BSL and lists infisical/openbao as alternatives — needs rewriting since we're now committing to OpenBao
- Line 73: "Done features" says "Integrate HashiCorp Vault" — update to OpenBao
- Line 61: mentions "redesign the code with an interface" — decide if this is still a goal

### Step 7 — CLI command naming decision

After migration, `edgectl vault` as a subcommand name is potentially confusing since we're no longer using Vault. Options:
1. Keep `edgectl vault` — users are familiar with the term, low friction
2. Rename to `edgectl bao` — matches the actual backend
3. Rename to `edgectl secrets` — generic, future-proof if interface pattern is added

This is a UX decision. Recommend option 3 if an interface pattern is planned, option 1 if not.

---

## What does NOT change

- All KV v2 paths (`kv/data/rke2/...`) stay the same — OpenBao is wire-compatible with Vault on the server side
- Token auth method unchanged
- `.goreleaser.yml` and `makefile` have no Vault references
- All business logic in `pkg/rke2/` and `pkg/lb/` unchanged (code logic stays the same, only strings/comments change)
- `renovate.json` — existing config groups all `github.com/` deps and auto-merges patches; will automatically cover `openbao/openbao` updates

---

## OpenBao server setup (for reference)

OpenBao replaces the `vault` binary with `bao`:

```bash
# Docker
docker run --cap-add=IPC_LOCK \
  -e BAO_DEV_ROOT_TOKEN_ID="root" \
  -p 8200:8200 \
  quay.io/openbao/openbao:latest server -dev

# Environment
export BAO_ADDR="http://127.0.0.1:8200"
export BAO_TOKEN="root"
```

The CLI commands are identical, just replace `vault` with `bao`:
```bash
bao kv put kv/test foo=bar
bao kv get kv/test
```

---

## Estimated effort

| Task | Effort |
|------|--------|
| SDK swap + go mod tidy | 5 min |
| Import updates (5 files) | 5 min |
| Env var update (1 file, 2 lines) | 2 min |
| User-facing strings + comments (~30 occurrences across ~10 files) | 20 min |
| Docs rewrite (docs/vault/ + root readme.md) | 15 min |
| CLI command naming (if renaming `vault` subcommand) | 10 min |
| `go build` + `go vet` + smoke test | 10 min |
| **Total** | **~70 min** |
