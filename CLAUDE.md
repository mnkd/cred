# cred

A Go CLI tool for centrally managing API keys and other secrets via the OS keyring (macOS Keychain).

## Build & Test

```bash
make build       # go mod tidy + go build
make test        # go test ./...
make install     # go install .
```

## Architecture

- `cmd/` — cobra subcommands
- `internal/keyring/` — timeout wrapper for zalando/go-keyring
- `internal/config/` — read/write ~/.config/cred/config.yml
- `internal/env/` — .env file parser

## Coding Rules

- macOS only. No Windows / Linux support needed
- Never store secret values in config files (store only in OS keyring)
- User-facing messages go to `os.Stderr` (`os.Stdout` is reserved for data output)
- Wrap errors with `fmt.Errorf("context: %w", err)`
- All code comments, commit messages, and documentation must be written in English

## Workflow

- Create a plan in docs/plans/ before starting implementation
- Follow TDD: write tests first -> confirm failure -> implement -> confirm tests pass
- After code changes, check whether README.md needs updating
  - Targets: usage, architecture diagram, comparison table, dependencies
