#!/usr/bin/env bash
# Attempt every cleanup even if keychain removal or a filesystem operation fails.
set -u
signing_directory="$RUNNER_TEMP/stuffstash-signing"
cleanup_status=0
rm -f "$KEY_PATH" || cleanup_status=$?
if [ -f "$signing_directory/installed-profile-path" ]; then
  profile_path=$(cat "$signing_directory/installed-profile-path")
  rm -f "$profile_path" || cleanup_status=$?
fi
if [ -f "$signing_directory/signing.keychain-db" ]; then
  security delete-keychain "$signing_directory/signing.keychain-db" || cleanup_status=$?
fi
rm -rf "$signing_directory" || cleanup_status=$?
exit "$cleanup_status"
