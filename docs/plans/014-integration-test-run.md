# 014: Integration test for `cred run`

## Goal

Create an integration test that verifies the full `cred run` flow:
config loading → keyring retrieval → environment injection → command execution.

## Approach

Build the `cred` binary and invoke it as a subprocess via `exec.Command`.
Use a temporary config directory (`XDG_CONFIG_HOME`) and the real OS keyring
to test end-to-end behavior.

## Test helper: `testcred`

A small Go binary (`testcred`) that prints specified environment variables
as `KEY=VALUE` lines to stdout. Used as the target command for `cred run`.

Location: `cmd/testcred/main.go`

```
Usage: testcred KEY1 KEY2 ...
Output: KEY1=value1\nKEY2=value2\n
```

If a requested key is not set, it prints `KEY=` (empty value).

## Test cases (`integration_test.go`)

All tests use `//go:build integration` build tag so they don't run in `make test`.

| # | Test | Description |
|---|------|-------------|
| 1 | Single key injection | Set one key, verify `cred run -- testcred KEY` outputs it |
| 2 | Multiple keys injection | Set multiple keys, verify all are injected |
| 3 | Namespace selection | Use `-n` flag to select a non-default namespace |
| 4 | Missing key in keyring | Config has a key but keyring doesn't → verify warning on stderr and other keys still work |
| 5 | Empty namespace | Namespace has no keys → verify warning and command still runs |
| 6 | Env override | Existing env var is overwritten by keyring value |

## Setup / Teardown

- **TestMain**: Build `cred` and `testcred` binaries into a temp directory
- **Each test**: Create temp config dir, write config.yml, set keyring values
- **Cleanup**: Delete keyring entries, remove temp dirs

## File structure

```
integration/
  integration_test.go   # Test cases
cmd/
  testcred/
    main.go             # Test helper binary
```

## Makefile

Add `make test-integration` target:

```makefile
test-integration:
	go test -tags integration -v ./integration/
```
