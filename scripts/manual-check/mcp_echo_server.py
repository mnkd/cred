"""Minimal stdio MCP server for checking cred run output masking.

Speaks newline-delimited JSON-RPC 2.0 on stdin/stdout (MCP stdio transport).
Tool show_token returns the DEMO_TOKEN environment variable, so a masked run
answers with "<concealed by cred>". The token is also logged to stderr, so
the client's MCP log shows whether stderr is masked too.
"""

import json
import os
import sys

TOOLS = [
    {
        "name": "show_token",
        "description": "Return the DEMO_TOKEN environment variable injected by cred run.",
        "inputSchema": {"type": "object", "properties": {}},
    },
    {
        "name": "ping_cred",
        "description": "Return a fixed string, to check the connection.",
        "inputSchema": {"type": "object", "properties": {}},
    },
]


def reply(msg_id, result=None, error=None):
    msg = {"jsonrpc": "2.0", "id": msg_id}
    if error is not None:
        msg["error"] = error
    else:
        msg["result"] = result
    sys.stdout.write(json.dumps(msg) + "\n")
    sys.stdout.flush()


def text(s):
    return {"content": [{"type": "text", "text": s}]}


def main():
    token = os.environ.get("DEMO_TOKEN", "")
    print(f"cred-manual-check: started, DEMO_TOKEN={token}", file=sys.stderr, flush=True)
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            req = json.loads(line)
        except json.JSONDecodeError:
            continue
        method, msg_id = req.get("method"), req.get("id")
        if msg_id is None:
            continue  # notification
        if method == "initialize":
            version = req.get("params", {}).get("protocolVersion", "2025-06-18")
            reply(msg_id, {
                "protocolVersion": version,
                "capabilities": {"tools": {}},
                "serverInfo": {"name": "cred-manual-check", "version": "1.0.0"},
            })
        elif method == "tools/list":
            reply(msg_id, {"tools": TOOLS})
        elif method == "tools/call":
            name = req.get("params", {}).get("name")
            if name == "show_token":
                reply(msg_id, text(f"DEMO_TOKEN={token}"))
            elif name == "ping_cred":
                reply(msg_id, text("pong from cred-manual-check"))
            else:
                reply(msg_id, error={"code": -32602, "message": f"unknown tool: {name}"})
        elif method == "ping":
            reply(msg_id, {})
        else:
            reply(msg_id, error={"code": -32601, "message": f"method not found: {method}"})


if __name__ == "__main__":
    main()
