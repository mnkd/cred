# Plan 020: Move GitHub Actions to Node.js 24

## Problem

CI warns that `actions/checkout@v4` and `actions/setup-go@v5` target Node.js 20, which is deprecated on GitHub Actions runners, and are being forced to run on Node.js 24. `goreleaser/goreleaser-action@v6` in the release workflow also targets Node.js 20.

## Solution

Use the latest major version of each action. All of them run on `node24`.

| Action | Before | After |
|---|---|---|
| `actions/checkout` | v4 | v7 |
| `actions/setup-go` | v5 | v7 |
| `goreleaser/goreleaser-action` | v6 | v7 |

Breaking changes in between, checked against the release notes:

- `actions/checkout`
  - v5: Node.js 24; requires runner v2.327.1 or later (GitHub-hosted runners qualify)
  - v6: persisted credentials move to a separate file. The release workflow pushes to the Homebrew tap with its own token, not with the persisted credentials
  - v7: refuses to check out fork PRs under `pull_request_target` and `workflow_run`. These triggers are not used
- `actions/setup-go`
  - v6: Node.js 24, and changed toolchain selection. The workflows use `go-version: stable`, which is unaffected
  - v7: ESM migration only
- `goreleaser/goreleaser-action`
  - v7: Node.js 24. Inputs used here (`version`, `args`) are unchanged

## Changes

1. `.github/workflows/ci.yml`
2. `.github/workflows/release.yml`
3. `.github/workflows/govulncheck.yml`

## Verification

- CI on the PR passes without the Node.js 20 deprecation annotation
- govulncheck runs on the PR because its workflow file changes
- The release workflow is checked on the next tag
