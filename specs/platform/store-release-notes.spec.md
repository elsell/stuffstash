# Store release notes

## Scope

Store delivery must use the shared release-note extraction backported from
make-app. Reviewed `Release notes:` bullets must take precedence; existing
`TestFlight notes:` bullets must remain supported. Feature/fix/performance subjects
must be the fallback, with deduplication and first-parent stable-tag ranges.
TestFlight must retain its version/build heading, changelog link and 4000-character
bound. Google Play must use the same highlights compacted to 500 Unicode code points.

## Delivery

TestFlight notes must publish automatically after successful upload, targeting the
exact bundle/version/build and verifying the saved notes. A separate manual notes
workflow must support rerunning publication without rebuilding. Existing signing
configuration and credentials must remain unchanged.

Play publication must be opt-in through a protected `play-store` environment and
manual/reusable notes workflow. It must use the Android package identity, explicit
track, singleton version code and locale. It must not upload binaries, infer latest,
change rollout state, or overwrite other releases/locales. A main-ancestor stable
published tag must supply the notes. An isolated edit must be verified, committed
once and verified again through a fresh edit. Matching notes must cause no commit.
OAuth and Publisher requests must use fixed Google origins, reject redirects,
sanitize errors and never automatically retry uncertain writes.

## Acceptance

Tests must cover reviewed/fallback extraction, Unicode bounds, real Git history,
exact release matching, preserved track state, idempotency, malformed responses,
permission/transport failures, safe authentication and workflow wiring. Tests must
not contact live stores. Android binary delivery remains product-owned: its caller
must pass the precise version code after assigning that build to its track.
