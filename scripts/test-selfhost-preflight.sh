#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

cp "$repo_root/.env.example" "$tmp_dir/.env"

SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --trial --env-file "$tmp_dir/.env" >/dev/null

if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --env-file "$tmp_dir/.env" >/dev/null 2>&1; then
  echo "strict preflight accepted unchanged example secrets" >&2
  exit 1
fi

sed -i.bak 's/stuffstash\.localhost/192.168.2.52/g' "$tmp_dir/.env"
if SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK=1 SELFHOST_PREFLIGHT_SKIP_PORT_CHECK=1 SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK=1 \
  "$repo_root/scripts/selfhost-preflight.sh" --trial --env-file "$tmp_dir/.env" >"$tmp_dir/output" 2>&1; then
  echo "preflight accepted an IP-literal OIDC hostname" >&2
  exit 1
fi
grep -q 'DNS hostname' "$tmp_dir/output"

echo "self-host preflight tests passed"
