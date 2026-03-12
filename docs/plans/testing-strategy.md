# Testing Strategy for edgectl

Tracking issue: [#35 — Automated unit / integration testing](https://github.com/michielvha/edgectl/issues/35)

> **Note:** Phase 2 (SecretStore interface) also lays the groundwork for
> [#30 — Make the secret backend modular](https://github.com/michielvha/edgectl/issues/30).
> Once all consumers accept `vault.SecretStore` instead of `*vault.Client`,
> swapping OpenBao for Infisical (or any other backend) is just a new struct
> implementing the same interface — no changes needed in `pkg/lb/`, `pkg/rke2/`, or `cmd/`.

---

## Phase 1: Pure Function Tests (no refactoring needed)

### New files
- **`pkg/lb/handler_test.go`** — Test `generateHAProxyConfig()` and `generateKeepalivedConfig()`
  - HAProxy: verify server lines for 6443/9345 backends with hostIPs map (no DNS hit), empty hosts, multiple hosts
  - Keepalived: verify MASTER/BACKUP state, priority 200/100, interface and VIP substitution
  - `addServersToBackend()` (unexported, accessible from same package) — test hostIPs lookup path

- **`pkg/vault/rke2_server_test.go`** — Test `getFirstMasterIP()`
  - Empty hosts returns currentIP
  - First host has entry in hostIPs map — returns that IP
  - First host missing from hostIPs — returns hostname as fallback

### Modified files
- **`makefile`** — Add `test` and `test-cover` targets

---

## Phase 2: Vault Interface + Dependency Injection

### Key design decision
One `SecretStore` interface (not many small ones) — all consumers use methods across multiple vault domains. ~15 methods total. The existing `*Client` struct already satisfies this interface implicitly.

### New files
- **`pkg/vault/interface.go`** — Define `SecretStore` interface (all public methods on `*Client`)
- **`pkg/vault/mock_store.go`** — Hand-written mock with function fields for each method

### Modified files (signature changes — add `vault.SecretStore` as first param)
- **`pkg/lb/handler.go`** — `CreateLoadBalancer(store, ...)`, `GetStatus(store, ...)`, `CleanupLoadBalancer(store, ...)`, `BootstrapLBFromSecretStore(store, ...)`
- **`pkg/rke2/server/install.go`** — `Install(store, ...)`, `FetchTokenFromSecretStore(store, ...)`
- **`pkg/rke2/agent/install.go`** — `Install(store, ...)`, `FetchToken(store, ...)`
- **`cmd/rke2/server/commands.go`** — Pass `vault.InitVaultClient()` to pkg functions
- **`cmd/rke2/agent/commands.go`** — Same
- **`cmd/rke2/lb/commands.go`** — Same
- **`cmd/rke2/system/commands.go`** — Same (kubeconfig/purge)
- **`cmd/secrets.go`** — Same

### New test files
- **`pkg/lb/handler_test.go`** (expand) — Test `CreateLoadBalancer` VIP priority logic with mock store
- **`pkg/rke2/server/install_test.go`** — Test host-list deduplication, VIP resolution with mock store
- **`pkg/rke2/agent/install_test.go`** — Test VIP resolution priority (store > flag > DNS) with mock store

---

## Phase 3: Integration Tests with Real OpenBao

### New files
- **`pkg/vault/integration_test.go`** (`//go:build integration`)
  - Uses `testcontainers-go` to spin up OpenBao in dev mode
  - Tests: generic CRUD round-trip, token store/retrieve, master info accumulation,
    kubeconfig VIP replacement, LB info with main/backup nodes, `DeleteClusterData` full cleanup

### Modified files
- **`go.mod`** — Add `testcontainers-go` as test dependency
- **`makefile`** — Add `test-integration` target

---

## Phase 4: DNS Abstraction (optional, low priority)

### Modified files
- **`pkg/lb/handler.go`** — Extract `var lookupIP = net.LookupIP` package-level var
- **`pkg/vault/rke2_server.go`** — Extract `var lookupHost = net.LookupHost` package-level var

### Expand test files
- Test `addServersToBackend` DNS fallback path with injected lookupIP
- Test `getHostIP` with injected lookupHost

---

## Design Decisions

1. **One interface (`SecretStore`), not many** — All consumers use methods across multiple domains. Splitting adds complexity without benefit at this scale.
2. **Interface lives in `pkg/vault/interface.go`** — Slightly non-idiomatic Go (interfaces usually live at the consumer), but practical since every consumer needs the same interface.
3. **Hand-written mock, not generated** — Keeps dependencies minimal. Switch to `moq`/`mockgen` if it becomes a burden.
4. **Consumers accept `vault.SecretStore` as a parameter** (DI via function arguments) — No need for structs/constructors, just add the interface as the first parameter.
5. **Integration tests use build tags** — `//go:build integration` keeps them out of `go test ./...`. CI runs them separately with Docker.
6. **Don't unit test thin vault wrappers** — `StoreJoinToken`, `RetrieveJoinToken`, etc. are 2-3 lines of path construction. Integration tests cover them better.

---

## Verification
```sh
# Phase 1 + 2 (unit tests)
go test ./... -v

# Phase 3 (integration, requires Docker)
go test ./pkg/vault/ -tags=integration -v -count=1
```
