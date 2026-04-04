# ADR-006: Defer code signing for Homebrew distribution

## Status

Accepted

## Context

macOS Keychain access behavior differs depending on whether the binary is code-signed:

| | Unsigned (formula / go install) | Signed (cask) |
|---|---|---|
| Keychain access | Permission dialog on first access | No dialog or fewer prompts |
| Functionality | Works correctly | Works correctly |

aws-vault distributes via `brew install --cask` specifically to ship a signed binary, avoiding Keychain permission dialogs for its large user base (~9,000 stars). 1Password CLI also uses cask, as it is a commercial closed-source product distributed only as a signed binary.

cred uses the macOS Keychain via `go-keyring`, so the same permission dialog behavior applies to unsigned builds.

## Decision

Do not code-sign the binary for now. Distribute as an unsigned binary via Homebrew formula (see ADR-005).

## Rationale

- Unsigned binaries work correctly with macOS Keychain — the only difference is a one-time permission dialog
- The current `go install` workflow is already unsigned, so the user experience does not regress
- Code signing requires an Apple Developer account ($99/year) and additional CI/CD complexity
- At the current project scale, the UX benefit does not justify the cost

## Consequences

- Users may see a Keychain permission dialog on first use after installation
- If the user base grows and the permission dialog becomes a friction point, revisit this decision and consider:
  - Signing the binary with an Apple Developer certificate
  - Switching from formula to cask for distribution
