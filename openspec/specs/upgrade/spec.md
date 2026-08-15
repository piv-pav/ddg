## Purpose

Self-upgrade ddg by checking the latest GitHub release tag and installing a newer version via `go install`.

## Requirements

### Requirement: Upgrade command
The `ddg` CLI SHALL provide an `upgrade` subcommand that compares the running version against the latest GitHub release tag and installs a newer version via `go install` when one exists.

#### Scenario: Newer version available
- **WHEN** `ddg upgrade` is run and the latest tag is newer than the running version
- **THEN** it runs `go install github.com/piv-pav/ddg@<tag>` and reports the upgraded version

#### Scenario: Already up to date
- **WHEN** `ddg upgrade` is run and the running version is equal to or newer than the latest tag
- **THEN** it reports "already up to date" and makes no changes

### Requirement: Latest version detection
The upgrade command SHALL fetch the latest version tag from the GitHub API (`api.github.com/repos/piv-pav/ddg/tags`) and use the first (most recent) tag.

#### Scenario: Tags fetched
- **WHEN** the GitHub API returns tags
- **THEN** the first tag name is used as the latest version

#### Scenario: GitHub unreachable
- **WHEN** the GitHub API request fails
- **THEN** the command fails with a clear error and makes no changes

### Requirement: Semantic version comparison
The upgrade command SHALL compare versions using semantic versioning, ignoring a `-dev` suffix on the running version.

#### Scenario: Dev build
- **WHEN** the running version is `v0.5.0-dev` and the latest tag is `v0.5.0`
- **THEN** the versions compare equal (the `-dev` suffix is stripped) and no upgrade occurs

#### Scenario: Unversioned build
- **WHEN** the running version is `dev` (unversioned local build)
- **THEN** it is treated as older than any tagged release and an upgrade proceeds

### Requirement: Consistent version string
The `ddg` version string SHALL be `v`-prefixed in all build paths, with a `-dev` suffix for local `just build` builds.

#### Scenario: Local build
- **WHEN** `ddg` is built via `just build`
- **THEN** `ddg --version` reports `v<version>-dev`

#### Scenario: Installed build
- **WHEN** `ddg` is installed via `go install`
- **THEN** `ddg --version` reports the tagged version with a `v` prefix
