# K3s Support Plan

> **Status: ✅ COMPLETE** — Implemented on branch `feat/k3s-support` (GitHub Issue #52). All code, tests, and shell scripts are merged and passing.

## Overview

Add K3s as a first-class distribution alongside RKE2. Both are Rancher Kubernetes distributions with nearly identical architectures — same config.yaml format, same Vault/secret store patterns, same join flow — but differ in paths, ports, service names, and install scripts.

## Architecture Comparison

| Aspect | RKE2 | K3s |
|--------|------|-----|
| Install script URL | `https://get.rke2.io` | `https://get.k3s.io` |
| Config path | `/etc/rancher/rke2/config.yaml` | `/etc/rancher/k3s/config.yaml` |
| Data dir | `/var/lib/rancher/rke2/` | `/var/lib/rancher/k3s/` |
| Node token path | `/var/lib/rancher/rke2/server/node-token` | `/var/lib/rancher/k3s/server/node-token` |
| Kubeconfig path | `/etc/rancher/rke2/rke2.yaml` | `/etc/rancher/k3s/k3s.yaml` |
| Server service | `rke2-server.service` | `k3s.service` |
| Agent service | `rke2-agent.service` | `k3s-agent.service` |
| Token env var | `RKE2_TOKEN` | `K3S_TOKEN` |
| Supervisor port | 9345 (separate) | 6443 (shared with API) |
| Uninstall script | `/usr/local/bin/rke2-uninstall.sh` | `/usr/local/bin/k3s-uninstall.sh` |
| Binary dir (PATH) | `/var/lib/rancher/rke2/bin` | Installs to `/usr/local/bin` directly |
| Profile script | `/etc/profile.d/rke2.sh` | Not needed (already in PATH) |
| Manifests dir | `/var/lib/rancher/rke2/server/manifests/` | `/var/lib/rancher/k3s/server/manifests/` |
| CIS hardening | Has CIS profile w/ sysctl | Not applicable |

**Key difference:** K3s has no supervisor port (9345). The API server on 6443 handles both. This means the HAProxy config for K3s omits the `rke2-supervisor-frontend/backend` block entirely.

## Implementation Strategy

The most effective approach is to **mirror the existing `rke2` structure** with a parallel `k3s` tree while extracting shared logic where possible. This minimizes risk and keeps the codebase navigable.

### Step 1 — Embedded shell scripts (`pkg/common/scripts/`) ✅

Created 4 new scripts mirroring the rke2 ones:

| New script | Based on | Key changes |
|-----------|----------|-------------|
| `k3s.sh` | `rke2.sh` | Paths, install URL, service names, no CIS hardening, no supervisor port |
| `k3s-bash.sh` | `rke2-bash.sh` | Not needed for K3s (binary already in PATH); make it a no-op or skip |
| `k3s-purge.sh` | `rke2-purge.sh` | Uninstall script path, service names |
| `k3s-status.sh` | `rke2-status.sh` | Service names (`k3s.service`, `k3s-agent.service`) |

### Step 2 — Package layer (`pkg/k3s/`) ✅

```
pkg/k3s/
├── server/
│   ├── install.go       # Mirrors pkg/rke2/server/install.go
│   └── install_test.go
└── agent/
    ├── install.go        # Mirrors pkg/rke2/agent/install.go
    └── install_test.go
```

Changes from rke2 equivalent (all implemented):
- `install.go` (server): Cluster ID prefix `k3s-{uuid}`, script `k3s.sh`/`install_k3s_server`, token path `/var/lib/rancher/k3s/server/node-token`, kubeconfig `/etc/rancher/k3s/k3s.yaml`, env vars `K3S_TOKEN`/`K3S_URL` (port 6443, not 9345)
- `install.go` (agent): Script `k3s.sh`/`install_k3s_agent`, env var `K3S_TOKEN`
- Tests: `install_test.go` for both server and agent packages

Everything else (vault operations, VIP logic, hostname lookup) is identical and calls the same `SecretStore` interface methods.

### Step 3 — Command layer (`cmd/k3s/`) ✅

```
cmd/
├── k3s.go                    # Top-level "edgectl k3s" command
└── k3s/
    ├── server/commands.go    # edgectl k3s server install [--cluster-id] [--vip]
    ├── agent/commands.go     # edgectl k3s agent install --cluster-id [--vip]
    ├── lb/commands.go        # edgectl k3s lb create/status --cluster-id [--vip]
    └── system/commands.go    # edgectl k3s system status/purge/config/bash
```

These are copies of `cmd/rke2/*` updated to:
- Call `k3s.Install()` instead of `rke2.Install()`
- Reference k3s scripts (`k3s-status.sh`, `k3s-purge.sh`)
- Update help text from "RKE2" to "K3s"
- Import aliases use `k3s` prefix (`k3sagentcmd`, `k3slbcmd`, etc.) to avoid Go package-level conflicts with rke2 aliases

### Step 4 — Load balancer config (`pkg/lb/handler.go`) ✅

The LB handler needs a **distribution-aware** HAProxy config. Two options:

**Implemented Option A:** Added `distro` parameter to `generateHAProxyConfig()`, `CreateLoadBalancer()`, and `BootstrapLBFromSecretStore()`. The `LoadBalancerConfig` struct has a `Distro string` field. HAProxy config conditionally includes supervisor frontend/backend on port 9345 only for RKE2. Frontend/backend names use generic `k8s-api-frontend`/`k8s-api-backend`.

### Step 5 — Vault paths (distro-aware prefix) ✅

Vault paths now use a `distro` parameter for clean path separation:
```
kv/data/rke2/{clusterID}/token    ← RKE2 clusters
kv/data/k3s/{clusterID}/token     ← K3s clusters
```

All `SecretStore` interface methods accept `distro` as the first parameter. The `rke2_*.go` vault files were renamed to distro-agnostic names (`token.go`, `server.go`, etc.). See `docs/plans/vault-path-distro-prefix.md` for the full implementation report.

### Step 6 — Root command registration (`cmd/root.go`) ✅

Added `k3sCmd` via `cmd/k3s.go` `init()` function calling `rootCmd.AddCommand(k3sCmd)`.

### Step 7 — Tests ✅

- ✅ Created `pkg/k3s/server/install_test.go` — tests `K3S_TOKEN` and `K3S_URL` env var setting
- ✅ Created `pkg/k3s/agent/install_test.go` — tests VIP priority and DNS resolution
- ✅ Updated `pkg/lb/handler_test.go` — renamed tests to `TestGenerateHAProxyConfig_RKE2_*`, added `TestGenerateHAProxyConfig_K3s_NoSupervisorPort` and `TestGenerateHAProxyConfig_K3s_SingleHost`
- ✅ All existing `pkg/vault/integration_test.go` tests work unchanged

## CLI UX

```bash
# K3s commands (mirror rke2)
edgectl k3s server install [--vip <ip>]
edgectl k3s server install --cluster-id <id> [--vip <ip>]
edgectl k3s agent install --cluster-id <id> [--vip <ip>]
edgectl k3s lb create --cluster-id <id> [--vip <ip>]
edgectl k3s lb status --cluster-id <id>
edgectl k3s system status
edgectl k3s system purge [--cluster-id <id>]
edgectl k3s system config
edgectl k3s system bash
```

## File Change Summary

| Category | New files | Modified files |
|----------|-----------|----------------|
| Shell scripts | 4 (`k3s.sh`, `k3s-bash.sh`, `k3s-purge.sh`, `k3s-status.sh`) | 0 |
| Go packages | 4 (`pkg/k3s/server/install.go`, `install_test.go`, `pkg/k3s/agent/install.go`, `install_test.go`) | 0 |
| Commands | 5 (`cmd/k3s.go`, `cmd/k3s/server/commands.go`, `cmd/k3s/agent/commands.go`, `cmd/k3s/lb/commands.go`, `cmd/k3s/system/commands.go`) | 1 (`cmd/root.go`) |
| LB handler | 0 | 1 (`pkg/lb/handler.go` — add distro param) |
| Tests | 2 | 1 (`pkg/lb/handler_test.go`) |
| **Total** | **15** | **3** |

## Execution Order

1. Shell scripts (independent, no Go changes)
2. `pkg/k3s/` packages + tests
3. LB handler distro parameter
4. `cmd/k3s/` commands + `cmd/k3s.go`
5. Register in `cmd/root.go`
6. Update docs (`docs/user/rke2.md` → consider a `docs/user/k3s.md`)

## Future Optimization (Optional)

After both rke2 and k3s work, consider extracting shared logic:
- A `pkg/distro` package with a `Distro` interface defining paths, service names, install script, etc.
- Single `pkg/cluster/server/install.go` that accepts a `Distro` config
- This would reduce duplication but adds abstraction complexity — only worth it if a third distro is planned (e.g., vanilla kubeadm)

## Implementation Notes

### Lessons Learned
- **Import aliases must be unique per package:** `cmd/k3s.go` and `cmd/rke2.go` both live in the `cmd` package, so import aliases like `servercmd` would conflict. K3s commands use `k3s`-prefixed aliases (`k3sagentcmd`, `k3slbcmd`, `k3sservercmd`, `k3ssystemcmd`).
- **K3s bash script is simpler:** K3s installs its binary to `/usr/local/bin` directly, so `k3s-bash.sh` only sets `KUBECONFIG` (no `PATH` manipulation needed, unlike RKE2).
- **K3s service naming:** K3s uses `k3s.service` for server and `k3s-agent.service` for agent (no `k3s-server.service`).

### Final File Count
| Category | New files | Modified files |
|----------|-----------|----------------|
| Shell scripts | 4 | 0 |
| Go packages + tests | 4 | 0 |
| Commands | 5 | 0 |
| LB handler | 0 | 2 (`handler.go`, `handler_test.go`) |
| RKE2 commands | 0 | 1 (`cmd/rke2/lb/commands.go` — added distro param) |
| **Total** | **13** | **3** |
