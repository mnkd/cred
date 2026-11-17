# Manual checks for `cred run`

These checks cover what `make test` and `make test-integration` cannot: real MCP clients and Ctrl-C in a terminal. Run them when process handling or output masking in `cred run` changes.

- An MCP server wrapped with `cred run` connects in Claude Code and Codex (CLI and desktop app), and its output is masked
- Ctrl-C stops a command running under masked `cred run`

The injected value is a dummy, not a secret.

## Files

- `setup.sh`
  - Builds cred and the MCP server from this checkout into `~/.cache/cred-manual-check/` (`$XDG_CACHE_HOME` if set)
  - Stores the dummy value for `cred://cred-manual-check/DEMO_TOKEN` in the keychain
  - Prints the commands to register the MCP server
- `mcpserver/` — a minimal stdio MCP server
  - `ping_cred` returns `pong from cred-manual-check`
  - `show_token` returns `DEMO_TOKEN=<concealed by cred>` when masking works
  - It writes `DEMO_TOKEN` to stderr at start, so the client's MCP log shows whether stderr is masked
- `ctrlc-check.sh` — the Ctrl-C check
- `cleanup.sh` — removes the dummy value and the cache directory

## 1. Set up

```bash
scripts/manual-check/setup.sh
```

## 2. Claude Code

Register the server with the `claude mcp add` command that `setup.sh` printed. The user scope is shared by the CLI and the desktop app.

- CLI
  - `claude mcp list` shows `cred-manual-check` as connected
  - In a new session, calling `ping_cred` of `cred-manual-check` returns `pong from cred-manual-check`
  - Calling `show_token` returns `DEMO_TOKEN=<concealed by cred>`
- Desktop app: the same two tools return the same results in a new session

## 3. Codex

Register the server with the `codex mcp add` command that `setup.sh` printed. It is added to `~/.codex/config.toml`, which the CLI and the desktop app share.

- CLI
  - `/mcp` in `codex` lists `cred-manual-check` and its two tools
  - `ping_cred` and `show_token` return the same results as in Claude Code
- Desktop app: the same two tools return the same results in a new thread

## 4. Ctrl-C

Run this in a terminal and press Ctrl-C once each time `press Ctrl-C` appears.

```bash
scripts/manual-check/ctrlc-check.sh
```

- Both cases print `token=<concealed by cred>`
- 1/2 (`sleep 30`) stops at once and reports exit code 130
- 2/2 (SIGINT handler) prints `got SIGINT` once, stops, and reports exit code 7
  - The command receives SIGINT from the terminal and from `cred run`; the handler exits on the first one

## 5. Clean up

Unregister the MCP server, then run `cleanup.sh`.

```bash
claude mcp remove cred-manual-check -s user
```

```bash
codex mcp remove cred-manual-check
```

```bash
scripts/manual-check/cleanup.sh
```

## Troubleshooting

- The server does not connect
  - Look at the server's stderr in the client's MCP log
  - `→ Resolved 1 reference(s)` means cred resolved the dummy value
- `Error: failed to load config: ... $HOME is not defined`
  - The client did not pass `HOME`. Add `HOME` to the server's `env` to narrow it down
