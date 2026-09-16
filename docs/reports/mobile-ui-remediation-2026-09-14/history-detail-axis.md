# History detail — 24-axis source review

R010, source953bb1ff plus M182. Inspected the route wrapper,
AssetHistoryDetailRouteScreen, AssetActivityQuery, AssetHistoryPresentation and
mounted cache/reversal cases. This is source evidence, not native acceptance.

The task is inspecting one substantial activity record and optionally reverting
its change. A pushed detail destination fits; technical metadata stays collapsed
in place. M182 aligns the readable actor fallback with History's list. This is a
project content decision. Apple's [Writing guidance](https://developer.apple.com/design/human-interface-guidelines/writing)
returned only a JavaScript shell during this follow-up; no new guideline wording
is asserted from that retrieval.

| Axis | Source decision and outstanding acceptance |
| --- | --- |
| Task | Shows action, exact time, actor and before/after changes. M182 removes opaque principal IDs from the primary actor fallback. |
| Navigation | Native stack destination with encoded route context. Missing entry offers Back to History. Cold-link return destination still needs native verification; router.back follows the entry stack. |
| Selection | No preference selection. Technical details toggle visibility in place, without pushing another route. |
| Modality | Revert opens native confirmation describing the outcome. Cancel preserves the record; accepted command remains on detail until completion. |
| Layout | ScrollView groups changes with bottom padding; no independent action overlay. Centered loading/error states use flex layout. Capture last-command reachability and safe areas at normal size. |
| Adaptation | Flexible vertical content; no explicit tablet width cap. Inspect long values on phone, iPad and split windows before claiming fit. |
| Typography | Headings and values can wrap; secondary styles use explicit line heights. No large-text acceptance inferred. |
| Appearance | Shared semantic palette and native command adapter. Contrast, disabled command rendering and clipping remain native checks. |
| Localization | Shared exact timestamp formatter uses device locale/time zone. Copy, arrow-delimited before/after values and source labels remain English; RTL needs verification. |
| Imagery | No content photos. Technical disclosure uses text plus/minus, with expanded state. Symbol appearance is not a native verification result. |
| Targets | Native Retry/Back/Revert; disclosure declares minimum height44. Actual hit bounds and last-action reachability remain open. |
| Gestures | Native scroll/Back; explicit buttons for all mutations. Verify swipe-back while reversal is pending. |
| Keyboard | N/A for entry: no editable fields. Technical values allow selection/copy; inspect native selection behavior. |
| Accessibility | Headings, error alerts, applied live region and disclosure expanded state are declared. Screen-reader order, confirmation focus and before/after speech remain unverified. |
| Motion | No custom motion. Native transitions and shared feedback require reduced-motion review. |
| Content | User changes precede collapsed allowlisted technical metadata. M182 preserves email when present and uses Someone with access otherwise. Detail consumes one record; direct-link query may scan scoped pages. |
| Search | N/A: inspect one identified event; discovery/filtering belongs to History list. |
| Loading | Loading activity text; cached scoped records can populate detail immediately. Failed background refresh retains data with explicit stale notice and Retry, while suppressing Revert. |
| Recovery | Distinguishes denied, missing, transport and stale-refresh states. Denial suppresses cached details. Recoverable failures retain Retry; reversal conflicts/denial retire unsafe Revert. |
| Editing | No text draft. Reversal is an explicit compensating command with confirmation, pending lock and applied state. Existing mounted cases cover retry and repeated acceptance. |
| Privacy | Scope includes session, tenant, inventory, asset and activity. Query allowlists technical keys; no secret editor or permission prompt. This review does not replace backend authorization tests or certify all value redaction. |
| Notifications | N/A: no OS notification entry or scheduling owned here. Shared notices are covered under recovery. |
| Media | N/A: no capture, picker or playback. Attachment metadata is read-only text. |
| Lifecycle | Focus-session and operation-scope ownership suppress stale completion navigation/notices. Scoped reads cancel when superseded. Mounted tests cover blur/return, changed record and teardown; cold start/background require native evidence. |

## Evidence and remaining work

Three mounted RED cases reproduce opaque IDs for absent/empty/whitespace email;
the contrasting real-email case passes before the fix. The candidate changes only
the detail fallback, preserving the existing trimmed-email path and audit identity.
All14 focused mounted/query cases, TypeScript and structural checks pass remotely
on paul (`/tmp/history-actor-green.log`); critic found no confirmed issues.
Run current native detail entry, disclosure, copy, unavailable/error/retry, reversal
cancel/success/conflict and navigation-return journeys. Source-reviewed cells do
not establish any of those native results.
