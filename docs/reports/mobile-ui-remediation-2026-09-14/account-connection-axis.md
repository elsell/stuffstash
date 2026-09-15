# Account and Connection settings review

R024/R026, source570b804c plus M141. Inspected route wiring, both screens,
SettingsQuery, SettingsList, SettingsRefreshNotice and shared styles. This is a
source review across all24 axes; runtime-sensitive cells remain pending.

| Axis | Source result and remaining evidence |
| --- | --- |
| Task | Account identifies the principal and signs out; Connection identifies the server and changes it. These are commands, not value selectors. Custom action-row rendering still needs native-adapter review (M142). |
| Navigation | Thin routes inject service actions; native stack owns Back. Confirmations retain their originating visit (M140). Actual Back/return remains pending. |
| Selection | No selectable setting or choice list. Displayed email and server address are read-only values. |
| Modality | Explicit native alerts explain sign-out/server-change consequences and provide Cancel. M140 rejects obsolete/reused acceptance. Native dismissal pending. |
| Layout | ScrollView contains grouped rows and explanatory footers; no fixed footer. Actual safe-area, bottom reachability and iPad widths pending. |
| Adaptation | Shared rows switch layout using font scale; native narrow/large-text/tablet evidence pending. Normal-size checks first. |
| Typography | Values shrink and are selectable. Long email/URL wrapping needs runtime verification; source rules alone do not close M33. |
| Appearance | Palette colors and shared disabled opacity. Native contrast/material appearance pending. |
| Localization | English prose; serverHostname derives display host while full URL remains available. Long strings, RTL and localization coverage pending. |
| Imagery | No screen-specific media or imagery; native navigation icons remain shared-shell work. |
| Targets | Action-row minimum52×44 is declared. Native bounds, focus and touch activation remain pending. |
| Gestures | No required hidden gesture. Explicit command and native Cancel/Back; interactive dismissal remains pending. |
| Keyboard | No editable field. Selectable values support copying; hardware focus and navigation remain pending. |
| Accessibility | Action rows declare button, busy and disabled; value rows have combined labels. Actual traversal, duplication and announcements pending. |
| Motion | No local animation; shared alert/notice transitions require platform checks. |
| Content | Small static groups. Account reads principal separately from inventory; Connection uses injected diagnostics. No pagination. |
| Search | Neither screen owns search or filtering. |
| Loading | Sign-out remains available when identity is pending/failed. Pending command prevents duplicate execution. Loading account identity falls back to Current account. |
| Recovery | M140 suppresses departed action errors and preserves retry. M141 corrects initial identity error copy; retained-data refresh copy stays unchanged. |
| Editing | No field draft. Single-use confirmations and current command eligibility are mounted-tested; successful session transition remains an integration gate. |
| Privacy | Service actions own session changes. Confirmation states what is kept/forgotten; no data-deletion promise beyond existing contract. No authorization rule is changed by M140/M141. |
| Notifications | No notification preference or push handler on these screens; interruption is shared lifecycle work. |
| Media | No acquisition, upload, playback or export action. |
| Lifecycle | Query/identity and focused visit own confirmation/error feedback. Cold start, system background and session transition still require native/integration evidence. |

M141 mounted recovery: initial principal read fails, accurate unavailable-details
copy is shown, sign-out remains present, retry loads principal and removes error.
61 settings tests, TypeScript and structural checks pass remotely. Native acceptance
is not inferred from these tests. The shared refresh notice's other callers keep
the same default message and behavior.
