#!/usr/bin/env sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmp_directory=$(mktemp -d)
trap 'rm -rf "$tmp_directory"' EXIT HUP INT TERM
output="$tmp_directory/StuffStash.entitlements"
writer="$repo_root/apps/mobile/scripts/write-ios-associated-domains.sh"

EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN='https://stash.example.test' "$writer" "$output"
plutil -lint "$output" >/dev/null
grep -q '<string>applinks:stash.example.test</string>' "$output"

EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN='' "$writer" "$output"
plutil -lint "$output" >/dev/null
if grep -q 'com.apple.developer.associated-domains' "$output"; then
  echo 'an unconfigured development build received an associated-domain entitlement' >&2
  exit 1
fi

if EAS_BUILD_PROFILE=production EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN='' "$writer" "$output" >/dev/null 2>&1; then
  echo 'a production iOS build accepted a missing invitation origin' >&2
  exit 1
fi

if EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN='https://stash.example.test:8443' "$writer" "$output" >/dev/null 2>&1; then
  echo 'an iOS build accepted a nonstandard invitation origin port' >&2
  exit 1
fi

for invalid_origin in \
  'https://user:secret@stash.example.test' \
  'https://stash.example.test/path' \
  'https://stash.example.test?query=yes' \
  'https://stash.example.test#fragment' \
  'https://stash example.test'; do
  if EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN="$invalid_origin" "$writer" "$output" >/dev/null 2>&1; then
    echo "an iOS build accepted invalid invitation origin: $invalid_origin" >&2
    exit 1
  fi
done

EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN='http://192.168.1.117:5173' \
  EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP=true \
  "$writer" "$output"
plutil -lint "$output" >/dev/null
if grep -q 'com.apple.developer.associated-domains' "$output"; then
  echo 'a private HTTP development origin received an associated-domain entitlement' >&2
  exit 1
fi

if EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN='http://192.168.1.117:5173' \
  EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP=false \
  "$writer" "$output" >/dev/null 2>&1; then
  echo 'an iOS build accepted private HTTP without explicit opt-in' >&2
  exit 1
fi

if EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN='http://8.8.8.8:5173' \
  EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP=true \
  "$writer" "$output" >/dev/null 2>&1; then
  echo 'an iOS build accepted public HTTP with local opt-in' >&2
  exit 1
fi

EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN='http://[::1]:5173' \
  EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP=true \
  "$writer" "$output"
plutil -lint "$output" >/dev/null
if grep -q 'com.apple.developer.associated-domains' "$output"; then
  echo 'an IPv6 loopback HTTP development origin received an associated-domain entitlement' >&2
  exit 1
fi

if EAS_BUILD_PROFILE=production \
  EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN='http://192.168.1.117:5173' \
  EXPO_PUBLIC_STUFF_STASH_INVITATION_ALLOW_INSECURE_LOCAL_HTTP=true \
  "$writer" "$output" >/dev/null 2>&1; then
  echo 'a production iOS build accepted private HTTP' >&2
  exit 1
fi

grep -q 'write-ios-associated-domains.sh' "$repo_root/apps/mobile/ios/StuffStash.xcodeproj/xcshareddata/xcschemes/StuffStash.xcscheme"
test "$(grep -c 'CODE_SIGN_ENTITLEMENTS = \"$(DERIVED_FILE_DIR)/StuffStash.entitlements\"' "$repo_root/apps/mobile/ios/StuffStash.xcodeproj/project.pbxproj")" -eq 2
