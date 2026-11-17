#!/bin/sh
# Prepares the manual check of cred run (see README.md in this directory).
# Builds cred and the MCP server from this checkout into a cache directory,
# so the paths that MCP clients keep do not point into a worktree that may
# be deleted.
set -e
SRC="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SRC/../.." && pwd)"
KIT="${XDG_CACHE_HOME:-$HOME/.cache}/cred-manual-check"

mkdir -p "$KIT"
(cd "$ROOT" && go build -o "$KIT/cred-dev" . && go build -o "$KIT/mcpserver" ./scripts/manual-check/mcpserver)
echo "built $KIT/cred-dev and $KIT/mcpserver"

# A dummy value, not a secret. References read the keychain directly, so
# ~/.config/cred/config.yml is not changed.
security add-generic-password -U -s cred:cred-manual-check -a DEMO_TOKEN -w cred-manual-check-dummy-token
echo "stored the dummy value for cred://cred-manual-check/DEMO_TOKEN"

cat <<MSG

Register the MCP server:

  claude mcp add cred-manual-check -s user -e DEMO_TOKEN=cred://cred-manual-check/DEMO_TOKEN -- '$KIT/cred-dev' run -- '$KIT/mcpserver'

  codex mcp add cred-manual-check --env DEMO_TOKEN=cred://cred-manual-check/DEMO_TOKEN -- '$KIT/cred-dev' run -- '$KIT/mcpserver'
MSG
