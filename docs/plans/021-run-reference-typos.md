# Plan 021: Reject reference typos and print errors once

## Problem

Feedback on 0.2.0:

1. A mistyped reference such as `cred:/work/KEY` or `cred//work/KEY` does not start with `cred://`, so it is treated as a plain value
   - The literal string reaches the command
   - With no reference left, `cred run` without `-n` injects the whole default namespace, which hands every key to an MCP server meant to receive one
2. Every error is printed twice: cobra prints `Error: ...` and `main.go` prints the same error again. This affects all commands

## Solution

1. Treat any value that starts with `cred:` or `cred//` as a reference
   - It must then parse as `cred://<namespace>/<KEY>`, or cred exits 1 without starting the command
   - Because it counts as a reference, it also turns off default namespace injection
   - A value that legitimately starts with `cred:` cannot be passed through `cred run`. This is accepted: such values are unlikely, and a silent typo is the larger risk
2. Set `SilenceErrors` on the root command so only `main.go` prints errors. `main.go` keeps the `Error:` prefix

Making namespace injection require `-n` (so a forgotten reference also fails) would break `cred run -- python app.py`, and is left for a separate decision.

## Changes

1. `internal/ref/ref.go` — `IsRef` matches `cred:` and `cred//`; `Parse` still requires `cred://`
2. `cmd/root.go` — `SilenceErrors: true`
3. `main.go` — print `Error: <err>`
4. Tests and README

## Steps

1. Write tests for `IsRef` / `buildRunEnv` with typos, and for a single error line, confirm failure
2. Implement
3. Update README
