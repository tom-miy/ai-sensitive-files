# Encryption

`encrypt: true` means the entry participates in encryption
and plaintext checks.

Use these path fields:

- `encrypted_path`: where the encrypted artifact is stored
- `decrypted_path`: where plaintext is written after decryption
- `crypto.secret_ref`:
  where an external secret manager stores the protected value

Use the `crypto` block to document the method and commands:

```yaml
crypto:
  method: "sops-age"
  recipients: ["age1exampleteampublickey...", "age1examplecipublickey..."]
  encrypt_command: "sops --encrypt --output {encrypted_path} {decrypted_path}"
  decrypt_command: "sops --decrypt --output {decrypted_path} {encrypted_path}"
  manual_edit: "decrypted"
```

Use `recipients` to record every age public key
that should be able to decrypt the file.
Examples include a team key and a CI key.
These age public keys are safe to commit
because they are not private identities.
SOPS stores recipients in `.sops.yaml`.
Keeping the same list in policy makes review
and generated crypto plans easier.

`manual_edit` controls sync checks.
`decrypted` allows local plaintext editing.
If plaintext is newer than the encrypted artifact, `check` fails.
It indicates that the file must be re-encrypted before commit.
With `encrypted` or `none`,
`check` fails when plaintext output remains.
When the encrypted artifact is newer than plaintext,
`check` warns that decryption should be refreshed.

For 1Password or Bitwarden,
omit `encrypted_path` and use `secret_ref`:

```yaml
crypto:
  method: "1password"
  secret_ref: "op://Engineering/App CI/.env"
  decrypt_command: "op read {secret_ref} > {decrypted_path}"
  manual_edit: "none"
```

In this mode, the password manager item is the source of truth.
`ai-sensitive-files` records the command and checks local plaintext handling.
It does not authenticate to the password manager or fetch secrets.

This repository expects SOPS and age to be installed through mise:

```bash
mise trust .
mise install
```

This installs SOPS and age for local use.
Install and authenticate `op` or `bw` separately
if the policy uses 1Password or Bitwarden.

The tool does not initialize SOPS, create keys, rotate keys,
or store private identities.
The example at `templates/sops/.sops.yaml.example` uses dummy recipients.

Lefthook can block private key files from being committed.
It cannot create keys, prevent key sharing,
or enforce local file permissions.

Operational guidance:

- Keep private age identities out of git.
- Mark age identity and private key paths with `commit_block: true`.
- Commit age public recipients when they are needed
  for `.sops.yaml` or policy review.
- Add all required age public recipients before sharing encrypted files
  with teammates or CI.
- Commit encrypted files only when your team has agreed
  on recovery and rotation.
- Keep `decrypted_path` plaintext files untracked
  unless there is a deliberate exception.
- Define organization key rotation outside this repository.
