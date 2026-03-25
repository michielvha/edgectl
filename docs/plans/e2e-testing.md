# E2E Testing Strategy for edgectl

Tracking issue: [#35 — Automated unit / integration testing](https://github.com/michielvha/edgectl/issues/35)

> **Context:** Unit tests (Phase 1–4) verify Go logic with mocks. Integration tests verify
> vault operations with a real OpenBao container. Neither validates that the actual install
> scripts work — downloading K3s/RKE2, configuring firewall rules, writing systemd units,
> and producing a healthy cluster node. E2E tests close this gap.

---

## Goals

1. Prove `edgectl k3s server install` and `edgectl rke2 server install` produce a running cluster node
2. Verify firewall rules are applied correctly per distro (UFW on Ubuntu, firewalld on Fedora/Rocky)
3. Verify `edgectl <distro> system status/purge/bash/kubeconfig` commands work against a real install
4. Verify load balancer config generation excludes supervisor port 9345 for K3s
5. Run automatically in CI on PRs that touch install-related code

---

## Phase 1: K3s E2E on GitHub Actions (ubuntu-latest)

K3s is lightweight and installs reliably on GitHub Actions runners. This is the quickest win.

### Workflow: `.github/workflows/e2e-test.yaml`

```yaml
name: E2E Tests

on:
  pull_request:
    branches: [main]
    paths:
      - 'pkg/k3s/**'
      - 'pkg/lb/**'
      - 'pkg/common/scripts/**'
      - 'cmd/k3s/**'
  workflow_dispatch:  # manual trigger for ad-hoc testing

jobs:
  k3s-ubuntu:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod

      - name: Build edgectl
        run: go build -o edgectl .

      - name: Install K3s server (skips vault)
        run: sudo ./edgectl k3s server install --vip 127.0.0.1 --skip-secret-store
        timeout-minutes: 5

      - name: Verify K3s service is running
        run: |
          sudo systemctl is-active k3s.service
          sudo k3s kubectl get nodes --no-headers | grep -q "Ready"

      - name: Verify firewall rules (UFW)
        run: |
          sudo ufw status | grep -q "6443"
          sudo ufw status | grep -q "10250"
          sudo ufw status | grep -q "2379"

      - name: Test system status command
        run: sudo ./edgectl k3s system status

      - name: Test system bash command
        run: sudo ./edgectl k3s system bash

      - name: Test system purge command
        run: sudo ./edgectl k3s system purge

      - name: Verify purge cleaned up
        run: |
          ! systemctl is-active k3s.service 2>/dev/null
          ! test -f /usr/local/bin/k3s
```

### Required code change: `--skip-secret-store` flag

The current `server install` flow requires a running OpenBao/Vault instance. For E2E tests
we need a `--skip-secret-store` flag that:
- Still runs the full install script (download, configure, firewall, systemd)
- Skips token/kubeconfig storage in the secret store
- Prints the generated cluster-id and node-token to stdout instead

This flag is also useful for users who want to manage secrets manually.

---

## Phase 2: RKE2 E2E on GitHub Actions

Same approach as K3s but for RKE2. RKE2 is heavier (downloads containerd + kubelet binaries)
so the job will take longer (~3-5 min).

```yaml
  rke2-ubuntu:
    runs-on: ubuntu-latest
    steps:
      # same pattern as k3s-ubuntu but with:
      - name: Install RKE2 server
        run: sudo ./edgectl rke2 server install --vip 127.0.0.1 --skip-secret-store
        timeout-minutes: 10

      - name: Verify RKE2 service
        run: |
          sudo systemctl is-active rke2-server.service
          sudo /var/lib/rancher/rke2/bin/kubectl --kubeconfig /etc/rancher/rke2/rke2.yaml get nodes | grep -q "Ready"

      - name: Verify firewall rules include 9345 (supervisor)
        run: |
          sudo ufw status | grep -q "9345"
          sudo ufw status | grep -q "6443"

      - name: Test system commands
        run: |
          sudo ./edgectl rke2 system status
          sudo ./edgectl rke2 system bash
          sudo ./edgectl rke2 system purge
```

---

## Phase 3: Firewalld testing (Fedora/Rocky)

GitHub Actions `ubuntu-latest` only covers UFW. For firewalld we need a different approach.

### Option A: Container-based (recommended)

Use a Fedora/Rocky container with systemd. Firewalld can run inside a privileged container.

```yaml
  k3s-fedora-firewalld:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6

      - name: Build edgectl
        run: |
          docker build -t edgectl-e2e-fedora -f tests/e2e/Dockerfile.fedora .

      - name: Run K3s install in Fedora container
        run: |
          docker run --privileged \
            --cgroupns=host \
            -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
            edgectl-e2e-fedora \
            /bin/bash -c "./edgectl k3s server install --vip 127.0.0.1 --skip-secret-store && \
              firewall-cmd --list-ports | grep -q 6443 && \
              systemctl is-active k3s.service"
```

With a corresponding `tests/e2e/Dockerfile.fedora`:
```dockerfile
FROM fedora:latest
RUN dnf install -y systemd firewalld curl
COPY edgectl /usr/local/bin/edgectl
CMD ["/sbin/init"]
```

### Option B: Self-hosted runner

If you have a Fedora/Rocky VM available, register it as a self-hosted GitHub Actions runner.
This gives full systemd and firewalld support without container hacks.

---

## Phase 4: Full cluster E2E (multi-node)

For testing agent join and load balancer creation, we need multiple nodes.
This is the most complex phase and could use:

- **k3d/kind** for lightweight multi-node simulation (limited — doesn't test our install scripts)
- **Multipass/Vagrant** on a self-hosted runner to spin up 3 VMs
- **Cloud VMs** (Azure/AWS) spun up via Terraform in CI, torn down after

### Test matrix:
```
| Test                        | Nodes | What it validates                              |
|-----------------------------|-------|------------------------------------------------|
| K3s server + agent join     | 2     | Token retrieval, agent install, node joins     |
| K3s 3-node HA + LB         | 3+1   | LB config (no 9345), VIP, HAProxy health check |
| RKE2 server + agent join    | 2     | Token retrieval, agent install, node joins     |
| RKE2 3-node HA + LB        | 3+1   | LB config (with 9345), VIP, keepalived         |
```

This phase is expensive — run on `workflow_dispatch` or nightly, not on every PR.

---

## Implementation Priority

1. **Phase 1** — Highest ROI. K3s on ubuntu-latest catches most regressions. Requires `--skip-secret-store` flag.
2. **Phase 2** — Straightforward extension of Phase 1 for RKE2.
3. **Phase 3** — Important for firewalld coverage but container setup is trickier.
4. **Phase 4** — Nice to have. Only needed when changing multi-node logic.

---

## Design Decisions

1. **`--skip-secret-store` flag** — Decouples install testing from vault availability. Also useful for users.
2. **Path-triggered** — E2E only runs when install-related code changes, not on doc edits.
3. **Privileged containers for firewalld** — Avoids needing non-Ubuntu runners while still testing firewalld.
4. **Timeout guards** — Install jobs get generous timeouts (5-10 min) to handle slow downloads.
5. **Purge as final step** — Validates cleanup works and leaves the runner clean.
