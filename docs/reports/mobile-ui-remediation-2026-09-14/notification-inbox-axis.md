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
| Layout | Automatically inset ScrollView, padded rows and trailing read action reserve 48 points. Long title/date/path overlap requires actual captures. |
| Adaptation | Wrapping title and vertical sections, no fixed screen height. Tablet density and narrow-window row fit remain open. |
| Typography | Read/unread weights differ with an additional marker and spoken state. Date follows calendar precision. Large-text work follows normal-size fixes. |
| Appearance | Semantic palette and native header/segment/recovery controls. M187 replaces the custom read-state accessory with SwiftUI/Compose buttons. Contrast and disabled rendering remain unverified. |
| Localization | Shared expiration formatter preserves month/day precision. English labels and date sentences still need locale/RTL review. |
| Imagery | Envelope/open-envelope signal the read command; breadcrumb text names locations. No asset photos. Native symbol consistency remains M187. |
| Targets | Read action reserves 48-point bounds and uses an actual native accessory in M187; declarations do not prove hit geometry. |
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
assistive behavior and physical push acceptance remain open.

M187 candidate uses a borderless SwiftUI icon button and Compose IconButton with
separate item-open behavior, a full spoken action and disabled callback guard.
Twenty focused checks plus TypeScript and structural validation pass on paul
(`/tmp/native-read-button-green.log`). The adapter check exercises iOS symbol
reversal and disabled activation; it does not verify native rendering or Android
assistive grouping. Current-build native geometry and traversal remain pending.
Code review caught an Android wrapper whose accessibility activation was iOS-only.
The candidate now places the content description on the Compose icon inside its
native button, matching the existing conversation adapter. Follow-up review found
no blocker; TypeScript and structural checks pass after the correction.

M242 follow-up closes the retained-action source gap: four mounted regressions
reproduced offscreen writes/navigation, and a fifth protects first loading after
an unfocused mount. Current focused commands still execute; retained callbacks
after teardown do not. All24 inbox checks, TypeScript and structural validation
pass on paul. Native interruption and current-build geometry remain open.

## Android native acceptance — September 16

Pixel6/API36 emulator,1080×2400 at420dpi, font1, light appearance. APK SHA256
`31457c4de8c7b78759039732b3b7adc9849a83daee44d448e5f7ba6b5f52be00` comes from
the disposable Android tree with M242 production screen and this fixture applied;
it is not a complete HEAD build. Initial build failed because that historical tree
lacked ManagedSearchPlacementFixture; copying the required fixture resolved the build.

Actual taps verified read/unread changes; breadcrumb navigation without changing
read state; item opening with read synchronization; settings navigation and return;
mark-all followed by empty Unread; and a denied mutation hiding all private rows
until an explicit successful Retry. The clickable read accessory parent measured
126×126 pixels (48dp), separate from the item target. The long title wraps and the
month/date/path content remains visible in the inspected initial capture.

Controlled repository denial tests client recovery, not backend access enforcement.
Fixture destination labels prove the resolved target, not production destination
content. This does not establish TalkBack traversal, dark appearance, tablet fit,
physical APNs behavior, interrupted in-flight completion or iOS acceptance.
The corresponding Swift journey is added but uncompiled/unexecuted here. Review
corrected asynchronous state waits before mark-all and Unread activation.

Evidence: [initial](evidence/android-inbox-initial.png),
[denied](evidence/android-inbox-denied.png),
[empty unread](evidence/android-inbox-unread-empty.png).
The script `/tmp/verify-inbox-native.py`, all four XML/capture pairs and build log
remain on paul. Six fixture installer checks, TypeScript and structural validation
pass; code critic has no remaining confirmed blocker.

### Compact dark comparison

The same APK also passes the complete interaction/recovery sequence at an Android
window override of840×1867 pixels,420dpi (320dp wide), normal font1, system dark.
The long title wraps to three lines without colliding with the 48dp read accessory.
The immediate-parent chip remains visible; the earlier ancestor initially extends
outside the horizontal trail viewport. Do not label that initial capture as showing
the entire path. Width and appearance are restored after the comparison.
[Compact entry](evidence/android-inbox-compact-dark.png) and
[compact denial](evidence/android-inbox-compact-denied.png) retain visual evidence.
This adds compact Android/dark coverage, not iOS, tablet or assistive verification.

A separate horizontal swipe reveals the earlier Hall medicine closet chip fully,
and an actual tap resolves `closet`. See
[revealed ancestor](evidence/android-inbox-ancestor-revealed.png). The follow-up
script is `/tmp/verify-inbox-trail.py`; the full compact sequence is
`/tmp/verify-inbox-compact.py`, with `/tmp/inbox-compact-dark.log` on paul.
Original1080×2400 size and light mode were explicitly confirmed restored.
