# Plan 015: Add Homebrew tap support

## Goal

Enable `brew install mnkd/tap/cred` by adding Homebrew formula generation to GoReleaser config.

## Current State

- `.goreleaser.yml` exists (builds darwin amd64/arm64)
- `.github/workflows/release.yml` exists (triggers on tag push)
- No Homebrew configuration in GoReleaser

## Tasks

1. Add `brews` section to `.goreleaser.yml`
2. Update `release.yml` to use a PAT (`HOMEBREW_TAP_TOKEN`) instead of `GITHUB_TOKEN` (needed for cross-repo push to `mnkd/homebrew-tap`)
3. Update `README.md` installation section to add `brew install` method
4. User creates `mnkd/homebrew-tap` repository on GitHub (manual)
5. User adds `HOMEBREW_TAP_TOKEN` secret to the cred repository (manual)
