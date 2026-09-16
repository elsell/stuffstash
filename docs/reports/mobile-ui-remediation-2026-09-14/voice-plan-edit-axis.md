# Voice proposal editing — 24 source axes

S122, source baseline4b988ab0 plus M173 candidate. Reviewed proposal composition,
EditablePlanCommandFields, ParentPicker, VoicePlanEdits, provider approval,
controller validation/decision lock and photo staging. This is source review;
native keyboard, selection and approval geometry remain unverified.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Review proposed inventory changes before explicit approval. Creation names and destinations are editable; other commands stay descriptive. |
| Navigation | Name edits stay inline. M174 opens hierarchical destination lookup in a native stack route; Back retains the draft. |
| Selection | M174 shows current destination, checkmarks, root, earlier proposed parents and existing candidates with disabled reasons. |
| Modality | Proposal stays in conversation; the M174 candidate replaces the flex sibling panel with a native stack route. Native sheet-to-stack transition remains unverified. |
| Layout | Scrolling proposal plus native bottom decision controls. Destination selection now owns a separate viewport; native geometry remains open. |
| Adaptation | Flexible name text and wrapping descriptions; fixed inline command widths and parent header need native window checks. |
| Typography | Plain native text/input, bounded200-character name. Long summaries, risks and names require native acceptance. |
| Appearance | Native Approve/Cancel and staged-photo commands. M175 candidate also uses native inline name Save/Cancel commands. |
| Localization | Labels are English; name normalization preserves Unicode characters while collapsing whitespace. Long translations/RTL remain open. |
| Imagery | Local photo previews retain original URIs; numbered remove commands. Symbols distinguish editing and destination selection. |
| Targets | M175 replaces36×44 custom name icons with shared native commands below the field; actual native geometry remains unverified. |
| Gestures | Explicit edit, save, cancel and approve commands. No drag-only requirement. M174 replaces parent navigation chevrons with checked selection rows. |
| Keyboard | Name autofocus/Done commits; keyboard avoidance wraps conversation. M173 includes the visible pending name on Approve. M174 moves parent search into native navigation search. |
| Accessibility | Named edit/location/photo commands and disabled parent semantics. Parent state indication and inline name command geometry require corrections/acceptance. |
| Motion | No explicit proposal animation; parent appearance and keyboard transitions need native acceptance. |
| Content | Summary, ordered commands, placement, expiration and risks are shown before approval. Proposed IDs are not treated as saved asset links. |
| Search | Parent lookup debounces and isolates query results. M174 preserves loading/error/no-match states and offers native Retry; search uses the native navigation field. |
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

M175 adds four mounted name editor cases (Save, Done, Cancel, blank). All46 focused
editor/approval/presentation tests and static checks pass on paul; code critic
review is complete. Name ownership remains above the sheet. Native acceptance
remains open. M174 now has a route candidate with eight choice/recovery/ownership
checks; the new R142 route has a separate [24-axis source review](voice-location-axis.md).
The native fixture now exercises the production proposal, destination route and
returned draft; simulator execution remains pending.
