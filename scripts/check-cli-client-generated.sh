#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$root"
(cd apps/api && GOWORK=off go run ./cmd/stuff-stash-openapi) > "$tmp/openapi.json"
diff -u packages/api-client/openapi.json "$tmp/openapi.json"
scripts/generate-cli-client.sh "$tmp/generated" "$tmp/openapi.json"
diff -ru apps/cli/internal/adapters/httpapi/generated "$tmp/generated"
