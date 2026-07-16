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

grep -qE '^[[:space:]]+caddy:' "$compose_file" ||
  fail "compose.selfhost.yaml must include Caddy HTTPS edge"

grep -Fq '${STUFF_STASH_BIND_ADDRESS:?set STUFF_STASH_BIND_ADDRESS in .env}:' "$compose_file" ||
  fail "self-host published ports must require an explicit bind address"

grep -q '^STUFF_STASH_BIND_ADDRESS=127\.0\.0\.1$' .env.example ||
  fail "self-host ports must bind to loopback by default"

grep -q 'CADDY_IMAGE=caddy:.*@sha256:' .env.example ||
  fail "self-host Caddy image must be digest-pinned"

test -f deploy/selfhost/caddy/Caddyfile ||
  fail "self-host Caddyfile must exist"

grep -q 'tls internal' deploy/selfhost/caddy/Caddyfile ||
  fail "self-host Caddyfile must use Caddy local HTTPS"

grep -q 'aliases:' "$compose_file" && grep -q 'STUFF_STASH_SELFHOST_HOSTNAME' "$compose_file" ||
  fail "Caddy must expose the self-host hostname as a Docker network alias"

grep -q '127.0.0.1:5556/dex/.well-known/openid-configuration' "$compose_file" ||
  fail "self-host Dex healthcheck must verify OIDC discovery readiness"

grep -qE '^[[:space:]]+dex-config-bootstrap:' "$compose_file" ||
  fail "Compose must stage private Dex configuration"

grep -q 'selfhost-dex-config:/etc/dex:ro' "$compose_file" ||
  fail "Dex must read staged configuration from a named volume"

if grep -q 'DEX_CONFIG_PATH.*:/etc/dex/config.yaml:ro' "$compose_file"; then
  fail "Dex must not directly mount a mode-0600 host configuration"
fi

awk '
  /^[[:space:]]+app:$/ { in_app=1; next }
  /^[[:space:]][a-zA-Z0-9_-]+:$/ { if (in_app) exit }
  in_app && /^[[:space:]]+caddy-bootstrap:$/ { in_caddy_bootstrap_dep=1; next }
  in_caddy_bootstrap_dep && /condition:[[:space:]]+service_completed_successfully/ { found=1 }
  END { exit found ? 0 : 1 }
' "$compose_file" ||
  fail "self-host API must wait for Caddy HTTPS OIDC bootstrap"

grep -q 'well-known/openid-configuration' "$compose_file" ||
  fail "Caddy bootstrap must verify HTTPS OIDC discovery"

if awk '/^[[:space:]]+spicedb:/{in_spicedb=1; next} /^[[:space:]][a-zA-Z0-9_-]+:/{if(in_spicedb) exit} in_spicedb && /serve-testing/{found=1} END{exit found ? 0 : 1}' "$compose_file"; then
  fail "self-host SpiceDB must not use serve-testing"
fi

grep -q 'spicedb-postgres' "$compose_file" ||
  fail "self-host SpiceDB must use Postgres datastore service"

grep -q '^STUFF_STASH_API_IMAGE=ghcr.io/.*/stuffstash@sha256:' .env.example ||
  fail "self-host API image must default to a published digest-pinned image"

grep -q '^STUFF_STASH_WEB_IMAGE=ghcr.io/.*/stuffstash-web@sha256:' .env.example ||
  fail "self-host web image must default to a published digest-pinned image"

if awk '
  /^[[:space:]]+migration:$/ { in_target=1; target="migration"; next }
  /^[[:space:]]+app:$/ { in_target=1; target="app"; next }
  /^[[:space:]]+web:$/ { in_target=1; target="web"; next }
  /^[[:space:]][a-zA-Z0-9_-]+:$/ { if (in_target) in_target=0 }
  in_target && /^[[:space:]]+build:/ { found=target }
  END { exit found ? 0 : 1 }
' "$compose_file"; then
  fail "self-host API and web services must not build from source by default"
fi

awk '
  /^[[:space:]]+postgres:$/ { in_pg=1; service="postgres"; found=0; next }
  /^[[:space:]]+spicedb-postgres:$/ { in_pg=1; service="spicedb-postgres"; found=0; next }
  /^[[:space:]][a-zA-Z0-9_-]+:$/ {
    if (in_pg && !found) {
      printf "%s missing PGDATA under mounted volume\n", service > "/dev/stderr"
      exit 1
    }
    in_pg=0
  }
  in_pg && /PGDATA:[[:space:]]+\/var\/lib\/postgresql\/data\// { found=1 }
  END {
    if (in_pg && !found) {
      printf "%s missing PGDATA under mounted volume\n", service > "/dev/stderr"
      exit 1
    }
  }
' "$compose_file" ||
  fail "self-host Postgres services must set PGDATA under their mounted data volumes"

grep -q './deploy/selfhost/garage/garage.toml:/etc/garage.toml:ro' "$compose_file" ||
  fail "self-host Garage must use a self-host Garage config path"

test -f deploy/selfhost/garage/garage.toml ||
  fail "self-host Garage config file must exist"

test -f scripts/configure-garage-cors.mjs ||
  fail "Garage CORS configuration script must exist"

grep -q 'exec nginx -e /dev/stderr -c /tmp/nginx.conf -g "daemon off;"' deploy/web/start-web-runtime.sh ||
  fail "web runtime must send nginx startup errors to stderr for read-only containers"

grep -q 'SSL_CERT_FILE: /caddy-data/stuffstash-ca-certificates.crt' "$compose_file" ||
  fail "API must trust the Caddy local CA bundle for HTTPS OIDC discovery"

set -a
# shellcheck source=/dev/null
source .env.example
set +a

case "$STUFF_STASH_WEB_ORIGIN:$STUFF_STASH_API_ORIGIN:$STUFF_STASH_OIDC_ISSUER:$STUFF_STASH_S3_SECURE" in
  https://*:https://*:https://*:true) ;;
  *) fail "default self-host browser origins and OIDC issuer must use HTTPS" ;;
esac

case "$STUFF_STASH_S3_ENDPOINT:$STUFF_STASH_S3_PUBLIC_ENDPOINT" in
  "$STUFF_STASH_SELFHOST_HOSTNAME":*:"$STUFF_STASH_SELFHOST_HOSTNAME":*) ;;
  *) fail "default self-host S3 endpoints must use the Caddy HTTPS hostname" ;;
esac

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
  grep -q 'caddy:2.10.2@sha256:' "$config_output" ||
    fail "rendered Compose config must include pinned Caddy"
  grep -q 'PGDATA: /var/lib/postgresql/data/' "$config_output" ||
    fail "rendered Compose config must persist Postgres PGDATA under the mounted data volume"
  cleanup_compose_config
  trap - EXIT
fi

grep -q 'Docker Compose, Caddy HTTPS, Dex OIDC, Postgres, SpiceDB, and Garage' "$self_host_doc" ||
  fail "self-host docs must lead with durable Docker Compose, HTTPS, and bundled Dex"

if grep -q 'docker compose -f compose.selfhost.yaml up --build' "$self_host_doc"; then
  fail "self-host docs must not tell operators to build source images by default"
fi

if grep -q '^## Compose Evaluation' "$self_host_doc"; then
  fail "self-host docs must not present contributor evaluation as a public happy path"
fi

if grep -qi 'SQLite.*Compose.*self-host\|self-host.*SQLite' "$self_host_doc"; then
  fail "self-host docs must not advertise SQLite as a self-host path"
fi

grep -qi 'bundled Dex' "$topology_spec" ||
  fail "topology spec must require bundled Dex in the self-host happy path"

grep -qi 'Garage direct browser upload must work' specs/platform/self-hosting.spec.md ||
  fail "self-hosting spec must require Garage direct browser upload"

test -x scripts/selfhost-preflight.sh ||
  fail "operator preflight script must exist and be executable"

test -x scripts/build-selfhost-release.sh ||
  fail "self-host release bundle builder must exist and be executable"

test -x scripts/verify-selfhost-runtime.sh ||
  fail "self-host runtime verification script must exist and be executable"

grep -q 'scripts/verify-selfhost-runtime.sh' .github/workflows/ci.yml ||
  fail "CI must run the self-host topology"

grep -q 'scripts/build-selfhost-release.sh' .github/workflows/release.yml &&
  grep -q 'stuffstash-selfhost.tar.gz.sha256' .github/workflows/release.yml ||
  fail "releases must attach a checksummed self-host bundle"

scripts/test-selfhost-preflight.sh
scripts/test-selfhost-release-bundle.sh
