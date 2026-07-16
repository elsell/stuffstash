#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
env_file="$repo_root/.env"
trial=0

usage() {
  echo "Usage: scripts/selfhost-preflight.sh [--trial] [--env-file PATH]"
}

fail() {
  echo "Preflight failed: $*" >&2
  exit 1
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --trial) trial=1 ;;
    --env-file)
      shift
      [ "$#" -gt 0 ] || fail "--env-file needs a path"
      env_file="$1"
      ;;
    -h|--help) usage; exit 0 ;;
    *) fail "unknown option: $1" ;;
  esac
  shift
done

if [ "${SELFHOST_PREFLIGHT_SKIP_DOCKER_CHECK:-0}" != "1" ]; then
  command -v docker >/dev/null 2>&1 || fail "Docker is not installed"
  docker compose version >/dev/null 2>&1 || fail "Docker Compose is not available"
fi
[ -f "$env_file" ] || fail "copy .env.example to .env first"
[ -f "$repo_root/compose.selfhost.yaml" ] || fail "compose.selfhost.yaml is missing"

set -a
# shellcheck source=/dev/null
source "$env_file"
set +a

required=(
  STUFF_STASH_BIND_ADDRESS STUFF_STASH_SELFHOST_HOSTNAME
  STUFF_STASH_WEB_ORIGIN STUFF_STASH_API_ORIGIN STUFF_STASH_OIDC_ISSUER
  STUFF_STASH_WEB_PORT STUFF_STASH_HTTP_PORT DEX_HTTP_PORT GARAGE_S3_PORT
  STUFF_STASH_S3_ENDPOINT STUFF_STASH_S3_PUBLIC_ENDPOINT DEX_CONFIG_PATH
)
for name in "${required[@]}"; do
  [ -n "${!name:-}" ] || fail "$name is empty"
done

case "$STUFF_STASH_SELFHOST_HOSTNAME" in
  *:*) fail "use a DNS hostname, not an IP address" ;;
  *[!0-9.]* ) ;;
  *) fail "use a DNS hostname, not an IP address" ;;
esac

hostname_from_https_url() {
  local value="$1" remainder
  case "$value" in
    https://*) remainder="${value#https://}" ;;
    *) fail "$2 must use https://" ;;
  esac
  remainder="${remainder%%/*}"
  printf '%s\n' "${remainder%%:*}"
}

for name in STUFF_STASH_WEB_ORIGIN STUFF_STASH_API_ORIGIN STUFF_STASH_OIDC_ISSUER STUFF_STASH_WEB_OIDC_REDIRECT_URI STUFF_STASH_CORS_ALLOWED_ORIGINS; do
  value="${!name:-}"
  [ -n "$value" ] || fail "$name is empty"
  [ "$(hostname_from_https_url "$value" "$name")" = "$STUFF_STASH_SELFHOST_HOSTNAME" ] ||
    fail "$name must use $STUFF_STASH_SELFHOST_HOSTNAME"
done

for name in STUFF_STASH_S3_ENDPOINT STUFF_STASH_S3_PUBLIC_ENDPOINT; do
  value="${!name}"
  [ "${value%%:*}" = "$STUFF_STASH_SELFHOST_HOSTNAME" ] ||
    fail "$name must use $STUFF_STASH_SELFHOST_HOSTNAME"
done

case "$STUFF_STASH_BIND_ADDRESS" in
  127.0.0.1|::1) ;;
  *) echo "Notice: ports will be reachable beyond this machine at $STUFF_STASH_BIND_ADDRESS." >&2 ;;
esac

case "$DEX_CONFIG_PATH" in
  /*) dex_config="$DEX_CONFIG_PATH" ;;
  *) dex_config="$repo_root/${DEX_CONFIG_PATH#./}" ;;
esac
[ -f "$dex_config" ] || fail "Dex config not found: $dex_config"

example_secrets=0
for value in \
  "${POSTGRES_PASSWORD:-}" "${SPICEDB_POSTGRES_PASSWORD:-}" \
  "${STUFF_STASH_SPICEDB_PRESHARED_KEY:-}" "${STUFF_STASH_S3_ACCESS_KEY:-}" \
  "${STUFF_STASH_S3_SECRET_KEY:-}" "${STUFF_STASH_PROVIDER_CREDENTIAL_KEY:-}"; do
  case "$value" in
    change-me-*|AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=) example_secrets=1 ;;
  esac
done
if [ "$example_secrets" -eq 1 ]; then
  if [ "$trial" -eq 1 ]; then
    echo "Trial mode: example secrets are still in use. Replace them before household use." >&2
  else
    fail "example secrets remain; use --trial only for evaluation"
  fi
fi

port_in_use() {
  local port="$1"
  if command -v lsof >/dev/null 2>&1; then
    lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
  elif command -v ss >/dev/null 2>&1; then
    ss -ltn | awk -v port=":$port" '$4 ~ port "$" { found=1 } END { exit found ? 0 : 1 }'
  else
    return 1
  fi
}

if [ "${SELFHOST_PREFLIGHT_SKIP_PORT_CHECK:-0}" != "1" ]; then
  for port in "$STUFF_STASH_WEB_PORT" "$STUFF_STASH_HTTP_PORT" "$DEX_HTTP_PORT" "$GARAGE_S3_PORT"; do
    port_in_use "$port" && fail "port $port is already in use"
  done
fi

if [ "${SELFHOST_PREFLIGHT_SKIP_COMPOSE_CHECK:-0}" != "1" ]; then
  docker compose --env-file "$env_file" -f "$repo_root/compose.selfhost.yaml" config --quiet ||
    fail "Compose configuration is invalid"
fi

echo "Preflight passed."
