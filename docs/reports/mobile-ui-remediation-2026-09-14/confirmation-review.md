# Confirmation call-site review

Reviewed at64ec341d. This review covers the named paths below, not every dialog
or every axis of their screens. `confirmation-call-sites.csv` remains the inventory;
rows without evidence stay pending. Native dialog appearance, VoiceOver order and
actual dismissal timing require their separate acceptance traces.

| Path | Task and pattern fit | Ownership evidence | Remaining verification |
| --- | --- | --- | --- |
| Asset Archive / Restore / Delete permanently | Existing project policy uses a native alert that names the asset and consequence; permanent removal is distinguished from reversible lifecycle changes. | AssetDetailRouteScreen captures the command visit and consumes acceptance once. Mounted tests cover earlier visit, teardown and reuse after completed Archive/Restore/Delete. | Native action-menu-to-alert flow, cancel and correct return; no broad runtime claim. |
| Invitation cancellation | Native confirmation names the affected email and explains loss of invitation access; Keep Invitation is the cancel path. | InventorySharingScreen captures the initiating visit, consumes acceptance once and locks each invitation independently. Mounted tests cover leave/return and simultaneous cancellations of different invitations. | Native menu, alert, pending row and retry on the corrected build. |
| History reversal | Reversal is an explicit compensating change, not a generic dismiss; native confirmation explains the recorded action. | AssetHistoryDetailRouteScreen binds the callback to its presentation session and operation scope, rejects unavailable/pending reversal and suppresses late navigation/feedback. Mounted tests exercise different activities, leaving, late completion and failure. | Physical alert interaction and return destination. The helper only presents; ownership resides in the screen. Do not infer generic single-use behavior from the helper. |
| Provider credential / prompt discard | Native discard confirmation is appropriate for losing a replacement draft; Keep Editing preserves it. | useProviderEditorExit checks saving and captured presentation before acceptance, consumes the exit and disarms removal before dispatch. Mounted tests cover blur/refocus, duplicate acceptance and immediate Back during saving. | Native navigation removal and draft retention on the corrected build. |
| Push-open failure | An acknowledgement explains why an explicitly requested destination cannot open. No mutation callback is attached to OK. | PushNotificationNavigation aborts an earlier request and its effect cleanup; only a non-aborted result can navigate or show the dialog. | Physical cold/warm notification taps and session replacement. Source review does not replace physical push evidence. |

The shared AppFeedback dialog adapter forwards caller actions to the system alert;
it does not establish their task ownership. Each mutating caller therefore remains
responsible for its own validity and lifetime checks. AppServicesFeedbackGate,
provider archive, customization dialogs, photo-source selection and sheet operation
failure alerts still need completion of their individual reviews. Already-corrected
photo, expiration, account, conversation and Edit discard entries retain their
finding-specific evidence rather than being relabeled as fully verified here.

Validation on paul:121 tests across AssetDetailRouteScreen,
AssetHistoryDetailRouteScreen, AssetHistoryRevertAction, InventorySharingScreen
and SettingsScreens pass at64ec341d. Log: /tmp/confirmation-audit-64ec341d.log.
The tests substantiate the specific ownership behaviors above; they do not render
native dialogs or close the physical acceptance gaps.
