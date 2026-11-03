# Plan 017: Fix vulnerable dependencies and build with the latest Go

## Problem

- `govulncheck` reports a vulnerability in `golang.org/x/sys` v0.42.0 (GO-2026-5024, fixed in v0.44.0). It affects `golang.org/x/sys/windows` only and cred does not call it, but the fix is a patch-level update.
- `go.mod` declares `go 1.25.0`, and CI and release builds read it through `go-version-file`. setup-go installs exactly 1.25.0, so release binaries miss standard library security fixes. Go 1.25 is no longer supported.

## Solution

1. Update `golang.org/x/sys` to v0.44.0
2. Raise the `go` directive to 1.26.0
   - The `go` directive states the minimum version. 1.26 is the oldest supported release
   - govulncheck v1.9.0 also requires Go 1.26 or later
3. Use `go-version: stable` in CI and release builds so they always use the latest patch release

## Trade-offs

- CI no longer builds and tests with the minimum Go version itself
  - Standard library APIs newer than the `go` directive are still caught: `go vet` (stdversion) reports them, and the compiler rejects newer language features
  - What remains unchecked is behavior that differs between toolchain versions, such as bugs and GODEBUG defaults
- The release toolchain is not pinned. Re-running the release for an old tag builds with whatever Go is stable at that time. Each binary records the version it was built with (`go version <binary>`)

## Steps

1. Update dependencies and the `go` directive
2. Confirm `go build`, `go test`, `go vet`, and govulncheck pass with the latest Go
3. Confirm `go build` passes with Go 1.26.0
4. Check if README.md needs updating
