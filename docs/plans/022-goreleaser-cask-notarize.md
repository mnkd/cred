# Plan 022: Distribute a signed and notarized Homebrew Cask

## Problem

GoReleaser prints two deprecation warnings on every release:

1. `archives.format` was renamed to `formats`
2. `brews` is deprecated in favor of `homebrew_casks`
   - The formula GoReleaser generates only downloads a prebuilt binary. Homebrew formulae are meant to build from source; prebuilt binaries belong in casks

## Constraints

- Homebrew applies the macOS quarantine attribute to cask downloads, so Gatekeeper checks the binary on first run. Formulae get no quarantine, which is why the current unsigned binary runs
- An unsigned or ad-hoc signed cask binary is blocked by Gatekeeper
- GoReleaser documents removing the quarantine attribute in a post-install hook. That bypasses Gatekeeper, and Homebrew states it does not normalise such bypasses. Other CLI casks (`1password-cli`, `codex`) ship binaries signed with a Developer ID and notarized
- Signing does not change Keychain behavior: go-keyring calls `/usr/bin/security`, so Keychain access control applies to that binary, not to cred

## Solution

1. Rename `archives.format` to `formats`
2. Replace `brews` with `homebrew_casks`
   - The cask goes to `Casks/cred.rb` in `mnkd/homebrew-tap` (the default directory)
   - No quarantine hook
3. Sign and notarize the darwin binaries with `notarize.macos` (quill, available in GoReleaser OSS)
   - `enabled: true`, not `isEnvSet`: a release without the secrets fails instead of publishing a cask that Gatekeeper blocks
   - Snapshot builds skip notarization on their own
4. Pass the signing secrets to GoReleaser in `release.yml`

### Secrets (repository secrets of `mnkd/cred`)

| Secret | Content |
|---|---|
| `MACOS_SIGN_P12` | Developer ID Application certificate and private key, `.p12`, base64 |
| `MACOS_SIGN_PASSWORD` | Password of the `.p12` |
| `MACOS_NOTARY_KEY` | App Store Connect API key, `.p8`, base64 |
| `MACOS_NOTARY_KEY_ID` | Key ID of the API key |
| `MACOS_NOTARY_ISSUER_ID` | Issuer ID of the API key |

## Migrating existing users (in `mnkd/homebrew-tap`, after the first cask release)

Homebrew migrates a formula to a cask in the same tap when the formula file is deleted and listed in `tap_migrations.json`:

1. Delete `cred.rb` (the formula)
2. Add `tap_migrations.json`:

   ```json
   {
     "cred": "mnkd/tap"
   }
   ```

On the next `brew update`, users who installed the formula are told that `mnkd/tap/cred` "has been migrated from a formula to a cask". Homebrew installs the cask automatically only if the cask is trusted (Tap Trust, Homebrew 6+). Users who trusted only the formula run:

```sh
brew trust --cask mnkd/tap/cred
brew install --cask mnkd/tap/cred
brew uninstall --formula --force cred
```

`/opt/homebrew/bin/cred` stays the same path, so MCP configs need no change.

Do this right after the first cask release. Until then the tap holds both a formula and a cask named `cred`.

## Changes

1. `.goreleaser.yml`
2. `.github/workflows/release.yml`
3. `README.md` — installation and migration note

## Verification

- `goreleaser check` reports no deprecations
- `goreleaser release --snapshot --clean` generates `dist/homebrew/Casks/cred.rb`
- `goreleaser release --clean --skip=publish,announce,validate` without the signing secrets fails at the notarize step
- On the next tag:
  - The release log shows signing and notarization for both darwin binaries
  - `codesign -dv` on the installed binary shows the Developer ID; `spctl -a -vv -t install` reports `Notarized Developer ID`
  - A quarantined install runs without a Gatekeeper prompt
