# Plan 016: Deterministic key iteration order in `cred run`

## Problem

`cmd/run.go` iterates `nsConfig.Keys` (a Go map) directly. Go map iteration
order is non-deterministic, so the injection order changes across runs.
(GitHub issue #7)

## Solution

Sort key names alphabetically before iterating. This guarantees deterministic
injection order with minimal code change.

Note: env_var collisions are impossible in the current design because the key
name IS the env var name, and map keys are unique by definition.

## Changes

1. `cmd/run.go` — collect and sort keys before the injection loop
2. `cmd/run_test.go` — add test verifying sorted iteration order

## Steps

1. Write test that verifies keys are injected in sorted order
2. Confirm test fails
3. Implement the fix (sort keys before iteration)
4. Confirm test passes
5. Check if README.md needs updating
