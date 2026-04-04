# Release Flow

## Versioning

Follows [Semantic Versioning](https://semver.org/):

| Type | When to bump | Example |
|---|---|---|
| MAJOR | Breaking changes | `1.0.0` → `2.0.0` |
| MINOR | New features (backward-compatible) | `0.1.0` → `0.2.0` |
| PATCH | Bug fixes | `0.1.0` → `0.1.1` |

The git tag is the single source of truth for the version. No source code changes are needed for a release.

## Steps

1. Merge all changes to `main`
2. Create a version tag:
   ```bash
   git tag v0.2.0
   ```
3. Push the tag:
   ```bash
   git push origin v0.2.0
   ```
4. GitHub Actions runs GoReleaser automatically:
   - Builds binaries (darwin amd64/arm64)
   - Creates a GitHub Release
   - Pushes the Homebrew formula to `mnkd/homebrew-tap`

## Verification

After the release workflow completes:

```bash
# Check the GitHub Release
gh release view v0.2.0

# Verify Homebrew installation
brew update
brew install mnkd/tap/cred
cred version
```

## How version is embedded

GoReleaser injects the version via `ldflags` at build time:

```
-X github.com/mnkd/cred/cmd.version={{.Version}}
```

- Release build (`brew install`, GitHub Release binary): shows the tag version (e.g., `0.2.0`)
- Development build (`go run .`, `make install`): shows `dev`
