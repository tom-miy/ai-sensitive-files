# Scope and Safety Checks

This page explains the project boundary in concrete terms.
The target workflow is simple:
a repository has known sensitive paths.
Developers want one policy that keeps AI ignore files,
gitignore entries, encryption/decryption plans,
and commit checks aligned.

Example paths:

- `.env`
- `.aws/credentials`
- `.agent-privacy-guard/entities.local.yaml`
- `../.runtime-secrets/.env` in a devcontainer-style setup
- local files materialized from 1Password or Bitwarden

## Checks This Tool Provides

- A policy entry is invalid or incomplete.
- A generated ignore/config file is stale.
- `.gitignore` does not include generated entries
  that policy says should be git-ignored.
- A `commit_block` path is already tracked by git.
- A decrypted plaintext file exists where policy says it should not.
- A decrypted plaintext file is newer than its encrypted artifact.
- The encrypted file should be refreshed before commit.
- An encrypted artifact is newer than local plaintext.
- Developers should decrypt again before editing.

## Work Left To Other Tools

- Prompt inspection and anonymization belong in a prompt gateway.
- Dangerous command blocking belongs in local agent hooks.
- Runtime file-read blocking belongs in a sandbox.
- It can also belong in devcontainer or agent-specific permissions.
- Unknown secret discovery belongs in secret scanners.
- SOPS/age keys stay in those tools.
- Password-manager access stays in those tools.
- Repository access stays in repository permissions.

Use this tool with secret scanners for unknown leaks.
Use SOPS/age or password managers for actual secret storage.
Use runtime controls such as devcontainers or local agent hooks
when AI file reads must be blocked at the OS/tool level.
