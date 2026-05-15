# Lefthook

Use Lefthook as a pre-commit gate:

```yaml
pre-commit:
  commands:
    ai-sensitive-files:
      run: ai-sensitive-files check --config .ai-sensitive-files/sensitive-files.yaml
```

The command checks that generated files match policy.
It checks that generated `.gitignore` entries were merged into `.gitignore`.
It checks that commit-blocked paths are not tracked.
It checks that plaintext output follows `crypto.manual_edit`.
It also checks that decrypted files are not newer
than encrypted artifacts before commit.

The installer does not edit existing Lefthook config.
Merge `templates/lefthook/lefthook.example.yml` manually
after reviewing the command.
