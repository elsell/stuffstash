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
