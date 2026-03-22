# CLI Enhancements Plan

## Overview

Enhance the `edgectl` CLI experience using two complementary libraries:
1. **charmbracelet/huh** — Interactive terminal forms for guided workflows and confirmations
2. **Viper (expanded)** — Persistent user configuration to reduce repetitive flag usage

Both are already Go ecosystem staples. Viper is already integrated (for `--verbose` only); `huh` would be a new dependency.

## 1. Interactive Forms with charmbracelet/huh

### Library Summary

`huh` (v2, 6.7k stars) provides interactive terminal form components: `Input`, `Text`, `Select`, `MultiSelect`, `Confirm`, `FilePicker`, and `Spinner`. It supports dynamic forms (fields update based on prior input), built-in themes (Charm, Dracula, Catppuccin, Base16, Default), and a first-class accessible mode for screen readers.

### Use Cases for edgectl

#### A. Interactive Install Wizard

When `edgectl rke2 server install` or `edgectl k3s server install` is run **without flags**, present an interactive form instead of failing or using defaults blindly:

```go
import "charm.land/huh/v2"

var (
    distro    string
    clusterID string
    vip       string
    isNew     bool
)

form := huh.NewForm(
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("Select Kubernetes distribution").
            Options(
                huh.NewOption("K3s (lightweight)", "k3s"),
                huh.NewOption("RKE2 (hardened)", "rke2"),
            ).
            Value(&distro),

        huh.NewConfirm().
            Title("Create a new cluster?").
            Affirmative("Yes, new cluster").
            Negative("No, join existing").
            Value(&isNew),
    ),

    huh.NewGroup(
        huh.NewInput().
            Title("Cluster ID").
            Description("Enter the cluster ID to join").
            Value(&clusterID),
    ).WithHideFunc(func() bool { return isNew }), // skip if new cluster

    huh.NewGroup(
        huh.NewInput().
            Title("Virtual IP (optional)").
            Description("VIP for the load balancer / TLS SANs").
            Value(&vip),
    ),
)
```

This replaces the current pattern where users must know the exact flags. Flags would still work for scripting/automation (non-interactive mode).

#### B. Destructive Operation Confirmations

Add `huh.NewConfirm()` before destructive operations:

```go
// Before system purge
var confirmed bool
huh.NewConfirm().
    Title("⚠️  This will completely uninstall K3s and remove all data. Continue?").
    Affirmative("Yes, purge").
    Negative("Cancel").
    Value(&confirmed).
    Run()

if !confirmed {
    fmt.Println("Aborted.")
    return
}
```

Apply to:
- `edgectl {rke2,k3s} system purge`
- `edgectl {rke2,k3s} lb cleanup`

#### C. First-Time Config Setup

An `edgectl init` command that uses `huh` forms to create `$HOME/.edgectl.yaml`:

```go
form := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Title("OpenBao / Vault address").
            Description("e.g., https://vault.example.com:8200").
            Value(&vaultAddr).
            Validate(validateURL),

        huh.NewSelect[string]().
            Title("Default Kubernetes distribution").
            Options(
                huh.NewOption("K3s", "k3s"),
                huh.NewOption("RKE2", "rke2"),
            ).
            Value(&defaultDistro),

        huh.NewInput().
            Title("Default cluster ID (optional)").
            Value(&defaultClusterID),
    ),
)
```

#### D. Spinner for Long Operations

Wrap long-running operations (install scripts, vault lookups) with `huh/spinner`:

```go
import "charm.land/huh/v2/spinner"

err := spinner.New().
    Title("Installing K3s server...").
    Action(func() { err = server.Install(store, clusterID, isExisting, vip) }).
    Run()
```

### Implementation Approach

- Add `charm.land/huh/v2` as a dependency
- Create a `pkg/tui/` package with reusable form builders
- Detect interactive vs non-interactive (piped) terminal: skip forms when stdin is not a terminal
- Keep all flags working for automation — forms are the fallback when flags are absent

### Non-interactive Detection

```go
import "golang.org/x/term"

func isInteractive() bool {
    return term.IsTerminal(int(os.Stdin.Fd()))
}
```

When running in a pipeline or CI, `isInteractive()` returns false, and the CLI falls back to requiring flags as it does today.

## 2. Expanded Viper Configuration

### Current State

Viper is integrated in `cmd/root.go` but only binds the `verbose` flag. The config file `$HOME/.edgectl.yaml` is read but nothing beyond `verbose` is stored.

### Proposed Config Schema

```yaml
# $HOME/.edgectl.yaml
verbose: false

# Default distribution (used when distro-specific command is not given)
default-distro: k3s

# Default cluster ID (saves passing --cluster-id every time)
default-cluster-id: my-edge-cluster

# OpenBao / Vault configuration
vault:
  address: https://vault.example.com:8200
  # token is read from VAULT_TOKEN env var (never stored in config)

# Default VIP for lb commands
default-vip: 192.168.1.100
```

### Implementation Steps

1. **Bind new config keys to flags** in each command's `init()`:
   ```go
   func init() {
       createCmd.Flags().String("cluster-id", "", "Cluster ID")
       viper.BindPFlag("default-cluster-id", createCmd.Flags().Lookup("cluster-id"))
   }
   ```

2. **Add fallback logic** in command `Run` functions:
   ```go
   clusterID, _ := cmd.Flags().GetString("cluster-id")
   if clusterID == "" {
       clusterID = viper.GetString("default-cluster-id")
   }
   ```

3. **Vault address from config**:
   ```go
   // In vault.InitVaultClient()
   addr := os.Getenv("VAULT_ADDR")
   if addr == "" {
       addr = viper.GetString("vault.address")
   }
   ```

### Priority of Config Sources (Viper default)

1. CLI flags (highest)
2. Environment variables
3. Config file (`$HOME/.edgectl.yaml`)
4. Defaults (lowest)

## 3. New Commands

| Command | Purpose | Uses huh? |
|---------|---------|-----------|
| `edgectl init` | Interactive first-time setup, generates `$HOME/.edgectl.yaml` | Yes — full form |
| `edgectl config show` | Display current effective configuration | No |
| `edgectl config set <key> <value>` | Set a config value | No |

## Implementation Order

1. **Viper expansion** — add config keys, fallback logic (no new dependency)
2. **huh dependency** — `go get charm.land/huh/v2`
3. **Destructive confirmations** — add `Confirm` to purge/cleanup commands (quick wins)
4. **`edgectl init`** — interactive config setup
5. **Interactive install wizard** — fallback forms when flags are missing
6. **Spinners** — wrap long operations

## File Change Summary

| Category | New files | Modified files |
|----------|-----------|----------------|
| TUI package | 1 (`pkg/tui/forms.go`) | 0 |
| Commands | 2 (`cmd/init.go`, `cmd/config.go`) | 0 |
| Config expansion | 0 | 1 (`cmd/root.go`) |
| Confirmations | 0 | 4 (`cmd/{rke2,k3s}/system/commands.go`, `cmd/{rke2,k3s}/lb/commands.go`) |
| Install wizards | 0 | 4 (`cmd/{rke2,k3s}/{server,agent}/commands.go`) |
| Dependencies | 0 | 2 (`go.mod`, `go.sum`) |
| **Total** | **3** | **11** |

## Design Principles

- **Flags always work** — interactive forms supplement, never replace, flag-based usage
- **Non-interactive safe** — detect piped stdin and skip TUI elements
- **Config is optional** — everything works without `$HOME/.edgectl.yaml`
- **Progressive disclosure** — simple commands stay simple; complexity is behind `--help` or interactive prompts
