# Integration

Recommended setup in a target repository:

1. Run `mise trust .` in this repository.
2. Approve the local tool definition.
3. Run `mise install` to install Go, SOPS, age, and Lefthook.
4. Run `mise run install-cli` to build `.bin/ai-sensitive-files`.
5. Run `bash install.sh --target /path/to/app` from this repository.
6. Review `.ai-sensitive-files/sensitive-files.yaml`.
7. Run `ai-sensitive-files validate --config .ai-sensitive-files/sensitive-files.yaml`.
8. Run `ai-sensitive-files generate --config .ai-sensitive-files/sensitive-files.yaml --out .`.
9. Review `.gitignore.ai-sensitive-files`.
10. Merge entries into `.gitignore` as needed.
11. Merge the Lefthook example if the repository uses Lefthook.

Generated files should be reviewed like any other security-relevant config.
Do not rely on a single generated ignore file as the only control.
