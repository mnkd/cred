# 013: Standardize env var override behavior

## Problem

`cmd/run.go` appends credentials to the environ slice via `append()`.
When a variable with the same name already exists, both entries coexist
and which value wins is OS/runtime-dependent.

Tools like envchain (`setenv(overwrite=1)`) and aws-vault explicitly
overwrite existing variables to guarantee deterministic behavior.

## Solution

Replace the `append` call with a helper that scans the existing environ
slice for a matching `KEY=` prefix. If found, overwrite in place;
otherwise append.

## Files

- `cmd/run.go` — use `setEnv` helper instead of raw `append`
- `cmd/run_test.go` — add tests for overwrite and append cases
