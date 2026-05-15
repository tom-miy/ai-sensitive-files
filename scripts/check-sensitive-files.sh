#!/usr/bin/env sh
set -eu

config="${1:-.ai-sensitive-files/sensitive-files.yaml}"
ai-sensitive-files check --config "$config"
