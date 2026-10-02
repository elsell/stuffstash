#!/usr/bin/env bash
set -euo pipefail
work="${RUNNER_TEMP:?}/connected-native"
args=(
  -workspace apps/mobile/ios/StuffStash.xcworkspace -scheme StuffStashAudit
  -configuration Release -destination 'platform=iOS Simulator,name=iPhone 17'
  CODE_SIGNING_ALLOWED=NO ONLY_ACTIVE_ARCH=YES
)
# Compilation runs before authentication; its diagnostics are safe to inspect.
xcodebuild build-for-testing "${args[@]}"
# Runtime logs can include login entry. Do not publish these logs or raw xcresult.
xcodebuild test-without-building "${args[@]}" \
  -test-timeouts-enabled YES -maximum-test-execution-time-allowance 600 \
  -only-testing:StuffStashAuditTests/ConnectedAuditTests/testRealSignInPersistenceAndPrincipalIsolation \
  -resultBundlePath "$work/results.xcresult" > "$work/xcodebuild-test.log" 2>&1
