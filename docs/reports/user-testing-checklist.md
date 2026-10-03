# Tests to try when convenient

Pending checks do not hold up development or releases. This is the single list
of tests that need your device or judgment; no need to answer each one now.
Unchecked items are **unverified**; known failures awaiting a fix are labeled separately.

When reporting a result, include the app version/build, device, and iOS or Android
version. “Passed” is enough for a successful check. For a failure, describe the
last action and what happened; a screenshot or recording helps when convenient.
Use a small test inventory for changes. Do not remove your original inventory.

## 1. Save and restore an archive on iPhone

Status: pending. Archive support shipped in v0.28.0, TestFlight build 162.1;
record the actual newer build you test. Connected browser and Android tests
already cover a round trip, but Android required an ADB file-copy bridge and
therefore did not verify saving through a real share recipient.

- [ ] In Inventory Settings, create an export archive with Photos and Other files
  included. Use a test inventory containing a photo and, if available, a tag,
  custom field and another file. Wait for Ready, share it, and save to Files.
  **Expected:** a ZIP is saved and remains available after leaving the app.
- [ ] Choose that ZIP through the restore file picker. Review the inventory
  contents before approving a restore into a new inventory.
  **Expected:** the review matches the source; no existing inventory is replaced.
- [ ] Open the restored inventory and a photo. Compare item names, locations,
  tags and any custom fields/files included in the test inventory.
  **Expected:** the data matches and included media opens. The original inventory
  is still intact.
- [ ] Separately try a JSON or CSV export; dismiss the share sheet once, then
  export again and save the file.
  **Expected:** cancel returns normally and the second export saves successfully.
  CSV is a flat report, not a complete backup; use the archive for restoration.

Evidence: [archive acceptance](spec-gap-evidence-2026-10-01/android-onboarding-archive/README.md).

## 2. iPhone sign-in and return

Status: pending. Android and browser sign-in have connected evidence; iPhone
runner authentication stopped before completing the system-browser journey.

- [ ] If convenient, sign out and sign back in through the system browser.
  **Expected:** the browser returns to Stuff Stash, the correct household and
  inventory open, and closing/reopening the app preserves the session.
- [ ] Open Browse, search for an item, open Details, then return to Browse and Home.
  **Expected:** the correct inventory stays selected, results remain usable, and
  returning does not leave a stuck refresh spinner.

Evidence: [connected native limitations](spec-gap-evidence-2026-10-01/connected-native/README.md)
and [Android workflow results](spec-gap-evidence-2026-10-01/core-workflows/android-connected/README.md).

## 3. Screen-reader use, if available

Status: pending. Emulator TalkBack showed focus, but did not establish complete
activation or spoken output. This does not require you to learn a screen reader;
skip it unless you already use VoiceOver or TalkBack or want to try it.

- [ ] With VoiceOver or TalkBack, navigate Home → Browse → an item's Details → Back.
  **Expected:** controls announce useful names and state; focus follows a sensible
  order, actions activate, and nothing needed to return is unreachable.
- [ ] Open Filters, choose an option, apply it, then clear it.
  **Expected:** the selected option is announced and results are reachable after
  the sheet closes; focus does not remain trapped in the closed sheet.

## 4. Return cancellation after the API fix is deployed

Status: the reported failure was reproduced against PostgreSQL. The fix passed
backend regression tests and is deployed to Paul in API v0.28.26. The device
retest below remains unverified. No new iPhone build is required.

- [ ] With a checked-out test item, tap Return on Home, then Cancel return in the
  Return details sheet.
  **Expected:** the sheet closes and the same item is checked out again; no
  “Could not cancel return” message appears. Reopen Details to confirm its state.

## 5. Invitation cancellation after the direct-command update

Status: the direct Cancel invitation command shipped in v0.28.29.
Use that version or newer for this check.
Known issue [#239](https://github.com/elsell/stuffstash/issues/239) remains open.

- [ ] In a test inventory's Sharing page, enter an email, then cancel a pending
  test invitation using its Cancel invitation command. First choose Keep
  invitation, then reopen the confirmation and cancel.
  **Expected:** confirmation remains reachable without keyboard obstruction;
  keeping preserves the invitation and cancelling updates its status.
- [ ] If a request fails naturally, retry and then navigate away and back.
  **Expected:** a readable error, usable actions and no stuck keyboard or spinner.
  There is no need to deliberately break connectivity.

Evidence: [bounded native results](spec-gap-evidence-2026-10-01/sharing-direct-confirmation/README.md).

## 6. Voice after the origin configuration correction

Status: the reported voice request was rejected with HTTP403 before session
creation. Infra commit 07396de adds the exact public API origin behind TLS
termination. No app update is required. Authenticated speech remains unverified.

- [ ] Reopen Conversation and speak a short inventory question.
  **Expected:** it connects and responds without the connection-interrupted error.
  If it fails, note the time so the server request can be correlated.

## Judgments awaiting your preference

These are decisions, not failed tests, and do not block other work.

- Android: the List/Map switcher shifts sideways when Filters disappears in Map.
  Should it stay anchored in the same position?
- Web: returning from a filtered Browse → Details → cancelled Move can return to
  Home. Should Back instead restore the filtered Browse view?

## Already confirmed or tracked elsewhere

- Notification delivery: user confirmed; no repeat requested.
- Photo double-tap zoom: user confirmed working.
- Photo swiping briefly closing the viewer: user confirmed fixed.
- iOS sharing actions obscured by the keyboard: known release follow-up
  [#239](https://github.com/elsell/stuffstash/issues/239), not an unverified test.
  The targeted post-release check is listed above.

Local-model acceptance needs a new model/host decision after its bounded failed
comparison. It is an engineering follow-up, not a device test for this checklist.

## Label downloads and scanning

Status: unverified on native devices and the physical Brother QL-800. Record the
build and device used; browser/type checks do not establish physical print quality.

- [ ] On an item, container, and location, open More → Label options. Switch QR
  with title / QR only and reference visibility; save PNG and PDF. **Expected:**
  preview matches the selected layout, files open, and changing choices during
  loading never shows an older result.
- [ ] Open system printing, cancel once, then print the PDF at actual size / 100%
  on 29 × 90 mm stock. **Expected:** text and QR fit without clipping; scanning
  succeeds. Saving/opening a print dialog must not display “printed.”
- [ ] From Browse, scan a label on iOS and Android. Deny camera permission once
  and paste its link instead. **Expected:** a clear recovery action, one asset
  navigation, camera stops, and Back returns to Browse.
- [ ] Open a label while signed out, then sign in. Try an archived asset and a
  label from a different instance. **Expected:** intended label is retained;
  authorized archived detail opens; foreign-instance data never appears. Changing
  server must reauthorize rather than forwarding credentials to the printed URL.
- [ ] Scan a QR with an old hostname after restoring the same instance to a new
  configured server. **Expected:** in-app scan still opens the asset without
  contacting the old hostname. A normal phone camera needs the old URL to remain
  reachable; this cannot be repaired by the app automatically.
- [ ] Open HTTPS labels with the app installed and uninstalled, including a
  self-hosted domain without verified app association. **Expected:** usable web
  sign-in fallback and an explicit Open in Stuff Stash action; invitations and
  OIDC callbacks still work.

## Brother QL-800 labels — pending hardware verification

Implementation tests use a stateful USB protocol fake. Read-only inspection of
Paul found no connected Brother printer; no physical print was performed.
Verify once the connector worker is integrated:

- Linux `usblp` binding and device permissions permit the registered QL-800 to open.
- The 29 × 90 mm roll produces a readable title and scannable QR in the intended orientation.
- Completion and waiting status frames match actual output; submission alone never appears as completed.
- Power-off, USB disconnect, empty roll, and cutter errors show useful status.
- Disconnect or restart during output produces uncertainty without duplicate labels.

These checks do not block independent software delivery.

After an ambiguous print, verify the connector can read a fresh idle status from
the QL-800 without printing another label. The queue must stay paused until an
editor acknowledges the unknown outcome. That acknowledgement must preserve the
uncertain attempt and permit a separate, explicit reprint. Idle confirmation is
covered by protocol-fake tests; its physical-device behavior remains unverified.
### Mobile label printing and inventory defaults

Status: source and stateful-fake checks only; native device and USB acceptance not
performed for the mobile printing controls candidate.

- On iPhone/iPad and Android, open Inventory Settings → Printers. Confirm the
  registered Brother and 29 × 90 mm media appear, and connector registration/last
  seen remain distinct from printer readiness. Unplug USB and refresh; the screen
  must not say a label completed. Check narrow layout and enlarged text.
- Change the default layout and automatic-print switch, Save, leave and return.
  Confirm saved values. Back before Save must leave stored defaults unchanged.
  With a viewer account, inspect printers without editable defaults.
- Open an active item, container and location → More → Print label. Select the
  printer, preview the QR/title label and print. Follow queued/preparing/printing
  to the reported result. Confirm physical paper output on the QL-800 separately.
- Cancel a queued job while the printer is unavailable. For uncertain output,
  inspect the printer and verify the app explains the paused queue without
  automatically issuing another label. Background/return and screen-reader
  traversal must preserve accessible controls and avoid private preview leakage.

For the create-and-print follow-up, open Add item with automatic printing enabled.
Confirm the native switch starts on, stays off after you turn it off and navigate
away/back, and resets to the inventory default for a new cleared draft. Enable it,
save an item and follow View print job. Confirm one item and one physical label.
When a submission response is lost, Retry Save must retain the original item and
label request; fields stay locked until its outcome is recovered. Check this on
iPhone/iPad and Android with the keyboard visible. These device checks remain
unverified; controlled source tests cover initialization, retained requests,
prepared tag identity and cross-inventory rejection.

For mobile uncertainty recovery, open a job with uncertain output. Before its
latest attempt reports idle, confirm Resolve job is disabled and the guidance
asks you to check the printer/connector. Once idle is reported, choose what you
observed, explicitly acknowledge uncertainty, and resolve. Confirm the screen
says Uncertainty acknowledged and retains your report without claiming confirmed
completion or printing again. Lose a response once and retry the same
acknowledgement. Verify viewer accounts have no recovery controls, and exercise
the outcome picker and acknowledgement with enlarged text and VoiceOver/TalkBack.
Status: source/stateful-fake evidence only; native acceptance remains unverified.

For mobile reprints and diagnostics, open a completed/failed/canceled job and
choose Reprint label. Pick current printer/template settings, preview an asset
label, and submit. Confirm the new job links to its predecessor and the original
job is unchanged. Uncertain and active jobs must not offer reprint. Lose a response,
leave the task, and return: Retry must recover the same job without another label.
In Inventory Settings → Printers, Print test label must enqueue one diagnostic
label only when pressed; entry and refresh must do nothing. Diagnostic reprints
explain fixed test content without pretending to preview an asset. Verify viewer
accounts cannot issue either command. Check back navigation, narrow/enlarged text,
and VoiceOver/TalkBack. Status: source and controlled-fake checks only; native
runtime and physical output remain unverified.
