# Notification inbox — 24-axis source review

R022, source17f9e2b7 plus M186. Inspected route scope loading, inbox component,
NotificationInboxQueries, HTTP repository/error mapping and mounted tests. Root
push delivery and reminder setup remain in [notification ownership](notifications-axis.md).

| Axis | Source findings and native acceptance still required |
| --- | --- |
| Task | Inbox reviews delivered reminders; expiration workspace owns all dated items. Opening resolves the currently authorized item and marks the reminder read. |
| Navigation | Native Notifications header, settings destination, item and breadcrumb navigation. M55 owns delayed-open navigation; native return/interruption remains open. |
| Selection | All/Unread is a native segmented choice. Read state is a separate per-row action, not card selection. |
| Modality | Standard pushed inbox and destination routes. No extra confirmation for reversible read-state changes. |
| Layout | Automatically inset ScrollView, padded rows and trailing read action reserve44 points. Long title/date/path overlap requires actual captures. |
| Adaptation | Wrapping title and vertical sections, no fixed screen height. Tablet density and narrow-window row fit remain open. |
| Typography | Read/unread weights differ with an additional marker and spoken state. Date follows calendar precision. Large-text work follows normal-size fixes. |
| Appearance | Semantic palette and native header/segment/recovery controls. M187 tracks the remaining custom read-state accessory. Contrast and disabled rendering remain unverified. |
| Localization | Shared expiration formatter preserves month/day precision. English labels and date sentences still need locale/RTL review. |
| Imagery | Envelope/open-envelope signal the read command; breadcrumb text names locations. No asset photos. Native symbol consistency remains M187. |
| Targets | Read action declares44-point minimum bounds; these do not prove hit geometry. M187 will replace its custom implementation with an actual native accessory. |
| Gestures | Native scroll and explicit pull; no swipe-only operation. Read state, paging and settings remain explicit commands. |
| Keyboard | N/A: no text entry in the inbox. |
| Accessibility | Item labels include expiration and read state; independent named read command, error alert and native filter are declared. Native traversal/focus/announcements remain open. |
| Motion | No custom animation. Pull and initial progress indicators need reduced-motion review. |
| Content | Paginated rows merge by ID; sparse transport pages are traversed by the query. Empty All/Unread messages require an exhausted successful read. ScrollView rendering is not virtualized; large retained-page performance is unverified. |
| Search | N/A: All/Unread refines this delivered-reminder feed; item discovery is Browse's task. |
| Loading | Initial progress is labeled; a single operation lock disables competing mutations. Explicit pull owns its indicator, not background loading. |
| Recovery | Safe error plus native Retry; ordinary failed refresh keeps useful rows. M186 clears denied rows/cursor/read markers and prevents a false empty-success message. |
| Editing | No text draft. Open marks read; read/unread and mark-all refresh current filter and reconcile count. No optimistic claim on failed commands. |
| Privacy | M186 covers401/403 through the real generated HTTP client, adapter and application query with synthetic fetch responses. Denial hides titles and pagination through a subsequent503; fresh200 list restores data. This verifies the mobile response boundary, not backend authorization decisions. |
| Notifications | This is in-app read state; physical APNs delivery, cold push and badges remain root/device acceptance work. Inbox reads do not establish delivered-push correctness. |
| Media | N/A: no acquisition/playback. |
| Lifecycle | Route keys by session/tenant/inventory; teardown aborts work, focus session guards delayed open navigation. Retained other action callbacks and native interruption still need follow-up. |

Eight M186 RED cases reproduced stale private titles after401/403 during refresh,
open, read-state or mark-all. All31 focused checks and static validation pass on
paul (`/tmp/inbox-access-reviewed.log`); critic found no blocker. Native geometry,
assistive behavior and physical push acceptance remain open. M187 is not fixed yet.
