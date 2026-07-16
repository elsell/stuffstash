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

while IFS= read -r line || [ -n "$line" ]; do
  line="${line%$'\r'}"
  case "$line" in
    ''|'#'*) continue ;;
  esac
  name="${line%%=*}"
  [ "$name" != "$line" ] || fail "invalid line in $env_file: $line"
  [[ "$name" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || fail "invalid variable name in $env_file: $name"
  value="${line#*=}"
  if [[ "$value" =~ ^\"(.*)\"$ ]] || [[ "$value" =~ ^\'(.*)\'$ ]]; then
    value="${BASH_REMATCH[1]}"
  fi
  printf -v "$name" '%s' "$value"
done < "$env_file"

required=(
  STUFF_STASH_BIND_ADDRESS STUFF_STASH_SELFHOST_HOSTNAME
  STUFF_STASH_WEB_ORIGIN STUFF_STASH_API_ORIGIN STUFF_STASH_OIDC_ISSUER
  STUFF_STASH_WEB_PORT STUFF_STASH_HTTP_PORT DEX_HTTP_PORT GARAGE_S3_PORT
  STUFF_STASH_S3_ENDPOINT STUFF_STASH_S3_PUBLIC_ENDPOINT DEX_CONFIG_PATH
)
for name in "${required[@]}"; do
  [ -n "${!name:-}" ] || fail "$name is empty"
done

valid_dns_hostname() {
  local hostname="$1" label
  local labels=()
  [ "${#hostname}" -le 253 ] || return 1
  [[ "$hostname" == *.* ]] || return 1
  [[ "$hostname" != .* && "$hostname" != *. ]] || return 1
  [[ ! "$hostname" =~ ^[0-9.]+$ ]] || return 1
  IFS='.' read -r -a labels <<< "$hostname"
  for label in "${labels[@]}"; do
    [ "${#label}" -ge 1 ] && [ "${#label}" -le 63 ] || return 1
    [[ "$label" =~ ^[A-Za-z0-9]$|^[A-Za-z0-9][A-Za-z0-9-]*[A-Za-z0-9]$ ]] || return 1
  done
}

valid_dns_hostname "$STUFF_STASH_SELFHOST_HOSTNAME" ||
  fail "use a valid DNS hostname, not an IP address or URL"

expect_value() {
  local name="$1" expected="$2"
  [ "${!name:-}" = "$expected" ] || fail "$name must be $expected"
}

expected_web_origin="https://$STUFF_STASH_SELFHOST_HOSTNAME:$STUFF_STASH_WEB_PORT"
expect_value STUFF_STASH_WEB_ORIGIN "$expected_web_origin"
expect_value STUFF_STASH_API_ORIGIN "https://$STUFF_STASH_SELFHOST_HOSTNAME:$STUFF_STASH_HTTP_PORT"
expect_value STUFF_STASH_OIDC_ISSUER "https://$STUFF_STASH_SELFHOST_HOSTNAME:$DEX_HTTP_PORT/dex"
expect_value STUFF_STASH_WEB_OIDC_REDIRECT_URI "$expected_web_origin/callback"
expect_value STUFF_STASH_CORS_ALLOWED_ORIGINS "$expected_web_origin"
expect_value STUFF_STASH_S3_ENDPOINT "$STUFF_STASH_SELFHOST_HOSTNAME:$GARAGE_S3_PORT"
expect_value STUFF_STASH_S3_PUBLIC_ENDPOINT "$STUFF_STASH_SELFHOST_HOSTNAME:$GARAGE_S3_PORT"

case "$STUFF_STASH_BIND_ADDRESS" in
  127.0.0.1|::1) ;;
  *)
    [ "$trial" -ne 1 ] || fail "trial mode requires a loopback bind address"
    echo "Notice: ports will be reachable beyond this machine at $STUFF_STASH_BIND_ADDRESS." >&2
    ;;
esac

case "$DEX_CONFIG_PATH" in
  /*) dex_config="$DEX_CONFIG_PATH" ;;
  *) dex_config="$repo_root/${DEX_CONFIG_PATH#./}" ;;
esac
[ -f "$dex_config" ] || fail "Dex config not found: $dex_config"

if [ "$trial" -ne 1 ]; then
  if cmp -s "$dex_config" "$repo_root/deploy/selfhost/dex/config.yaml"; then
    fail "create a private Dex config and replace the bundled identities"
  fi
  if mode="$(stat -c '%a' "$dex_config" 2>/dev/null)"; then
    :
  else
    mode="$(stat -f '%Lp' "$dex_config" 2>/dev/null || true)"
  fi
  [ "$mode" = "600" ] || fail "private Dex config must have mode 600"
  if grep -Fq \
    -e 'owner@example.com' \
    -e 'viewer@example.com' \
    -e 'stuff-stash-local-secret' \
    -e '$2a$10$2b2cU8CPhOTaGrs1HRQuAueS7JTT5ZHsHSzYiFPm1leZck7Mc8T4W' \
    -e '11111111-1111-1111-1111-111111111111' \
    -e '22222222-2222-2222-2222-222222222222' \
    "$dex_config"; then
    fail "private Dex config still contains example identities or client secrets"
  fi
  dex_has_line() {
    awk -v expected="$1" '
      { line=$0; sub(/^[[:space:]]+/, "", line) }
      line == expected { found=1 }
      END { exit found ? 0 : 1 }
    ' "$dex_config"
  }
  dex_has_line "issuer: $STUFF_STASH_OIDC_ISSUER" ||
    fail "private Dex issuer must match STUFF_STASH_OIDC_ISSUER"
  dex_has_line "- $STUFF_STASH_WEB_ORIGIN" ||
    fail "private Dex allowed origin must match STUFF_STASH_WEB_ORIGIN"
  awk -v target_id="$STUFF_STASH_WEB_OIDC_CLIENT_ID" -v target_redirect="$STUFF_STASH_WEB_OIDC_REDIRECT_URI" '
    {
      line=$0
      sub(/^[[:space:]]+/, "", line)
    }
    line ~ /^- id: / {
      in_target=(line == "- id: " target_id)
      if (in_target) seen=1
      next
    }
    in_target && line == "public: true" { public_client=1 }
    in_target && line == "- " target_redirect { redirect=1 }
    END { exit seen && public_client && redirect ? 0 : 1 }
  ' "$dex_config" ||
    fail "private Dex web client must be public and use STUFF_STASH_WEB_OIDC_REDIRECT_URI"
fi

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
