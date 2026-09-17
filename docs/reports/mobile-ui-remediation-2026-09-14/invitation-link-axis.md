# System invitation entry — all 24 axes

S131, source checkpoint c090a391. Reviewed InventoryInvitationLinkContext,
useInvitationLinkSnapshot, PendingInventoryInvitation, InvitationLinkParser,
app configuration, AppServicesContent, invitations/accept, route actions and the
current InventoryInvitationScreen. This extends the separate R019 screen review
to external entry. It establishes no physical universal-link acceptance.

| Axis | Evidence and remaining work |
| --- | --- |
| Task | An external invitation opens a review task with explicit Join, followed by Open inventory. Receiving a URL does not itself accept membership. |
| Navigation | Expo receives initial and foreground URLs; the configured app scheme and HTTPS route lead to invitation review. Initialized route parameters are replaced with the bare route. Real OS handoff, account onboarding and return need runtime. |
| Selection | Access is granted by the sender and displayed as Viewer/Editor; no editable picker belongs in acceptance. |
| Modality | Review uses a root stack screen, appropriate for a consequential external task. Done/Not now clear and return Home. System Back behavior still needs native review. |
| Layout | Destination uses a flexible ScrollView and bounded card, without a fixed bottom action overlay. No external-entry-specific layout is introduced. Phone/iPad insets remain unverified. |
| Adaptation | Destination has narrow-width and enlarged-text branches. Normal-size captures and full onboarding-to-review composition remain missing. |
| Typography | Inventory and recovery text wrap. Long external names, expiry strings and clipping require native captures. |
| Appearance | Current destination uses theme tokens and NativeCommandButton. Earlier R019 discussion of custom command controls is historical M133 evidence, not current implementation. Light/dark/disabled appearance remains unverified. |
| Localization | Expiry formatting follows the device locale. Copy is English; RTL and translated labels remain outside verified evidence. |
| Imagery | Mail/success imagery supplements explicit review/result text; external link metadata does not introduce fetched artwork. Reading order remains unverified. |
| Targets | Current command adapter supplies platform controls. No physical target measurements are established by the parser or route tests. |
| Gestures | Explicit Join/Open/Done/Not now commands exist. OS Back/swipe dismissal and repeated activation need native acceptance. |
| Keyboard | Review has no input; a keyboard may be present during preceding onboarding/sign-in. Entry must be checked for dismissal and focus restoration. |
| Accessibility | Destination has named progress and a polite result region. Screen reader focus after external entry, error and account switching is unverified. |
| Motion | No custom external-entry animation; navigation and named progress are delegated to existing components. Reduced Motion remains a runtime check. |
| Content | Review exposes inventory, access and expiry. Invalid/expired/revoked/cancelled states explain why Join is unavailable. There is no list pagination. |
| Search | N/A: receiving and reviewing one invitation has no search task. |
| Loading | Initial URL lookup settles on rejection; later foreground delivery remains supported. Preview, Joining and Opening have named states. Six mounted link-source cases cover ordering/cleanup. |
| Recovery | Malformed links become safe invalid state; foreground delivery supersedes late initial lookup. Account mismatch and accepted-but-opening-failed states have distinct recovery. Real browser/app/sign-in transitions remain unverified. |
| Editing | No editable invitation draft. Clear invalidates a late initial result; route completion cannot clear a replacement invitation. Pending operation and system dismissal need runtime checks. |
| Privacy | Parser enforces trusted HTTPS origin or the app scheme, exact fields and bounded identifiers/token; opt-in private HTTP is constrained. Token remains in reference memory, not parser errors. Client adapter tests are not backend authorization or OS logging certification. |
| Notifications | N/A: push notification entry is S130; this surface receives invitation URLs. |
| Media | N/A: no media capture, upload or playback in link intake. |
| Lifecycle | Link subscription is removed on cleanup; foreground capture wins over late startup success/failure; clearing blocks resurrection. Route completion is focus/reference-owned. Physical cold/warm, suspended and signed-out entry remain missing. |

Remote validation on paul: 64 cases across parser, pending invitation, link-source
hook, route actions, API adapter, onboarding and progress suites pass. Log:
`/tmp/invitation-entry-audit.log`. No endpoint or authorization behavior changed.
These checks do not certify association-file delivery, platform link registration,
physical browser handoff, navigation restoration, or backend access isolation.

No additional confirmed product defect was established in this source pass.
Retained callback and rapid-repeat behavior at the full route boundary remains a
test gap; do not infer it from the current late-completion tests.
