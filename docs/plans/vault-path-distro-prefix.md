# Vault Path Distro-Aware Prefix — Implementation Report

## Problem

All vault secret paths are hardcoded to `kv/data/rke2/` and `kv/metadata/rke2/` regardless of cluster type. When a K3s cluster is installed, its secrets end up stored as:

```
kv/data/rke2/k3s-abc12345/token        # K3s cluster, stored under rke2/ prefix
kv/data/rke2/k3s-abc12345/kubeconfig
kv/data/rke2/k3s-abc12345/masters
kv/data/rke2/k3s-abc12345/lb/hostname
```

**Expected** (organized by cluster type):

```
kv/data/k3s/k3s-abc12345/token
kv/data/k3s/k3s-abc12345/kubeconfig
kv/data/k3s/k3s-abc12345/masters
kv/data/k3s/k3s-abc12345/lb/hostname
```

The `docs/plans/k3s-support.md` (line 96-98) acknowledged this as a "naming artifact" and deferred it. This plan addresses it.

---

## Custom Prefix Option — Analysis

The user considered allowing a custom top-level value (e.g., a domain name) so users can group clusters differently. **Recommendation: skip this for now.** Reasons:

1. The distro prefix (`rke2`/`k3s`) is already implicitly set — the tool knows the distro at install time. No new flags needed.
2. A custom prefix adds a new required parameter that must be stored, passed through every command, and kept consistent across all nodes in a cluster. It would need to be persisted (e.g., in `/etc/edgectl/vault-prefix`) and validated on every operation.
3. The cluster ID already provides uniqueness. Users who need custom grouping can use vault policies or mount points rather than path prefixes.
4. If needed later, the same `distro` parameter introduced here can trivially be replaced with a user-supplied string without further architectural changes.

---

## Approach

Add a `distro` string parameter to every vault method that constructs a path. The distro value (`"rke2"` or `"k3s"`) replaces the hardcoded `"rke2"` in all `fmt.Sprintf` path constructions. The callers (install commands, CLI handlers) already know the distro.

### Option chosen: Parameter on each method

The `SecretStore` interface methods gain a `distro` parameter. This is preferred over storing distro on the `Client` struct because:
- The `Client` is a generic vault client — it shouldn't be distro-aware
- The interface stays explicit about what data it needs
- Tests remain simple (no client configuration needed)

---

## Files to Change

### 1. `pkg/vault/interface.go` — Add `distro` parameter to interface

**Current** (lines 17-35):
```go
StoreJoinToken(clusterID, token string) error
RetrieveJoinToken(clusterID string) (string, error)
StoreMasterInfo(clusterID, hostname string, hosts []string, vip string) error
RetrieveMasterInfo(clusterID string) (hosts []string, vip string, hostIPs map[string]string, err error)
RetrieveFirstMasterIP(clusterID string) (string, error)
StoreKubeConfig(clusterID, kubeconfigPath, vip string) error
RetrieveKubeConfig(clusterID, destinationPath string) error
StoreLBInfo(clusterID, hostname, vip string, isMain bool) error
RetrieveLBInfo(clusterID string) (nodes []map[string]interface{}, vip string, err error)
RemoveLBNode(clusterID, hostname string) error
DeleteClusterData(clusterID string) error
```

**Target**: Every method that constructs a vault path gets `distro` as the **first** parameter:
```go
StoreJoinToken(distro, clusterID, token string) error
RetrieveJoinToken(distro, clusterID string) (string, error)
StoreMasterInfo(distro, clusterID, hostname string, hosts []string, vip string) error
RetrieveMasterInfo(distro, clusterID string) (hosts []string, vip string, hostIPs map[string]string, err error)
RetrieveFirstMasterIP(distro, clusterID string) (string, error)
StoreKubeConfig(distro, clusterID, kubeconfigPath, vip string) error
RetrieveKubeConfig(distro, clusterID, destinationPath string) error
StoreLBInfo(distro, clusterID, hostname, vip string, isMain bool) error
RetrieveLBInfo(distro, clusterID string) (nodes []map[string]interface{}, vip string, err error)
RemoveLBNode(distro, clusterID, hostname string) error
DeleteClusterData(distro, clusterID string) error
```

Also update the comments from "RKE2 token management" etc. to "Cluster token management" (generic).

---

### 2. `pkg/vault/rke2_token.go` — Use `distro` in paths

**Lines 21, 29**: Replace hardcoded `"rke2"` with `distro` parameter.

```go
// Before (line 21):
func (c *Client) StoreJoinToken(clusterID, token string) error {
    return c.StoreSecret(fmt.Sprintf("kv/data/rke2/%s/token", clusterID), ...)

// After:
func (c *Client) StoreJoinToken(distro, clusterID, token string) error {
    return c.StoreSecret(fmt.Sprintf("kv/data/%s/%s/token", distro, clusterID), ...)
```

Same pattern for `RetrieveJoinToken` (line 28-29).

**Also**: Consider renaming this file from `rke2_token.go` to `token.go` since it's now distro-agnostic. (Optional, but cleaner.)

---

### 3. `pkg/vault/rke2_server.go` — Use `distro` in paths

**4 path constructions** need updating:

| Line | Current | Target |
|------|---------|--------|
| 27 | `func (c *Client) StoreMasterInfo(clusterID, hostname string, ...)` | `func (c *Client) StoreMasterInfo(distro, clusterID, hostname string, ...)` |
| 37 | `fmt.Sprintf("kv/data/rke2/%s/masters", clusterID)` | `fmt.Sprintf("kv/data/%s/%s/masters", distro, clusterID)` |
| 58 | `fmt.Sprintf("kv/data/rke2/%s/masters", clusterID)` | `fmt.Sprintf("kv/data/%s/%s/masters", distro, clusterID)` |
| 100 | `func (c *Client) RetrieveMasterInfo(clusterID string)` — line 101 path | `func (c *Client) RetrieveMasterInfo(distro, clusterID string)` — use `distro` in path |
| 137 | `func (c *Client) RetrieveFirstMasterIP(clusterID string)` — line 138 path | `func (c *Client) RetrieveFirstMasterIP(distro, clusterID string)` — use `distro` in path |

**Also**: Consider renaming from `rke2_server.go` to `server.go`. (Optional.)

---

### 4. `pkg/vault/rke2_kubeconfig.go` — Use `distro` in paths

| Line | Current | Target |
|------|---------|--------|
| 24 | `func (c *Client) StoreKubeConfig(clusterID, kubeconfigPath, vip string)` | Add `distro` as first param |
| 41 | `fmt.Sprintf("kv/data/rke2/%s/kubeconfig", clusterID)` | `fmt.Sprintf("kv/data/%s/%s/kubeconfig", distro, clusterID)` |
| 47 | `func (c *Client) RetrieveKubeConfig(clusterID, destinationPath string)` | Add `distro` as first param |
| 48 | `fmt.Sprintf("kv/data/rke2/%s/kubeconfig", clusterID)` | `fmt.Sprintf("kv/data/%s/%s/kubeconfig", distro, clusterID)` |

**Also**: Consider renaming from `rke2_kubeconfig.go` to `kubeconfig.go`. (Optional.)

---

### 5. `pkg/vault/rke2_loadbalancer.go` — Use `distro` in paths

| Line | Function | Path change |
|------|----------|-------------|
| 22 | `StoreLBInfo(clusterID, hostname, vip string, isMain bool)` | Add `distro` first param |
| 23 | `fmt.Sprintf("kv/data/rke2/%s/lb/%s", clusterID, hostname)` | `fmt.Sprintf("kv/data/%s/%s/lb/%s", distro, clusterID, hostname)` |
| 32 | `RetrieveLBInfo(clusterID string)` | Add `distro` first param |
| 34 | `fmt.Sprintf("kv/metadata/rke2/%s/lb", clusterID)` | `fmt.Sprintf("kv/metadata/%s/%s/lb", distro, clusterID)` |
| 53 | `fmt.Sprintf("kv/data/rke2/%s/lb/%s", clusterID, key)` | `fmt.Sprintf("kv/data/%s/%s/lb/%s", distro, clusterID, key)` |
| 81 | `RemoveLBNode(clusterID, hostname string)` | Add `distro` first param |
| 83 | `fmt.Sprintf("kv/metadata/rke2/%s/lb/%s", clusterID, hostname)` | `fmt.Sprintf("kv/metadata/%s/%s/lb/%s", distro, clusterID, hostname)` |

**Also**: Consider renaming from `rke2_loadbalancer.go` to `loadbalancer.go`. (Optional.)

---

### 6. `pkg/vault/rke2_cluster.go` — Use `distro` in paths

| Line | Change |
|------|--------|
| 20 | `func (c *Client) DeleteClusterData(clusterID string)` → add `distro` first param |
| 21 | `fmt.Sprintf("kv/metadata/rke2/%s", clusterID)` → `fmt.Sprintf("kv/metadata/%s/%s", distro, clusterID)` |

**Also**: Consider renaming from `rke2_cluster.go` to `cluster.go`. (Optional.)

---

### 7. Callers — Pass `distro` to vault methods

Every call site that invokes `SecretStore` methods must now pass the distro string.

#### `pkg/rke2/server/install.go`
All `store.StoreJoinToken(...)`, `store.StoreMasterInfo(...)`, etc. calls need `"rke2"` as first arg.

| Line | Current | Target |
|------|---------|--------|
| 69 | `store.StoreJoinToken(clusterID, token)` | `store.StoreJoinToken("rke2", clusterID, token)` |
| 79 | `store.StoreKubeConfig(clusterID, kubeconfigPath, vip)` | `store.StoreKubeConfig("rke2", clusterID, kubeconfigPath, vip)` |
| 93 | `store.RetrieveMasterInfo(clusterID)` | `store.RetrieveMasterInfo("rke2", clusterID)` |
| 122 | `store.StoreMasterInfo(clusterID, hostname, hosts, existingVIP)` | `store.StoreMasterInfo("rke2", clusterID, hostname, hosts, existingVIP)` |
| 37 | `store.RetrieveMasterInfo(clusterID)` (in existing-cluster VIP fetch) | `store.RetrieveMasterInfo("rke2", clusterID)` |

#### `pkg/rke2/server/install.go` — `FetchTokenFromSecretStore`
| Line | Current | Target |
|------|---------|--------|
| 140 | `store.RetrieveJoinToken(clusterID)` | `store.RetrieveJoinToken("rke2", clusterID)` |
| 155 | `store.RetrieveFirstMasterIP(clusterID)` | `store.RetrieveFirstMasterIP("rke2", clusterID)` |

#### `pkg/k3s/server/install.go`
Same pattern, but pass `"k3s"` instead of `"rke2"`.

| Line | Current | Target |
|------|---------|--------|
| 37 | `store.RetrieveMasterInfo(clusterID)` | `store.RetrieveMasterInfo("k3s", clusterID)` |
| 69 | `store.StoreJoinToken(clusterID, token)` | `store.StoreJoinToken("k3s", clusterID, token)` |
| 79 | `store.StoreKubeConfig(clusterID, kubeconfigPath, vip)` | `store.StoreKubeConfig("k3s", clusterID, kubeconfigPath, vip)` |
| 93 | `store.RetrieveMasterInfo(clusterID)` | `store.RetrieveMasterInfo("k3s", clusterID)` |
| 122 | `store.StoreMasterInfo(clusterID, hostname, hosts, existingVIP)` | `store.StoreMasterInfo("k3s", clusterID, hostname, hosts, existingVIP)` |

#### `pkg/k3s/server/install.go` — `FetchTokenFromSecretStore`
| Line | Current | Target |
|------|---------|--------|
| 138 | `store.RetrieveJoinToken(clusterID)` | `store.RetrieveJoinToken("k3s", clusterID)` |
| 155 | `store.RetrieveFirstMasterIP(clusterID)` | `store.RetrieveFirstMasterIP("k3s", clusterID)` |

#### Additional callers to find and update

Search for all `SecretStore` method invocations across the codebase. Key locations likely include:

| Caller pattern | Files to check |
|---------------|----------------|
| `store.StoreJoinToken(` | `pkg/rke2/server/`, `pkg/k3s/server/` |
| `store.RetrieveJoinToken(` | `pkg/rke2/server/`, `pkg/k3s/server/`, `pkg/rke2/agent/`, `pkg/k3s/agent/` |
| `store.StoreMasterInfo(` | `pkg/rke2/server/`, `pkg/k3s/server/` |
| `store.RetrieveMasterInfo(` | `pkg/rke2/server/`, `pkg/k3s/server/`, `cmd/` |
| `store.RetrieveFirstMasterIP(` | `pkg/rke2/server/`, `pkg/k3s/server/`, `pkg/rke2/agent/`, `pkg/k3s/agent/` |
| `store.StoreKubeConfig(` | `pkg/rke2/server/`, `pkg/k3s/server/` |
| `store.RetrieveKubeConfig(` | `cmd/rke2/`, `cmd/k3s/` |
| `store.StoreLBInfo(` | `pkg/lb/` |
| `store.RetrieveLBInfo(` | `pkg/lb/` |
| `store.RemoveLBNode(` | `pkg/lb/`, `cmd/` |
| `store.DeleteClusterData(` | `cmd/rke2/system/`, `cmd/k3s/system/` |

**Important**: The `pkg/lb/` package already has a `distro` parameter (the `Distro` field on `LoadBalancerConfig`). Reuse that value when calling vault methods from load balancer code.

Use this grep to find all call sites:
```bash
rg 'store\.(Store|Retrieve|Remove|Delete)(JoinToken|MasterInfo|FirstMasterIP|KubeConfig|LBInfo|LBNode|ClusterData)\(' --type go
```

---

### 8. Tests to Update

#### `pkg/vault/integration_test.go`
All integration test calls need the `distro` parameter added. Use `"rke2"` (or `"test"`) as the distro value in tests.

| Line | Function | Change |
|------|----------|--------|
| 165 | `client.StoreJoinToken(clusterID, token)` | Add `"rke2"` as first arg |
| 170 | `client.RetrieveJoinToken(clusterID)` | Add `"rke2"` as first arg |
| 187 | `client.StoreMasterInfo(clusterID, "master1", ...)` | Add `"rke2"` as first arg |
| 192 | `client.RetrieveMasterInfo(clusterID)` | Add `"rke2"` as first arg |
| 204 | `client.StoreMasterInfo(clusterID, "master2", ...)` | Add `"rke2"` as first arg |
| 209 | `client.RetrieveMasterInfo(clusterID)` | Add `"rke2"` as first arg |
| 222 | `client.RetrieveFirstMasterIP(clusterID)` | Add `"rke2"` as first arg |
| 255 | `client.StoreKubeConfig(clusterID, ...)` | Add `"rke2"` as first arg |
| 262 | `client.RetrieveKubeConfig(clusterID, ...)` | Add `"rke2"` as first arg |
| 294 | `client.StoreLBInfo(clusterID, ...)` | Add `"rke2"` as first arg |
| 300 | `client.StoreLBInfo(clusterID, ...)` | Add `"rke2"` as first arg |
| 305 | `client.RetrieveLBInfo(clusterID)` | Add `"rke2"` as first arg |
| 334 | `client.RemoveLBNode(clusterID, ...)` | Add `"rke2"` as first arg |
| 339 | `client.RetrieveLBInfo(clusterID)` | Add `"rke2"` as first arg |
| 356-367 | All calls in `TestIntegration_DeleteClusterData` | Add `"rke2"` as first arg |
| 370 | `client.RetrieveJoinToken(clusterID)` | Add `"rke2"` as first arg |
| 376 | `client.DeleteClusterData(clusterID)` | Add `"rke2"` as first arg |
| 382 | `client.RetrieveJoinToken(clusterID)` | Add `"rke2"` as first arg |
| 387 | `client.RetrieveMasterInfo(clusterID)` | Add `"rke2"` as first arg |

**Consider adding**: A new test `TestIntegration_K3sDistroPathSeparation` that stores secrets for both `"rke2"` and `"k3s"` distros with the same cluster ID suffix and verifies they don't collide.

#### `pkg/vault/rke2_server_test.go`
No path-dependent code in unit tests (they test pure helper functions), so **no changes needed**.

#### `pkg/k3s/server/install_test.go` and `pkg/rke2/server/install_test.go`
These likely use mock `SecretStore` implementations. The mock's method signatures must be updated to match the new interface. Check mock structs for:
- `StoreJoinToken` → add `distro` param
- `RetrieveJoinToken` → add `distro` param
- `RetrieveMasterInfo` → add `distro` param
- etc.

#### `pkg/lb/handler_test.go`
If LB tests call vault methods through mocks, update those signatures too.

---

### 9. Documentation to Update

| File | Line(s) | Change |
|------|---------|--------|
| `docs/user/secret-management.md` | 117-121 | Change `kv/data/rke2/<cluster-id>/...` to show both distros: `kv/data/<distro>/<cluster-id>/...` where `<distro>` is `rke2` or `k3s` |
| `docs/user/secret-management.md` | 123 | Update purge path reference |
| `docs/user/rke2.md` | 28, 81 | Update vault path references |
| `docs/plans/k3s-support.md` | 88-98 | Mark Step 5 as now implemented with distro-aware paths |
| `docs/plans/openbao-migration.md` | 37, 158 | Update path references |
| `docs/plans/todo-fixes.md` | 50-53 | Update path references |
| `cmd/secrets.go` | 23 | Update example path in help text: `kv/data/rke2/my-cluster/token` → show generic `kv/data/<distro>/my-cluster/token` |

---

### 10. File Renaming (Optional, Recommended)

Since the vault files are no longer RKE2-specific, rename for clarity:

| Current | Proposed |
|---------|----------|
| `pkg/vault/rke2_token.go` | `pkg/vault/token.go` |
| `pkg/vault/rke2_server.go` | `pkg/vault/server.go` |
| `pkg/vault/rke2_kubeconfig.go` | `pkg/vault/kubeconfig.go` |
| `pkg/vault/rke2_loadbalancer.go` | `pkg/vault/loadbalancer.go` |
| `pkg/vault/rke2_cluster.go` | `pkg/vault/cluster.go` |

Update file-level doc comments to remove "RKE2" references. Since Go uses package-level naming (not file-level), renaming files has zero impact on imports or the API.

---

## Summary of All Hardcoded Path Occurrences

Total: **14 `fmt.Sprintf` calls** with hardcoded `rke2` in vault paths:

| File | Line | Path pattern |
|------|------|-------------|
| `pkg/vault/rke2_token.go` | 21 | `kv/data/rke2/%s/token` |
| `pkg/vault/rke2_token.go` | 29 | `kv/data/rke2/%s/token` |
| `pkg/vault/rke2_server.go` | 37 | `kv/data/rke2/%s/masters` |
| `pkg/vault/rke2_server.go` | 58 | `kv/data/rke2/%s/masters` |
| `pkg/vault/rke2_server.go` | 101 | `kv/data/rke2/%s/masters` |
| `pkg/vault/rke2_server.go` | 138 | `kv/data/rke2/%s/masters` |
| `pkg/vault/rke2_kubeconfig.go` | 41 | `kv/data/rke2/%s/kubeconfig` |
| `pkg/vault/rke2_kubeconfig.go` | 48 | `kv/data/rke2/%s/kubeconfig` |
| `pkg/vault/rke2_loadbalancer.go` | 23 | `kv/data/rke2/%s/lb/%s` |
| `pkg/vault/rke2_loadbalancer.go` | 34 | `kv/metadata/rke2/%s/lb` |
| `pkg/vault/rke2_loadbalancer.go` | 53 | `kv/data/rke2/%s/lb/%s` |
| `pkg/vault/rke2_loadbalancer.go` | 83 | `kv/metadata/rke2/%s/lb/%s` |
| `pkg/vault/rke2_cluster.go` | 21 | `kv/metadata/rke2/%s` |

All 14 become `kv/data/%s/%s/...` or `kv/metadata/%s/%s/...` with `distro` as the first `%s`.

---

## Migration / Backwards Compatibility

**Existing clusters** have secrets stored under `kv/data/rke2/`. After this change, new K3s clusters will store under `kv/data/k3s/`, and new RKE2 clusters will continue under `kv/data/rke2/`. Existing K3s clusters (with data at `kv/data/rke2/k3s-*/`) will **not** automatically migrate.

Options:
1. **Do nothing** — existing K3s clusters keep working because the cluster ID is the real key. The `rke2` prefix is cosmetic for existing data. New clusters get clean paths.
2. **Provide a migration command** — `edgectl vault migrate --cluster-id k3s-abc12345` that copies data from `kv/data/rke2/k3s-abc12345/` to `kv/data/k3s/k3s-abc12345/` and deletes the old path. Only needed if the user wants a clean vault namespace.

**Recommendation**: Option 1 for now. Document the artifact in release notes. Add migration later if users request it.

---

## Verification Checklist

After implementation:
- [ ] `go build ./...` compiles cleanly
- [ ] `go vet ./...` passes
- [ ] Unit tests pass: `go test ./pkg/vault/... ./pkg/rke2/... ./pkg/k3s/... ./pkg/lb/...`
- [ ] Integration tests pass: `go test -tags integration ./pkg/vault/...`
- [ ] New integration test confirms K3s secrets land at `kv/data/k3s/...`
- [ ] New integration test confirms RKE2 secrets still land at `kv/data/rke2/...`
- [ ] `rg 'kv/data/rke2|kv/metadata/rke2' --type go` returns 0 matches (all hardcoded paths replaced)
- [ ] Docs updated
