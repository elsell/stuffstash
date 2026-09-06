#!/usr/bin/env bash
set -eu
script_directory=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
directory=$(mktemp -d)
trap 'rm -rf "$directory"' EXIT
export RUNNER_TEMP="$directory"
export KEY_PATH="$directory/key.p8"
mkdir -p "$directory/stuffstash-signing" "$directory/bin"
touch "$KEY_PATH" "$directory/profile.mobileprovision" "$directory/stuffstash-signing/identity.p12" "$directory/stuffstash-signing/signing.keychain-db"
printf '%s' "$directory/profile.mobileprovision" > "$directory/stuffstash-signing/installed-profile-path"
# Controlled security adapter simulates an unavailable keychain service.
printf '#!/bin/sh\nexit 7\n' > "$directory/bin/security"
chmod +x "$directory/bin/security"
result=0
PATH="$directory/bin:$PATH" bash "$script_directory/cleanup-ios-signing.sh" || result=$?
test "$result" = 7
test ! -e "$KEY_PATH"
test ! -e "$directory/profile.mobileprovision"
test ! -e "$directory/stuffstash-signing"
bash "$script_directory/cleanup-ios-signing.sh"
