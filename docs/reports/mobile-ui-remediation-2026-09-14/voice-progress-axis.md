# Proposal progress and photo recovery — 24 source axes

S123 atcd353ee2 plus M176 candidate. Reviewed VoicePlanProgress and presentation,
the active proposal and historical exchange consumers, provider retry ownership,
and controller attachment progress/failure/retry. Native screenshots, announcement
timing and physical uploads are not established by this review.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Distinguishes applying an approved change from adding photos after the change is saved. Partial upload failure must not imply the inventory write failed. |
| Navigation | Progress remains inside the active exchange and retained history. Retry addresses the original plan, not whichever exchange is newest. |
| Selection | Not applicable: progress/retry offers no value selection. |
| Modality | No additional modal is introduced for background attachment status. Native conversation dismissal/return still needs acceptance. |
| Layout | Inline progress and retry remain within the conversation scroll viewport. No absolute overlay; keyboard/small-detent visibility needs native inspection. |
| Adaptation | Title flexes alongside indicator and percentage. Long descriptions and narrow phone/tablet layouts remain unverified. |
| Typography | Shared text scaling, wrapping details and no fixed row height. Normal-size failure reasons need native acceptance before larger-text work. |
| Appearance | Palette-backed track, text and status. M176 uses a warning icon for terminal attachment failures, retaining the saved-change title. Contrast remains unmeasured. |
| Localization | English status/count strings, plural handling and numeric percentage; translation/RTL are not certified. |
| Imagery | M176 replaces the success checkmark for terminal photo failure with a warning symbol. Text also names the attention state, so meaning does not depend on symbol/color. |
| Targets | Active/history Retry photos uses native command adapter. Source tests verify callback IDs; actual native hit region remains pending. |
| Gestures | Explicit Retry; no gesture-only recovery. Scroll position is owned by conversation. |
| Keyboard | Status itself has no text input; retry coexists with conversation keyboard. Reachability and focus retention require native testing. |
| Accessibility | Grouped label/value and polite live region expose task, counts and measured percentage. M176 restores partial-failure reason in accessible text. Native announcements/order remain open. |
| Motion | Indeterminate native activity while saving; measured fill only for counted attachments. Text provides a static alternative; Reduce Motion runtime behavior remains pending. |
| Content | Cumulative attached/total/failure counts survive retries; no fabricated time percentage during inventory save. Both active/history use the shared presentation. |
| Search | Not applicable: progress has no searchable choices or filter scope. |
| Loading | Saving and photo uploading are separate stages. Retry marks progress and suppresses duplicate retry in provider; applying the plan is not repeated. |
| Recovery | M176 fixes loss of safe failure detail when only some attachments fail. Retry keeps successful attachments and attempts failures only. Permanent missing-intent failures remain counted. |
| Editing | Execution status changes only on confirmed events. Attachment retry does not reopen proposal editing or repeat the inventory mutation. |
| Privacy | Failure reasons follow the existing allowlist; arbitrary exception text becomes generic. New mixed-success test includes a synthetic private URL and token to verify redaction. No authorization policy changes. |
| Notifications | Not applicable: the progress widget does not register notifications or own notification entry. |
| Media | Attachment retry uses retained photo inputs and reports saved/partial outcomes. Physical library access and upload finalization are outside mounted evidence. |
| Lifecycle | Provider lifetime/plan guards reject old completions and update matching history. Scope reset/disposal clears retry state. Native background/dismissal is still unverified. |

M176 initially failed three remote regression cases: safe mixed-success detail,
arbitrary exception redaction after mixed success, and visible/accessibility text.
The existing cumulative-retry case now checks counts plus the intended failure
message. All1,769 mobile tests across278 files, TypeScript and structural checks
pass remotely on paul (`/tmp/mobile-m176-full.log`). Code critic found no blocker.
Source coverage is not a native pass.
