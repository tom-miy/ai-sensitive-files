# Architecture

`ai-sensitive-files` has one source of truth:
`.ai-sensitive-files/sensitive-files.yaml` in the target repository.

The source repository keeps examples under `templates/`.
The target repository keeps live policy under `.ai-sensitive-files/`.
This avoids mixing tool-provided policy examples
with application-owned directories such as `configs/`.

The CLI is intentionally small:

- `internal/domain`: policy types and validation rules
- `internal/infra`: policy file loading
- `internal/usecase`: generated file building, checks, and list output
- `internal/interface/cli`: command parsing and output

The tool does not scan arbitrary secrets, manage keys,
or enforce AI agent runtime behavior.
It generates and checks files that other tools consume.
