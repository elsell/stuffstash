# Focused filter evidence — run35159542174

Source624bcf29. iPad job105008548645 completes4/5 cases. Phone subsequently reaches the job timeout; see the terminal result below.
Browse last-tag selection now passes its explicit checked-state observation and
final applied-tag result. Browse search, availability and expiration search pass.
This is native assertion evidence; it does not prove the earlier intermittent
last-tag failure has a production fix.

The calendar popover dismisses successfully. The failure is later, waiting for
Choose date range after pressing Back. The reviewed final capture shows the fixture
launcher. The dismissal capture hierarchy already contains Choose date range and a
Filters navigation title: the outside tap also returned to the filter overview.
Its computed point(372,873.75) is inside the underlying Back frame(102,820.5,540,54).
The next Back therefore closes the sheet. The new dismissal target avoided the
calendar but failed to avoid an underlying command. Do not classify this as a
proven broken production Back handler. The test needs a non-command outside target
and must verify it remains on Date range immediately after calendar dismissal.

Retained evidence: [dismissal geometry](evidence/ipad-calendar-geometry-351595.txt),
[post-dismissal hierarchy](evidence/ipad-calendar-dismissed-351595.txt),
[final capture](evidence/ipad-calendar-return-351595.png), and
[selected-tag observation](evidence/ipad-last-tag-selection-351595.txt).
Full artifact10472872529 remains on paul at `/tmp/native351595-filters-ipad.zip`;
log `/tmp/native351595-filters-ipad.log` is local. Phone geometry remains needed
before changing the production keyboard inset.

Candidate test correction: restrict outside-popover targets to the current Date
range navigation bar and exclude all button/title bounds. Record geometry before
requiring a candidate, then assert the Date range navigation bar remains after
popover dismissal. The native failure above supplies the red case; eight fixture
preparation checks and mobile structural checks pass remotely. Critic review found
and corrected missing geometry on the no-candidate path. Swift compilation and
native acceptance remain pending. Production controls are unchanged.

## Terminal phone result

Job105008548835 is cancelled at the120-minute workflow limit. Its first last-tag
case fails in setUp at app.launch: Xcode reports an application launch timeout.
The next Browse keyboard case starts, then the job remains in the test step until
cancellation. This is not a completed filter journey and gives no new M249 geometry.
Do not infer a production filter or input failure from this run; launch root cause
is unknown. The earlier reproduced phone overlap remains valid evidence.

Export fails because results.xcresult has no Info.plist. Artifact10476406134
(8,761,603 bytes) retains build/test output and a diagnostics log containing only
a uid lookup message and header. No usable app trace or screenshots are available.
Local archive: `/tmp/native351595-filters-phone.zip`; complete job log:
`/tmp/native351595-filters-phone.log`.

Release-corrections35165286470 at966158e6 has now started on both devices. A fresh
focused filter run35168921441 at that same source is queued to recover the missing
phone probe evidence, including the already-corrected calendar target. The active
correction and full audit jobs were not cancelled or restarted.
