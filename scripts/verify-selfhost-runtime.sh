#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

[ ! -e .env ] || {
  echo "runtime verification requires a workspace without .env" >&2
  exit 1
}

tmp_dir="$(mktemp -d)"
project="stuffstash-runtime-${GITHUB_RUN_ID:-$$}"
compose=(docker compose --project-name "$project" -f compose.selfhost.yaml)

cleanup() {
  "${compose[@]}" down --volumes --remove-orphans >/dev/null 2>&1 || true
  rm -f .env
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

cp deploy/selfhost/dex/config.yaml "$tmp_dir/dex-config.yaml"
chmod 600 "$tmp_dir/dex-config.yaml"
sed "s|^DEX_CONFIG_PATH=.*|DEX_CONFIG_PATH=$tmp_dir/dex-config.yaml|" .env.example > .env

"${compose[@]}" up --detach

ready=0
for _ in $(seq 1 90); do
  if curl --insecure --fail --silent --show-error \
      https://stuffstash.localhost:8080/healthz >/dev/null &&
    curl --insecure --fail --silent --show-error \
      https://stuffstash.localhost:5556/dex/.well-known/openid-configuration >/dev/null; then
    ready=1
    break
  fi
  sleep 2
done

if [ "$ready" -ne 1 ]; then
  "${compose[@]}" ps >&2
  "${compose[@]}" logs --no-color >&2
  echo "self-host runtime did not become ready" >&2
  exit 1
fi

mode_owner="$("${compose[@]}" run --rm --no-deps --entrypoint sh dex-config-bootstrap -c \
  'stat -c "%a:%u:%g" /staged/config.yaml')"
[ "$mode_owner" = "600:1001:1001" ] || {
  echo "staged Dex config has unexpected mode or owner: $mode_owner" >&2
  exit 1
}

echo "self-host runtime verification passed"
