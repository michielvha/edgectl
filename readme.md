# edge-cli

A CLI tool to manage the edge cloud. Comparable to `awscli` or `azure-cli`.

## features

- Auto Create an edge cloud kubernetes cluster powered by rke2
    - Fetch & Add kubeconfig file to current context
    - Auto setup ArgoCD - Add helmchart to directory on host with rke2 helm integration
    - Automated lifecycle management of edge cloud
- Use Cobra for cli

## Use edgectl

install the cli with the following commands:
```shell
go install github.com/michielvha/edgectl@latest
edgectl version
```

## Changelog & TODO

- [x] Create pipeline to auto release with goreleaser.
- [x] Create version command using cobra and the variable is dynamically set at build time, in pipeline this is integrated with GitVersion.

- [ ] Change to a logging library to support log levels and better logging. Support an `--debug / --verbose` flag to enable debug logging.

- [ ] Create commands to call bash scripts for admin tasks, rke2 install etc.
- [ ] Integrate HashiCorp Vault for secret management. Auto save & fetch secrets to Add agents to workers automatically
  - [ ] add some kind of clusterID generation to be able to tell what to join with what.. I'm thinking based of hostname and then handle the hostname per customer.
  - so create an id on master creation. Always ask for the cluster id when joining a worker, all other logic can be handled based of that in the background.
- [ ] Fetch kubeconfig automatically. like in ``azure-cli``

- [ ] Auto Bootstrap ArgoCD for automated dev setup
  - [ ] Add helm chart to directory on host with rke2 helm integration

- [ ] Integrate viper for environment variables and config file support
- [ ] use viper for global flags.
  - [ ] Add support for `--dry-run` to all commands

- [x] update gitVersion to be like chartFetch with release branch strategy



- [ ] rewrite the file structure to be like this 

```
edgectl/
├── cmd/                   # All cobra commands live here
│   ├── root.go            # Entry point, adds subcommands
│   ├── version.go         # `edgectl version`
│   ├── server/            # Subcommand group (optional, for modularity)
│   │   ├── install.go     # `edgectl server install`
│   │   └── purge.go       # `edgectl server purge`
│   └── agent/             # Another group if needed
├── scripts/               # Embedded or external bash scripts
├── pkg/                   # Public libraries usable outside CLI
│   ├── vault/             # Logic to interact with Vault
│   ├── common/            # Common handlers
│   ├── rke2/              # rke2 handler
├── go.mod
├── main.go                # Just sets version and calls cmd.Execute()
└── README.md
```