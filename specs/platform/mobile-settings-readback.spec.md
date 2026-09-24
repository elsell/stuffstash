# Connected settings readback audit

The existing native editor fixture mutates one repository and then replaces itself
with an independently seeded collection. A saved notice and return therefore do
not establish collection readback or reopening persisted values.

Add an isolated connected tag journey using the production collection/editor,
real queries/commands and one shared in-memory repository/query client across
routes. Verify collection → edit → rejected save → retry → updated collection →
reopen → cancel. Also verify creation appears and archive removes the tag from
the same collection. Capture normal-text collection and editor composition.

Fixture-only state must never reach a production session or API. Preserve scope
checks and reject unknown IDs. No fixture callback may directly update the visible
collection in place of the normal query invalidation path. These checks establish
controlled-client persistence, not server persistence or physical-device quality.
This is subsequent audit work and does not expand the frozen selection release.

Run35930694492 exposes missing fixture invalidation: compose the fake repository
through the same ObservedCustomizationRepository and query-client mutation observer
as production, using the persistent journey scope. A cache-backed regression must
prove rejected commands keep cached reads intact and successful commands invalidate
and refresh them. Never directly set the collection after save. Also require the
returned row to clear the header and remain hittable; the failed capture shows an
old row behind the header, which needs verification after correcting the wiring.

Run35937082802 now returns the updated tag on both devices, proving the corrected
fixture follows cache invalidation. The row remains at y24 behind the native
header and is not hittable. Keep this distinct from persistence. The collection
currently replaces a non-scrolling loading root with a scroll view. Preserve one
scrolling root through loading, ready and retry states so native inset ownership
is stable; do not add a guessed header-height offset or force a navigation-return
scroll. Keep the native row hittability/header-clearance and full reopen/create/
archive journey unchanged. This candidate requires native verification.

Run35939654846 rejects the stable-scroll-root hypothesis: the updated phone row
still starts at y24 under the Tags header ending at y116. The header is already
opaque in the fixture and production stack; changing transparency would not
distinguish a cause. Use one explicit iOS viewport owner instead: reserve the
measured native header above the collection ScrollView and disable that scroll
view's automatic content insets. The viewport, including its refresh control,
starts below navigation; do not force content offsets or guess a device height.
Respect the bottom safe area. Android retains its existing inset behavior.
Keep loading, ready and retry ownership stable and retain the unchanged native
Save/reopen/create/archive assertions. This is the second layout candidate and
requires native proof before release; do not repeat the first candidate.

## Subsequent release batch

Combine this collection correction and clearer settings action grouping with Move
completion emphasis, based on the integrated filter/detail branch. Preserve its
current dependency patch and locks; the earlier Settings branch's obsolete sheet
patch is not part of this batch. The native `settings-move-release` subset exercises
connected collection Save/reopen/create/archive, native collection search/Add,
and the four existing Move/Move
Here journeys, retaining their search, creation, selection, rejection and retry
assertions. Review phone/iPad screenshots as well as terminal results. Dispatch
only after the integrated follow-up has passed. The failed Settings layout
hypothesis is replaced by the explicit viewport decision above; do not repeat it.
Require the initial and returned first row to be hittable and within48 points
below the header, catching doubled spacing as well as overlap. Reuse the existing
search/Add case to verify expanded search, clear/cancel and restored rows.
