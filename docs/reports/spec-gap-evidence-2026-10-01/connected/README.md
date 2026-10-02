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

## Connected export recovery

CI [36912314042](https://github.com/elsell/stuffstash/actions/runs/36912314042)
passed at `b6b07b1c`: real Dex sign-in, UI item creation, actual JSON/CSV downloads
with item-content and CSV-header assertions, persisted `inventory.exported`
history for both formats, and absent/other-principal access rejection. The test
runs against migrated PostgreSQL and SpiceDB, without route interception.
[Inspected export screen](browser-export.png) shows the download confirmation;
the file-content and audit assertions, not the screenshot, prove exported data.

The earlier HTTP500 was caused by the missing PostgreSQL audit-action constraint
migration. PR #222 merged `39700a21`; main now requires the connected browser
check. Release run36913294557 succeeded as v0.26.2, including TestFlight upload;
that delivery result is separate from runtime acceptance.
This does not verify saving exports into a physical iOS receiving application.


## October 2 connected everyday web workflows

Tested the isolated Paul server with normal English text in Chromium 1228 at
1440×1000 and 390×844. Server images were the pinned v0.27.3 API
`5d07b3ea41cb2236ed335b79816683c356c0d2a21536ee41da53e46edba0bfed` and web
`402aac247ff17a6debe72bbee482f0a31e60a794ca9564142611f264dfe69630`.
Signed in through real Dex; no injected session or intercepted API responses.
The browser trusted the exact server certificate public key after verification
against the isolated server's CA. No general TLS-verification bypass was used.
Only synthetic inventory data was changed.

Observed: search and Availability filtering returned the item; editing its name
survived reload; moving it to Inventory root and back to a searched location
survived reload; Map search revealed the containing location and item. Inspected
[Filters](paul-filters-narrow.png) and [Move](paul-move-narrow.png) at narrow width
show reachable footer actions. This does not establish large-text, assistive,
physical-mobile or connected native acceptance.

**Finding awaiting user decision:** [filtered Browse](paul-filtered-browse.png)
→ item → Move → Cancel → item Back ends at [Home](paul-detail-back-home.png),
not the originating filtered results. This was reproduced twice. Returning to
Browse through navigation subsequently showed an empty search query. Source
`assetDetailBackHref` deliberately falls back to Home when no location is selected;
the current tests assert that fallback. The observation is not a flaky selector
or a failed Move persistence operation. No fix has been applied pending the user's
requested confirmation of newly observed UI problems.
