# GitVersion Implementation Guide

## Overview

This document describes the **GitVersion** implementation for the `edgectl` repository. **GitVersion** is used to automatically generate semantic version numbers based on Git history and branch structure, eliminating the need for manual version management.

## Configuration

The project uses a custom **GitVersion** configuration defined in the root `gitversion.yml` file. This configuration is based on the **GitHubFlow** workflow with several customizations to meet the specific versioning needs of the project.

### Key Configuration Components

```yaml
workflow: GitHubFlow/v1

# Custom strategies
strategies:
- MergeMessage
- TaggedCommit
- TrackReleaseBranches
- VersionInBranchName

branches:
  release:
    # Custom release branch configuration
    regex: ^release/(?<BranchName>[0-9]+\.[0-9]+\.[0-9]+)$
    label: ''
    increment: None
    prevent-increment:
      when-current-commit-tagged: true
    is-release-branch: true
    mode: ContinuousDelivery
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
- Use ContinuousDelivery mode for predictable version numbers
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

which we've ofcourse done in our [workflow](../../.github/workflows/binary-release.yaml) you can use my [custom action](https://github.com/michielvha/gitversion-tag-action)

## Sample Output

When running `gitversion`, you will get JSON output similar to:

```json
{
  "MajorMinorPatch": "0.2.0",
  "SemVer": "0.2.0-45",
  "BranchName": "release/0.2.0",
  "Sha": "dba550f2c2bd7bfd4e4c56c5ee920a41dab5d866",
  "ShortSha": "dba550f",
  "UncommittedChanges": 1
  // Additional properties not shown for brevity
}
```

## References

- [GitVersion Documentation](https://gitversion.net/docs/)
- [Configuration Options](https://gitversion.net/docs/reference/configuration)
- [Version Strategies](https://gitversion.net/docs/reference/version-increments)

## Troubleshooting

If GitVersion is not behaving as expected:

1. Use `gitversion /showconfig` to verify the current configuration
2. Check that your branch naming follows the expected patterns
3. Verify that Git tags are properly formatted
4. Ensure your commit messages use the correct format if you're using commit message versioning