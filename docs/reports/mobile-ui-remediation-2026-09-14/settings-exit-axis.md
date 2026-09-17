# Settings dirty-editor exit

S110 source review at0f378691 with M155 candidate. Reviewed
CustomizationEditorScreen, CustomizationEditorWorkflow, useProviderEditorExit,
and the mounted customization navigation tests. This reviews the exit interaction;
each editor's content and mutation boundary remains a separate surface.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Prevent accidental loss of an unsaved settings draft; confirmation is appropriate only when dirty. |
| Navigation | Navigation action is retained, authorized once, then dispatched after guard removal. Provider exit separately disarms its native guard. Native Back timing remains open. |
| Selection | Keep Editing cancels; Discard authorizes the captured exit. M155 rejects a stale confirmation after ownership changes. |
| Modality | System Alert owns the confirmation; no nested custom selection screen. Native alert dismissal and interruption need verification. |
| Layout | System alert controls its geometry; editor header suppression is source behavior, not evidence of usable native Back layout. |
| Adaptation | No fixed alert width; native phone/tablet presentation remains pending. |
| Typography | Short title/message and two action labels; long editor content is outside the confirmation. Native clipping remains unverified. |
| Appearance | System alert uses semantic destructive/cancel styles. Actual contrast and appearance require runtime evidence. |
| Localization | English strings; localized action expansion and RTL behavior remain unverified. |
| Imagery | N/A: confirmation has no image or icon semantics. |
| Targets | System action buttons; no source-only target-size claim. |
| Gestures | Dirty customization editor disables route gestures and supplies a Back action. Native swipe, tab change and Android Back need runtime review. |
| Keyboard | Confirmation does not clear the draft; focus/keyboard restoration after Keep Editing remains a native check. |
| Accessibility | Native alert supplies title, message and named actions. Focus announcement/return has not been observed. |
| Motion | No custom confirmation animation. System Reduce Motion behavior remains pending. |
| Content | Warning describes unsaved changes; provider warning describes a replacement. |
| Search | N/A: exit confirmation has no search. |
| Loading | Save/lifecycle progress belongs to the editor. Provider guard explicitly blocks saving; customization suppresses dirty guard during save and relies on disabled controls. Native navigation during save needs separate verification. |
| Recovery | Keep Editing retains the draft. M155 preserves drafts when stale Discard is accepted after blur/return/resource replacement. |
| Editing | Successful save bypasses discard prompt; current Discard dispatches once. Regression tests cover these transitions. |
| Privacy | Confirmation does not mutate remote settings or grant access. Underlying editors enforce scope/permissions separately. |
| Notifications | N/A: no system notification behavior. |
| Media | N/A: no media acquisition or presentation. |
| Lifecycle | Focus/resource ownership now guards customization Discard as it already guards asynchronous save/lifecycle presentation. Physical background interruption and native alert lifetime remain pending. |

Both stale blur/return cases failed before M155. The replacement-resource case
also verifies that accepting the old alert cannot discard the new draft. These are
mounted/source checks, not native confirmation screenshots or gesture acceptance.

All52 customization tests pass remotely; TypeScript and the mobile structural
check passed for the implementation. Code critic found no confirmed scoped
regression. Existing test act warnings remain recorded rather than hidden.
