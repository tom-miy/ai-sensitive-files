# ai-sensitive-files

`ai-sensitive-files` is a small security-focused developer tool
for managing sensitive file paths in a repository.

As a portfolio project, it focuses on a small operational problem
that appears in real projects.
Files such as `.env`, `.aws/credentials`, local entity dictionaries,
and decrypted password-manager exports can contain credentials
or organization-specific information.
If those files enter AI context or git history,
the result can be data exposure, missed review,
and hard-to-clean repository history.

In practice, the rules for those files are often scattered
across ignore files, encryption commands, and hooks.

This repo puts those rules in `.ai-sensitive-files/sensitive-files.yaml`.
It then generates AI/editor ignore files, gitignore entries,
crypto plans, and commit checks from the same source.

What this demonstrates:

- policy-driven generation instead of hand-maintained ignore files
- conservative installer behavior that prints follow-up steps
  instead of editing user-owned config
- SOPS/age and 1Password/Bitwarden integration points
  without owning key management
- commit-time checks for tracked plaintext files
  and stale generated output
- clear separation from prompt sanitization, command blocking,
  and secret scanning tools

日本語版: [README.ja.md](README.ja.md)

## What It Generates

From `.ai-sensitive-files/sensitive-files.yaml`, the CLI writes:

- `.aiignore`, `.cursorignore`, `.copilotignore`:
  common AI/editor ignore intent
- `.gitignore.ai-sensitive-files`:
  entries to review and merge into `.gitignore`
- `generated/claude-code-deny-read.json`:
  Claude Code `denyRead` snippet
- `generated/ai-agent-guidance.md`:
  Codex / Cursor / Copilot guidance snippet
- `generated/ai-sensitive-files.summary.md`:
  human-readable policy summary
- `generated/encryption-targets.txt`: SOPS/age target list
- `generated/decryption-targets.txt`:
  encrypted-to-decrypted path mapping
- `generated/secret-sources.txt`:
  1Password / Bitwarden references used to create local plaintext files
- `generated/crypto-plan.md`:
  configured encrypt/decrypt commands and manual-edit policy

Existing `.gitignore` is never edited directly.
`check` verifies that entries generated in `.gitignore.ai-sensitive-files`
have been merged into `.gitignore`.

## Commands

Install the development and integration tools with mise:

```bash
mise trust .
mise install
mise run install-cli
```

`mise trust .` approves this repository's local tool definition.
`mise install` installs Go, SOPS, age, and Lefthook.
They are used for local validation and examples.
`mise run install-cli` builds this repository's CLI into
`.bin/ai-sensitive-files`.
mise adds `.bin` to PATH while you are in this repo.
These commands do not create encryption keys
or trust application hook config.

Validate a policy before trusting generated output:

```bash
ai-sensitive-files validate --config .ai-sensitive-files/sensitive-files.yaml
```

Generate ignore files and snippets from the policy:

```bash
ai-sensitive-files generate --config .ai-sensitive-files/sensitive-files.yaml --out .
```

Check whether commit-blocked files are tracked.
Also check plaintext targets, `.gitignore`, and generated files:

```bash
ai-sensitive-files check --config .ai-sensitive-files/sensitive-files.yaml
```

Show the policy in a review-friendly format:

```bash
ai-sensitive-files list --config .ai-sensitive-files/sensitive-files.yaml
```

All commands support `--json`.

## Install Into An App Repo

### 1. Install the sample policy into the target app repository

```bash
bash install.sh --target /path/to/app
```

This creates `/path/to/app/.ai-sensitive-files/`.
It copies `templates/sensitive-files.example.yaml` to
`/path/to/app/.ai-sensitive-files/sensitive-files.yaml`.
It does not create `configs/`.

### 2. Validate the policy inside the target app repository

```bash
cd /path/to/app
ai-sensitive-files validate --config .ai-sensitive-files/sensitive-files.yaml
```

### 3. Generate ignore files and derived snippets from the policy

```bash
ai-sensitive-files generate --config .ai-sensitive-files/sensitive-files.yaml --out .
```

### 4. Apply follow-up changes manually

`install.sh` does not overwrite existing policy files.
It does not append to `.gitignore`, initialize SOPS/age,
or edit Lefthook config.
After generation, read the printed guidance.
Add only the entries you want to `.gitignore` or `lefthook.yml`.

## Policy Example

This example manages one logical secret, represented by two paths:

- `.env`:
  local plaintext used by the app and developers.
  It should be hidden from AI tools and blocked from commits
- `.env.sops.yaml`:
  encrypted copy of the same data.
  This is the file intended to be committed

The `action` block says where this path is used:
AI ignore output, gitignore output, encryption/decryption checks,
and commit blocking.

```yaml
sensitive_files:
  - path: ".env"                      # sensitive plaintext path to hide from AI and block from commits
    encrypted_path: ".env.sops.yaml"  # encrypted artifact managed by SOPS/age
    decrypted_path: ".env"            # local plaintext output after decryption
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

The YAML policy is the source of truth.
It keeps ignore files, gitignore entries, encryption targets,
decrypted plaintext paths, encrypt/decrypt commands,
manual-edit rules, and commit checks on the same boundary.

For secrets stored outside git, use `crypto.secret_ref` instead of `encrypted_path`:

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

With this shape, the source of truth is the secret manager item.
It is not a repo-local encrypted file.
The generated plan records how to materialize local plaintext.
`check` still blocks that plaintext from commits.

## Lefthook

Merge this into an existing `lefthook.yml` after review:

```yaml
pre-commit:
  commands:
    ai-sensitive-files:
      run: ai-sensitive-files check --config .ai-sensitive-files/sensitive-files.yaml
```

The hook is a commit gate.
It catches stale generated files, missing `.gitignore` entries,
tracked `commit_block` paths,
plaintext files that policy says should not exist,
and decrypted files newer than their encrypted artifacts.

## Secret Storage

`encrypt: true` marks an entry as protected by encryption
or an external secret manager.

`crypto.method` records the mechanism.
Examples: `sops-age`, `1password`, `bitwarden`.

`crypto.encrypt_command` and `crypto.decrypt_command`
record how encryption and decryption are run.
This makes local plaintext creation reviewable.
It also shows how plaintext should be re-encrypted before commit.

When no encrypted artifact is stored in the repository,
use `crypto.secret_ref`.
For 1Password or Bitwarden,
`secret_ref` points to the protected value's external location.

Supported policy shapes:

- SOPS/age:
  `encrypted_path` is a repo-local encrypted artifact.
  `decrypted_path` is the local plaintext output
- 1Password / Bitwarden:
  `secret_ref` points to the external secret.
  `decrypt_command` materializes `decrypted_path` locally

For stronger separation, use a devcontainer.
Write decrypted environment files outside the workspace:

```yaml
sensitive_files:
  - path: ".env"                              # path that must not appear in the repo
    encrypted_path: ".env.sops.yaml"          # encrypted artifact in the repo
    decrypted_path: "/workspaces/.runtime-secrets/app.env" # env file loaded by devcontainer
    crypto:
      method: "sops-age"
      decrypt_command: "umask 077; mkdir -p /workspaces/.runtime-secrets; sops --decrypt {encrypted_path} > {decrypted_path}"
      encrypt_command: "umask 077; sops --encrypt --output {encrypted_path} {decrypted_path}"
      manual_edit: "decrypted"
    action:
      aiignore: true
      gitignore: true
      encrypt: true
      commit_block: true
```

In that layout, devcontainer loads `/workspaces/.runtime-secrets/app.env`.
It uses `runArgs: ["--env-file", "/workspaces/.runtime-secrets/app.env"]`.
The project workspace only contains the encrypted artifact
and generated policy files.
`ai-sensitive-files` still checks the external plaintext path.
It does not add that outside-workspace path to `.gitignore`
or AI ignore files.

`crypto.manual_edit` controls local editing:

- `decrypted`:
  local plaintext editing is allowed.
  If plaintext is newer than the encrypted file, `check` fails.
  The file must be re-encrypted before commit
- `encrypted`:
  the encrypted artifact is the edit target.
  If plaintext output remains, `check` fails
- `none`:
  neither side should be edited manually.
  If plaintext output remains or changes, `check` fails

If an encrypted file is newer than the decrypted file,
`check` prints a warning to refresh decryption.
This helps catch updates after pulling changes.

For password managers, the generated plan records the `op` or `bw` command.
The tool does not log in, fetch, or store password-manager credentials.

Install SOPS and age with `mise install` when using SOPS/age.
Install and authenticate `op` or `bw` separately when using 1Password or Bitwarden.

This repo does not implement key management.
age public keys, also called recipients, are safe to commit
because they are public.
age identities and other private keys must not be committed.

Lefthook can block private key files from being committed.
It cannot create keys, prevent key sharing,
or enforce local file permissions.
The provided `.sops.yaml` example uses dummy recipients
for a team key and a CI key.
Teams should define rotation and recovery outside this repo.

Encrypted files may be tracked when your policy allows it.
Plaintext sensitive files should generally not be tracked.

## Demo

The demo validates the sample policy and generates files.
It creates a plaintext `.agent-privacy-guard/entities.local.yaml`.
Then it shows that `check` blocks it and prints the Claude Code snippet.

```bash
bash scripts/demo.sh
```

With mise:

```bash
mise run demo
```

## Related Repositories

| Repository | Responsibility |
|---|---|
| agent-privacy-guard | Inspecting, anonymizing, and enforcing policy for prompts sent to external LLMs / MCP |
| secure-dev-hooks | Guardrails for AI agent local operations, dangerous commands, and file access |
| ai-sensitive-files | YAML-managed files hidden from AI, encryption targets, and commit-block targets |

## Project Boundary

This project starts after a team already knows which paths are sensitive.

It answers practical repository questions:

- Should `.env` be hidden from AI tools?
- Should `.env` be ignored by git?
- Where is the encrypted or password-manager-backed source?
- Where does decrypted plaintext appear during local development?
- Should commit checks fail if that plaintext is tracked or out of sync?

Finding unknown leaked credentials is a different job.
Use gitleaks, trufflehog, or GitHub secret scanning for that.
Access to SOPS, age, 1Password, Bitwarden,
and the repository is also handled by those systems.
`ai-sensitive-files` keeps the repository-side file rules explicit and testable.
