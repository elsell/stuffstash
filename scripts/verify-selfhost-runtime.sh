#!/usr/bin/env bash
set -euo pipefail

script_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
repo_root="${SELFHOST_RUNTIME_ROOT:-$script_root}"
cd "$repo_root"

[ ! -e .env ] || {
  echo "runtime verification requires a workspace without .env" >&2
  exit 1
}
[ ! -e .stuffstash/selfhost/dex/config.yaml ] || {
  echo "runtime verification requires a workspace without a generated Dex config" >&2
  exit 1
}

project="stuffstash-runtime-${GITHUB_RUN_ID:-$$}"
compose=(docker compose --project-name "$project" -f compose.selfhost.yaml)

cleanup() {
  "${compose[@]}" down --volumes --remove-orphans >/dev/null 2>&1 || true
  rm -f .env
  rm -f .stuffstash/selfhost/dex/config.yaml
  rmdir .stuffstash/selfhost/dex .stuffstash/selfhost .stuffstash 2>/dev/null || true
}
trap cleanup EXIT

scripts/configure-selfhost.sh
scripts/selfhost-preflight.sh
host="$(awk -F= '$1 == "STUFF_STASH_SELFHOST_HOSTNAME" { print $2 }' .env)"
[ -n "$host" ] || {
  echo "LAN setup did not write STUFF_STASH_SELFHOST_HOSTNAME" >&2
  exit 1
}

"${compose[@]}" up --detach

ready=0
for _ in $(seq 1 90); do
  if curl --insecure --fail --silent --show-error \
      "https://$host:8080/healthz" >/dev/null &&
    curl --insecure --fail --silent --show-error \
      "https://$host:5556/dex/.well-known/openid-configuration" >/dev/null; then
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
