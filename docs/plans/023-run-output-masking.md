# Plan 023: Mask secrets in `cred run` output

## Problem

When an AI agent runs a command through `cred run` and the command prints an injected value (`env`, `printenv`, `curl -v`, `set -x`, ...), the secret ends up in the conversation history. 1Password's `op run` conceals secrets on stdout and stderr by default and offers `--no-masking` to turn this off. cred should do the same.

## Current behavior

- `cred run` replaces itself with the command via `syscall.Exec`
  - Exit code and signals belong to the command without any code in cred
  - The command's output is not modified
- Plan 019 recorded "no masking" and "`syscall.Exec` stays as is". This plan replaces that decision

## Terms

- Secret: a value cred read from the keychain for this run
  - Values injected from a namespace (`-n`, or the default namespace when no reference and no `--env-file` is used)
  - Values looked up for `cred://` references, in the inherited environment or in `--env-file`
  - A namespace value later overridden by `--env-file` is still a secret. It was read from the keychain, and masking it costs nothing
- Plain value: a value written as is in `--env-file` or inherited from the environment. Not a secret
- Child: the command started by `cred run`

## Behavior

- By default, `cred run` replaces secrets in the child's stdout and stderr with `<concealed by cred>`
- `--no-masking` turns this off
- Matching is on the exact bytes of a secret
  - Encoded forms (base64, URL encoding, JSON escaping, ...) are not masked
- Every byte covered by an occurrence of any secret is concealed
  - Overlapping occurrences form one run and are replaced by one `<concealed by cred>`
    - This covers a shorter secret inside a longer one (the longer one wins) and the end of one secret overlapping the start of another
  - Occurrences that only touch, without overlapping, are replaced separately
- Secrets shorter than 8 bytes are not masked
  - Replacing a few common bytes would corrupt unrelated output
  - They are skipped silently; the minimum is documented in the help and README
- stdout and stderr are masked separately. A secret split between the two streams is not masked

## Streaming

`cred run` wraps MCP stdio servers, so output must not be delayed.

- Each chunk read from the child is masked and written at once, except for a held-back tail
- Held-back tail: the bytes from the earliest position where the rest of the chunk is a proper prefix of some secret
  - If an occurrence run overlaps that position, the tail starts at the start of the run, because a later byte may extend it
  - This lets a secret split across two reads still be masked. `op run` documents that it does not handle this case
- A JSON-RPC message ends with a newline, so it is not delayed unless a secret contains a newline (e.g. a PEM key stored without trimming)
- A prompt that does not end with a newline is held back if its end is a prefix of a secret, until more output arrives. Use `--no-masking` for interactive commands
- At EOF, the held-back tail is masked with no partial match left open and written

## Process handling

Masking needs a child whose stdout and stderr are pipes.

- When there is nothing to mask, cred keeps using `syscall.Exec`, so the child keeps the terminal and the exact old behavior
  - Nothing to mask: `--no-masking`, or no secret of 8 bytes or more
  - `op run` likewise leaves stdout and stderr as a terminal when there is nothing to conceal
- Otherwise cred starts the child and stays as its parent
  - argv is kept as typed (`exec.Cmd{Path: binary, Args: args}`, reusing the `LookPath` result)
  - stdin is the same file descriptor as cred's stdin (no copying)
  - If cred's stdout and stderr are the same file (e.g. `2>&1` or a terminal), the child gets one pipe for both, so their order is kept. Otherwise each gets its own pipe
  - The child stays in cred's process group
  - A start error is reported as `failed to start <command>: <err>` with exit code 1

### Signals

- SIGINT, SIGTERM, SIGHUP, SIGQUIT, SIGUSR1, SIGUSR2 and SIGWINCH received by cred are forwarded to the child
  - This is the set `op run` 2.41.1 forwards (checked by sending each signal to `op run`)
  - `signal.Notify` is called on a buffered channel before the child starts, so a signal in between is not lost
  - Errors from forwarding to a child that has already exited are ignored
  - SIGTSTP and SIGCONT are not forwarded; the terminal sends them to the whole process group
- On Ctrl-C in a terminal, the child receives SIGINT twice: from the terminal and from cred. `op run` behaves the same. `--no-masking` keeps the old behavior
- If cred itself is killed with SIGKILL, the child is not signalled. It ends on stdin EOF or on its next write to the closed pipe. This is accepted

### Waiting and exit code

- cred waits for the child to exit, not for the pipes to reach EOF
  - After the child exits, cred keeps reading for up to 1 second, then stops, so a background process that inherited the pipes cannot keep cred alive
  - Each stream's held-back tail is then written
- cred exits with the child's exit code
- If the child is killed by a signal, cred exits with 128 + the signal number (the shell convention)
  - `op run` exits with 255 here, which hides the signal
  - Re-raising the signal on cred itself is not reliable for signals the Go runtime handles
- The exit code reaches `main` as an error type that `main` turns into `os.Exit(code)` without printing `Error:`

### Write errors

- If writing to cred's stdout or stderr fails with EPIPE, Go's default applies: cred is killed by SIGPIPE
  - This matches the old behavior, where the child itself was killed by SIGPIPE
  - The child then gets EPIPE or SIGPIPE on its next write
- Any other write error stops reading that stream and closes it, so the child gets EPIPE on its next write. cred still exits with the child's exit code

## TTY

With masking, the child's stdout and stderr are pipes, not a terminal, so some commands drop colors or interactive output. This is the same as `op run`; use `--no-masking` for such commands. stdin is still the terminal.

## Scope and limits

Masking prevents accidental disclosure in logs and conversation history. It is not a security boundary.

- A command that wants to reveal a secret can still do so, e.g. by encoding it, writing it to a file, or writing to `/dev/tty`
- Secrets that reach the command another way (e.g. `export $(cred get ...)` in a shell profile) are not known to cred and are not masked
- An MCP tool that returns an injected secret returns `<concealed by cred>` instead. The replacement has no `"`, `\` or control characters, so JSON stays valid

## Alternatives considered

- A pseudo-terminal for the child would keep colors, but brings terminal modes, window size and line-ending translation into cred. `--no-masking` covers the cases that need a terminal
- Line buffering would delay any output without a trailing newline
- An idle flush timer would avoid the held-back prompt case, but needs a timer and a lock, and lets a split secret leak

## Changes

1. `internal/mask` (new)
   - `const Replacement = "<concealed by cred>"`, `const MinLen = 8`
   - `Values(values []string) []string`
     - Drops values shorter than `MinLen` and duplicates, and sorts longest first (order among equal lengths does not matter)
   - `NewWriter(w io.Writer, values []string) *Writer`
     - Applies `Values` itself, so empty or short values can never match
   - `(*Writer).Write(p)`
     - Returns `len(p)` on success, including bytes held back
     - Returns the underlying writer's error as is
   - `(*Writer).Flush() error` masks and writes the held-back tail
   - A `Writer` is not safe for concurrent use; cred uses one per stream
2. `cmd/run.go`
   - `buildRunEnv` returns `runEnv{Environ, Secrets []string}`
     - `injectNamespace` adds every value it reads; `resolveReferences` adds every value it looks up
     - No filtering here; `mask.Values` does that
   - `--no-masking` flag
   - `len(mask.Values(secrets)) == 0` or `--no-masking` → `syscall.Exec`; otherwise run the child as above
   - Help (`Long`) and flag text
3. `cmd/run_child.go` (new) — starting the child, forwarding signals, reading the pipes and computing the exit code. Takes stdin, stdout and stderr as `*os.File` so it can be unit-tested
4. `cmd/exit.go` (new) and `main.go` — the exit-code error type and its handling in `main`
5. Docs
   - README
     - The `Output` section: masking by default, `--no-masking`, the minimum length, TTY, scope and limits
     - The examples `cred run -- env | grep OPENAI` and `sh -c 'echo "key=$OPENAI_API_KEY" ...'`: they now print `<concealed by cred>`, which still shows that the variable is set
     - The `testcred` section: output is masked unless `--no-masking` is given
     - Architecture tree: `internal/mask/`
   - `CLAUDE.md` architecture list: `internal/mask/`
   - Release notes for 0.3.0 (GoReleaser writes the GitHub Release; the note goes in the PR description for the release): masking is on by default and changes output; `--no-masking` restores the old behavior

The version is set from the git tag at release time; this change ships as 0.3.0.

## Tests

1. `internal/mask` (already written)
   - Plain match, every occurrence, touching occurrences replaced separately
   - Secret split across chunks, byte by byte
   - Only a partial match is held back; a partial match that does not complete is written
   - A newline-terminated line is written at once
   - Longer secret wins within one chunk and across chunks; the shorter one is masked at `Flush`
   - To add: the end of one secret overlapping the start of another, within one chunk and split right after the first secret
   - To add: repeated occurrences of a secret whose suffix is also its prefix
   - To add: a masked JSON line stays valid (`json.Valid`)
   - Short values are not masked; `Values` filters, deduplicates and sorts
2. `cmd`
   - `buildRunEnv` returns namespace and reference values as secrets, not plain values (already written)
   - Child runner: exit code passed on, 128 + n on a signal, stdout and stderr masked, the shared-pipe case
3. Integration (run locally; CI does not run them)
   - Existing tests assert exact lines such as `KEY=<concealed by cred>`
     - `TestSingleKeyInjection`, `TestNamespaceSelection`, `TestMissingKeyInKeyring`, `TestEnvOverride`, `TestReferenceResolution`, `TestEnvFileReference`
     - Plain `--env-file` values stay visible in `TestEnvFileReference`
   - `TestMultipleKeysInjection` keeps its 7-byte values on purpose: they are below the minimum, so the output is unmasked
   - New
     - `--no-masking` prints each value exactly
     - `sh -c 'exit 3'` with a secret → exit code 3
     - `sh -c 'kill -TERM $$'` with a secret → exit code 143
     - SIGTERM sent to cred reaches the child
     - stderr is masked

## Steps

1. Add the missing `internal/mask` tests and confirm they fail (the others already fail)
2. Implement `internal/mask` and pass its tests
3. `buildRunEnv`: pass the existing tests
4. Child runner and exit code: tests, then implementation
5. Integration tests: update, then pass
6. README, help, `CLAUDE.md`
7. `make test`, `make test-integration`, `make build`

## Manual verification

- `cred run -- env` shows `<concealed by cred>` for secrets
- `cred run --no-masking -- env` shows the values
- Ctrl-C stops a long-running command
- A command that colors output on a terminal prints without color, and with color under `--no-masking`
- MCP servers wrapped with `cred run` still connect in Claude Code and Codex (CLI and desktop app). Checked by the user
