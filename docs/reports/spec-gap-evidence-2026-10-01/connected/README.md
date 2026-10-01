# Connected authentication and native image evidence

October 1, 2026. Evidence for the scoped V1/G7/G8 follow-up in PR #218.

## Real browser sign-in and isolation

[CI 36905702456](https://github.com/elsell/stuffstash/actions/runs/36905702456)
passed at `36f9ecec`, including the connected browser journey and existing Go,
client, security and regression checks. Chromium signed in through the pinned
local Dex fixture using the real authorization-code/PKCE callback. The app created
the first workspace and retained Browse after reload. The same inventory returned
200 for its owner, 401 without authentication, and denial for a separately signed-in
principal. No injected browser session or intercepted API response supplied this
result. [Inspected post-login screenshot](browser-browse.png).

This proves an isolated local stack, not production-provider connectivity or all
browser workflows. CI does not retain sign-in traces or general error-context files.

## Native gallery callbacks

[Native run 36904211794](https://github.com/elsell/stuffstash/actions/runs/36904211794)
uses source `4ef2f027`. Its iPhone 17 simulator check passed. It renders the
production detail gallery and uses the production mobile telemetry session with
the real runtime clock and a controlled delivery sink. The image is the repository's
bundled `stuff-stash-glyph.png`; the failure source appends `.missing` to that URI.
Cache state was not controlled. This is one repetition, not a latency benchmark.

Inspected [loaded image](iphone-image-success.png) and
[unavailable preview](iphone-image-failure.png) match the tested states.
[Raw allowlisted samples](iphone-image-measurements.json) include success
(39.44 ms), failure (0.0023 ms) and cancellation (1.72 ms). The cancellation is
retained rather than discarded. These native callback measurements do not identify
network latency or establish cold/warm percentiles, rendering overhead, physical
performance or real telemetry-server delivery. XCTest checks exact keys, platform,
surface, variant, outcomes and finite durations bounded to 60,000 ms.

This record covers iPhone; the linked run also schedules iPad verification as a
separate release check. Connected native sign-in, broader assistive/physical
acceptance and production performance evidence remain separate outstanding work.

## Application boundary extraction

Voice vocabulary projection and resolution now live in the agent-model application
package. Source comparison verified identical bodies after explicit symbol changes.
Existing isolation and expiration-capability tests passed in the CI run above.
This completes that extraction; broader realtime orchestration remains tracked.
