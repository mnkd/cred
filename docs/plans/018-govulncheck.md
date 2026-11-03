# Plan 018: Run govulncheck weekly and on dependency changes

## Problem

Nothing checks dependencies or the standard library for known vulnerabilities.

## Solution

1. Add `make vulncheck` (version pinned with `GOVULNCHECK_VERSION`)
2. Add `.github/workflows/govulncheck.yml`
   - Triggers: weekly on main, PRs that change `go.mod` / `go.sum` / `Makefile` / the workflow itself, and manual runs
   - Use `go-version: stable`, as release builds do, because the standard library is checked against the Go version that runs govulncheck
3. Do not run it on every PR and do not make it a required check
   - The result depends on that day's vulnerability database, so a new advisory would block unrelated PRs

## Trade-offs

- `golang/govulncheck-action` is not used. It installs govulncheck `@latest`, so local runs and CI could use different versions
- A PR that only changes `Makefile` may get a red check from a newly published vulnerability. It does not block merging

## Steps

1. Add the `vulncheck` target and the workflow
2. Confirm `make vulncheck` passes
3. Run `actionlint` on the workflow
4. Check if README.md needs updating
