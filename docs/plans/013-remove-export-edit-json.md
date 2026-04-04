# 013: Remove export, edit, and --format json

## Goal

Reduce maintenance cost by removing non-essential features.

## Features to remove

### 1. `cred export`

- **Why**: `cred run -- env` covers the use case. Promoting `run` as the primary workflow makes `export` redundant.
- **Files to delete**: `cmd/export.go`, `cmd/export_test.go` (if exists)
- **Files to edit**: `cmd/root.go` (remove subcommand registration if any), `README.md`

### 2. `cred edit`

- **Why**: `cred set KEY newvalue` already overwrites. The editor flow adds complexity (temp file handling) for minimal benefit.
- **Files to delete**: `cmd/edit.go`, `cmd/edit_test.go` (if exists)
- **Files to edit**: `cmd/root.go` (remove subcommand registration if any), `README.md`

### 3. `--format json`

- **Why**: Adds output branching to every display command. Low usage for a personal CLI tool.
- **Files to edit**:
  - `cmd/root.go`: Remove `outputFormat` var, `--format` flag, `isJSON()` func
  - `cmd/get.go`: Remove JSON output branch
  - `cmd/list.go`: Remove JSON output branch
  - `cmd/namespace.go`: Remove JSON output branch
  - `README.md`: Remove JSON section and comparison table entry

## Steps

1. Delete `cmd/export.go` and related test files
2. Delete `cmd/edit.go` and related test files
3. Remove `--format json` flag and all JSON output branches
4. Update `README.md` (usage sections, architecture diagram, comparison table)
5. Run `make test` to confirm nothing breaks
6. Delete plan files for removed features: `005-export.md`, `006-edit.md`, `012-format-json.md`

## Done criteria

- `cred export`, `cred edit`, `--format json` no longer exist
- All tests pass
- README reflects the current feature set
