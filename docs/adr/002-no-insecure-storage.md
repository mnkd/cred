# ADR-002: Do not implement --insecure-storage flag

## Status

Accepted

## Context

gh CLI has an `--insecure-storage` flag that provides a plaintext file fallback for environments where the keyring is unavailable (headless servers, CI). We considered whether cred needs a similar feature.

### Two approaches for fallback implementation

1. **Plaintext file + file permissions (gh CLI approach)**: No encryption. Protected only by `0600` permissions
2. **Encrypted with master password**: Key derivation via Argon2id + AES-GCM encryption. Requires password input every time

## Decision

Do not implement the `--insecure-storage` flag.

## Rationale

- cred's target environment is limited to local machines (MCP, scripts, web development, Docker), where the OS keyring is always available
- The plaintext file approach risks leaving secrets on disk
- The encryption approach requires password input every time, contradicting cred's development motivation that `op signin` is inconvenient
- There are no use cases where the keyring would be unavailable

## Consequences

- cred does not work in environments where the keyring is unavailable (intentional limitation)
- Secret management in CI environments assumes the use of dedicated mechanisms such as GitHub Secrets
