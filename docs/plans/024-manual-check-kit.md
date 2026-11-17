# Plan 024: Kit for manual checks of `cred run`

## Problem

Two behaviors of `cred run` cannot be covered by `make test` or `make test-integration`, and were checked by hand for plan 023:

- An MCP server wrapped with `cred run` connects in real clients (Claude Code and Codex, CLI and desktop app), and its output is masked
- Ctrl-C in a terminal stops a command running under masked `cred run`

The scripts used for that check lived outside the repository, with absolute paths of one machine. They will be needed again whenever process handling or masking changes.

## Solution

Add the kit under `scripts/manual-check/`, without machine-specific paths.

- `setup.sh`
  - Builds cred from the checkout into `${XDG_CACHE_HOME:-$HOME/.cache}/cred-manual-check/cred-dev`, and copies the MCP server there
    - MCP clients keep the registered path, so it must not point into a worktree that may be deleted
  - Stores a dummy value (not a secret) for the reference `cred://cred-manual-check/DEMO_TOKEN` in the keychain
    - References read the keychain directly, so `~/.config/cred/config.yml` is not changed
  - Prints the `claude mcp add` and `codex mcp add` commands with the resolved paths
- `mcp_echo_server.py` — a minimal stdio MCP server (Python standard library only)
  - `ping_cred` returns a fixed string, to check the connection
  - `show_token` returns `DEMO_TOKEN`, which is `<concealed by cred>` when masking works
  - Writes `DEMO_TOKEN` to stderr at start, so the client's MCP log shows whether stderr is masked
- `ctrlc-check.sh` — two cases to stop with Ctrl-C: a plain `sleep`, and a command with a SIGINT handler
- `cleanup.sh` — removes the dummy keychain value and the cache directory, and prints the commands to unregister the MCP server
- `README.md` — the procedure and the expected results

Python is used for the server because `/usr/bin/python3` comes with the Command Line Tools or Xcode, which a macOS development machine usually has (e.g. for git). Building cred itself does not need them. Without them, `/usr/bin/python3` only offers to install them, so the README says to run `xcode-select --install`. A Go server under `cmd/` would add a binary that only this manual check uses.

## Changes

1. `scripts/manual-check/` (new) — the files above
2. `README.md` — a pointer in the Testing section
3. `CLAUDE.md` — `scripts/manual-check/` in the architecture list

## Steps

1. Write the scripts
2. Run them: `setup.sh`, an MCP exchange through `cred-dev` with a small client, `ctrlc-check.sh` syntax, `cleanup.sh`
3. Update README and CLAUDE.md
