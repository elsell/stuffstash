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
health=""
for attempt in $(seq 1 20); do
  health="$(request GET /healthz 2>/dev/null || true)"
  health_status="$(printf '%s\n' "$health" | head -n 1)"
  if [ "$health_status" = "200" ]; then
    break
  fi
  sleep 0.5
done
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
list_response="$(request GET "/tenants/${tenant_id}/inventories?limit=50" "$auth_header")"
list_status="$(printf '%s\n' "$list_response" | head -n 1)"
[ "$list_status" = "200" ] || fail "expected inventory list status 200, got ${list_status}"
printf '%s\n' "$list_response" | tail -n +2 | grep -q "\"id\":\"${inventory_id}\"" || fail "inventory list did not include ${inventory_id}"
printf '%s\n' "$list_response" | tail -n +2 | grep -q '"pagination"' || fail "inventory list did not include pagination metadata"

echo "creating location asset in inventory ${inventory_id}"
location_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/assets" "$auth_header" '{"kind":"location","title":"Garage"}')"
location_status="$(printf '%s\n' "$location_response" | head -n 1)"
[ "$location_status" = "201" ] || fail "expected location asset create status 201, got ${location_status}"
location_id="$(printf '%s\n' "$location_response" | tail -n +2 | extract_first_id)"
[ -n "$location_id" ] || fail "location asset create response did not include an id"

echo "creating custom field definitions"
tenant_field_response="$(request POST "/tenants/${tenant_id}/custom-field-definitions" "$auth_header" '{"key":"serial","displayName":"Serial","type":"text"}')"
tenant_field_status="$(printf '%s\n' "$tenant_field_response" | head -n 1)"
[ "$tenant_field_status" = "201" ] || fail "expected tenant custom field create status 201, got ${tenant_field_status}"
printf '%s\n' "$tenant_field_response" | tail -n +2 | grep -q '"key":"serial"' || fail "tenant custom field response did not include serial key"

inventory_field_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/custom-field-definitions" "$auth_header" '{"key":"condition","displayName":"Condition","type":"enum","enumOptions":["new","used"]}')"
inventory_field_status="$(printf '%s\n' "$inventory_field_response" | head -n 1)"
[ "$inventory_field_status" = "201" ] || fail "expected inventory custom field create status 201, got ${inventory_field_status}"
printf '%s\n' "$inventory_field_response" | tail -n +2 | grep -q '"key":"condition"' || fail "inventory custom field response did not include condition key"

echo "listing effective custom field definitions"
field_list_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/custom-field-definitions?limit=50" "$auth_header")"
field_list_status="$(printf '%s\n' "$field_list_response" | head -n 1)"
[ "$field_list_status" = "200" ] || fail "expected custom field list status 200, got ${field_list_status}"
printf '%s\n' "$field_list_response" | tail -n +2 | grep -q '"key":"serial"' || fail "custom field list did not include serial"
printf '%s\n' "$field_list_response" | tail -n +2 | grep -q '"key":"condition"' || fail "custom field list did not include condition"
printf '%s\n' "$field_list_response" | tail -n +2 | grep -q '"pagination"' || fail "custom field list did not include pagination metadata"

echo "creating item asset under location ${location_id}"
asset_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/assets" "$auth_header" "{\"kind\":\"item\",\"title\":\"Fertilizer\",\"parentAssetId\":\"${location_id}\",\"customFields\":{\"serial\":\"bag-1\",\"condition\":\"new\"}}")"
asset_status="$(printf '%s\n' "$asset_response" | head -n 1)"
[ "$asset_status" = "201" ] || fail "expected item asset create status 201, got ${asset_status}"
asset_id="$(printf '%s\n' "$asset_response" | tail -n +2 | extract_first_id)"
[ -n "$asset_id" ] || fail "item asset create response did not include an id"
printf '%s\n' "$asset_response" | tail -n +2 | grep -q '"serial":"bag-1"' || fail "item asset response did not include serial custom field"

echo "listing assets"
asset_list_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/assets?limit=50" "$auth_header")"
asset_list_status="$(printf '%s\n' "$asset_list_response" | head -n 1)"
[ "$asset_list_status" = "200" ] || fail "expected asset list status 200, got ${asset_list_status}"
printf '%s\n' "$asset_list_response" | tail -n +2 | grep -q "\"id\":\"${location_id}\"" || fail "asset list did not include ${location_id}"
printf '%s\n' "$asset_list_response" | tail -n +2 | grep -q "\"id\":\"${asset_id}\"" || fail "asset list did not include ${asset_id}"
printf '%s\n' "$asset_list_response" | tail -n +2 | grep -q '"pagination"' || fail "asset list did not include pagination metadata"

viewer_principal="${STUFF_STASH_VERIFY_VIEWER_PRINCIPAL:-user-two}"
viewer_auth_header="Authorization: Bearer dev:${viewer_principal}"

echo "granting viewer access to ${viewer_principal}"
grant_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/access-grants" "$auth_header" "{\"principalId\":\"${viewer_principal}\",\"relationship\":\"viewer\"}")"
grant_status="$(printf '%s\n' "$grant_response" | head -n 1)"
[ "$grant_status" = "201" ] || fail "expected access grant status 201, got ${grant_status}"
printf '%s\n' "$grant_response" | tail -n +2 | grep -q "\"principalId\":\"${viewer_principal}\"" || fail "grant response did not include ${viewer_principal}"
printf '%s\n' "$grant_response" | tail -n +2 | grep -q '"relationship":"viewer"' || fail "grant response did not include viewer relationship"

echo "listing access grants"
grant_list_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/access-grants?limit=50" "$auth_header")"
grant_list_status="$(printf '%s\n' "$grant_list_response" | head -n 1)"
[ "$grant_list_status" = "200" ] || fail "expected access grant list status 200, got ${grant_list_status}"
printf '%s\n' "$grant_list_response" | tail -n +2 | grep -q "\"principalId\":\"${viewer_principal}\"" || fail "grant list did not include ${viewer_principal}"
printf '%s\n' "$grant_list_response" | tail -n +2 | grep -q '"pagination"' || fail "grant list did not include pagination metadata"

echo "checking granted viewer can read but not mutate or share"
viewer_asset_list_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/assets?limit=50" "$viewer_auth_header")"
viewer_asset_list_status="$(printf '%s\n' "$viewer_asset_list_response" | head -n 1)"
[ "$viewer_asset_list_status" = "200" ] || fail "expected viewer asset list status 200, got ${viewer_asset_list_status}"

viewer_field_list_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/custom-field-definitions?limit=50" "$viewer_auth_header")"
viewer_field_list_status="$(printf '%s\n' "$viewer_field_list_response" | head -n 1)"
[ "$viewer_field_list_status" = "200" ] || fail "expected viewer custom field list status 200, got ${viewer_field_list_status}"

viewer_asset_create_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/assets" "$viewer_auth_header" '{"kind":"item","title":"Unauthorized"}')"
viewer_asset_create_status="$(printf '%s\n' "$viewer_asset_create_response" | head -n 1)"
[ "$viewer_asset_create_status" = "403" ] || fail "expected viewer asset create status 403, got ${viewer_asset_create_status}"

viewer_field_create_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/custom-field-definitions" "$viewer_auth_header" '{"key":"viewer-field","displayName":"Viewer Field","type":"text"}')"
viewer_field_create_status="$(printf '%s\n' "$viewer_field_create_response" | head -n 1)"
[ "$viewer_field_create_status" = "403" ] || fail "expected viewer custom field create status 403, got ${viewer_field_create_status}"

viewer_share_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/access-grants" "$viewer_auth_header" '{"principalId":"user-three","relationship":"viewer"}')"
viewer_share_status="$(printf '%s\n' "$viewer_share_response" | head -n 1)"
[ "$viewer_share_status" = "403" ] || fail "expected viewer share status 403, got ${viewer_share_status}"

echo "local API verification passed"
