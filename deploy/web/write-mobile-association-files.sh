#!/usr/bin/env sh
set -eu

output_directory="${1:-/tmp}"
ios_app_id="${STUFF_STASH_MOBILE_IOS_APP_ID:-}"
android_package="${STUFF_STASH_MOBILE_ANDROID_PACKAGE:-app.stuffstash.mobile}"
android_fingerprint="${STUFF_STASH_MOBILE_ANDROID_SHA256_CERT_FINGERPRINT:-}"

if [ -n "$ios_app_id" ] &&
  ! printf '%s' "$ios_app_id" | grep -Eq '^[A-Z0-9]{10}\.[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)+$'; then
  echo "STUFF_STASH_MOBILE_IOS_APP_ID must contain a 10-character App ID prefix and bundle ID" >&2
  exit 1
fi

if ! printf '%s' "$android_package" | grep -Eq '^[A-Za-z][A-Za-z0-9_]*(\.[A-Za-z][A-Za-z0-9_]*)+$'; then
  echo "STUFF_STASH_MOBILE_ANDROID_PACKAGE must be a valid application ID" >&2
  exit 1
fi

if [ -n "$android_fingerprint" ] &&
  ! printf '%s' "$android_fingerprint" | grep -Eq '^([A-Fa-f0-9]{2}:){31}[A-Fa-f0-9]{2}$'; then
  echo "STUFF_STASH_MOBILE_ANDROID_SHA256_CERT_FINGERPRINT must be a colon-separated SHA-256 fingerprint" >&2
  exit 1
fi

mkdir -p "$output_directory"

if [ -n "$ios_app_id" ]; then
  cat > "$output_directory/apple-app-site-association" <<EOF
{"applinks":{"details":[{"appIDs":["$ios_app_id"],"components":[{"/":"/invitations/accept","comment":"Stuff Stash inventory invitations"}]}]}}
EOF
else
  printf '{"applinks":{"details":[]}}\n' > "$output_directory/apple-app-site-association"
fi

if [ -n "$android_fingerprint" ]; then
  cat > "$output_directory/assetlinks.json" <<EOF
[{"relation":["delegate_permission/common.handle_all_urls"],"target":{"namespace":"android_app","package_name":"$android_package","sha256_cert_fingerprints":["$android_fingerprint"]}}]
EOF
else
  printf '[]\n' > "$output_directory/assetlinks.json"
fi
