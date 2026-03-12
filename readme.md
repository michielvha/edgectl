# edge-cli

[![Build and Release](https://github.com/michielvha/edgectl/actions/workflows/binary-release.yaml/badge.svg)](https://github.com/michielvha/edgectl/actions/workflows/binary-release.yaml)
[![Release](https://img.shields.io/github/release/michielvha/edgectl.svg?style=flat-square)](https://github.com/michielvha/edgectl/releases/latest)
[![Go Report Card]][go-report-card]

<!-- CodeQL]][code-ql] -->
<!-- [![codecov]][code-cov] When we have testing we should include something like codecov to help scan -->

<div align="center">
  <img src="./docs/edge-cloud.png" alt="EdgeCloud Logo" width="250"/>
</div>

A CLI tool to manage the edge cloud. Comparable to `awscli` or `azure-cli`.

## Features

- **RKE2 Cluster Management** — Bootstrap, join, and manage Kubernetes clusters powered by RKE2
  - Automated server and agent installation with embedded bash scripts
  - Cluster ID-based node joining (no manual token handling)
  - Fetch & merge kubeconfig into your local context
- **Secret Management** — Powered by [OpenBao](https://openbao.org/) (Linux Foundation fork of HashiCorp Vault, MPL-2.0)
  - Automatic token storage and retrieval for cluster operations
  - Generic `get`/`set` commands for ad-hoc secret management
- **Load Balancer** — Automated HAProxy + Keepalived setup for HA clusters
  - Primary/backup node configuration with VIP failover
  - Status monitoring and cleanup commands
- **Logging** — Structured logging with `zerolog`, `--verbose` flag for debug output
- **Cross-platform releases** — GoReleaser with Homebrew tap support

## Install

```bash
go install github.com/michielvha/edgectl@latest
edgectl version
```

Or download a pre-built binary from the [releases page](https://github.com/michielvha/edgectl/releases/latest).

## Quick Start

```bash
# Set up secret store connection
export BAO_ADDR="https://your-openbao-instance:8200"
export BAO_TOKEN="your-token"

# Bootstrap a new cluster
sudo edgectl rke2 server install --vip 172.16.12.232

# Join worker nodes
sudo edgectl rke2 agent install --cluster-id <cluster-id>

# Fetch kubeconfig
edgectl rke2 system kubeconfig --cluster-id <cluster-id>
```

## Documentation

See [docs/README.md](docs/README.md) for full documentation including architecture, development setup, and project plans.

## Roadmap

- [ ] Enable [encryption at rest](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/) for secrets
- [ ] [#31](https://github.com/michielvha/edgectl/issues/31) — Add support for Fedora-based architectures
- [ ] [#35](https://github.com/michielvha/edgectl/issues/35) — Create unit/integration tests
- [ ] Auto-bootstrap ArgoCD for automated dev setup
- [ ] Add `--dry-run` support to all commands
- [ ] Interface pattern for pluggable secret backends (infisical, AWS Secrets Manager, Azure Key Vault, etc.)

---

[![Go Doc](https://pkg.go.dev/badge/github.com/michielvha/edgectl.svg)](https://pkg.go.dev/github.com/michielvha/edgectl)
[![license](https://img.shields.io/github/license/michielvha/edgectl.svg?style=flat-square)](LICENSE)

[Go Report Card]: https://goreportcard.com/badge/github.com/michielvha/edgectl
[go-report-card]: https://goreportcard.com/report/github.com/michielvha/edgectl
[CodeQL]: https://github.com/michielvha/edgectl/actions/workflows/github-code-scanning/codeql/badge.svg?branch=main
[code-ql]: https://github.com/michielvha/edgectl/actions/workflows/github-code-scanning/codeql
[codecov]: https://codecov.io/gh/michielvha/edgectl/branch/main/graph/badge.svg
[code-cov]: https://codecov.io/gh/michielvha/edgectl
