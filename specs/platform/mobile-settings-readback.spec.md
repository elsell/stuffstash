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

## Subsequent release batch

Combine this collection correction and clearer settings action grouping with Move
completion emphasis, based on the integrated filter/detail branch. Preserve its
current dependency patch and locks; the earlier Settings branch's obsolete sheet
patch is not part of this batch. The native `settings-move-release` subset exercises
connected collection Save/reopen/create/archive and the four existing Move/Move
Here journeys, retaining their search, creation, selection, rejection and retry
assertions. Review phone/iPad screenshots as well as terminal results. Dispatch
only after the existing Settings candidate and integrated follow-up have passed;
do not use this batch to repeat an unchanged failing hypothesis.
