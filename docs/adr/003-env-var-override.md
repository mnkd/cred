# ADR-003: Explicitly overwrite existing environment variables in `cred run`

## Status

Accepted

## Context

`cred run` builds the child process environment by appending credentials to the `os.Environ()` slice:

```go
environ = append(environ, envVar+"="+value)
```

When a variable with the same name already exists, both entries coexist in the slice. Which value the child process sees is OS/runtime-dependent — this is non-deterministic behavior.

Comparable tools handle this differently:

| Tool | Method | Behavior |
|---|---|---|
| envchain | `setenv(overwrite=1)` | Explicit overwrite |
| aws-vault | `os.Setenv` | Explicit overwrite |
| **cred** | `append` | **OS-dependent (non-deterministic)** |

### Options considered

1. **Always overwrite (envchain / aws-vault approach)**: Unconditionally replace existing values
2. **Overwrite with warning**: Replace and print a warning to stderr
3. **Flag control (`--override` / `--no-override`)**: Let users choose via CLI flag

## Decision

Option 1: Always overwrite existing environment variables.

## Rationale

- The purpose of `cred run` is to inject keyring credentials — keyring values should take precedence
- Matches established conventions (envchain, aws-vault)
- Simple and deterministic — no ambiguity about which value wins
- If users don't want keyring injection, they simply don't use `cred run`
- Adding a flag would be YAGNI and increase CLI complexity

## Consequences

- Users cannot override a keyring-managed credential via `export VAR=value` when using `cred run`
- To use a different value temporarily, users should update the keyring or run the command without `cred run`
