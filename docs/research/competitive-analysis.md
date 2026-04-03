# Competitive Analysis: OS Keyring-based Secret Injection CLI Tools

## Overview

Survey of CLI tools that manage secrets and inject them as environment variables when running commands. Focused on tools that use the OS keyring (macOS Keychain, etc.) as the primary store.

## Most Similar Tools

### envchain (Closest Competitor)

- **Repository**: https://github.com/sorah/envchain
- **Stars**: ~1,284
- **Language**: C
- **How it works**: Stores secrets in macOS Keychain or gnome-keyring, organized by namespace. Injects them as environment variables via `envchain <namespace> <command>`.
- **Differences from cred**:
  - No config file for key-to-env-var mapping
  - No `.env` file integration
  - Written in C (harder to extend)
  - Simpler but less structured CLI

### aws-vault (AWS-specific)

- **Repository**: https://github.com/99designs/aws-vault
- **Stars**: ~8,964
- **Language**: Go
- **How it works**: Stores AWS credentials in OS keyring. Injects temporary STS credentials via `aws-vault exec <profile> -- <command>`.
- **Differences from cred**: AWS-only, not general-purpose.

## Tools with Env Injection but Without OS Keyring

| Tool | Stars | Backend | Usage |
|---|---|---|---|
| [teller](https://github.com/tellerops/teller) | ~3,189 | Multi-cloud (AWS SSM, Vault, GCP, etc.) | `teller run -- <command>` |
| [chamber](https://github.com/segmentio/chamber) | ~2,584 | AWS SSM Parameter Store | `chamber exec <service> -- <command>` |
| [Doppler CLI](https://github.com/DopplerHQ/cli) | ~358 | Doppler SaaS | `doppler run -- <command>` |
| [1Password CLI](https://developer.1password.com/docs/cli/reference/commands/run/) | N/A | 1Password vault | `op run -- <command>` (Touch ID support) |
| [sops](https://github.com/getsops/sops) | ~21,365 | Encrypted files (KMS/PGP/age) | No `run` subcommand |
| [gopass](https://github.com/gopasspw/gopass) | ~6,773 | GPG-encrypted files in git | Env injection via direnv integration |

## Comparison Matrix

| Tool | OS Keyring | Env Inject (`run`) | Local-only | General Purpose |
|---|---|---|---|---|
| **cred** | Yes (macOS Keychain) | Yes | Yes | Yes |
| **envchain** | Yes (Keychain, gnome-keyring) | Yes | Yes | Yes |
| **aws-vault** | Yes | Yes | Yes | No (AWS only) |
| **teller** | No | Yes | No | Yes |
| **chamber** | No | Yes | No | Yes |
| **Doppler CLI** | No | Yes | No | Yes |
| **1Password CLI** | No (own vault) | Yes | No | Yes |
| **sops** | No | No | Yes | Yes |
| **gopass** | No (GPG files) | Partial (direnv) | Yes | Yes |

## cred's Differentiators

1. **Config-driven mapping** (`config.yml`): Declarative key-to-env-var mapping per namespace
2. **`.env` file parsing**: Bridge between file-based and keyring-based workflows
3. **Go-based**: Easier to extend and cross-compile than envchain (C)
4. **Structured CLI**: Subcommands (list, get, set, run, etc.) with cobra

## Conclusion

The niche of "OS keyring + env var injection + config-driven + general purpose" is underserved. Most popular tools either rely on cloud backends (Doppler, chamber, teller) or are domain-specific (aws-vault). envchain is the only direct competitor, but lacks config-driven mapping and `.env` support.

---

*Researched: 2026-04-04*
