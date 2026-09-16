# iPad native run350806

September16, iPad mini(A17 Pro), job104756069021, source da1a3e38
(merge c04b6a5bac882cf50df6a7926e415a421e8c13d7): **74/81 pass,7 fail**.
Both fixture jobs are now terminal; the newer3dc2bab6 run35091949871 is active.
Log `/tmp/native350806-ipad.log`; artifact10444464015 retained on paul at
`/tmp/native350806-ipad.zip`. No TestFlight acceptance follows from these samples.

Color ordinary opening/clear, all nine delivered-touch probes, single accessible
name and target activation pass. Both Place search variants, settings collection
search and Sharing recovery pass. The remaining seven failures are Add hidden-header
entry, enlarged Edit metadata/tags and Move Here, pushed notice geometry, ordinary
controlled text entry (`Native draeft nam`), and proposal-location search clearing.

[Conversation entry](evidence/ipad-voice-context-350806.png) now places inventory
context minY380.5 below header maxY378.5. The location control and proposal actions
are visible; the preceding user message is partially clipped by the scrolled
transcript viewport. This resolves the observed entry context overlap in this
sample, not all detents, lengths or full conversation acceptance.

## Focused search clearing investigation

The proposal-location journey passes lookup retry and unmatched search. Before
Clear text, the hierarchy exposes the search field as Keyboard Focused/Focused,
value `missing`. Clear text bounds are(695.5,73.5,20.5,20.5); the synthesized tap
lands at their center(705.75,83.75). After that tap, the
[final capture](evidence/ipad-location-cleared-350806.png) shows restored Garage bin
and a reachable-looking header Search icon, with no search field or keyboard.
The original field-preservation assertion fails. Do not claim the subsequent
fresh-query/selection/Back steps passed; they were never reached.

This is a runtime observation requiring investigation, not yet a root-cause claim.
The existing spec accepts collapse for an *unfocused* field. The retained pre-tap
hierarchy here reports focus, so that exception cannot simply waive this failure.
A static native search diagnostic now types/clears while focused, records whether
the field remains or collapses, then exercises a fresh query. It has no query-state
callbacks but uses a different presentation context from Conversation, so it cannot
alone isolate the cause. The production focused-clear assertion remains intact.
The new Swift diagnostic awaits macOS execution; no production workaround was added.

Selected evidence is under `/tmp/ipad350806-selected/` and
`/tmp/ipad350806-context/`. Full ZIP preserves event and screenshot metadata.
## Pushed notice observation

The failed predicate's last sample has valid app/content/header bounds but zero
rectangles and empty identifiers for all three notice controls. The
[retained final capture](evidence/ipad-notice-final-350806.png) and hierarchy instead
show notice(16,96,712,70), dismiss(33,107,489.5,48), action(532.5,107,178.5,48),
all below header bottom86 and inside content(0,86,744,1047). The log records repeated
failed element resolutions during the geometry predicate. This contradicts a
claim of visible header overlap in the final state; it does not prove the earlier
bounded observation passed or establish the cause of the zero rectangles.

The acceptance gate now reads app, content, header and control frames from a single
public XCTest snapshot per sample. Missing elements, snapshot errors, empty bounds
and failed full containment still fail. The five-second timeout and later actual
command/dismiss/navigation checks remain. This changes measurement consistency,
not product layout or acceptance requirements. Swift compilation/native execution
remain pending on macOS; remote mobile structural check passes. Local evidence:
`/tmp/ipad350806-notice/`; source log: `/tmp/notice-snapshot-structural.log`.
