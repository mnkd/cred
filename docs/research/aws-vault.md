# aws-vault - Detailed Analysis

## 1. Basic Info

- **Author**: 99designs (Australian design marketplace company)
- **License**: MIT
- **Repository**: https://github.com/99designs/aws-vault
- **Stars**: ~8,000+
- **First release**: ~2016
- **Latest release**: v7.2.0 (March 26, 2023) -- last release from 99designs
- **Maintenance status**: **Effectively stalled since 2023.** No releases since March 2023. Community fork exists at [ByteNess/aws-vault](https://github.com/ByteNess/aws-vault). Suggested alternative: [common-fate/granted](https://github.com/common-fate/granted).

## 2. Architecture

- **Language**: Go
- **Key dependency**: `github.com/99designs/keyring` v1.2.2 -- uniform Go API across OS credential stores
- **Supported keyring backends**:
  - macOS Keychain
  - Windows Credential Manager
  - Linux Secret Service / GNOME Keyring / KWallet
  - Pass (password-store)
  - Encrypted file
  - KeyCtl
- **Supported OS**: macOS, Linux, Windows (cross-platform via keyring abstraction)
- **Other dependencies**: AWS SDK for Go, cobra (CLI framework)

## 3. Features / Subcommands

| Subcommand | Purpose |
|---|---|
| `add` | Store AWS credentials (Access Key ID + Secret Access Key) |
| `list` | Display stored profiles and session status |
| `exec` | Execute command with temporary AWS credentials injected |
| `login` | Open AWS Console in browser via federated login URL |
| `rotate` | Rotate IAM access keys |
| `remove` | Delete stored credentials |
| `export` | Export credentials for piping to other tools |
| `clear` | Remove cached temporary sessions |

## 4. Env Var Injection Mechanism

- **Process execution**: `syscall.Exec()` -- true exec (process replacement)
- **Two credential delivery modes**:
  1. **Environment variables (default)**: Sets `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_REGION`, `AWS_VAULT`
  2. **Local metadata server (`--server` flag)**: Starts a local ECS credential endpoint with auto-refresh
- **Conflict handling**: If `AWS_VAULT` is already set (nested session), warns/errors. Injected vars override existing `AWS_*` env vars.

## 5. Security Model

- **Keyring access control**: Delegates to OS keyring. macOS Keychain prompts for permission.
- **Secret values in memory**: Long-term credentials stored in OS keyring, not plaintext on disk. STS tokens held in process memory for subprocess duration. No explicit memory encryption (mlock).
- **Audit logging**: None built-in. AWS CloudTrail captures STS API calls.
- **Key design principle**: Minimizes long-term credential exposure by using short-lived STS tokens (default 1 hour). `~/.aws/credentials` file is not used.

## 6. UX/DX

- **Setup**:
  1. `brew install aws-vault`
  2. `aws-vault add <profile>` -- enter Access Key ID and Secret Access Key
  3. Configure `~/.aws/config` with profiles, roles, MFA serial
  4. `aws-vault exec <profile> -- <command>`
- **Daily workflow**:
  ```
  aws-vault exec dev -- aws s3 ls
  aws-vault exec prod -- terraform plan
  aws-vault login staging  # opens browser
  ```
- **Error messages**: Generally clear, but edge cases (nested sessions, macOS Keychain `-25308` errors) can be cryptic.

## 7. Limitations / Known Issues

- **Stalled maintenance**: No release since March 2023
- **Nested session confusion**: `AWS_VAULT` env var not cleared properly (Issue #905)
- **macOS Keychain permission errors**: `-25308` error in non-interactive contexts (CI/CD, cron)
- **credential_process compatibility**: Issues with nested aws-vault calls (Issue #1176)
- **AWS SSO workflow**: New AWS SSO flows may not be fully supported (Issue #1261)
- **No multi-cloud support**: AWS only
- **No access control / audit**: No built-in RBAC or logging
- **No automatic secret rotation lifecycle**: `rotate` exists but is manual

## 8. AWS-Specific Features

- **STS integration**: `GetSessionToken` (MFA sessions) and `AssumeRole` (cross-account/role-based)
- **MFA support**: Prompts for MFA token; caches session within TTL
- **Session management**: Caches STS session tokens with configurable TTL
- **Role assumption**: Full support for IAM role chaining via `~/.aws/config`
- **AWS SSO**: Supports AWS SSO (Identity Center) login flow
- **Console login**: `aws-vault login` generates federated sign-in URL
- **Credential process**: Can act as `credential_process` provider in `~/.aws/config`

## 9. General-Purpose Use: Not Viable

aws-vault is deeply coupled to AWS:

- `add` only accepts AWS Access Key ID + Secret Access Key pairs
- All credential operations go through AWS STS
- Data model is AWS profiles, not generic key-value pairs
- No API to store/retrieve arbitrary secrets

The underlying `99designs/keyring` Go library is general-purpose and reusable, but aws-vault itself cannot be repurposed.

---

*Researched: 2026-04-04*
