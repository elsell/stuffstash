# Home return details — 24 source axes

## Current native evidence, September 16

[Run351404 Home review](native-home-351404.md) records seven passing Home scenarios
on each iOS target and inspected header/tab-return captures. It also links the
existing Android header checks. Remaining gaps and fixture limits are explicit
there; the earlier source-checkpoint observations below are historical.


## Cancellation follow-up S139

Reviewed atdc901a08 against all24 axes below: HomeReturnDetailsRouteScreen,
HomeReturnDetailsSheet and useHomeReturnActions. Cancel return invokes the
recorded compensating undo operation; it is not a discard-only dismissal. It
shares the Save lock, preserves the editor after undo failure, and reconciles the
restored checkout after success. No-undo and revoked-access states offer Close.
Route Back goes through the same task owner; gesture dismissal remains disabled.
No additional cancellation defect was established. The existing phone/iPad
testHomeReturnCancelRestoresCheckout pass remains partial runtime evidence; it
does not verify permission change, background interruption or all failure states.
Search, media, notifications and imagery add no cancellation-specific controls.
Normal-size reachability, busy feedback, localized copy and assistive-technology
output remain covered by the outstanding native acceptance axes below.

## Optional details and pending/recovery follow-up S137/S138

Reviewed at5013f1af against all24 axes below. S137 owns optional note entry after
the return; S138 owns pending Save/Cancel and recovery. Source inspection confirms
the native-owned multiline seed is keyed by return session, while the application
draft receives edits. Save/Cancel share a synchronous lock. Failed save retains
the note and permits retry; unavailable undo gives Close instead of suggesting
that cancellation remains possible. Permission loss retains read-only text with
an explicit non-mutating Close. No additional source defect was established.

Native run35046586497 atb6321dcb passes both HomeReturnCancelRestoresCheckout and
HomeReturnDetailsRecoverInsideSheet on iPhone17 and iPad mini. The current complete
error-frame assertion also exists in that tested source. Inspected failed-save
captures show the entire error below the header, Returned clean retained, keyboard
dismissed, and Save/Cancel unobscured. The named M169 overlap is corrected in these
normal-size light-appearance states. Evidence:
[phone capture](phone-return-error-350465.png),
[phone hierarchy](phone-return-error-350465.txt),
[iPad capture](ipad-return-error-350465.png),
[iPad hierarchy](ipad-return-error-350465.txt).

The native retry completes and dismisses the sheet. Its synthetic save port still
does not assert persisted note contents; mounted tests assert submitted details.
This combination supports partial runtime layout/recovery evidence, not full
certification of alternate appearances, permission revocation, background/return,
physical device keyboard behavior or enlarged text. These cases remain on the
acceptance worklist. Existing finding cells stay linked to their original issues.

## Checked-out entry S065 follow-up

S065 reviewed at d6130098 across the24 axes below, with these entry-specific
differences: Home shows at most three checked-out cards, each with a named Return
action; View all navigates to Browse's checked-out scope. Permission to return is
independent of creation permission. The card itself opens Details; Return starts
the one-tap command and subsequent optional-details task. Entry has no text input,
search field or independent modal; the details route owns those interactions.

Source inspection includes HomeScreen, useHomeReturnActions and task presentation.
Tests cover duplicate callbacks, already-returned checkout identity, later checkouts,
permission loss, failed detail save/cancel and reconciliation after departure.
The shared89-case completion run passes remotely; native card target geometry,
sheet navigation, keyboard, lifecycle and reconciliation remain unverified. No
additional defect was established; prior native recovery failures stay open.

## Details route R140

R140 atcf05fb1a. Reviewed HomeReturnDetailsRouteScreen, HomeReturnTaskPresentation,
HomeReturnDetailsSheet, useHomeReturnActions, route registration, checkout spec,
HomeScreen and presentation tests. This is a production route for the existing
one-tap Return flow, not a runner-only fixture despite its name.

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | Return already completed; the sheet adds optional details or undoes the return. This sequence is explicit in asset-checkout.spec.md. |
| Navigation | Native titled sheet; task owner stays on Home. Missing task dismisses only when focused, or replaces with Home without a back destination. |
| Selection | No value picker; note entry and commands only. |
| Modality | Navigation removal requests owner cancellation/undo; gestures disabled. This is not ordinary form cancellation because the return already happened. Native behavior remains pending. |
| Layout | Direct ScrollView owns automatic insets and keyboard adjustment; content and action row scroll together. Actual sheet bounds/keyboard reachability require runtime evidence. |
| Adaptation | Action row wraps with available width. Normal-size phone/tablet checks remain required; enlarged layout is not certified. |
| Typography | Item title, optional-note label and status/error text. Long title/error wrapping requires native captures. |
| Appearance | Semantic palette and native command buttons. Sheet dark/light appearance remains unverified. |
| Localization | English action and recovery copy; freeform note preserved. RTL layout and input unverified. |
| Imagery | No images or icon-only commands. |
| Targets | Native commands in flexing wrappers; verify actual hit regions and keyboard overlap on each device. |
| Gestures | Visible Cancel return/Close and Save; interactive dismissal is disabled to avoid bypassing the owned operation. |
| Keyboard | Multiline input seeded once per session; native text owns the caret while application draft receives changes. Diagnostic single-line failures elsewhere do not prove this input safe. |
| Accessibility | Input named Optional return details; errors and access-change text marked alerts. Focus movement, VoiceOver order and command reading remain open. |
| Motion | Native presentation; no custom timed transitions. Reduced-motion behavior needs runtime review. |
| Content | Explicitly says the item is already returned after lost access; missing undo ID warns that cancellation is unavailable. Save/Cancel purpose follows the checkout spec. |
| Search | No search task within this note form. |
| Loading | Save and undo share a lock, disable entry/actions, and expose Saving/Canceling return labels. Initial return pending belongs to Home. |
| Recovery | Inline save/undo errors retain the note and permit retry. Failed reconciliation notices are visit-scoped. |
| Editing | Per-session note seed; pending lock rejects repeated changes. Save reads application draft; native complete-text input remains an acceptance requirement. |
| Privacy | Current permission gates edits and operations; revocation retains read-only note and allows Close. Server authorization remains a separate boundary. |
| Notifications | No screen-owned notification action. Interruption/resume behavior needs lifecycle acceptance. |
| Media | No microphone, camera or upload interaction. |
| Lifecycle | Mounted/session ownership prevents old callbacks changing a later return editor. Owner disappearance behind another screen waits for focus before dismissing. Late mutation/error and native return need acceptance. |

Existing mounted Home and route tests cover duplicate return, late completion,
retained failed details, Back routed through undo, access revocation, stale editor
callbacks and missing-owner recovery. They are included in the 1,734-test full
checkpoint after M166. No new native pass is inferred from that checkpoint.

Run35029854251 at e8b3d42dccf3f13428fb26bbb1cfd85ea8b0dd9e passed both
testHomeReturnCancelRestoresCheckout and testHomeReturnDetailsRecoverInsideSheet
on iPhone17 and iPad mini. These check reachable cancellation/restored checkout,
complete `Returned clean` text, keyboard dismissal, retained failed-save text,
retry and sheet dismissal. The synthetic save port does not inspect the submitted
note, so this is not proof of persisted note contents. Initial and failed-save
iPhone screenshots were inspected. Initial form and commands are visible. After
failed Save, the error heading partly sits under the navigation blur (M169).
The test only checks existence, so its pass does not establish error visibility.
See [retained error screenshot](phone-return-error-350298.png). Terminal evidence
adds partial navigation/editing/recovery coverage only.
