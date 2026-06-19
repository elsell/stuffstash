#!/usr/bin/env bash
set -euo pipefail

base_url="${STUFF_STASH_VERIFY_BASE_URL:-http://localhost:8080}"
principal="${STUFF_STASH_VERIFY_PRINCIPAL:-user-one}"
auth_header="${STUFF_STASH_VERIFY_AUTH_HEADER:-Authorization: Bearer dev:${principal}}"

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

base64_one_line() {
  base64 | tr -d '\n'
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

echo "creating shelf asset under location ${location_id}"
shelf_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/assets" "$auth_header" "{\"kind\":\"location\",\"title\":\"Shelf\",\"parentAssetId\":\"${location_id}\"}")"
shelf_status="$(printf '%s\n' "$shelf_response" | head -n 1)"
[ "$shelf_status" = "201" ] || fail "expected shelf asset create status 201, got ${shelf_status}"
shelf_id="$(printf '%s\n' "$shelf_response" | tail -n +2 | extract_first_id)"
[ -n "$shelf_id" ] || fail "shelf asset create response did not include an id"

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

echo "moving item asset to shelf ${shelf_id}"
asset_update_response="$(request PATCH "/tenants/${tenant_id}/inventories/${inventory_id}/assets/${asset_id}" "$auth_header" "{\"title\":\"Fertilizer Bag\",\"parentAssetId\":\"${shelf_id}\",\"customFields\":{\"serial\":\"bag-1\",\"condition\":\"used\"}}")"
asset_update_status="$(printf '%s\n' "$asset_update_response" | head -n 1)"
[ "$asset_update_status" = "200" ] || fail "expected item asset update status 200, got ${asset_update_status}"
printf '%s\n' "$asset_update_response" | tail -n +2 | grep -q "\"parentAssetId\":\"${shelf_id}\"" || fail "item asset update response did not include shelf parent"
printf '%s\n' "$asset_update_response" | tail -n +2 | grep -q '"title":"Fertilizer Bag"' || fail "item asset update response did not include updated title"
printf '%s\n' "$asset_update_response" | tail -n +2 | grep -q '"condition":"used"' || fail "item asset update response did not include updated condition"

echo "uploading attachment for item asset ${asset_id}"
attachment_content_base64="iVBORw0KGgo="
attachment_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/assets/${asset_id}/attachments" "$auth_header" "{\"fileName\":\"receipt.png\",\"contentType\":\"image/png\",\"contentBase64\":\"${attachment_content_base64}\"}")"
attachment_status="$(printf '%s\n' "$attachment_response" | head -n 1)"
[ "$attachment_status" = "201" ] || fail "expected attachment create status 201, got ${attachment_status}"
attachment_id="$(printf '%s\n' "$attachment_response" | tail -n +2 | extract_first_id)"
[ -n "$attachment_id" ] || fail "attachment create response did not include an id"
printf '%s\n' "$attachment_response" | tail -n +2 | grep -q '"fileName":"receipt.png"' || fail "attachment response did not include file name"
printf '%s\n' "$attachment_response" | tail -n +2 | grep -q '"contentType":"image/png"' || fail "attachment response did not include content type"

echo "listing attachments"
attachment_list_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/assets/${asset_id}/attachments?limit=50" "$auth_header")"
attachment_list_status="$(printf '%s\n' "$attachment_list_response" | head -n 1)"
[ "$attachment_list_status" = "200" ] || fail "expected attachment list status 200, got ${attachment_list_status}"
printf '%s\n' "$attachment_list_response" | tail -n +2 | grep -q "\"id\":\"${attachment_id}\"" || fail "attachment list did not include ${attachment_id}"
printf '%s\n' "$attachment_list_response" | tail -n +2 | grep -q '"pagination"' || fail "attachment list did not include pagination metadata"

echo "downloading attachment"
download_file="$(mktemp)"
download_status="$(curl -sS -o "$download_file" -w "%{http_code}" -H "$auth_header" "${base_url}/tenants/${tenant_id}/inventories/${inventory_id}/assets/${asset_id}/attachments/${attachment_id}/content")"
[ "$download_status" = "200" ] || fail "expected attachment download status 200, got ${download_status}"
download_base64="$(base64_one_line < "$download_file")"
[ "$download_base64" = "$attachment_content_base64" ] || fail "attachment download content did not match upload"
rm -f "$download_file"

echo "listing inventory audit records"
audit_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/audit-records?limit=50" "$auth_header")"
audit_status="$(printf '%s\n' "$audit_response" | head -n 1)"
[ "$audit_status" = "200" ] || fail "expected inventory audit list status 200, got ${audit_status}"
printf '%s\n' "$audit_response" | tail -n +2 | grep -q '"action":"asset.created"' || fail "audit list did not include asset.created"
printf '%s\n' "$audit_response" | tail -n +2 | grep -q '"action":"asset.updated"' || fail "audit list did not include asset.updated"
printf '%s\n' "$audit_response" | tail -n +2 | grep -q '"action":"asset.moved"' || fail "audit list did not include asset.moved"
printf '%s\n' "$audit_response" | tail -n +2 | grep -q '"action":"attachment.created"' || fail "audit list did not include attachment.created"
printf '%s\n' "$audit_response" | tail -n +2 | grep -q '"pagination"' || fail "audit list did not include pagination metadata"

echo "listing assets"
asset_list_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/assets?limit=50" "$auth_header")"
asset_list_status="$(printf '%s\n' "$asset_list_response" | head -n 1)"
[ "$asset_list_status" = "200" ] || fail "expected asset list status 200, got ${asset_list_status}"
printf '%s\n' "$asset_list_response" | tail -n +2 | grep -q "\"id\":\"${location_id}\"" || fail "asset list did not include ${location_id}"
printf '%s\n' "$asset_list_response" | tail -n +2 | grep -q "\"id\":\"${shelf_id}\"" || fail "asset list did not include ${shelf_id}"
printf '%s\n' "$asset_list_response" | tail -n +2 | grep -q "\"id\":\"${asset_id}\"" || fail "asset list did not include ${asset_id}"
printf '%s\n' "$asset_list_response" | tail -n +2 | grep -q '"pagination"' || fail "asset list did not include pagination metadata"

echo "searching assets"
custom_field_search_response="$(request GET "/tenants/${tenant_id}/search/assets?q=bag-1&mode=exact&limit=50" "$auth_header")"
custom_field_search_status="$(printf '%s\n' "$custom_field_search_response" | head -n 1)"
[ "$custom_field_search_status" = "200" ] || fail "expected custom field search status 200, got ${custom_field_search_status}"
printf '%s\n' "$custom_field_search_response" | tail -n +2 | grep -q "\"id\":\"${asset_id}\"" || fail "custom field search did not include ${asset_id}"
printf '%s\n' "$custom_field_search_response" | tail -n +2 | grep -q '"field":"custom_field"' || fail "custom field search did not include custom field match"
printf '%s\n' "$custom_field_search_response" | tail -n +2 | grep -q '"pagination"' || fail "custom field search did not include pagination metadata"

attachment_search_response="$(request GET "/tenants/${tenant_id}/search/assets?q=receipt&limit=50" "$auth_header")"
attachment_search_status="$(printf '%s\n' "$attachment_search_response" | head -n 1)"
[ "$attachment_search_status" = "200" ] || fail "expected attachment search status 200, got ${attachment_search_status}"
printf '%s\n' "$attachment_search_response" | tail -n +2 | grep -q "\"id\":\"${asset_id}\"" || fail "attachment search did not include ${asset_id}"
printf '%s\n' "$attachment_search_response" | tail -n +2 | grep -q '"field":"attachment_file_name"' || fail "attachment search did not include attachment file name match"

viewer_auth_header="${STUFF_STASH_VERIFY_VIEWER_AUTH_HEADER:-}"
viewer_principal="${STUFF_STASH_VERIFY_VIEWER_PRINCIPAL:-}"
if [ -z "$viewer_auth_header" ]; then
  viewer_principal="${viewer_principal:-user-two}"
  viewer_auth_header="Authorization: Bearer dev:${viewer_principal}"
elif [ -z "$viewer_principal" ]; then
  viewer_me="$(request GET /me "$viewer_auth_header")"
  viewer_me_status="$(printf '%s\n' "$viewer_me" | head -n 1)"
  [ "$viewer_me_status" = "200" ] || fail "expected viewer /me status 200, got ${viewer_me_status}"
  viewer_principal="$(printf '%s\n' "$viewer_me" | tail -n +2 | extract_first_id)"
  [ -n "$viewer_principal" ] || fail "viewer /me response did not include an id"
fi

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

viewer_search_response="$(request GET "/tenants/${tenant_id}/search/assets?q=receipt&limit=50" "$viewer_auth_header")"
viewer_search_status="$(printf '%s\n' "$viewer_search_response" | head -n 1)"
[ "$viewer_search_status" = "200" ] || fail "expected viewer search status 200, got ${viewer_search_status}"
printf '%s\n' "$viewer_search_response" | tail -n +2 | grep -q "\"id\":\"${asset_id}\"" || fail "viewer search did not include ${asset_id}"

viewer_field_list_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/custom-field-definitions?limit=50" "$viewer_auth_header")"
viewer_field_list_status="$(printf '%s\n' "$viewer_field_list_response" | head -n 1)"
[ "$viewer_field_list_status" = "200" ] || fail "expected viewer custom field list status 200, got ${viewer_field_list_status}"

viewer_audit_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/audit-records?limit=50" "$viewer_auth_header")"
viewer_audit_status="$(printf '%s\n' "$viewer_audit_response" | head -n 1)"
[ "$viewer_audit_status" = "200" ] || fail "expected viewer inventory audit list status 200, got ${viewer_audit_status}"
printf '%s\n' "$viewer_audit_response" | tail -n +2 | grep -q '"action":"inventory_access.granted"' || fail "viewer audit list did not include inventory access grant"

viewer_attachment_list_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/assets/${asset_id}/attachments?limit=50" "$viewer_auth_header")"
viewer_attachment_list_status="$(printf '%s\n' "$viewer_attachment_list_response" | head -n 1)"
[ "$viewer_attachment_list_status" = "200" ] || fail "expected viewer attachment list status 200, got ${viewer_attachment_list_status}"

viewer_download_file="$(mktemp)"
viewer_download_status="$(curl -sS -o "$viewer_download_file" -w "%{http_code}" -H "$viewer_auth_header" "${base_url}/tenants/${tenant_id}/inventories/${inventory_id}/assets/${asset_id}/attachments/${attachment_id}/content")"
[ "$viewer_download_status" = "200" ] || fail "expected viewer attachment download status 200, got ${viewer_download_status}"
viewer_download_base64="$(base64_one_line < "$viewer_download_file")"
[ "$viewer_download_base64" = "$attachment_content_base64" ] || fail "viewer attachment download content did not match upload"
rm -f "$viewer_download_file"

viewer_tenant_audit_response="$(request GET "/tenants/${tenant_id}/audit-records?limit=50" "$viewer_auth_header")"
viewer_tenant_audit_status="$(printf '%s\n' "$viewer_tenant_audit_response" | head -n 1)"
[ "$viewer_tenant_audit_status" = "403" ] || fail "expected viewer tenant audit list status 403, got ${viewer_tenant_audit_status}"

viewer_asset_create_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/assets" "$viewer_auth_header" '{"kind":"item","title":"Unauthorized"}')"
viewer_asset_create_status="$(printf '%s\n' "$viewer_asset_create_response" | head -n 1)"
[ "$viewer_asset_create_status" = "403" ] || fail "expected viewer asset create status 403, got ${viewer_asset_create_status}"

viewer_asset_update_response="$(request PATCH "/tenants/${tenant_id}/inventories/${inventory_id}/assets/${asset_id}" "$viewer_auth_header" '{"title":"Unauthorized"}')"
viewer_asset_update_status="$(printf '%s\n' "$viewer_asset_update_response" | head -n 1)"
[ "$viewer_asset_update_status" = "403" ] || fail "expected viewer asset update status 403, got ${viewer_asset_update_status}"

viewer_attachment_upload_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/assets/${asset_id}/attachments" "$viewer_auth_header" "{\"fileName\":\"blocked.png\",\"contentType\":\"image/png\",\"contentBase64\":\"${attachment_content_base64}\"}")"
viewer_attachment_upload_status="$(printf '%s\n' "$viewer_attachment_upload_response" | head -n 1)"
[ "$viewer_attachment_upload_status" = "403" ] || fail "expected viewer attachment upload status 403, got ${viewer_attachment_upload_status}"

viewer_field_create_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/custom-field-definitions" "$viewer_auth_header" '{"key":"viewer-field","displayName":"Viewer Field","type":"text"}')"
viewer_field_create_status="$(printf '%s\n' "$viewer_field_create_response" | head -n 1)"
[ "$viewer_field_create_status" = "403" ] || fail "expected viewer custom field create status 403, got ${viewer_field_create_status}"

viewer_share_response="$(request POST "/tenants/${tenant_id}/inventories/${inventory_id}/access-grants" "$viewer_auth_header" '{"principalId":"user-three","relationship":"viewer"}')"
viewer_share_status="$(printf '%s\n' "$viewer_share_response" | head -n 1)"
[ "$viewer_share_status" = "403" ] || fail "expected viewer share status 403, got ${viewer_share_status}"

viewer_revoke_response="$(request DELETE "/tenants/${tenant_id}/inventories/${inventory_id}/access-grants/${viewer_principal}/viewer" "$viewer_auth_header")"
viewer_revoke_status="$(printf '%s\n' "$viewer_revoke_response" | head -n 1)"
[ "$viewer_revoke_status" = "403" ] || fail "expected viewer revoke status 403, got ${viewer_revoke_status}"

echo "revoking viewer access from ${viewer_principal}"
revoke_response="$(request DELETE "/tenants/${tenant_id}/inventories/${inventory_id}/access-grants/${viewer_principal}/viewer" "$auth_header")"
revoke_status="$(printf '%s\n' "$revoke_response" | head -n 1)"
[ "$revoke_status" = "204" ] || fail "expected revoke status 204, got ${revoke_status}"

revoke_again_response="$(request DELETE "/tenants/${tenant_id}/inventories/${inventory_id}/access-grants/${viewer_principal}/viewer" "$auth_header")"
revoke_again_status="$(printf '%s\n' "$revoke_again_response" | head -n 1)"
[ "$revoke_again_status" = "204" ] || fail "expected idempotent revoke status 204, got ${revoke_again_status}"

revoked_viewer_asset_list_response="$(request GET "/tenants/${tenant_id}/inventories/${inventory_id}/assets?limit=50" "$viewer_auth_header")"
revoked_viewer_asset_list_status="$(printf '%s\n' "$revoked_viewer_asset_list_response" | head -n 1)"
[ "$revoked_viewer_asset_list_status" = "403" ] || fail "expected revoked viewer asset list status 403, got ${revoked_viewer_asset_list_status}"

echo "local API verification passed"
