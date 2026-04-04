# ADR-004: Remove the `env_var` field from KeyMeta

## Status

Accepted

## Context

Each key in the config has an `env_var` field intended to allow injecting a secret under a different environment variable name than the key itself:

```yaml
keys:
  OPENAI_API_KEY:
    env_var: OPENAI_API_KEY   # always identical to the key name
```

In practice, `cred set` and `cred import` both copy the key name into `env_var` automatically, and there is no CLI flag to specify a different value. The only way to set a custom `env_var` is to manually edit the YAML — an undocumented workflow.

### Use cases considered

1. **Inject under a different name** (e.g., key `openai-dev`, env var `OPENAI_API_KEY`): No CLI support exists for this. Requires manual YAML editing.
2. **Inject the same secret under multiple names**: The 1:1 key-to-env_var design cannot support this regardless.

### Impact on `cred list`

The `KEY` and `ENV_VAR` columns in `cred list` output are virtually always identical, adding visual noise without informational value.

## Decision

Remove the `env_var` field from `KeyMeta`. Use the key name directly as the environment variable name in `cred run`.

## Rationale

- The field is effectively unused — no CLI path sets a value different from the key name
- Removing it simplifies the config schema, `cred list` output, and internal code
- The feature can be re-introduced with proper CLI support if a real need arises
- YAGNI: do not maintain code for a hypothetical use case with no current users

## Consequences

- Config migration: existing `env_var` fields in YAML will be ignored / removed on next write
- `cred list` output drops the `ENV_VAR` column
- `cred run` uses the key name directly as the environment variable name
- Users who manually edited `env_var` to differ from the key name will lose that customization
