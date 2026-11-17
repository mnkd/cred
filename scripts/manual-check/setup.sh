#!/bin/sh
# Prepares the manual check of cred run (see README.md in this directory).
# Builds cred from this checkout into a cache directory, so the path that
# MCP clients keep does not point into a worktree that may be deleted.
set -e
SRC="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SRC/../.." && pwd)"
KIT="${XDG_CACHE_HOME:-$HOME/.cache}/cred-manual-check"

mkdir -p "$KIT"
(cd "$ROOT" && go build -o "$KIT/cred-dev" .)
cp "$SRC/mcp_echo_server.py" "$KIT/"
echo "built $KIT/cred-dev"

# A dummy value, not a secret. References read the keychain directly, so
# ~/.config/cred/config.yml is not changed.
security add-generic-password -U -s cred:cred-manual-check -a DEMO_TOKEN -w cred-manual-check-dummy-token
echo "stored the dummy value for cred://cred-manual-check/DEMO_TOKEN"

cat <<MSG

Register the MCP server:

  claude mcp add cred-manual-check -s user -e DEMO_TOKEN=cred://cred-manual-check/DEMO_TOKEN -- '$KIT/cred-dev' run -- /usr/bin/python3 '$KIT/mcp_echo_server.py'

  codex mcp add cred-manual-check --env DEMO_TOKEN=cred://cred-manual-check/DEMO_TOKEN -- '$KIT/cred-dev' run -- /usr/bin/python3 '$KIT/mcp_echo_server.py'
MSG
