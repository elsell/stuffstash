#!/usr/bin/env sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
runtime_node=${NODE_BINARY:-node}
tmp_directory=$(mktemp -d)
trap 'rm -rf "$tmp_directory"' EXIT HUP INT TERM

if grep -q 'NODE_BINARY=.*CODEX_RUNTIME_NODE_BIN.*test-mobile-release' "$repo_root/Makefile"; then
  echo 'mobile release tests must not force the Codex desktop Node path in CI' >&2
  exit 1
fi

mkdir -p "$tmp_directory/apps/mobile/ios/StuffStash.xcodeproj"
cat > "$tmp_directory/apps/mobile/ios/StuffStash.xcodeproj/project.pbxproj" <<'EOF'
MARKETING_VERSION = 0.0.0;
CURRENT_PROJECT_VERSION = 1;
MARKETING_VERSION = 0.0.0;
CURRENT_PROJECT_VERSION = 1;
EOF

"$runtime_node" "$repo_root/scripts/prepare-mobile-release.mjs" \
  --repository-root "$tmp_directory" --tag v1.2.3 --build-number 6789.2

test "$(grep -c 'MARKETING_VERSION = 1.2.3;' "$tmp_directory/apps/mobile/ios/StuffStash.xcodeproj/project.pbxproj")" -eq 2
test "$(grep -c 'CURRENT_PROJECT_VERSION = 6789.2;' "$tmp_directory/apps/mobile/ios/StuffStash.xcodeproj/project.pbxproj")" -eq 2

if "$runtime_node" "$repo_root/scripts/prepare-mobile-release.mjs" \
  --repository-root "$tmp_directory" --tag v1.2.3-beta.1 --build-number 6790.1 >/dev/null 2>&1; then
  echo 'release preparation accepted a prerelease tag' >&2
  exit 1
fi
if "$runtime_node" "$repo_root/scripts/prepare-mobile-release.mjs" \
  --repository-root "$tmp_directory" --tag v1.2.3 --build-number 10000.1 >/dev/null 2>&1; then
  echo 'release preparation accepted a build number beyond the compatibility bound' >&2
  exit 1
fi
if "$runtime_node" "$repo_root/scripts/prepare-mobile-release.mjs" \
  --repository-root "$tmp_directory" --tag v1.2.3 --build-number invalid >/dev/null 2>&1; then
  echo 'release preparation accepted an invalid Apple build number' >&2
  exit 1
fi

workflow="$repo_root/.github/workflows/testflight.yml"
grep -q 'workflow_call:' "$workflow"
if grep -q 'types: \[published\]' "$workflow"; then
  echo 'TestFlight workflow must not expose signing secrets to arbitrary published tags' >&2
  exit 1
fi
grep -q 'runs-on: macos-26' "$workflow"
grep -q '/Applications/Xcode_26.6.app/Contents/Developer' "$workflow"
grep -q 'pod --version.*1.17.0' "$workflow"
grep -q 'pod install --deployment' "$workflow"
grep -q '^COCOAPODS: 1.17.0$' "$repo_root/apps/mobile/ios/Podfile.lock"
publish_environment=$(awk '
  $0 == "  publish:" { in_publish = 1 }
  in_publish && $0 == "    env:" { in_environment = 1; next }
  in_environment && $0 == "    steps:" { exit }
  in_environment { print }
' "$workflow")
printf '%s\n' "$publish_environment" | grep -q 'NODE_ENV: production'
printf '%s\n' "$publish_environment" | grep -q 'STUFF_STASH_MOBILE_RELEASE_TAG:.*inputs.release_tag'
if grep -q '^[[:space:]]*MOBILE_RELEASE_TAG:' "$workflow"; then
  echo 'TestFlight workflow uses the obsolete release-tag environment name' >&2
  exit 1
fi
if grep -q '\$MOBILE_RELEASE_TAG' "$workflow"; then
  echo 'TestFlight workflow references the obsolete release-tag environment name' >&2
  exit 1
fi
grep -q 'GITHUB_RUN_NUMBER.*GITHUB_RUN_ATTEMPT' "$workflow"
grep -q 'xcodebuild archive' "$workflow"
grep -q 'xcodebuild -exportArchive' "$workflow"
grep -q -- '-authenticationKeyPath' "$workflow"
grep -q 'APP_STORE_CONNECT_API_KEY_BASE64' "$workflow"
grep -q 'EX_DEV_CLIENT_NETWORK_INSPECTOR: "false"' "$workflow"
grep -q 'EXDevLauncher\|EXDevMenu' "$workflow"
grep -q 'CFBundleIdentifier' "$workflow"
grep -q 'com.apple.developer.associated-domains' "$workflow"
grep -q 'uses: ./\.github/workflows/testflight.yml' "$repo_root/.github/workflows/release.yml"
grep -q 'release_tag:.*needs.validate.outputs.next_tag' "$repo_root/.github/workflows/release.yml"
if grep -Eqi 'eas |EXPO_TOKEN|STUFF_STASH_EXPO_PROJECT_ID|expo-github-action' "$workflow"; then
  echo 'TestFlight workflow still depends on EAS or an Expo account' >&2
  exit 1
fi
grep -q 'EXPO_PUBLIC_STUFF_STASH_API_BASE_URL: ""' "$workflow"
grep -q 'EXPO_PUBLIC_STUFF_STASH_TENANT_ID: ""' "$workflow"
grep -q 'EXPO_PUBLIC_STUFF_STASH_INVITATION_ORIGIN: ""' "$workflow"
grep -q 'EXPO_PUBLIC_STUFF_STASH_VOICE_DIAGNOSTICS_ENABLED: "false"' "$workflow"
grep -q 'STUFF_STASH_MOBILE_GENERAL_DISTRIBUTION_BUILD: "true"' "$workflow"
if grep -q 'secrets: inherit' "$repo_root/.github/workflows/release.yml"; then
  echo 'TestFlight caller inherits unrelated repository secrets' >&2
  exit 1
fi

app_config="$repo_root/apps/mobile/app.config.js"
grep -q "bundleIdentifier: 'org.stuffstash.mobile'" "$app_config"
if grep -q 'projectId\|EAS_BUILD_PROFILE' "$app_config"; then
  echo 'mobile application configuration still depends on EAS' >&2
  exit 1
fi
test ! -e "$repo_root/apps/mobile/eas.json"

export_options="$repo_root/apps/mobile/ios/ExportOptions-TestFlight.plist"
grep -q '<string>app-store-connect</string>' "$export_options"
grep -q '<string>upload</string>' "$export_options"
grep -q '<string>7585W4AG8C</string>' "$export_options"

info_plist="$repo_root/apps/mobile/ios/StuffStash/Info.plist"
grep -q '<string>org.stuffstash.mobile</string>' "$info_plist"
grep -q '<string>$(MARKETING_VERSION)</string>' "$info_plist"
grep -q '<string>$(CURRENT_PROJECT_VERSION)</string>' "$info_plist"
grep -q 'Stuff Stash uses the local network to connect to self-hosted servers you choose.' "$info_plist"
if grep -q '_expo\._tcp\|Expo Dev Launcher' "$info_plist"; then
  echo 'production iOS metadata still advertises Expo developer discovery' >&2
  exit 1
fi
if grep -q 'exp+stuff-stash' "$info_plist"; then
  echo 'production iOS metadata still registers the Expo development URL scheme' >&2
  exit 1
fi

python3 - "$info_plist" <<'PYTHON'
import plistlib
import sys
with open(sys.argv[1], 'rb') as source:
    info = plistlib.load(source)
assert info.get('ITSAppUsesNonExemptEncryption') is False, 'Native iOS metadata must declare no non-exempt encryption'
PYTHON
grep -q "info.get('ITSAppUsesNonExemptEncryption') is not False" "$workflow"

echo 'mobile release tests passed'
