# iPad native audit — run35121454700

Source1a15ca11, tested mergec5b6f1b110135c3d80f1433429c84dcd248d8fdb.
Job104881256483 ended cancelled after exceeding the90-minute budget; GitHub's
check annotation explicitly confirms that cause. No manual cancellation occurred.
The complete log is retained locally at `/tmp/native351214-ipad-complete.log`.

The log records85 completed tests:79 passed and6 failed in4086.168 seconds.
The final test ended17:56:43 UTC, cancellation occurred17:56:58, and screenshot
export failed because the result bundle lacked Info.plist. Artifact upload also
failed; the run's artifact inventory contains phone fixtures and both onboarding
jobs, but no iPad fixture artifact. These are log results, not inspected captures.

## Failures and next acceptance

- Add draft in navigation stack: UI snapshot query timed out at line1518 of the
  tested source. The test took386 seconds. No screenshot is available to establish
  whether a product hang or the automation query caused it.
- Controlled text without accessory: retained assertion value `Natve draft namei`
  differs from `Native draft name`. Character ordering remains unresolved.
- Inbox: AX glyph measured24.5 points against the old44-point assertion. The
  current delivered-touch candidate is newer; this failure does not prove a small
  delivered touch region.
- Enlarged-text Edit metadata, Edit tags and Move Here fail; deferred until the
  normal-text findings are addressed, per user priority.

Sharing recovery, ordinary color opening, the color touch-region probes, Add photo
return, preconfigured Place search and Voice proposal retry/return pass in the log.
Their passes do not clear phone failures or imply uninspected visual acceptance.
The job remains a failed/incomplete evidence gate.

Future jobs receive120 minutes to preserve result finalization and export. Run
35130374705 at0586f845 started automatically after this job ended and retains its
existing90-minute budget; leave it undisturbed. It includes the corrected inbox
and Add-return assertions and the managed-search/header-action comparison.
