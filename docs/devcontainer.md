# Devcontainer Runtime Secrets

When AI tooling runs inside the project workspace,
the strongest practical layout is to keep decrypted plaintext
outside that workspace.

Recommended shape:

- Store the encrypted artifact in the repository policy.
- Or store the secret reference in the repository policy.
- Write decrypted plaintext outside the project.
- Example: `/workspaces/.runtime-secrets/app.env`.
- Configure devcontainer to load that file as an environment file.
- Keep `path` as the repo path that must not appear.
- Example: `.env`.

Policy example:

```yaml
sensitive_files:
  - path: ".env"
    encrypted_path: ".env.sops.yaml"
    decrypted_path: "/workspaces/.runtime-secrets/app.env"
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

Devcontainer example:

```jsonc
{
  "runArgs": [
    "--env-file",
    "/workspaces/.runtime-secrets/app.env"
  ],
  "postStartCommand": "ai-sensitive-files check --config .ai-sensitive-files/sensitive-files.yaml"
}
```

Run the decrypt command before starting the app process.
Then let the container runtime load the file as environment variables.
This avoids requiring the app to read a secret file path inside the project.

This repository does not create the devcontainer or mount layout.
It records the intended paths.
It checks that plaintext does not drift back into the repository.
