#!/bin/sh
# Removes what setup.sh added. Unregister the MCP server first.
KIT="${XDG_CACHE_HOME:-$HOME/.cache}/cred-manual-check"
security delete-generic-password -s cred:cred-manual-check -a DEMO_TOKEN >/dev/null 2>&1 && echo "removed the dummy keychain value"
rm -rf "$KIT" && echo "removed $KIT"

cat <<MSG

If still registered, unregister the MCP server:

  claude mcp remove cred-manual-check -s user
  codex mcp remove cred-manual-check
MSG
