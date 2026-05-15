# Policy Config

Policy lives at `.ai-sensitive-files/sensitive-files.yaml`
in each target repository.

## File Shape

The file has one top-level key, `sensitive_files`.
Each item describes one sensitive path policy.

```yaml
sensitive_files:
  - path: ".env"
    reason: "local environment secrets"
    tags: ["env", "secret"]
    action:
      aiignore: true
      gitignore: true
      encrypt: false
      commit_block: true
```

When the file is encrypted or restored from a secret manager,
add `decrypted_path` and `crypto`.

```yaml
sensitive_files:
  - path: ".env"
    encrypted_path: ".env.sops.yaml"
    decrypted_path: ".env"
    reason: "local environment secrets"
    tags: ["env", "secret"]
    crypto:
      method: "sops-age"
      recipients: ["age1exampleteampublickey...", "age1examplecipublickey..."]
      encrypt_command: "sops --encrypt --output {encrypted_path} {decrypted_path}"
      decrypt_command: "sops --decrypt --output {decrypted_path} {encrypted_path}"
      manual_edit: "decrypted"
    action:
      aiignore: true
      gitignore: true
      encrypt: true
      commit_block: true
```

For external secret managers,
use `crypto.secret_ref` instead of `encrypted_path`.

```yaml
sensitive_files:
  - path: ".env.ci"
    decrypted_path: ".env.ci"
    reason: "CI-only environment secrets fetched from 1Password"
    crypto:
      method: "1password"
      secret_ref: "op://Engineering/App CI/.env"
      decrypt_command: "op read {secret_ref} > {decrypted_path}"
      manual_edit: "none"
    action:
      aiignore: true
      gitignore: true
      encrypt: true
      commit_block: true
```

## Fields

Top-level field:

- `sensitive_files`: top-level list

Required fields in each `sensitive_files` item:

- `path`: relative path or basic wildcard pattern
- `reason`: why this file is sensitive
- `action`: which generated files and checks should use this path
- `crypto`: required when `action.encrypt` is true

Optional path fields:

- `encrypted_path`: encrypted artifact path, for example `.env.sops.yaml`
- `decrypted_path`:
  plaintext output path after decryption,
  required when `action.encrypt` is true
- `crypto.secret_ref`:
  external secret reference for 1Password / Bitwarden entries
  that do not have a repo-local encrypted artifact

`decrypted_path` may point outside the repository.
Example: `/workspaces/.runtime-secrets/app.env`

Use that shape when a devcontainer or similar runtime boundary
keeps plaintext out of the project workspace
and loads it as an environment file.
Outside-workspace paths are checked for plaintext drift.
They are not emitted into `.gitignore` or AI ignore files.

Required crypto fields when `action.encrypt` is true:

- `method`: encryption method name, for example `sops-age`
- `recipients`:
  age public keys allowed to decrypt the file.
  Required for `sops-age` entries.
  Useful for team and CI keys.
  Safe to commit because they are public keys
- `encrypt_command`:
  command template using `{encrypted_path}` and `{decrypted_path}`.
  Required for repo-local encrypted artifacts
- `decrypt_command`:
  command template using `{encrypted_path}`, `{decrypted_path}`,
  and optionally `{secret_ref}`
- `manual_edit`: one of `decrypted`, `encrypted`, or `none`

`action` flags:

- `aiignore: true`:
  add the path to AI/editor ignore outputs.
  Examples: `.aiignore`, `.cursorignore`, `.copilotignore`,
  and Claude `denyRead`
- `gitignore: true`:
  add the path to `.gitignore.ai-sensitive-files`.
  `check` also verifies that `.gitignore` contains the generated entry
- `encrypt: true`:
  treat the path as protected by SOPS/age, 1Password, Bitwarden,
  or another configured method.
  Include it in generated crypto plans.
  Check plaintext at `decrypted_path`
- `commit_block: true`:
  fail `check` if `path` or a repo-relative `decrypted_path`
  is git tracked

At least one `action` flag must be true for each entry.

## Path Rules

Paths are written from the target repository root.
Use `/` as the separator.

`path` appears once per `sensitive_files` item.
`encrypted_path` and `decrypted_path` are separate fields
for the same item.

Allowed path values:

| Field | Example | Meaning |
|---|---|---|
| `path` | `.env` | Repo path that should not be exposed or committed |
| `path` | `.agent-privacy-guard/entities.local.yaml` | Nested repo path |
| `path` | `secrets/**` | Basic wildcard for a repo directory |
| `encrypted_path` | `.env.sops.yaml` | Encrypted artifact stored in the repo |
| `decrypted_path` | `/workspaces/.runtime-secrets/app.env` | Devcontainer env file outside the workspace |

Rejected examples:

| Example | Why rejected |
|---|---|
| `/Users/alice/app/.env` | Absolute host path |
| `C:\Users\alice\.env` | Backslashes / Windows-style path |
| `secrets//token.env` | Empty path segment |

For normal repository files, keep paths inside the repository.
Use `/workspaces/.runtime-secrets/...` only when a devcontainer
or similar runtime boundary intentionally keeps decrypted plaintext
outside the project workspace and loads it as environment variables.
