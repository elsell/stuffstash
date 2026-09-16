# Voice proposal editing — 24 source axes

S122, source baseline4b988ab0 plus M173 candidate. Reviewed proposal composition,
EditablePlanCommandFields, ParentPicker, VoicePlanEdits, provider approval,
controller validation/decision lock and photo staging. This is source review;
native keyboard, selection and approval geometry remain unverified.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Review proposed inventory changes before explicit approval. Creation names and destinations are editable; other commands stay descriptive. |
| Navigation | Name edits stay inline. Hierarchical destination lookup justifies a selection surface; M174 addresses its current custom inline panel. |
| Selection | Root, earlier proposed commands and existing assets are distinct destination choices. Disabled existing parents preserve reasons. Current selection lacks a checkmark (M174). |
| Modality | Proposal stays in conversation. ParentPicker is a conditional flex sibling, not native sheet/stack navigation (M174). |
| Layout | Scrolling proposal plus native bottom decision controls. Parent panel competes for the same vertical area; geometry must be tested after M174. |
| Adaptation | Flexible name text and wrapping descriptions; fixed inline command widths and parent header need native window checks. |
| Typography | Plain native text/input, bounded200-character name. Long summaries, risks and names require native acceptance. |
| Appearance | Native Approve/Cancel and staged-photo commands. Inline name Save/Cancel still use custom icon Pressables (M175). |
| Localization | Labels are English; name normalization preserves Unicode characters while collapsing whitespace. Long translations/RTL remain open. |
| Imagery | Local photo previews retain original URIs; numbered remove commands. Symbols distinguish editing and destination selection. |
| Targets | Inline name commands reserve36×44 points with no hitSlop (M175); native geometry unverified. |
| Gestures | Explicit edit, save, cancel and approve commands. No drag-only requirement. Parent rows currently use navigation chevrons for value selection (M174). |
| Keyboard | Name autofocus/Done commits; keyboard avoidance wraps conversation. M173 includes the visible pending name on Approve. Parent search uses a custom text input (M174). |
| Accessibility | Named edit/location/photo commands and disabled parent semantics. Parent state indication and inline name command geometry require corrections/acceptance. |
| Motion | No explicit proposal animation; parent appearance and keyboard transitions need native acceptance. |
| Content | Summary, ordered commands, placement, expiration and risks are shown before approval. Proposed IDs are not treated as saved asset links. |
| Search | Parent lookup debounces and isolates query results. Loading/error/no-match state is discarded by the screen (M174). |
| Loading | Pending review decisions suppress Approve/Cancel through presentation; actual execution progress is separate. |
| Recovery | Validation returns to review; submission failures retain staged drafts. M173 stores the visible name before attempt; M174 covers lookup recovery. |
| Editing | Inline drafts live above the sheet. M173 prevents Approve silently using an older committed name; blank visible names block submission. |
| Privacy | Parent lookup uses scoped server query; approval travels through existing transport validation. This review does not replace authorization boundary tests. |
| Notifications | No proposal-owned notification control; interruption handling belongs to lifecycle acceptance. |
| Media | Photos attach after approval; failed proposals retain read-only previews. Source chooser callbacks are visit/plan-owned; physical camera remains unverified. |
| Lifecycle | Plan identity/status changes clear local selector and replace drafts; controller locks duplicate decisions. Native dismissal/return with edited proposal remains open. |

M173 has mounted provider/controller evidence for blank rejection, latest name,
preserved placement/other edits and failed-submission draft retention. Source
validation does not prove native action visibility, reading order or safe-area fit.
