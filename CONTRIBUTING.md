# Contributing to cred

Thank you for your interest in contributing to cred!

## Getting Started

### Prerequisites

- Go 1.25 or later
- macOS (cred uses macOS Keychain)

### Build

```bash
git clone https://github.com/mnkd/cred.git
cd cred
make build
```

### Test

```bash
make test
```

Tests use a mock keyring backend — they do not access the actual OS keyring.

## Development Workflow

1. Fork and clone the repository
2. Create a feature branch from `main`
3. Write tests first (TDD)
4. Implement the feature
5. Run `make test` to ensure all tests pass
6. Commit with [Conventional Commits](https://www.conventionalcommits.org/) format
7. Open a pull request against `main`

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `ci`, `perf`

## Code Style

- All comments, commit messages, and documentation in English
- User-facing messages go to `os.Stderr` (`os.Stdout` is for data output only)
- Wrap errors with `fmt.Errorf("context: %w", err)`
- Run `go vet ./...` before submitting

## Reporting Issues

Please open an issue on GitHub with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Your macOS version and Go version
