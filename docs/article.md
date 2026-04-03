# cred — Centrally Manage API Keys with macOS Keychain

> [!NOTE]
> The issues I mention about 1Password CLI (`op`) in this article may simply be due to my lack of understanding of `op`. There may also be security issues in cred's design and implementation. If you notice anything, please let me know.

## Introduction

Have you ever experienced any of these when managing API keys in your local development environment?

- `.env` files scattered across multiple projects, with no clear idea which one is up to date
- Nearly (or actually) committing `.env` to Git by accident
- Using 1Password CLI (`op`) but getting frustrated by `op signin` every time the session expires
- Manually updating `.env` files across multiple projects whenever a new API key is issued

I built **cred** to solve these problems.

## What is cred?

A Go CLI tool that **centrally manages** secrets such as API keys **using macOS Keychain**. You can use it alongside `.env` files, or operate entirely without them.

### Features

| Feature | Description |
|---------|-------------|
| **No session management** | Uses macOS Keychain — no re-authentication like `op signin` required |
| **macOS only** | Focused on local development (macOS), prioritizing simplicity |
| **Namespace organization** | Group credentials by project or environment |
| **Env injection** | Automatically inject as environment variables with `cred run -- python app.py` |
| **.env compatible** | Import from and export to existing `.env` files |

## Inspired by gh CLI's Credential Management

cred's design is inspired by the credential management architecture of GitHub CLI (`gh`).

### How gh CLI Manages Tokens

gh CLI securely manages GitHub access tokens using the OS keyring. Analyzing its implementation reveals the following design:

**Multi-layer fallback strategy:**

```
1. Environment variables (GH_TOKEN / GITHUB_TOKEN)  ← For CI/CD
2. OS keyring (macOS Keychain, etc.)                 ← Default
3. Plaintext config file (~/.config/gh/)             ← Legacy compatibility
```

**Keyring access with timeout:**

gh CLI doesn't use [zalando/go-keyring](https://github.com/zalando/go-keyring) directly. Instead, it implements a wrapper with a 3-second timeout using goroutines and channels. This prevents hanging when the Keychain daemon is unresponsive.

**Keyring storage format:**

```
service: "gh:{hostname}"     (e.g., "gh:github.com")
user:    "{username}"        (e.g., "octocat")
secret:  "gho_xxxxxxxxxxxx"  (token)
```

### What I Noticed About gh CLI

gh CLI's credential management is robust and well-designed. However, I also noticed some limitations:

- **GitHub tokens only**: Cannot manage other API keys (OpenAI, AWS, Slack, etc.)
- **No env injection**: `gh` has no mechanism to pass secrets as environment variables
- **No `.env` integration**: No way to migrate from existing `.env` files

"What if I generalized this design to make local secret management easier?" — That's what motivated me to create cred.

### Comparing gh CLI and cred

While learning a lot from gh CLI's design, cred extends it as a general-purpose tool.

**In common:**

| Feature | gh CLI | cred |
|---------|--------|------|
| Keyring library | zalando/go-keyring | zalando/go-keyring (same) |
| Timeout mechanism | goroutine + 3s timeout | Same pattern |
| Metadata management | YAML config file | YAML config file |
| Secret storage | OS keyring only | OS keyring only |

**What cred adds:**

| Feature | gh CLI | cred |
|---------|--------|------|
| Key design | `gh:{hostname}` + `{username}` | `cred:{namespace}` + `{key_name}` |
| Purpose | GitHub tokens only | General-purpose (any API key) |
| Namespace | Per host | Freely created |
| Env injection | N/A | `cred run -- command` |
| .env import | N/A | `cred import .env` |
| .env export | N/A | `cred export` |
| JSON output | N/A | `--format json` |
| Diagnostics | N/A | `cred doctor` |

## Installation

```bash
go install github.com/mnkd/cred@latest
```

## Basic Usage

### Storing and Retrieving Secrets

```bash
# Store (enter value via prompt — not saved in shell history)
cred set OPENAI_API_KEY

# Read from file (not saved in shell history)
cat token.txt | cred set OPENAI_API_KEY -

# Retrieve (output to stdout)
cred get OPENAI_API_KEY

# Assign to shell variable
export OPENAI_API_KEY=$(cred get OPENAI_API_KEY)
```

> [!WARNING]
> Passing the value directly as an argument like `cred set KEY sk-xxxx` will leave the secret in your shell history (`~/.zsh_history`, etc.). For storing secrets, use `cred set KEY` (prompted input) or `cat file | cred set KEY -` instead.

### Importing from .env Files

If you have existing `.env` files, you can bulk import them. After importing, you can continue using `.env` alongside cred, or switch to cred entirely and eliminate `.env` files.

```bash
# Preview
cred import .env --dry-run

# Import
cred import .env
```

### Organizing by Project with Namespaces

```bash
# Create namespaces
cred ns create work -d "Work projects"
cred ns create personal -d "Personal projects"

# Store and retrieve with a specific namespace
cred set API_KEY -n work
cred get API_KEY -n work

# Inject all credentials from a namespace as environment variables
cred run -n work -- python app.py
```

### Automatic Environment Variable Injection

With `cred run`, you can pass environment variables to commands without `.env` files.

```bash
cred run -- python app.py
cred run -- node server.js
cred run -n production -- ./deploy.sh
```

## Use Case: Integration with a Next.js App

Here's how cred is used in an internal dashboard built with Next.js.

### Design Approach: The "Key Name Reference" Pattern

In `.env.local`, store **only the cred key names, not the actual secret values**.

```env
# .env.local.example

# key name in cred command
JIRA_TOKEN_REF=JIRA_TOKEN
GEMINI_API_KEY_REF=GEMINI_API_KEY
```

The key point is the `*_REF` naming convention. These variables contain cred key names, and the actual values are retrieved via cred at runtime.

### Secret Retrieval Helper

```typescript
// src/lib/secrets.ts
import { execSync } from "child_process";

const cache = new Map<string, string>();

export async function getSecret(ref: string): Promise<string> {
  if (cache.has(ref)) return cache.get(ref)!;
  const value = execSync(`cred get "${ref}"`, {
    encoding: "utf-8",
    timeout: 10_000,
  }).trim();
  cache.set(ref, value);
  return value;
}
```

This function:
1. Retrieves secrets from macOS Keychain via the `cred get` command
2. Caches retrieved values in memory to speed up subsequent calls within the same process

### Usage in API Clients

```typescript
// Jira client
const tokenRef = process.env.JIRA_TOKEN_REF;    // "JIRA_TOKEN"
const token = await getSecret(tokenRef);          // cred get JIRA_TOKEN
const auth = Buffer.from(`${email}:${token}`).toString("base64");

// Backlog client
const apiKeyRef = process.env.BACKLOG_API_KEY_REF; // "BACKLOG_API_KEY"
const apiKey = await getSecret(apiKeyRef);           // cred get BACKLOG_API_KEY
```

### Benefits of This Pattern

| Aspect | Direct values in `.env` | cred + `*_REF` pattern |
|--------|------------------------|------------------------|
| **Git safety** | Risk of accidentally committing `.env` | `.env` contains only key names — no value leak even if committed |
| **Secret rotation** | Manually update `.env` in every project | Update once with `cred set`, reflected across all projects |
| **Team sharing** | `.env.example` contains only dummy values | `.env.example` can list key names, making setup steps clear |
| **New members** | Copy `.env` and replace values | Just register each key with `cred set` |

### Note: Secret Protection with Claude Code

When using AI coding tools, handling `.env` files becomes even more important. Claude Code has guards for `.env` files, but they are not foolproof.

**You can deny `.env` reads with permission rules:**

```json
{
  "permissions": {
    "deny": ["Read(./.env)", "Read(./.env.*)"]
  }
}
```

However, the official documentation notes:

> Read and Edit deny rules apply to Claude's built-in file tools, not to Bash subprocesses. A `Read(./.env)` deny rule blocks the Read tool but does not prevent `cat .env` in Bash. For OS-level enforcement that blocks all processes from accessing a path, enable the sandbox.
>
> — [Claude Code Permissions](https://docs.anthropic.com/en/docs/claude-code/permissions)

In other words, deny rules block Claude's Read/Edit tools, but cannot prevent `cat .env` via Bash. For complete protection, you need OS-level restrictions via the [sandbox feature](https://docs.anthropic.com/en/docs/claude-code/security).

With cred's `*_REF` pattern, since `.env` contains only key names, the risk of secret value leakage is significantly reduced — even if Claude reads the `.env` file or if it gets committed accidentally.

## Other Useful Tips

### Global Setup in .zshrc

```bash
# .zshrc
export OPENAI_API_KEY=$(cred get OPENAI_API_KEY)
export ANTHROPIC_API_KEY=$(cred get ANTHROPIC_API_KEY)
```

### Registering GitHub Secrets

```bash
gh secret set OPENAI_API_KEY --body "$(cred get OPENAI_API_KEY)"
```

### Copy to Clipboard

```bash
cred get OPENAI_API_KEY | pbcopy
```

### Diagnostics

```bash
cred doctor
# ✓ Keyring: accessible
# ✓ Config: ~/.config/cred/config.yml
# ✓ default/OPENAI_API_KEY
# ⚠ default/OLD_KEY: in config but not in keyring
```

## Conclusion

cred was born from learning gh CLI's robust credential management design and thinking, "What if I generalized this to make things easier?"

- Stable keyring access with the same go-keyring + timeout wrapper as gh CLI
- Extended as a general-purpose tool with namespaces, `cred run`, and `.env` import/export
- The `*_REF` pattern allows safe secret references from application code
- Reduces the risk of accidental `.env` commits and centralizes secret management across projects

If you're struggling with secret management in your local development environment, give it a try.

```bash
go install github.com/mnkd/cred@latest
```
