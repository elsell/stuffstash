#!/usr/bin/env bash
set -euo pipefail

base_url="${STUFF_STASH_VERIFY_BASE_URL:-http://localhost:8080}"
principal="${STUFF_STASH_VERIFY_PRINCIPAL:-user-one}"
auth_header="Authorization: Bearer dev:${principal}"

fail() {
  echo "verification failed: $*" >&2
  exit 1
}

request() {
  local method="$1"
  local path="$2"
  local auth="${3:-}"
  local body="${4:-}"
  local response_file status
  local curl_args

  response_file="$(mktemp)"
  curl_args=(-sS -o "$response_file" -w "%{http_code}" -X "$method" "${base_url}${path}")
  if [ -n "$auth" ]; then
    curl_args+=(-H "$auth")
  fi
  if [ -n "$body" ]; then
    curl_args+=(-H "Content-Type: application/json" -d "$body")
  fi
  status="$(curl "${curl_args[@]}")"

  printf '%s\n' "$status"
  cat "$response_file"
  rm -f "$response_file"
}

extract_first_id() {
  sed -n 's/.*"id":"\([^"]*\)".*/\1/p' | head -n 1
}

echo "checking health at ${base_url}"
health="$(request GET /healthz)"
health_status="$(printf '%s\n' "$health" | head -n 1)"
[ "$health_status" = "200" ] || fail "expected health status 200, got ${health_status}"

echo "checking unauthenticated rejection"
unauthenticated="$(request GET /me)"
unauthenticated_status="$(printf '%s\n' "$unauthenticated" | head -n 1)"
[ "$unauthenticated_status" = "401" ] || fail "expected /me without auth status 401, got ${unauthenticated_status}"

echo "checking authenticated identity"
me="$(request GET /me "$auth_header")"
me_status="$(printf '%s\n' "$me" | head -n 1)"
[ "$me_status" = "200" ] || fail "expected /me status 200, got ${me_status}"

echo "creating tenant"
tenant_response="$(request POST /tenants "$auth_header" '{"name":"Home"}')"
tenant_status="$(printf '%s\n' "$tenant_response" | head -n 1)"
[ "$tenant_status" = "201" ] || fail "expected tenant create status 201, got ${tenant_status}"
tenant_id="$(printf '%s\n' "$tenant_response" | tail -n +2 | extract_first_id)"
[ -n "$tenant_id" ] || fail "tenant create response did not include an id"

echo "creating inventory in tenant ${tenant_id}"
inventory_response="$(request POST "/tenants/${tenant_id}/inventories" "$auth_header" '{"name":"Tools"}')"
inventory_status="$(printf '%s\n' "$inventory_response" | head -n 1)"
[ "$inventory_status" = "201" ] || fail "expected inventory create status 201, got ${inventory_status}"
inventory_id="$(printf '%s\n' "$inventory_response" | tail -n +2 | extract_first_id)"
[ -n "$inventory_id" ] || fail "inventory create response did not include an id"

echo "listing inventories"
list_response="$(request GET "/tenants/${tenant_id}/inventories" "$auth_header")"
list_status="$(printf '%s\n' "$list_response" | head -n 1)"
[ "$list_status" = "200" ] || fail "expected inventory list status 200, got ${list_status}"
printf '%s\n' "$list_response" | tail -n +2 | grep -q "\"id\":\"${inventory_id}\"" || fail "inventory list did not include ${inventory_id}"

echo "local API verification passed"
