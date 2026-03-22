# How to Lint

We're using **golangci-lint** in our pipeline to integrate linting into our workflow. This will help us catch errors early and keep our codebase clean. You'll want to manually run **golangci-lint** and fix any linting errors before commiting to main.

## Quick Start

```shell
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
# brew install golangci-lint
golangci-lint run
golangci-lint run --fix     # Try this first to autofix
gofmt -w .                  # Then fix remaining formatting
```

## Extra Linters

```shell
go install mvdan.cc/gofumpt@latest
go install github.com/segmentio/golines@latest
golines --max-len=100 --base-formatter=gofumpt -w
```

## Pre-Commit Setup

To automatically check your code before committing, you can set up pre-commit hooks:

### Option 1: Git Hooks (manual setup)

Create a file `.git/hooks/pre-commit` with:

```bash
#!/bin/bash
set -e

# Run golangci-lint
echo "Running golangci-lint..."
golangci-lint run

# If we got here, it passed
echo "✅ Linting successful!"
```

Make it executable:

```shell
chmod +x .git/hooks/pre-commit
```

### Option 2: Using pre-commit Framework

1. Install pre-commit: 
   ```shell
   # macOS
   brew install pre-commit
   
   # pip
   pip install pre-commit
   ```

2. Create `.pre-commit-config.yaml` in the project root:

```yaml
repos:
-   repo: https://github.com/tekwizely/pre-commit-golang
    rev: v1.0.0-rc.1
    hooks:
    -   id: go-mod-tidy
    -   id: golangci-lint
    -   id: go-test-mod
```

3. Install the hooks:
```shell
pre-commit install
```

## Common Linting Errors and Fixes

Here are some common linting errors and how to fix them:

### errcheck: Unchecked Errors
Always check error returns from functions:

```go
// Bad
viper.BindPFlag("flag", cmd.Flags().Lookup("flag"))

// Good
if err := viper.BindPFlag("flag", cmd.Flags().Lookup("flag")); err != nil {
    log.Printf("error binding flag: %v", err)
}
```

### gocritic: Use More Efficient Functions
Use the most appropriate function for your needs:

```go
// Bad
strings.Replace(s, old, new, -1)

// Good
strings.ReplaceAll(s, old, new)
```

### staticcheck: Empty Branches
Don't leave empty if/else branches:

```go
// Bad
if err := doSomething(); err == nil {
    // Nothing here
}

// Good
if err := doSomething(); err == nil {
    log.Println("Operation successful")
}
```

## Configuring golangci-lint

Our project uses a `.golangci.yml` configuration file to customize which linters are enabled and configure their behavior. See the [official documentation](https://golangci-lint.run/usage/configuration/) for details.

## VSCode Integration

To get real-time linting in VSCode:

1. Install the Go extension
2. Add these settings to your `settings.json`:

```json
{
    "go.lintTool": "golangci-lint",
    "go.lintFlags": [
        "--fast"
    ]
}
```