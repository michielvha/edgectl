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

## features
**TODO: Provide proper list of commands & purposes**

- Auto Create an edge cloud kubernetes cluster powered by rke2
    - Fetch & Add kubeconfig file to current context
    - Auto setup ArgoCD - Add helm chart to directory on host with rke2 helm integration
    - Automated lifecycle management of edge cloud
- Use Cobra for cli

## Use edgectl

install the cli with the following commands:
```shell
go install github.com/michielvha/edgectl@latest
edgectl version
```

## Changelog & TODO


- [ ] [#31](https://github.com/michielvha/edgectl/issues/31) - Add support for fedora based architectures.
- [ ] write some kind of var or file to determine which system is which role ( server, agent, lb ) or find another way. VIPER?
- [ ] [#35](https://github.com/michielvha/edgectl/issues/35) - Create unit/integration tests file so we can test before release
  
### Logging

- Integrated viper for environment variables and config file support
  - [ ] Create logic to allow Adding ClusterID to the config file so you don't have to manually specify it
  - [ ] Store Cluster-ID & role in `/etc/environment` so it's available everywhere.
- using viper for global flags.
  - [ ] Add support for `--dry-run` to all commands
  
### managed rke2

- [ ] Auto Bootstrap ArgoCD for automated dev setup
  - [ ] Add helm chart to directory on host with rke2 helm integration
- [ ] Some kind of debug command that will verify connectivity etc, when an install fails..?
- [ ] optionally Auto install rancher as a management interface.

### Secret Management

Hashicorp vault is not able to be provided by us as a managed service because of it's license. We can keep it as a bring your own vault thing.

For a fully managed service look into [infisical](https://github.com/Infisical/infisical?tab=License-1-ov-file) via the [infisical go sdk](https://infisical.com/docs/sdks/languages/go).

We'll have to redesign the code with an interface to easily allow for bringing your own secretManagement tool.

---

# Done features

Check what should be retained in the features list & remove everything else:

- [x] Create pipeline to auto release with goreleaser.
- [x] Create version command using cobra and the variable is dynamically set at build time, in pipeline this is integrated with GitVersion.
- [x] Ensure scalable file structure.
- [x] Create commands to call bash scripts for admin tasks, rke2 install etc.
- [x] Integrate HashiCorp Vault for secret management. 
  - [x] Auto save & fetch secrets to Add agents to workers automatically
  - [x] add some kind of clusterID generation to be able to tell what to join with what.. I'm thinking based of hostname and then handle the hostname per customer. so create an id on master creation if cluster id provided don't create new one. Always ask for the cluster id when joining a worker, all other logic can be handled based of that in the background.
- [x] Fetch kubeconfig automatically. like in ``azure-cli``
- [x] update gitVersion to be like chartFetch with release branch strategy
-  Changed to a logging library (zerolog) and wrote log.go pkg to support log levels and better logging. Supporting a `--verbose` flag to enable debug logging. **We'll only use `logger.debug` for debug logging everything else will be stdout on cli.**

---


[![Go Doc](https://pkg.go.dev/badge/github.com/michielvha/edgectl.svg)](https://pkg.go.dev/github.com/michielvha/edgectl)
[![license](https://img.shields.io/github/license/michielvha/edgectl.svg?style=flat-square)](LICENSE)

[Go Report Card]: https://goreportcard.com/badge/github.com/michielvha/edgectl
[go-report-card]: https://goreportcard.com/report/github.com/michielvha/edgectl
[CodeQL]: https://github.com/michielvha/edgectl/actions/workflows/github-code-scanning/codeql/badge.svg?branch=main
[code-ql]: https://github.com/michielvha/edgectl/actions/workflows/github-code-scanning/codeql
[codecov]: https://codecov.io/gh/michielvha/edgectl/branch/main/graph/badge.svg
[code-cov]: https://codecov.io/gh/michielvha/edgectl
