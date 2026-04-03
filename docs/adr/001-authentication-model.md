# ADR-001: No authentication mechanism in cred itself

## Status

Accepted

## Context

cred is a CLI tool for managing secrets such as API keys in a local development environment.
We need to decide which layer handles authentication.

### Target environments

- Reference during MCP and script execution on a local machine
- Web apps under development
- Docker (local execution)

Execution outside local environments (CI, shared servers, etc.) is not targeted.

## Decision

cred does not implement its own authentication mechanism; it delegates to the OS keyring's authentication.

- macOS Keychain: Protected by the user's login password. The OS displays an access permission dialog on first access
- Linux Secret Service: Protected by the user session (login)
- Windows Credential Manager: Protected by Windows login

We adopt the model: "logged-in user = authenticated."

## Rationale

- gh CLI uses the same architecture (`zalando/go-keyring` + no session auth) to manage GitHub access tokens, proving the model's validity
- The development motivation for cred was the inconvenience of 1Password CLI (`op`) session management (`op signin`); adding custom authentication contradicts this motivation
- The OS keyring is isolated per user, and cross-user protection is handled by the OS
- Since the target environment is limited to local machines, an additional authentication layer is unnecessary

## Consequences

- cred stores secret values only in the OS keyring (never in config files)
- Memory protection between processes and access control are delegated to the OS
- Support for non-local environments is out of scope
