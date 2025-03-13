# edge-cli


So we write a CLI like thing in go like azure cli that does stuff like

auto add agents to workers
auto fetch and add kubeconfig to current context for remote admin
we use our existing bash the go is more of a wrapper
auto bootstrap ArgoCD etc, this binary can then be used in a pipeline or something for a 1 click installed devEnv with ArgoCD
it's all about automated lifecycle just like awscli & azure cli

- Use Cobra for cli
- Integrate with hashicorp vault for secret management

## Changelog & TODO

- [x] Create pipeline to auto release with goreleaser.
- [x] Create version command using cobra and the variable is dynamically set at build time, in pipeline this is integrated with GitVersion.

- [ ] Create commands to call bash scripts for admin tasks, rke2 install etc.
- [ ] Integrate HashiCorp Vault for secret management. Auto save & fetch secrets to Add agents to workers automatically
- [ ] Fetch kubeconfig automatically. like in azurecli
- [ ] Bootstrap ArgoCD via edgecli