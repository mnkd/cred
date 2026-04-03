# Troubleshooting: macOS Keychain Access Issues

cred depends on macOS Keychain. It will not work if the Keychain is inaccessible.

## Access Availability

| Condition | Accessible? |
|-----------|-------------|
| After Mac restart (normal login) | Yes |
| First access after binary rebuild | Permission dialog may appear |
| Keychain manually locked | No |
| Keychain daemon unresponsive | Times out after 3 seconds |

## After Rebuilding the Binary

cred internally uses [zalando/go-keyring](https://github.com/zalando/go-keyring), which accesses the Keychain via `/usr/bin/security` on macOS. When you rebuild the binary with `go build` or `go install`, macOS may treat it as a different application, causing a permission dialog to appear on first access.

## Timeout Mechanism

cred applies a 3-second timeout to all Keychain operations. If the Keychain daemon is unresponsive, cred returns an error instead of hanging indefinitely. This follows the same design as gh CLI.

## Diagnosing with `cred doctor`

Run `cred doctor` to check Keychain availability. It performs a Write → Read → Delete cycle with a test entry.

```bash
cred doctor
# ✓ Keyring: accessible        ← OK
# ✗ Keyring: write failed: ... ← Not accessible
```

## Debugging Keychain Access Failures

You can reproduce an inaccessible Keychain state by locking the login Keychain with the `security` command.

```bash
# 1. Lock the Keychain
security lock-keychain ~/Library/Keychains/login.keychain-db

# 2. Test cred behavior
cred doctor

# 3. Unlock immediately after testing (requires login password)
security unlock-keychain ~/Library/Keychains/login.keychain-db
```

Example output when the Keychain is locked:

```
Checking keyring...
  ✗ Keyring: write failed: keyring operation timed out
Checking config...
  ✓ Config: /Users/you/.config/cred/config.yml
  ✓ Config: version 1, 2 namespace(s)
Checking consistency...
  ✗ default/BACKLOG_API_KEY: keyring error: keyring operation timed out
  ✗ default/JIRA_TOKEN: keyring error: keyring operation timed out
```

When the Keychain is locked, each operation fails after the 3-second timeout. The more registered keys you have, the longer the total wait time.

> [!WARNING]
> `security lock-keychain` locks the **entire** login Keychain. This affects not only cred but all applications that depend on the Keychain — Safari saved passwords, Mail.app account authentication, Wi-Fi passwords, `gh` CLI tokens, etc. Keep the lock duration to a minimum.
