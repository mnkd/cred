# Env Var Injection Mechanism & Security Model: cred vs envchain vs aws-vault

## Env Var Injection Mechanism

| | cred | envchain | aws-vault |
|---|---|---|---|
| **Process execution** | `syscall.Exec()` | `setenv()` + `execvp()` | `syscall.Exec()` |
| **Execution method** | exec (process replacement) | exec (process replacement) | exec (process replacement) |
| **Env var construction** | `os.Environ()` + `setEnv` (overwrite) | `setenv(overwrite=1)` | `os.Environ()` + overwrite |
| **Conflict with existing vars** | Explicit overwrite | Explicit overwrite | Explicit overwrite |
| **Command resolution** | `exec.LookPath()` | `execvp()` (auto PATH search) | `exec.LookPath()` |

### Common Pattern

All three tools use exec (process replacement). The cred/envchain/aws-vault process replaces itself with the target command -- no parent process remains. This is the standard Unix pattern for injecting environment variables.

### Env Var Conflict Handling

All three tools now explicitly overwrite existing environment variables. cred uses a `setEnv` helper that scans the environ slice for a matching `KEY=` prefix and overwrites in place (see ADR-003).

## Security Model

| | cred | envchain | aws-vault |
|---|---|---|---|
| **Storage** | macOS Keychain (via zalando/go-keyring) | macOS Keychain (Security.framework) / libsecret | macOS Keychain (via 99designs/keyring) |
| **Service name** | `cred:{namespace}` | `envchain-{NAMESPACE}` | Profile-name based |
| **Secrets on disk** | None (config.yml = metadata only) | None | None |
| **Secrets in memory** | Go string → discarded on exec | C string → discarded on execvp | Go string → discarded on exec |
| **Memory zeroing** | No | No | No |
| **Passphrase control** | Delegates to OS | `--require-passphrase` flag | Delegates to OS |
| **Timeout** | **3-second timeout** | None | None |
| **Audit logging** | None | None | None (CloudTrail for STS calls) |
| **Encryption** | Delegates to OS Keychain | Delegates to OS Keychain | Delegates to OS Keychain |

### Common Security Model

All three tools share the same fundamental security model:

1. Secret values are stored only in the OS Keychain (never in files)
2. Retrieved values temporarily exist in process memory
3. `exec` replaces the process, discarding cred/envchain/aws-vault's own memory
4. However, secrets remain in the child process environment (readable via `/proc/PID/environ` on Linux)
5. None of the tools perform explicit memory zeroing

### cred-specific Characteristics

- **Timeout mechanism** (`keyring.go:18`): 3-second timeout on Keychain operations. envchain and aws-vault lack this, meaning the CLI hangs if the Keychain dialog is left unattended.
- **No passphrase control**: envchain provides `--require-passphrase` to set per-item access policy. cred delegates entirely to OS default behavior.

## Potential Improvements for cred

1. ~~**Env var conflict handling**~~: Resolved in ADR-003 — now explicitly overwrites existing variables.
2. **Passphrase control**: Evaluate whether an envchain-like `--require-passphrase` feature is needed.

---

*Researched: 2026-04-04*
