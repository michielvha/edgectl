# GitVersion Implementation

> [!IMPORTANT]
> These docs are based on gitVersion version 6.x

## Overview

This document describes the **GitVersion** implementation for the `edgectl` repository. **GitVersion** is used to automatically generate semantic version numbers based on Git history and branch structure, eliminating the need for manual version management.

## Configuration

The project uses a custom **GitVersion** configuration defined in the root `gitversion.yml` file. This configuration is based on the **GitHubFlow** workflow with several customizations to meet the specific versioning needs of the project.

### Key Configuration Components

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
    prevent-increment:
      of-merged-branch: true
    track-merge-target: false
    track-merge-message: true
    is-main-branch: true
    mode: ContinuousDeployment
  release:
    regex: ^release/(?<BranchName>[0-9]+\.[0-9]+\.[0-9]+)$
    label: ''
    increment: None
    prevent-increment:
      when-current-commit-tagged: true
      of-merged-branch: true
    is-release-branch: true
    mode: ContinuousDeployment
    source-branches:
    - main

assembly-versioning-scheme: MajorMinorPatch
```

## Versioning Strategy

Our versioning strategy combines several approaches:

1. **MergeMessage**: Analyzes merge commit messages for version increments.
2. **TaggedCommit**: Uses tagged commits to determine versions.
3. **TrackReleaseBranches**: Tracks release branches to derive version information.
4. **VersionInBranchName**: Extracts version information from the branch name.

### Branch Strategy

- **Main Branch**: Source for new development.
- **Release Branches**: Named as `release/x.y.z` where x.y.z is the version number.
- **Feature Branches**: For new features before merging to main.

### Version Bumping

Version increments can be triggered by:

- Adding specific tags to commit messages:
  - `+semver: major` or `+semver: breaking` - Bump major version
  - `+semver: minor` or `+semver: feature` - Bump minor version 
  - `+semver: patch` or `+semver: fix` - Bump patch version
  - `+semver: none` or `+semver: skip` - Don't bump version

### Release Branch Configuration

The release branch configuration is customized to:
- Match branches named `release/x.y.z`
- Prevent version increments when commits are tagged
- Use ContinuousDeployment mode for predictable version number
- Source from the main branch

## Usage

### Command Line

Check the current version:
```bash
gitversion
```

View the full configuration:
```bash
gitversion /showconfig
```

### Integration with CI/CD

GitVersion can be integrated into CI/CD pipelines to automatically:
- Generate version numbers for builds
- Tag releases
- Provide version information for artifacts

which we've provided in our [workflow](../../.github/workflows/binary-release.yaml) you can use my [custom action](https://github.com/michielvha/gitversion-tag-action)

## Our versioning strategy

- **Patch**: For runtime fixes or small adjustments (e.g., bug fixes or performance improvements). (RUN)
- **Minor**: For introducing new features or changes that are backward-compatible. (CHANGE)
- **Major**: For breaking changes that are not backward-compatible. (BREAKING)


## References

- [GitVersion Documentation](https://gitversion.net/docs/)
- [Configuration Options](https://gitversion.net/docs/reference/configuration)
- [Version Strategies](https://gitversion.net/docs/reference/version-increments)