#!/usr/bin/env bash
set -euo pipefail

schema_path="packages/api-client/src/generated/schema.d.ts"
tmp_schema="$(mktemp)"

cleanup() {
  rm -f "$tmp_schema"
}
trap cleanup EXIT

cp "$schema_path" "$tmp_schema"

"${PNPM:-pnpm}" --dir packages/api-client generate >/dev/null

if ! diff -u "$tmp_schema" "$schema_path"; then
  echo "generated API client schema is out of date; run make api-client-generate" >&2
  exit 1
fi
