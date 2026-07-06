#!/usr/bin/env bash
set -euo pipefail

project="${STUFF_STASH_DEX_VERIFY_PROJECT:-stuffstash-dex-oidc-$$}"
api_port="${STUFF_STASH_DEX_VERIFY_HTTP_PORT:-18081}"
postgres_port="${STUFF_STASH_DEX_VERIFY_POSTGRES_PORT:-15433}"
spicedb_port="${STUFF_STASH_DEX_VERIFY_SPICEDB_PORT:-15052}"
dex_port="${STUFF_STASH_DEX_VERIFY_DEX_PORT:-15556}"
base_url="http://localhost:${api_port}"
dex_base_url="http://localhost:${dex_port}/dex"
client_id="stuff-stash-local"
client_secret="stuff-stash-local-secret"
mobile_client_id="stuff-stash-mobile-local"
wrong_audience_client_id="stuff-stash-wrong-audience"
wrong_audience_client_secret="stuff-stash-wrong-audience-secret"

fail() {
  echo "dex oidc verification failed: $*" >&2
  exit 1
}

compose() {
  STUFF_STASH_HTTP_PORT="$api_port" \
    POSTGRES_PORT="$postgres_port" \
    SPICEDB_GRPC_PORT="$spicedb_port" \
    DEX_HTTP_PORT="$dex_port" \
    STUFF_STASH_AUTH_MODE=oidc \
    STUFF_STASH_AUTHZ_MODE=spicedb \
    STUFF_STASH_SPICEDB_TLS_ENABLED=false \
    STUFF_STASH_SPICEDB_BOOTSTRAP_SCHEMA=true \
    STUFF_STASH_SPICEDB_SCHEMA_PATH=/deploy/spicedb/schema.zed \
    STUFF_STASH_OIDC_ISSUER=http://dex:5556/dex \
    STUFF_STASH_OIDC_CLIENT_ID="$client_id" \
    STUFF_STASH_OIDC_CLIENT_IDS="${client_id},stuff-stash-web-local,${mobile_client_id}" \
    STUFF_STASH_OIDC_MOBILE_CLIENT_ID="${mobile_client_id}" \
    STUFF_STASH_OIDC_MOBILE_REDIRECT_URI=stuffstash://auth/callback \
    STUFF_STASH_OIDC_MOBILE_SCOPES=openid,email,profile,offline_access \
    docker compose -p "$project" -f compose.yaml -f compose.oidc.yaml "$@"
}

cleanup() {
  local status=$?
  if [ "$status" -ne 0 ]; then
    compose logs --no-color app dex spicedb migration >&2 || true
  fi
  compose down -v --remove-orphans >/dev/null 2>&1 || true
  exit "$status"
}
trap cleanup EXIT

extract_json_string() {
  local key="$1"
  python3 -c 'import json,sys; key=sys.argv[1]; print(json.load(sys.stdin).get(key, ""))' "$key"
}

request_token() {
  local email="$1"
  local requested_client_id="${2:-$client_id}"
  local requested_client_secret="${3:-$client_secret}"
  local token_response token

  token_response="$(
    curl -sS -u "${requested_client_id}:${requested_client_secret}" \
      -H "Content-Type: application/x-www-form-urlencoded" \
      -d "grant_type=password" \
      -d "username=${email}" \
      -d "password=password" \
      -d "scope=openid profile email" \
      "${dex_base_url}/token"
  )"
  token="$(printf '%s\n' "$token_response" | extract_json_string id_token)"
  [ -n "$token" ] || fail "Dex did not return an id_token for ${email}: ${token_response}"
  printf '%s\n' "$token"
}

request_public_client_token() {
  local email="$1"
  local requested_client_id="$2"
  local token_response token

  token_response="$(
    curl -sS \
      -H "Content-Type: application/x-www-form-urlencoded" \
      -d "grant_type=password" \
      -d "client_id=${requested_client_id}" \
      -d "username=${email}" \
      -d "password=password" \
      -d "scope=openid profile email offline_access" \
      "${dex_base_url}/token"
  )"
  token="$(printf '%s\n' "$token_response" | extract_json_string id_token)"
  [ -n "$token" ] || fail "Dex did not return an id_token for ${email} public client ${requested_client_id}: ${token_response}"
  printf '%s\n' "$token"
}

assert_status() {
  local expected="$1"
  local description="$2"
  local header="${3:-}"
  local status
  local curl_args=(-sS -o /dev/null -w "%{http_code}" "${base_url}/me")

  if [ -n "$header" ]; then
    curl_args+=(-H "$header")
  fi

  status="$(curl "${curl_args[@]}")"
  [ "$status" = "$expected" ] || fail "expected ${description} status ${expected}, got ${status}"
}

require_python() {
  if ! command -v python3 >/dev/null 2>&1; then
    fail "python3 is required to parse Dex token responses"
  fi
}

require_python

echo "starting Dex, Postgres, and SpiceDB for project ${project}"
compose up -d --build postgres spicedb dex

echo "waiting for Dex discovery"
dex_discovery=""
for _ in $(seq 1 60); do
  dex_discovery="$(curl -fsS "${dex_base_url}/.well-known/openid-configuration" 2>/dev/null || true)"
  if [ "$(printf '%s\n' "$dex_discovery" | extract_json_string issuer 2>/dev/null || true)" = "http://dex:5556/dex" ]; then
    break
  fi
  sleep 1
done
[ "$(printf '%s\n' "$dex_discovery" | extract_json_string issuer 2>/dev/null || true)" = "http://dex:5556/dex" ] || fail "Dex discovery did not become ready"

echo "requesting Dex ID tokens"
owner_token="$(request_token owner@example.com)"
viewer_token="$(request_token viewer@example.com)"
mobile_owner_token="$(request_public_client_token owner@example.com "$mobile_client_id")"
wrong_audience_token="$(request_token owner@example.com "$wrong_audience_client_id" "$wrong_audience_client_secret")"

echo "starting migration and API in OIDC mode"
compose up -d --build migration app

echo "checking OIDC boundary rejects bad tokens"
for _ in $(seq 1 60); do
  if curl -fsS "${base_url}/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
assert_status 401 "missing OIDC bearer token"
assert_status 401 "malformed OIDC bearer token" "Authorization: Bearer not-a-jwt"
assert_status 401 "unsigned OIDC token" "Authorization: Bearer eyJhbGciOiJub25lIn0.eyJpc3MiOiJodHRwOi8vZGV4OjU1NTYvZGV4Iiwic3ViIjoib3duZXIiLCJhdWQiOiJzdHVmZi1zdGFzaC1sb2NhbCIsImV4cCI6NDEwMjQ0NDgwMH0."
assert_status 401 "wrong-audience OIDC token" "Authorization: Bearer ${wrong_audience_token}"
assert_status 200 "mobile-audience OIDC token" "Authorization: Bearer ${mobile_owner_token}"

echo "running full API user-flow verification with Dex tokens"
STUFF_STASH_VERIFY_BASE_URL="$base_url" \
  STUFF_STASH_VERIFY_AUTH_HEADER="Authorization: Bearer ${owner_token}" \
  STUFF_STASH_VERIFY_VIEWER_AUTH_HEADER="Authorization: Bearer ${viewer_token}" \
  scripts/verify-local-api.sh

echo "Dex OIDC API verification passed"
