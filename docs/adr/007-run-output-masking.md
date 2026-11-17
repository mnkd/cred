# ADR-007: Mask secrets in `cred run` output by default

## Status

Accepted

## Context

`cred run` is used by AI agents and as the command of MCP stdio servers. When the command prints an injected value (`env`, `printenv`, `curl -v`, `set -x`, ...), the secret ends up in logs and in the agent's conversation history.

Until 0.2.x, `cred run` replaced itself with the command via `syscall.Exec` and did not touch its output (plan 019). Exit codes, signals and the terminal belonged to the command with no code in cred.

1Password CLI's `op run` conceals secrets on stdout and stderr by default and offers `--no-masking`. Its documentation calls masking best effort and says a secret split across separate writes can appear in plaintext.

Masking requires reading the command's output, so the command can no longer replace cred: cred has to start it as a child, read its stdout and stderr through pipes, and take over what `exec` gave for free.

### Options considered

1. **Keep `syscall.Exec` and no masking** — No change, but secrets keep leaking into agent transcripts
2. **Masking as an opt-in flag** — Safe for existing users, but the users who most need it (agents, MCP configs written once) would not turn it on
3. **Masking by default, `--no-masking` to turn it off** — The same choice as `op run`
4. **Run the command under a pseudo-terminal** — Keeps colors and interactive output, but brings terminal modes, window size and line-ending translation into cred

## Decision

Mask secrets in the command's stdout and stderr by default, and offer `--no-masking` (option 3).

- Masked values are the values cred read from the keychain for the run
  - Values injected from a namespace and values of resolved `cred://` references
  - Plain values in `--env-file` are not masked
  - Values shorter than 8 bytes are not masked
- Every byte covered by an occurrence of any value is replaced; each run of overlapping occurrences becomes one `<concealed by cred>`
- Output is streamed
  - Only a tail that is a proper prefix of some value is held back until the next read, so a value split across two reads is still masked
  - Newline-terminated output, such as JSON-RPC messages, is never delayed unless a value contains a newline
- When there is something to mask, cred starts the command as a child and stays as its parent
  - It forwards SIGINT, SIGTERM, SIGHUP, SIGQUIT, SIGUSR1, SIGUSR2 and SIGWINCH, the set `op run` 2.41.1 forwards
  - Signals ignored when cred started (e.g. under `nohup`) stay ignored for the command
  - It exits with the command's exit code, or 128 + n when the command was killed by signal n
  - After the command exits, it reads until the pipes have been silent for 1 second, so a background process that inherited them cannot keep cred alive
- With `--no-masking`, or when no value is long enough to mask, cred still replaces itself via `syscall.Exec`

## Rationale

- Agents and MCP clients are the main users of `cred run`, and they rarely change the flags of a configured command. Only a default protects them
- Matching `op run` keeps one mental model for users of both tools, including the flag name
- Keeping `syscall.Exec` when nothing needs masking leaves the old behavior wherever masking cannot help
- 128 + n is the shell convention and keeps the signal visible; `op run` exits with 255, which hides it. Re-raising the signal on cred itself is not reliable for signals the Go runtime handles
- A pseudo-terminal adds platform-specific code for the cases `--no-masking` already covers

## Consequences

- Output of `cred run` changes for existing users; this ships as 0.3.0 with a release note
- The command's stdout and stderr are pipes, not a terminal, so some commands drop colors or interactive output. `--no-masking` restores the terminal
- A prompt that does not end with a newline is held back while its end could be the start of a secret
- On Ctrl-C in a terminal, the command receives SIGINT twice: from the terminal and from cred. `op run` behaves the same
- If cred is killed with SIGKILL, the command is not signalled; it ends on stdin EOF or its next write
- Masking prevents accidental disclosure only. It is not a security boundary: a command can still encode a value or write it elsewhere, and secrets passed to the command outside cred (e.g. `export $(cred get ...)`) are not known to cred
- Process handling that `exec` used to provide is now cred's code (`cmd/run_child.go`). Changes to it are checked by hand with `scripts/manual-check/`, since unit and integration tests cannot drive real MCP clients or a terminal
- Implementation details are in `docs/plans/023-run-output-masking.md`
