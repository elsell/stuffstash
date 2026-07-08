#!/usr/bin/env bash
set -euo pipefail

compose_file="compose.selfhost.yaml"
self_host_doc="docs/src/content/docs/self-hosting.md"
topology_spec="specs/platform/local-development-topology.spec.md"

fail() {
  echo "self-host happy path check failed: $*" >&2
  exit 1
}

grep -qE '^[[:space:]]+dex:' "$compose_file" ||
  fail "compose.selfhost.yaml must include bundled Dex"

grep -q 'dexidp/dex:.*@sha256:' "$compose_file" ||
  fail "self-host Dex image must be digest-pinned"

grep -q '127.0.0.1:5556/dex/.well-known/openid-configuration' "$compose_file" ||
  fail "self-host Dex healthcheck must verify OIDC discovery readiness"

awk '
  /^[[:space:]]+app:$/ { in_app=1; next }
  /^[[:space:]][a-zA-Z0-9_-]+:$/ { if (in_app) exit }
  in_app && /^[[:space:]]+dex:$/ { in_dex_dep=1; next }
  in_dex_dep && /condition:[[:space:]]+service_healthy/ { found=1 }
  END { exit found ? 0 : 1 }
' "$compose_file" ||
  fail "self-host API must wait for healthy Dex before OIDC startup"

if awk '/^[[:space:]]+spicedb:/{in_spicedb=1; next} /^[[:space:]][a-zA-Z0-9_-]+:/{if(in_spicedb) exit} in_spicedb && /serve-testing/{found=1} END{exit found ? 0 : 1}' "$compose_file"; then
  fail "self-host SpiceDB must not use serve-testing"
fi

grep -q 'spicedb-postgres' "$compose_file" ||
  fail "self-host SpiceDB must use Postgres datastore service"

grep -q './deploy/selfhost/garage/garage.toml:/etc/garage.toml:ro' "$compose_file" ||
  fail "self-host Garage must use a self-host Garage config path"

test -f deploy/selfhost/garage/garage.toml ||
  fail "self-host Garage config file must exist"

test -f scripts/configure-garage-cors.mjs ||
  fail "Garage CORS configuration script must exist"

grep -q 'exec nginx -e /dev/stderr -c /tmp/nginx.conf -g "daemon off;"' deploy/web/start-web-runtime.sh ||
  fail "web runtime must send nginx startup errors to stderr for read-only containers"

set -a
# shellcheck source=/dev/null
source .env.example
set +a

grep -q "issuer: ${STUFF_STASH_OIDC_ISSUER}" deploy/selfhost/dex/config.yaml ||
  fail "default Dex issuer must match .env.example"

grep -q -- "- ${STUFF_STASH_WEB_OIDC_REDIRECT_URI}" deploy/selfhost/dex/config.yaml ||
  fail "default Dex web redirect URI must match .env.example"

grep -q -- "- ${STUFF_STASH_WEB_ORIGIN}" deploy/selfhost/dex/config.yaml ||
  fail "default Dex allowed origin must match .env.example"

if docker compose version >/dev/null 2>&1; then
  config_output="$(mktemp)"
  created_env_file=0
  if [ ! -f .env ]; then
    cp .env.example .env
    created_env_file=1
  fi
  cleanup_compose_config() {
    rm -f "$config_output"
    if [ "$created_env_file" -eq 1 ]; then
      rm -f .env
    fi
  }
  trap cleanup_compose_config EXIT
  docker compose --env-file .env.example -f "$compose_file" config >"$config_output"
  grep -q 'condition: service_healthy' "$config_output" ||
    fail "rendered Compose config must keep Dex health dependency"
  grep -q 'spicedb-postgres' "$config_output" ||
    fail "rendered Compose config must include SpiceDB datastore Postgres"
  cleanup_compose_config
  trap - EXIT
fi

grep -q 'Docker Compose, Dex OIDC, Postgres, SpiceDB, and Garage' "$self_host_doc" ||
  fail "self-host docs must lead with durable Docker Compose plus bundled Dex"

if grep -q '^## Compose Evaluation' "$self_host_doc"; then
  fail "self-host docs must not present contributor evaluation as a public happy path"
fi

if grep -qi 'SQLite.*Compose.*self-host\|self-host.*SQLite' "$self_host_doc"; then
  fail "self-host docs must not advertise SQLite as a self-host path"
fi

grep -qi 'bundled Dex' "$topology_spec" ||
  fail "topology spec must require bundled Dex in the self-host happy path"
