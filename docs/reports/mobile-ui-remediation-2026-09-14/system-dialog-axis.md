# System dialogs

S129, source 53e3915c plus M189. This is the source review of the app's native
alert pattern and its inventoried callers, using confirmation-review.md and the
finding-specific reports. It does not certify all modal sheets or system prompts.

Additional caller review: AppServicesFeedbackGate's session-expired dialog is an
acknowledgement, with no mutation on Continue. Its composition visit retires once;
old composition errors cannot present a fresh prompt. Customization lifecycle
confirmation captures focus/resource ownership and a workflow that separates
confirmation, mutation and completion. Cancel/dismiss resets only confirmation;
pending mutation blocks repeated acceptance. Provider Archive already consumes
acceptance once and checks current presentation. The shared AppFeedback adapter
only forwards native alert actions; those guarantees belong to callers.

M189 found that photo-removal errors checked resource lifetime but not focused
visit. Two mounted cases reproduced a late alert after blur and after blur/refocus.
The candidate uses the existing visit predicate only for error presentation;
completion and lock cleanup still run, and a fresh attempt on return remains usable.

| Axis | Source result and remaining acceptance |
| --- | --- |
| Task | Destructive confirmation, discard, acknowledgement and Android photo choice are different tasks; reviewed separately in the caller inventory. |
| Navigation | Alerts retain the underlying route. Mutation/discard callbacks own any later navigation. M189 prevents departed error presentation. |
| Selection | Android source choice is selection; destructive actions are commands. No shared alert is used as the short settings value picker. |
| Modality | System Alert owns presentation. An acknowledgement is retained for explicitly requested operations that cannot open or finish; no new modal route introduced. |
| Layout | Native alert lays out title, message and actions; actual truncation/long names and screen occlusion still need captures. |
| Adaptation | Platform layout owns phone/tablet behavior. iPad action-sheet anchoring is separately pending in photo-source review. |
| Typography | System alert typography; long user-provided names and large text need native checks. |
| Appearance | Native alert materials/action styles. Destructive and cancel semantics specified at individual callers; contrast remains runtime work. |
| Localization | English copy with inserted resource names. RTL and translated action fit unverified. |
| Imagery | N/A: current Alert calls have no custom image content. |
| Targets | System action buttons; native hit-test evidence required, especially the known invitation cancellation failure. |
| Gestures | Explicit buttons exist. Android dismissal is caller-specific; customization onDismiss only cancels confirmation. |
| Keyboard | No alert text entry. Focus/keyboard restoration to underlying dirty editors still requires native observation. |
| Accessibility | System semantics plus named actions; no assumption that native implies verified reading order or focus restoration. |
| Motion | System transitions; no authored animation. Reduced motion acceptance pending. |
| Content | Destructive messages identify resource/consequence; source review distinguishes permanent deletion, reversal, archive and draft loss. |
| Search | N/A: not a search surface. |
| Loading | Alerts obtain consent before pending commands; screens own pending locks, not duplicated alert spinners. |
| Recovery | Acknowledgements preserve underlying task/retry. M189 rejects late photo errors while releasing pending state. Native retry journeys remain open. |
| Editing | Discard guards are visit/resource-bound in reviewed editors. Keep Editing retains drafts; this is not proof of native dismissal behavior. |
| Privacy | Resource/account ownership belongs to each caller. AppFeedback is not an authorization boundary or generic session guard. Backend security is outside this presentation review. |
| Notifications | Push-open failure is an acknowledgement with abort ownership; physical cold/warm push evidence remains separate. |
| Media | Photo-source chooser and removal consent/error have separate lifetime checks. Physical camera/library behavior remains unverified. |
| Lifecycle | Caller-specific visit/resource and single-use guards are recorded in confirmation-call-sites.csv. M189 corrects the newly observed source gap; native interruption remains pending. |

Current-build native alert activation, cancellation, focus return and interruption
are still required. Source coverage does not turn historical native failures green.
All 101 related detail/photo checks, TypeScript and structural validation pass on
paul (`/tmp/photo-removal-visit-green.log`). Code critic found no blocker.
