# TODO Fixes Plan

## TODO 1 — Consistent Vault client init (`pkg/vault/generic.go:47`)

**Problem:** Two patterns exist for initializing the Vault client:
- `vault.InitVaultClient()` — prints error and returns nil; used in `cmd/vault.go`
- `vault.NewClient()` + manual error handling — used in `cmd/rke2/system/commands.go:67`, `cmd/rke2/lb/commands.go:68`, and throughout `pkg/lb/handler.go` and `pkg/rke2/agent/install.go`

**Fix:** Clarify the two patterns as intentional with different scopes:
- `NewClient()` is correct for `pkg/` code that propagates errors — no change needed
- `InitVaultClient()` should be used in all `cmd/` handlers that currently do `vault.NewClient()` + inline error print + `os.Exit(1)`

**Files to update:**
- `cmd/rke2/system/commands.go:67` — replace `vault.NewClient()` + error block with `vault.InitVaultClient()` + nil check
- `cmd/rke2/lb/commands.go:68` — same

**Additional code smell (not a TODO but worth fixing):** Both `server/install.go` and `agent/install.go` create a vault client at the top of `Install()`, then call `FetchTokenFromVault()`/`FetchToken()` which creates a SECOND client internally. This wastes connections. Consider passing the client as a parameter:
- `pkg/rke2/server/install.go:22` creates client → `FetchTokenFromVault()` at line 144 creates another
- `pkg/rke2/agent/install.go:18` creates client → `FetchToken()` at line 49 creates another

Remove the TODO comment once done.

---

## TODO 2 — Vault CLI too RKE2-specific (`cmd/vault.go:13`)

**Problem:** `edgectl vault upload/fetch` only work with RKE2 join tokens. The referenced `pkg/vault/handler.go` was never created. The commands are useful for testing but not generic.

**Fix:** Add two generic subcommands alongside the existing ones:

```
edgectl vault get  --path <kv-path> --key <field>
edgectl vault set  --path <kv-path> --key <field> --value <value>
```

These wrap `StoreSecret` and `RetrieveSecret` directly, enabling ad-hoc interaction with any KV path. Remove the stale `handler.go` reference from the comment.

**Files to update:**
- `cmd/vault.go` — add `vaultGetCmd` and `vaultSetCmd`, remove stale TODO/handler comment

---

## TODO 3 — `rke2 system purge` doesn't clean Vault (`cmd/rke2/system/commands.go:44`)

**Problem:** Purge only runs the local uninstall script. Cluster secrets remain in Vault indefinitely after a node is decommissioned.

**Fix:**
1. Add optional `--cluster-id` flag to `purgeCmd`
2. Add a `DeleteClusterData(clusterID string) error` method to `pkg/vault/generic.go` (or a new `rke2_cluster.go`) that deletes:
   - `kv/metadata/rke2/{clusterID}/token`
   - `kv/metadata/rke2/{clusterID}/kubeconfig`
   - `kv/metadata/rke2/{clusterID}/masters`
   - All LB entries under `kv/metadata/rke2/{clusterID}/lb/` (list then delete each)
3. In `purgeCmd`: if `--cluster-id` is provided, call `InitVaultClient()` and `DeleteClusterData()` after the bash script

The flag should be optional so purge still works without Vault access (e.g. Vault is down or was already wiped). Vault cleanup errors should be warn-and-continue, not fatal — the local purge is the primary operation.

**KV v2 deletion note:** For permanent deletion of all versions, use the `kv/metadata/` prefix (not `kv/data/`). The existing `DeleteSecret()` uses `.Logical().Delete()` which on `kv/data/` only soft-deletes the latest version. Targeting `kv/metadata/{path}` destroys all versions permanently.

**Files to update:**
- `cmd/rke2/system/commands.go` — add `--cluster-id` flag, vault cleanup call
- `pkg/vault/` — add `DeleteClusterData` method (new file `rke2_cluster.go` or append to `generic.go`)

---

## TODO 4 — LB status logic not in handler (`cmd/rke2/lb/commands.go:58`)

**Problem:** `statusCmd` directly creates a vault client and formats output inline. The handler package (`pkg/lb/handler.go`) should own this logic so it's testable and reusable.

**Fix:** Add `GetStatus(clusterID string) (vip string, nodes []LBNode, err error)` to `pkg/lb/handler.go`.

Define a typed struct instead of the current `map[string]interface{}`:
```go
type LBNode struct {
    Hostname string
    IsMain   bool
}
```

`statusCmd` calls `lb.GetStatus()` and formats the output. The raw vault client is no longer needed in the cmd layer.

**Files to update:**
- `pkg/lb/handler.go` — add `LBNode` struct and `GetStatus()` function
- `cmd/rke2/lb/commands.go` — replace inline vault logic with `lb.GetStatus()` call

---

## TODO 5 — No VIP fallback in agent install (`pkg/rke2/agent/install.go:38`)

**Problem:** If Vault has no VIP stored (e.g. cluster was bootstrapped without one), the agent installs without TLS SANs. The node may not be able to reach the API server through a later-added load balancer.

**Fix:** Add a `--vip` flag to the agent install command as an explicit override/fallback.

Flow:
1. Try to retrieve VIP from Vault (existing behaviour)
2. If not found in Vault **and** `--vip` flag was passed, use the flag value
3. If neither source has a VIP, log a warning and continue (existing behaviour — don't hard fail, some clusters don't use VIPs)

**Dead flag cleanup:** `cmd/rke2/agent/commands.go:54` already registers an `--lb-hostname` flag that is never read or passed to `agent.Install()`. Either remove this dead flag or repurpose it alongside the new `--vip` flag.

**Files to update:**
- `cmd/rke2/agent/commands.go` — add `--vip` flag, remove or repurpose dead `--lb-hostname` flag, pass vip to `agent.Install()`
- `pkg/rke2/agent/install.go` — accept optional vip parameter, use flag value as fallback

---

## TODO 6 — Script permission bug for multi-user environments (`pkg/common/embedded.go:18`)

**Problem:** Scripts are extracted to `/tmp/{scriptName}` (e.g. `/tmp/rke2.sh`). If user A runs first, the file is owned by A with `0o777`. When user B runs, `os.WriteFile` fails because they don't own the file, so the old (potentially stale) script from user A executes instead — or the write fails entirely.

**Fix:** Extract to a user-namespaced temp directory using the current user's UID:

```go
import "os/user"

func ExtractEmbeddedScript(scriptName string) string {
    u, err := user.Current()
    uid := "unknown"
    if err == nil {
        uid = u.Uid
    }

    dir := filepath.Join("/tmp", "edgectl-"+uid)
    if err := os.MkdirAll(dir, 0o700); err != nil {
        fmt.Printf("❌ Failed to create script dir: %v\n", err)
        os.Exit(1)
    }

    scriptPath := filepath.Join(dir, scriptName)
    // ... rest unchanged
}
```

This gives each user their own `/tmp/edgectl-{uid}/` directory, isolated from other users. No cleanup needed between runs.

**Files to update:**
- `pkg/common/embedded.go` — update `ExtractEmbeddedScript` to use UID-namespaced path

---

---

## Bonus — Makefile bugs (not TODOs but broken)

Two pre-existing bugs in `makefile`:
1. **Line 54:** `.PHONY: config` but the target on line 55 is `status:` — should be `.PHONY: status`
2. **Line 59:** `test func:` has a space in the target name — invalid make syntax. Should be `test-func:` (matching the `.PHONY: test-func` on line 58)

Additionally, `make purge` (line 48) should be updated if TODO 3 adds `--cluster-id` to the purge command.

---

## Execution order

Suggested order to minimise conflicts when implementing:

1. **TODO 6** (embedded.go) — isolated change, no deps
2. **TODO 1** (vault client consistency) — small cmd-layer cleanup
3. **TODO 4** (lb status → handler) — adds struct + function, refactors one command
4. **TODO 3** (purge vault cleanup) — adds new vault method + flag
5. **TODO 5** (agent VIP fallback) — small flag + param change
6. **TODO 2** (generic vault CLI) — additive, no breaking changes
7. **Makefile fixes** — can be done at any point
