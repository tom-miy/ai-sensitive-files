#!/usr/bin/env sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
target="."

while [ "$#" -gt 0 ]; do
  case "$1" in
    --target)
      target="$2"
      shift 2
      ;;
    -h|--help)
      echo "usage: bash install.sh [--target /path/to/app]"
      exit 0
      ;;
    *)
      echo "unknown option: $1" >&2
      exit 2
      ;;
  esac
done

mkdir -p "$target/.ai-sensitive-files/generated"

policy="$target/.ai-sensitive-files/sensitive-files.yaml"
if [ -e "$policy" ]; then
  echo "kept existing $policy"
else
  cp "$root/templates/sensitive-files.example.yaml" "$policy"
echo "created $policy"
fi

echo
echo "Toolchain expectation:"
echo "  Run mise trust . && mise install in the ai-sensitive-files repo to install Go, SOPS, age, and Lefthook."
echo "  Run mise run install-cli in the ai-sensitive-files repo to build .bin/ai-sensitive-files."
echo "  This installer does not create age identities or initialize SOPS."
echo
echo "Add these generated entries to .gitignore after reviewing them:"
echo
echo "  # ai-sensitive-files generated gitignore entries"
echo "  # Run: ai-sensitive-files generate --config .ai-sensitive-files/sensitive-files.yaml --out ."
echo "  .env"
echo "  .agent-privacy-guard/entities.local.yaml"
echo "  secrets/**"

if [ -e "$target/lefthook.yml" ] || [ -e "$target/lefthook.yaml" ]; then
  echo
  echo "Lefthook appears to be installed. Merge this example manually:"
  echo "  templates/lefthook/lefthook.example.yml"
fi

echo
echo "SOPS/age is installed through mise, but encryption is not initialized automatically. Review:"
echo "  templates/sops/.sops.yaml.example"
echo "  For 1Password or Bitwarden policies, install and authenticate op or bw separately."
echo
echo "Next commands:"
echo "  cd $target"
echo "  ai-sensitive-files validate --config .ai-sensitive-files/sensitive-files.yaml"
echo "  ai-sensitive-files generate --config .ai-sensitive-files/sensitive-files.yaml --out ."
