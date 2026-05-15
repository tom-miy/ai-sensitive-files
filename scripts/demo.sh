#!/usr/bin/env sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
tmp="${TMPDIR:-/tmp}/ai-sensitive-files-demo"
rm -rf "$tmp"
mkdir -p "$tmp/.agent-privacy-guard" "$tmp/secrets"
git -C "$tmp" init -q
GOCACHE="${GOCACHE:-/private/tmp/ai-sensitive-files-gocache}" go build -o "$tmp/ai-sensitive-files" "$root/cmd/ai-sensitive-files"

cp "$root/templates/sensitive-files.example.yaml" "$tmp/policy.yaml"
printf 'Acme Corp\n' > "$tmp/.agent-privacy-guard/entities.local.yaml"

echo "1. validate sample policy"
"$tmp/ai-sensitive-files" validate --config "$tmp/policy.yaml"

echo
echo "2. generate ignore/config files"
"$tmp/ai-sensitive-files" generate --config "$tmp/policy.yaml" --out "$tmp"
cp "$tmp/.gitignore.ai-sensitive-files" "$tmp/.gitignore"

echo
echo "3. list policy entries"
"$tmp/ai-sensitive-files" list --config "$tmp/policy.yaml"

echo "4. check detects plaintext sensitive file"
if "$tmp/ai-sensitive-files" check --config "$tmp/policy.yaml" --repo "$tmp"; then
  echo "expected check to fail but it passed" >&2
  exit 1
else
  echo "check blocked plaintext sensitive file as expected"
fi

echo
echo "5. generated Claude Code denyRead snippet"
cat "$tmp/generated/claude-code-deny-read.json"

echo
echo "6. generated decryption target mapping"
cat "$tmp/generated/decryption-targets.txt"

echo
echo "7. generated crypto plan"
sed -n '1,80p' "$tmp/generated/crypto-plan.md"
