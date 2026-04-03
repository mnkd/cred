# 013: Standardize env var override behavior

## Decision

ADR-003: Always overwrite existing environment variables (envchain/aws-vault approach).

## Problem

`cmd/run.go` appends credentials to the environ slice via `append()`.
When a variable with the same name already exists, both entries coexist
and which value wins is OS/runtime-dependent.

## Solution

Add a `setEnv` helper that scans the environ slice for a matching
`KEY=` prefix. If found, overwrite in place; otherwise append.

## Tasks

1. [x] Add `setEnv(environ []string, key, value string) []string` helper in `cmd/environ.go`
2. [x] Add unit tests for `setEnv` in `cmd/environ_test.go`
   - Overwrite: existing `KEY=old` is replaced with `KEY=new`
   - Append: new key is added when not present
   - Prefix safety: `API_KEY_V2` is not matched by `API_KEY`
   - Empty value: `KEY=` is handled correctly
3. [x] Replace `append(environ, envVar+"="+value)` with `setEnv(environ, envVar, value)` in `cmd/run.go`
4. [x] Run tests to confirm all tests pass
5. [x] Update `docs/research/injection-and-security-comparison.md` to reflect the change
