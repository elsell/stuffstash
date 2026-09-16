# Notification tap entry

S130, source abdbf5d5. Reviewed root installation, composition lifetime,
PushNotificationNavigation, ExpoPushNotificationResponses, OpenPushNotification,
NotificationInboxQueries and direct-APNs tests. This is a source review of entry
coordination, not a physical push test or a new notification settings review.

| Axis | Source result and evidence limits |
| --- | --- |
| Task | An explicit tap opens the current authorized asset referenced by the reminder. It does not trust a payload asset ID as a destination. |
| Navigation | Handler waits for root navigation readiness, resolves current item, selects its inventory and pushes asset detail. Actual back stack after cold/warm entry remains unverified. |
| Selection | No choice control; inventory selection follows validated reminder scope. Current-selection reconciliation has its own command/observer. |
| Modality | Successful entry pushes detail. Failed explicit entry presents a native acknowledgement; no additional selection sheet. |
| Layout | Entry coordinator renders no content. Destination and failure-alert geometry are owned by their surfaces and remain runtime checks. |
| Adaptation | Shared route resolution on phone/tablet; actual native stack and supported Android entry need verification. |
| Typography | N/A for this zero-content coordinator. Destination/alert typography is covered separately. |
| Appearance | N/A for coordinator rendering; notification appearance and destination are separate surfaces. |
| Localization | Safe failure copy is English. No arbitrary transport message is shown by the fallback; localized notification content remains separate. |
| Imagery | N/A: entry routing does not render artwork or icons. |
| Targets | N/A for coordinator: OS notification target and destination controls are outside its rendered content. |
| Gestures | User taps the OS notification. Back/swipe behavior after navigation needs native acceptance; no authored entry gesture. |
| Keyboard | Coordinator does not request keyboard focus. Entry over a dirty keyboard-visible editor needs runtime verification of navigation/draft preservation. |
| Accessibility | Native error acknowledgement has explicit title/message/OK. Destination focus and notification announcements need VoiceOver/TalkBack evidence. |
| Motion | No custom animation; navigation transition belongs to the native stack. Reduced-motion entry remains open. |
| Content | Routing requires nonempty bounded strings, valid server URL and matching principal. Direct APNs fallback is only used when ordinary data is absent. |
| Search | N/A: explicit notification identity resolution, not search. |
| Loading | Resolution is asynchronous with no coordinator-owned visual progress. Slow-network perception is an unverified UX risk; not claimed as a observed defect or a passing scenario. |
| Recovery | Typed failures use safe notification copy; unknown errors direct users to the inbox. Later deliberate taps can retry the same notification after handling settles. |
| Editing | Opening marks the currently resolved reminder read before inventory selection. Failed later selection may leave it read; it does not claim mutation rollback. Underlying dirty-editor preservation needs native acceptance. |
| Privacy | Server/principal validated before inbox access; server resolves current authorization and item. Malformed/wrong-recipient synthetic native payloads are rejected in tests. This is not proof of all backend authorization paths or OS lock-screen privacy. |
| Notifications | Launch response and live response share the adapter. New live response takes precedence over delayed launch read; concurrent duplicates are suppressed. Default action only. APNs delivery/badges remain physical acceptance gaps. |
| Media | N/A: routing does not acquire or play media. |
| Lifecycle | New response aborts previous navigation work; effect teardown aborts/unsubscribes. Shared response adapter retains launch consumption across subscriptions and ignores disposed callbacks. Cancellation cannot undo a server-side read already committed. |

Existing controlled tests cover direct APNs launch/live mapping, malformed/server/
principal rejection, content-data precedence, duplicate/repeated responses,
disposal, current-item resolution and cancellation before inventory selection.
They do not mount PushNotificationNavigation or exercise Expo Router's real stack.
All 18 checks across these four adapter/application files pass remotely on paul
(`/tmp/push-entry-abdbf5d5.log`).

Required native/physical scenarios: cold signed-in tap, warm tap from Home/Browse
and dirty editor, wrong account/server, expired session, inaccessible/removed item,
slow/failing network followed by repeat tap, newer tap overtaking older resolution,
session replacement during resolution, actual destination/back/focus, iPad and
Android equivalents. No whole-flow acceptance is asserted from adapter tests.
