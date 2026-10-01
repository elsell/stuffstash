#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
[ ! -e .env ] || { echo "Connected verification requires a workspace without .env" >&2; exit 1; }
project="stuffstash-browser-${GITHUB_RUN_ID:-$$}"
export STUFF_STASH_AUTH_MODE=oidc STUFF_STASH_AUTHZ_MODE=spicedb
export STUFF_STASH_SPICEDB_TLS_ENABLED=false STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA=true
export STUFF_STASH_SPICEDB_SCHEMA_PATH=/deploy/spicedb/schema.zed
export STUFF_STASH_CORS_ALLOWED_ORIGINS=http://localhost:5173
compose=(docker compose -p "$project" -f compose.yaml -f compose.oidc.yaml)
cleanup() { "${compose[@]}" down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT
"${compose[@]}" up -d --build postgres spicedb dex migration app
ready=false
for _ in $(seq 1 90); do
  if curl --fail --silent http://localhost:8080/healthz >/dev/null; then ready=true; break; fi
  sleep 2
done
if [ "$ready" != true ]; then echo 'Connected API did not become ready.' >&2; exit 1; fi
pnpm --dir apps/web exec playwright test --config playwright.connected.config.ts
