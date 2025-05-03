# 📦 Build and Release Pipeline Documentation

This document describes the structure, workflow, and key dependencies of the **Go binary Build & release pipeline** for the `edgectl` project.

## 🧪 Workflow Overview

The GitHub Actions workflow (`binary-release.yaml`) automates the build and release process of a Go Binary. It is triggered by:

- **Pushes** to the `main` branch (excluding changes to `docs/`, `readme.md`, `.gitignore`, etc.).
- **Manual workflow dispatches** via the GitHub UI.

### 👣 Workflow Steps

1. **Checkout Code**
   - Uses `actions/checkout@v4` with `fetch-depth: 0` to retrieve full history for version tagging.

2. **GitVersion Tagging**
   - Runs a custom GitHub Action (`michielvha/gitversion-tag-action@v3`) to compute semantic versioning using `gitversion.yml`.

3. **Setup Go**
   - Installs Go version `1.24.1` using `actions/setup-go@v5`.

4. **Linting**
   - Runs `golangci-lint` (`v1.64`) to check for common Go issues using a custom config (`.golangci.yml`).

5. **Cache GoReleaser**
   - Caches GoReleaser binary to speed up subsequent runs.

6. **PGP Key Import & Fingerprint Extraction**
   - Imports the GPG private key from a GitHub secret.
   - Extracts the fingerprint and sets it as an env variable for signing.

7. **Release with GoReleaser**
   - Runs `goreleaser/goreleaser-action@v6` to:
     - Build binaries across multiple OS/architecture targets.
     - Package, checksum, sign, and publish the release.

## 🧷 Dependent Files

### 1. `.goreleaser.yml`

Used by **GoReleaser** to define build and release behavior.

#### Highlights:

- **Multi-platform Builds:** Targets `linux`, `darwin`, and `windows` for `amd64` and `arm64`.
- **No CGO:** Uses `CGO_ENABLED=0` for portability in CI/CD environments.
- **LDFLAGS Injection:** Injects version and commit at build time:
  ```go
  -X main.version={{.Version}} -X main.commit={{.Commit}}
  ```
- **Checksum & Signing:**
  - Generates SHA256 checksum files.
  - Signs checksum files using GPG (non-interactive).
- **Changelog Generation:**
  - Extracted from Git commits.
  - Uses regex-based grouping:
    - 🚀 Features (`feat:`)
    - 🐛 Bug Fixes (`fix:`)
    - 🛠 Maintenance (`chore:`)

> See full config in [.goreleaser.yml](../../.goreleaser.yml).

### 2. `gitversion.yml`

Used to define GitVersion's behavior for semantic versioning.

#### Highlights:

- **Mainline Development Mode:** Uses `mode: Mainline` for automatic version bumping based on PR merges and commit messages.
- **Versioning Format:** MajorMinorPatchTag

> See full config in [gitversion.yml](../../gitversion.yml).

### 3. `.golangci.yml`

Configuration for `golangci-lint`.

#### Enabled Linters:

- `govet`, `errcheck`, `staticcheck`, `unused`
- `gofmt`, `goimports`, `gosimple`, `gocritic`

> Intended as a basic config. Consider expanding in the future.

## 🧩 Secrets & Environment Variables

| Name                 | Purpose                            |
|----------------------|------------------------------------|
| `PGP_PRIVATE_KEY`    | Used to sign the checksums         |
| `GITHUB_TOKEN`       | Provided by GitHub to create releases |
| `GPG_FINGERPRINT`    | Exported at runtime for signing    |

## 📌 Future Improvements

- ✅ **Add Unit Tests** using `gotestsum` and coverage report.
- ✅ **Add CodeQL** for automated security analysis.
- 🔜 Consider splitting the release and lint/test into separate workflows.