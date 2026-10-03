#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
output="${1:-$root/apps/cli/internal/adapters/httpapi/generated}"
schema="${2:-$root/packages/api-client/openapi.json}"
mkdir -p "$output"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
GOWORK=off go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0 \
  -config "$root/apps/cli/oapi-codegen.yaml" -o "$tmp" "$schema"
# Put the generator's own marker first for repository structural tooling.
python3 - "$tmp" "$output/client.gen.go" <<'PY'
from pathlib import Path
import sys
text = Path(sys.argv[1]).read_text()
lines = text.splitlines()
marker = next(line for line in lines if line.startswith('// Code generated '))
Path(sys.argv[2]).write_text(marker + '\n' + text)
PY
