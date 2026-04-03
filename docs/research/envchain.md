# envchain - Detailed Analysis

## 1. Basic Info

- **Author**: Sorah Fukumori ([sorah](https://github.com/sorah)), Japan-based Ruby/infrastructure engineer
- **License**: MIT
- **Repository**: https://github.com/sorah/envchain
- **Stars**: ~1,200
- **First release**: ~2014
- **Latest release**: v1.1.0
- **Maintenance status**: Mature, feature-complete, maintenance-mode. Commit activity is infrequent. Issues are rarely addressed.

## 2. Architecture

- **Language**: C (pure C, no external framework)
- **Source files**:
  - `envchain.c` -- main entry point, CLI argument parsing, prompt logic
  - `envchain_osx.c` -- macOS Keychain implementation (Security.framework)
  - `envchain_linux.c` -- Linux implementation (libsecret / D-Bus)
- **Dependencies**: macOS: system Security.framework only. Linux: libsecret. No third-party libraries.
- **Supported OS**: macOS, Linux (GNOME/D-Bus environments)
- **Build system**: Makefile (`make && sudo make install`)

## 3. Features

envchain uses flags, not subcommands:

| Flag/Usage | Description |
|---|---|
| `envchain --set NAMESPACE ENV [ENV...]` | Store secrets interactively (prompted) |
| `envchain NAMESPACE COMMAND [ARGS]` | Execute command with env vars injected |
| `envchain --list` | List all namespaces |
| `envchain --list NAMESPACE` | List keys in a namespace |
| `envchain --unset NAMESPACE ENV` | Remove a specific variable |
| `--noecho (-n)` | Hide input when prompting |
| `--require-passphrase (-p)` | Force Keychain passphrase prompt |
| `--no-require-passphrase (-P)` | Allow access without passphrase |

Multiple namespaces can be specified comma-separated: `envchain aws,hubot command`.

## 4. Env Var Injection Mechanism

- **Process execution**: `setenv()` + `execvp()` -- true exec (process replacement), not fork
  - No parent process remains running
  - PID is preserved
- **Env var priority**: `setenv()` with overwrite=1 -- **envchain values override existing env vars**
- **Conflict behavior**: If an env var already exists, envchain replaces it silently

## 5. Security Model

- **Keyring access control**: Items stored as generic passwords with service name `envchain-NAMESPACE`. `--require-passphrase` / `--no-require-passphrase` control Keychain ACL.
- **Secret values in memory**: Read from Keychain into process memory, set as env vars, then `execvp()` replaces the process. No explicit memory zeroing.
- **Audit logging**: None.
- **Storage**: Encrypted at rest by the OS Keychain. envchain itself does no encryption.

## 6. UX/DX

- **Setup**: `brew install envchain`, then `envchain --set NAMESPACE VAR`.
- **Daily workflow**:
  ```
  envchain aws aws s3 ls           # run aws CLI with credentials
  envchain myapp rails server      # run Rails with API keys
  envchain aws,slack ./deploy.sh   # multiple namespaces
  ```
- **Error messages**: Minimal. `execvp failed: No such file or directory` is cryptic when the command is not in PATH (Issue #13).

## 7. Limitations / Known Issues

| Issue | Description |
|---|---|
| [#13](https://github.com/sorah/envchain/issues/13) | Cryptic error when command not found in PATH |
| [#24](https://github.com/sorah/envchain/issues/24) | `--list` shows duplicate entries |
| [#14](https://github.com/sorah/envchain/issues/14) | direnv integration is awkward/unsupported |

Additional limitations:

- **No `get` command**: Cannot retrieve a single secret to stdout
- **No `.env` import/export**
- **No config/metadata**: No descriptions, timestamps, or annotations
- **No JSON output**
- **No edit command**: Must re-run `--set` to update
- **No diagnostics command**
- **No default namespace**: Must always specify namespace
- **Interactive-only input**: Cannot pass value as argument or via stdin pipe

## 8. Comparison with cred

| Feature | envchain | cred |
|---|---|---|
| **Language** | C | Go |
| **Keyring backend** | Security.framework / libsecret (direct) | zalando/go-keyring (wrapper) |
| **Supported OS** | macOS + Linux | macOS only |
| **CLI style** | Flags (`--set`, `--list`) | Subcommands (`set`, `get`, `list`) |
| **Set secret** | Interactive prompt only | Prompt, argument, stdin pipe |
| **Get secret (stdout)** | Not supported | `cred get KEY` |
| **Delete secret** | `--unset` | `cred delete` (with aliases) |
| **Env injection** | `envchain NS cmd` (exec) | `cred run -- cmd` (exec) |
| **Multiple namespaces** | Comma-separated (`aws,slack`) | Single namespace per invocation (`-n`) |
| **.env import/export** | Not supported | `cred import` / `cred export` |
| **Edit in editor** | Not supported | `cred edit KEY` |
| **Diagnostics** | Not supported | `cred doctor` |
| **Config file** | None | `~/.config/cred/config.yml` |
| **Key descriptions** | Not supported | `-d` flag |
| **JSON output** | Not supported | `--format json` |
| **Default namespace** | Not supported | `cred ns default` |
| **Passphrase control** | `--require-passphrase` | Delegates to OS keyring |

### envchain's advantages over cred

1. Linux support
2. Zero dependencies (pure C, system libraries only)
3. Tiny binary, minimal footprint
4. Multi-namespace in single invocation (comma-separated)
5. Explicit passphrase ACL control per item
6. Long track record (~1.2k stars, Homebrew formula)

---

*Researched: 2026-04-04*
