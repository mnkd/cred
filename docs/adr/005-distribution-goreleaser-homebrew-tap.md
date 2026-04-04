# ADR-005: Distribute via GoReleaser + Homebrew tap

## Status

Accepted

## Context

Currently, cred can only be installed via `go install github.com/mnkd/cred@latest`, which requires a Go toolchain. This limits adoption to users who already have Go installed.

### How competing tools are distributed

| Tool | Stars | Installation method |
|---|---|---|
| envchain | ~1,284 | `brew install envchain` (homebrew-core) |
| aws-vault | ~8,964 | `brew install --cask aws-vault` / GoReleaser |
| teller | ~3,189 | `brew install spectralops/tap/teller` (tap) |
| chamber | ~2,584 | `brew install chamber` (homebrew-core) |
| Doppler CLI | ~358 | `brew install dopplerhq/cli/doppler` (tap) |
| 1Password CLI | N/A | `brew install --cask 1password-cli` |
| sops | ~21,365 | `brew install sops` (homebrew-core) |
| gopass | ~6,773 | `brew install gopass` (homebrew-core) |

Nearly all tools in this space support Homebrew. High-star tools are in homebrew-core, while smaller or newer tools use a custom tap.

### Distribution options considered

1. **`go install` only (current)** — No additional setup, but requires Go.
2. **Homebrew tap (e.g., `mnkd/homebrew-tap`)** — No review process, anyone can publish immediately.
3. **homebrew-core** — `brew install cred` with no tap, but requires meeting acceptance criteria (star count, popularity, active maintenance).
4. **GoReleaser + Homebrew tap** — Automates cross-compilation, GitHub Release creation, and Homebrew formula generation on tag push via GitHub Actions.

## Decision

Adopt GoReleaser with a Homebrew tap (`mnkd/homebrew-tap`).

- Use GoReleaser to build release binaries and generate a Homebrew formula automatically.
- Use GitHub Actions to trigger the release pipeline on version tag push (e.g., `v0.1.0`).
- Create a separate repository `mnkd/homebrew-tap` to host the generated formula.
- Keep `go install` as an alternative installation method.

## Rationale

- GoReleaser is the de facto standard for Go CLI tool releases (used by aws-vault, goreleaser itself, etc.)
- A custom tap has no adoption barrier — homebrew-core can be pursued later if the project grows
- Automating the release pipeline reduces manual work and ensures reproducible builds
- Users can install with `brew install mnkd/tap/cred` without needing Go

## Consequences

- A new repository `mnkd/homebrew-tap` must be created and maintained
- A `.goreleaser.yml` config and GitHub Actions workflow must be added to this repository
- README.md installation instructions should be updated to show both `brew install` and `go install`
- A `GITHUB_TOKEN` with write access to the tap repository is needed for formula auto-updates
