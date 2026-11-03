# cred - Credential Manager for Local Development

A CLI tool that centrally manages API keys scattered across `.env` files using the OS keyring.
Inspired by the credential management architecture of GitHub CLI (`gh`).

## Features

- **No session management**: Uses macOS Keychain — no re-authentication like `op signin` required
- **macOS only**: Designed specifically for local development on macOS
- **Namespace organization**: Group credentials by project or environment
- **Env injection**: Automatically inject as environment variables with `cred run -- python app.py`
- **.env import**: Bulk migrate from existing `.env` files

## Installation

### Homebrew

```bash
brew install mnkd/tap/cred
```

### Go install

```bash
go install github.com/mnkd/cred@latest
```

### Build from source

```bash
git clone https://github.com/mnkd/cred.git
cd cred
go build -o cred .
```

## Usage

### `cred set` — Store a secret

```bash
# Enter value via input prompt (visible on screen)
cred set OPENAI_API_KEY

# Store with value specified directly
cred set OPENAI_API_KEY sk-xxxxxxxxxxxx

# Store with a description
cred set OPENAI_API_KEY sk-xxxxxxxxxxxx -d "OpenAI dev"

# Read from stdin (pipe)
echo "sk-xxxx" | cred set OPENAI_API_KEY -
cat token.txt | cred set GITHUB_TOKEN -

# Store in a specific namespace
cred set ANTHROPIC_API_KEY -n work

# Update an existing key's value (description is preserved)
cred set OPENAI_API_KEY sk-new-key
```

### `cred get` — Retrieve a secret

```bash
# Output value to stdout (no trailing newline)
cred get OPENAI_API_KEY

# Assign to a shell variable via command substitution
export OPENAI_API_KEY=$(cred get OPENAI_API_KEY)

# Copy to clipboard via pipe (macOS)
cred get OPENAI_API_KEY | pbcopy

# Retrieve from a specific namespace
cred get DB_PASSWORD -n work

# Use in scripts
curl -H "Authorization: Bearer $(cred get GITHUB_TOKEN)" https://api.github.com/user
```

#### Why a `%` appears at the end of the output

The value is printed without a trailing newline. Command substitution strips trailing newlines on its own, so the `$(cred get ...)` examples above behave the same either way. What the omission protects are the paths that carry the bytes through verbatim: piping to `pbcopy` / `xclip` and pasting the value somewhere else, or redirecting it to a file. A newline that survives those paths ends up inside the header value or the token being compared.

Because the output does not end with a newline, zsh appends a reverse-video `%` before drawing the next prompt. That marker comes from zsh (the `PROMPT_SP` option), not from `cred`, and is not part of the stored value. Other shells such as bash show nothing and simply continue with the prompt on the same line.

```bash
# Confirm there is no trailing newline (no 0a at the end of the output)
cred get OPENAI_API_KEY | xxd | tail -1

# Append a newline when reading the value in a terminal
cred get OPENAI_API_KEY && echo
```

### `cred delete` — Delete a secret

```bash
# Delete from both keyring and config file
cred delete OPENAI_API_KEY

# Aliases (rm / remove) work the same
cred rm OPENAI_API_KEY
cred remove OPENAI_API_KEY

# Delete from a specific namespace
cred delete AWS_SECRET_KEY -n work
```

### `cred list` — List registered keys

```bash
# List keys in the current namespace
cred list

# Alias
cred ls

# List keys in a specific namespace
cred list -n work

# List keys across all namespaces
cred list --all
cred ls -a
```

### `cred import` — Import from .env file

```bash
# Bulk register .env file contents into the keyring
cred import .env

# Preview (dry run, nothing is saved)
cred import .env --dry-run

# Overwrite existing keys
cred import .env --overwrite

# Import into a specific namespace
cred import .env.local -n dev

# Preview then overwrite
cred import .env --dry-run
cred import .env --overwrite
```

### `cred run` — Run a command with injected environment variables

```bash
# Inject all credentials from the current namespace as environment variables
cred run -- python app.py
cred run -- node server.js

# Inject from a specific namespace
cred run -n work -- ./run-local.sh

# Verify injected environment variables
cred run -- env | grep OPENAI

# Pass to Docker
cred run -- docker run --env OPENAI_API_KEY app:latest

# Run multiple commands via shell
cred run -- sh -c 'echo "key=$OPENAI_API_KEY" && python app.py'
```

### `cred namespace` — Namespace management

```bash
# List namespaces (* marks the current default)
cred ns list

# Create a new namespace
cred ns create work
cred ns create dev -d "API keys for development"

# Show the default namespace
cred ns default

# Switch the default namespace
cred ns default work

# Delete a namespace (with confirmation prompt)
cred ns delete work

# Skip confirmation and delete
cred ns delete staging --force

# Operate with a specific namespace
cred set API_KEY -n work
cred get API_KEY -n work
cred list -n work
cred run -n work -- python app.py
```

### Shell integration examples

```bash
# Auto-export in .bashrc / .zshrc
export OPENAI_API_KEY=$(cred get OPENAI_API_KEY)
export ANTHROPIC_API_KEY=$(cred get ANTHROPIC_API_KEY)

# Copy to clipboard via pipe
cred get OPENAI_API_KEY | pbcopy    # macOS
cred get OPENAI_API_KEY | xclip     # Linux

# Pass secrets to other tools
gh secret set OPENAI_API_KEY --body "$(cred get OPENAI_API_KEY)"

# Migrate from .env to cred
cred import .env --dry-run   # Preview first
cred import .env              # Import
rm .env                       # Remove the original file
```

### `cred doctor` — Environment diagnostics

```bash
# Check keyring availability, config file, and consistency
cred doctor
# ✓ Keyring: accessible
# ✓ Config: ~/.config/cred/config.yml
# ✓ default/OPENAI_API_KEY
# ⚠ default/OLD_KEY: in config but not in keyring
```

## Config File

Metadata is stored in `~/.config/cred/config.yml`.
**Secret values are never stored in the config file** — they are stored only in the OS keyring.

```yaml
version: 1
default_namespace: default
namespaces:
  default:
    keys:
      OPENAI_API_KEY:
        description: "OpenAI API Key (dev)"
        created_at: "2026-03-31T00:00:00Z"
        updated_at: "2026-03-31T00:00:00Z"
      ANTHROPIC_API_KEY: {}
  work:
    description: "Work credentials"
    keys:
      AWS_ACCESS_KEY_ID: {}
```

## Testing

### Unit tests

```bash
make test
```

### Integration tests

Integration tests verify the full `cred run` flow (config → keyring → env injection → command execution) using the real OS keyring.

```bash
make test-integration
```

### Vulnerability check

Reports known vulnerabilities in dependencies and the standard library that the code actually calls. The standard library is checked against the Go version you run it with. CI runs it weekly and on PRs that change `go.mod` or `go.sum`.

```bash
make vulncheck
```

### Manual verification with `testcred`

`testcred` is a helper that prints specified environment variables as `KEY=VALUE` lines. Use it to manually verify `cred run` behavior.

```bash
# Build testcred
go build -o testcred ./cmd/testcred

# Basic usage: prints the value of each specified key
TEST_VAR=hello ./testcred TEST_VAR
# TEST_VAR=hello

# Verify cred run injects credentials
cred run -- ./testcred OPENAI_API_KEY ANTHROPIC_API_KEY

# Verify a specific namespace
cred run -n work -- ./testcred AWS_ACCESS_KEY_ID
```

## Architecture

```
cred
├── main.go                       # Entry point
├── cmd/
│   ├── root.go                   # Root command + global flags
│   ├── set.go                    # cred set
│   ├── get.go                    # cred get
│   ├── delete.go                 # cred delete
│   ├── list.go                   # cred list
│   ├── run.go                    # cred run -- COMMAND
│   ├── import.go                 # cred import .env
│   ├── doctor.go                 # cred doctor
│   ├── namespace.go              # cred ns {list,create,default,delete}
│   └── testcred/
│       └── main.go               # Test helper: prints env vars as KEY=VALUE
├── integration/
│   └── integration_test.go       # Integration tests for cred run
└── internal/
    ├── keyring/
    │   └── keyring.go            # OS keyring wrapper (3-second timeout)
    ├── config/
    │   └── config.go             # YAML config file management
    └── env/
        └── env.go                # .env file parser
```

### Keyring storage format

```
service: "cred:{namespace}"    (e.g., "cred:default", "cred:work")
user:    "{key_name}"          (e.g., "OPENAI_API_KEY")
secret:  "{actual_value}"      (e.g., "sk-xxxxxxxxxxxx")
```

## Comparison with gh CLI

| Feature | gh CLI | cred |
|---------|--------|------|
| Backend | zalando/go-keyring | zalando/go-keyring (same) |
| Key design | `gh:{hostname}` + `{username}` | `cred:{namespace}` + `{key_name}` |
| Purpose | GitHub tokens only | General-purpose (any API key) |
| Env injection | N/A | `cred run` |
| .env import | N/A | `cred import` |

## Dependencies

- [spf13/cobra](https://github.com/spf13/cobra) — CLI framework
- [zalando/go-keyring](https://github.com/zalando/go-keyring) — Cross-platform keyring
- [go-yaml/yaml](https://github.com/go-yaml/yaml) — YAML parser
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) — Secure prompt input

## License

MIT
