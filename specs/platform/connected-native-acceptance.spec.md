# Connected native acceptance

## Purpose and scope

Close the October 1 audit's connected native evidence gap with a production mobile
client talking to an isolated real API through OIDC and SpiceDB. Fixture screens,
injected sessions and server-only checks do not satisfy this requirement.

The first bounded batch covers three gaps:
- V1 native sign-in and persistence: use the system authentication session, create
  an inventory item, relaunch the app, and reopen the saved item.
- V1 native identity isolation: sign out, sign in as a different principal, and
  verify the first principal's inventory is unavailable; corroborate with denied
  direct API access using the second principal's real token.
- D2 evidence integrity: retain source/server revisions, explicit assertion results
  and inspected post-authentication captures, with missing or failed steps marked
  unverified. Never retain credentials, tokens, sign-in screenshots or auth traces.

Normal English and iPhone come first. Broader tablet, physical, assistive and
performance acceptance remain separate requirements. This is not permission to
replace the existing production PostgreSQL/Garage browser checks with SQLite.

## Isolated runtime

Use a macOS GitHub runner for the iPhone simulator. Prefer native processes on the
same runner over adding remote network exposure or reusable tunnel credentials:
Stuff Stash API, real Dex and real SpiceDB. The API may use its supported SQLite
and filesystem adapters for this isolated journey; SpiceDB may use its actual
in-memory datastore. Do not substitute authorization or authentication fakes.

Use the repository's reviewed Dex and SpiceDB versions. Pin any downloaded source
or binary by immutable commit or digest, verify it before execution, and keep Go,
Node, Expo and Xcode selections consistent with existing workflows. Run compilation
on the hosted runner, never this Linux workstation. Each run gets disposable data,
loopback listeners, synthetic principals and explicit process cleanup. Configure
the API's required invitation origin explicitly as loopback HTTP with the existing
local-development opt-in; the harness must meet real startup validation.

The simulator uses real onboarding and configured OIDC discovery. Do not run the
fixture-route preparation script, replace repositories, intercept API responses,
or inject authentication state. Local HTTP is restricted to loopback within the
isolated verification run; production transport policy remains unchanged.

## Critical checks and stop conditions

Verify discovery and service readiness before building/running the app. Assert
missing-token API access is denied before the native journey. Bound startup and
workflow waits by the expected operation, and fail with a specific stage when it
cannot complete. Preserve the original failure result during cleanup.

Typing assertions wait for the full expected value and keyboard readiness before
submission; never infer successful entry from a screenshot alone. Tests must
observe server-persisted state after relaunch, not only optimistic mobile state.
Cross-principal access must use distinct actual identities and tenant membership.

One initial runtime attempt may distinguish infrastructure startup, system sign-in,
and product-workflow failures. A failed attempt requires a diagnosis and an
implementation decision before another run; unchanged retries are prohibited.
A sleeping script waits for the terminal workflow result. Upload only bounded,
allowlisted evidence. This spec defines intended verification, not achieved results.
