# Plan 019: Resolve `cred://` references in `cred run`

## Problem

MCP server configs (Claude Code `.mcp.json` / `~/.claude.json`, Codex `config.toml`) and scripts need tokens in environment variables. Today the only way to avoid writing tokens in plain text is `cred run -n <namespace>`, which injects every key in the namespace. That hands unrelated secrets to the launched command, cannot rename variables, and cannot combine keys from several namespaces.

## Goal

Like `op run`, configs hold only references (`cred://<namespace>/<KEY>`), and `cred run` replaces them with Keychain values before exec. The launched process receives only the keys that were referenced.

## Current behavior (facts)

- `-n` is optional. Without it, `cfg.ActiveNamespace` picks the default namespace and every key in it is injected
- All cred messages already go to stderr
- `syscall.Exec` replaces the cred process, so the exit code and signals already reach the command. This stays as is
- In namespace injection, a key missing from the keyring is skipped with a warning
- `keyring.Get` times out after 3 seconds

## Specification

### Reference syntax

- `cred://<namespace>/<KEY>`
  - The scheme is case-sensitive
  - `<namespace>` and `<KEY>` are non-empty and contain no `/`. Exactly one `/` separates them
- A value that starts with `cred://` but does not match is a parse error. It is never passed through as-is
- The environment variable name may differ from `<KEY>` (e.g. `API_TOKEN=cred://work/JIRA_TOKEN`)
- No escape mechanism, no variable expansion inside references

### `--env-file`

- `cred run --env-file <path> [--env-file <path>...] -- <command>`
- Parsed by the existing `env.ParseFile` (blank lines, `#` comments, surrounding quotes, `export ` prefix)
- A file that cannot be read stops cred before exec

### Building the environment

Layers are applied in this order. Later layers win.

1. Inherited environment
2. Namespace injection. It runs only when either:
   - `-n` is given explicitly, or
   - there are no references in the environment and no `--env-file` (backward-compatible behavior)
3. `--env-file` entries, in the order given
4. Every final value starting with `cred://` is resolved once. Resolved values are not resolved again

### Errors

- If any reference fails to parse or resolve, cred does not exec and exits with code 1
- All references are attempted, then all failures are reported together:
  - `✗ Cannot resolve API_TOKEN (cred://work/JIRA_TOKEN): not found in keyring`
- Failure reasons: not found, Keychain timed out or access denied, other Keychain error
- Messages contain only variable names and references, never values
- Namespace injection keeps skipping missing keys with a warning

### Output

- All cred messages go to stderr. stdout belongs to the command (MCP stdio servers use it for JSON-RPC)
- When references are resolved: `→ Resolved N reference(s)`
- `→ Injected N credential(s) from namespace ...` is printed only when namespace injection runs
- The command's stdout / stderr are not touched (no masking)

### Command lookup

`exec.LookPath` keeps using cred's own `PATH`. Desktop apps launch MCP servers with a minimal `PATH`, so the README recommends absolute paths for both `cred` and the command.

## Changes

1. `internal/ref/` — new package: parse references
2. `internal/env/env.go` — the invalid-line error no longer echoes the line content (it may hold a secret)
3. `cmd/run.go` — `--env-file` flag; extract environment building into `buildRunEnv`; keyring access through a replaceable variable for tests
4. `cmd/run_test.go` — unit tests for precedence, namespace-injection rules, and failure
5. `integration/integration_test.go` — reference resolution, no exec on failure, default namespace not injected with references
6. `README.md` and `cred run` help

## Steps

1. Write tests for `internal/ref`, confirm failure, implement
2. Update the `env.ParseFile` error test, confirm failure, implement
3. Write tests for `buildRunEnv`, confirm failure, implement
4. Add integration tests
5. Update README and help

## Manual verification

- MCP servers in the README form connect from Claude Code and Codex, both CLI and desktop app
- Whether a Keychain access dialog appears when launched from a desktop app, and whether the 3-second timeout is enough
