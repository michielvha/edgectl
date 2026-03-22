# Project Setup

This document consolidates the initial project bootstrap, Cobra CLI setup, and GitVersion configuration.

---

## Initial Setup

### Prerequisites

- Go 1.24+
- [Cobra CLI](https://github.com/spf13/cobra-cli) (for scaffolding new commands)
- [GitVersion](https://gitversion.net/) (for semantic versioning)

### Install & run

```bash
go install github.com/michielvha/edgectl@latest
edgectl version
```

### Build from source

```bash
go build -ldflags "-X main.Version=1.2.3 -X main.Commit=abcd1234" -o edgectl
./edgectl version
```

---

## Cobra CLI

Cobra is the CLI framework used for edgectl. The project was scaffolded with:

```bash
go install github.com/spf13/cobra-cli@latest
go mod init github.com/michielvha/edgectl
cobra-cli init
cobra-cli add version
```

### Key concepts

- **`rootCmd`** (`cmd/root.go`) — the base command (`edgectl`), defines persistent and local flags
- **`Execute()`** — parses user input and dispatches to the correct subcommand
- **Subcommands** are added via `rootCmd.AddCommand()` in each command file's `init()` function
- **Persistent flags** are global (available to all subcommands); **local flags** are scoped to a single command

### Adding a new command

```bash
cobra-cli add <command-name>
```

This generates a new file in `cmd/` with the boilerplate. Edit the `Run` function to implement the command logic.

---

## GitVersion

> Based on GitVersion 6.x

GitVersion automatically generates semantic version numbers from Git history and branch structure.

### Configuration

The project uses `gitversion.yml` in the repository root with the **GitHubFlow** workflow:

```yaml
workflow: GitHubFlow/v1
strategies:
  - MergeMessage
  - TaggedCommit
  - TrackReleaseBranches
  - VersionInBranchName

branches:
  main:
    regex: ^master$|^main$
    increment: Patch
    mode: ContinuousDeployment
    is-main-branch: true
  release:
    regex: ^release/(?<BranchName>[0-9]+\.[0-9]+\.[0-9]+)$
    label: ''
    increment: None
    is-release-branch: true
    mode: ContinuousDeployment
    source-branches:
      - main

assembly-versioning-scheme: MajorMinorPatch
```

### Version bumping

Add these tags to commit messages to control version increments:

| Tag | Effect |
|-----|--------|
| `+semver: major` or `+semver: breaking` | Bump major version |
| `+semver: minor` or `+semver: feature` | Bump minor version |
| `+semver: patch` or `+semver: fix` | Bump patch version |
| `+semver: none` or `+semver: skip` | No version bump |

### Versioning strategy

- **Patch**: Runtime fixes or small adjustments (bug fixes, performance improvements)
- **Minor**: New features or backward-compatible changes
- **Major**: Breaking changes that are not backward-compatible

### CI/CD integration

GitVersion is integrated into the [release workflow](../../.github/workflows/binary-release.yaml) using the [gitversion-tag-action](https://github.com/michielvha/gitversion-tag-action).

### Usage

```bash
gitversion              # Show current version
gitversion /showconfig  # Show full configuration
```

### References

- [GitVersion Documentation](https://gitversion.net/docs/)
- [Configuration Options](https://gitversion.net/docs/reference/configuration)
- [Version Strategies](https://gitversion.net/docs/reference/version-increments)
