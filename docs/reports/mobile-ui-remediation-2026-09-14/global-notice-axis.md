# Global notice review

Reviewed at `ab1861d6`, normal-text remediation priority. The
[consumer inventory](global-notice-consumers.csv) enumerates all 38 production
`.showNotice()` call sites in 14 source files under the mobile client. Test and
runner fixtures are excluded. The inventory records call sites and task families;
it is not a claim that each callback has native acceptance coverage.

AppFeedbackProvider owns the active notice outside AppServicesProviderInner and
its native Stack. AppNotice uses absolute positioning with only safe-area top plus
small spacing. It has no active header geometry or reserved content space. That
explains why a notice can occupy the navigation region. Sharing now avoids that
path, but the shared positioning defect remains M103.

Success notices are not all disposable decoration: Add offers View, asset detail
offers Undo, and History/customization completion deliberately navigates back.
Retain these actions and successful-operation identity in a shared handoff. Local
failure and incomplete-list notices can instead be owned by the affected form or
collection. Do not move everything into transient component-local state, replace
all notices with blocking alerts, or add a guessed constant header offset.
A shared placement repair must account for the active navigation/header and modal
content bounds, keyboard, Home/Browse chrome, and action reachability.

M104 is separate: AppFeedbackProvider survives transitions between ready services
and onboarding, while its ActiveNotice contains arbitrary action closures. Sign-out,
server change and session expiry update only the inner services state. There is no
notice reset/ownership boundary, so an already-visible notice can survive them.
This source finding does not establish that an old command can bypass server
permissions. Acceptance must prove both notice removal and invalidation of captured
actions; hiding text alone is insufficient.

## All 24 axes for S128

| Axis | Source conclusion and remaining evidence |
| --- | --- |
| Task | Shared nonmodal status/Undo/View has a purpose; 38 call sites include local recovery candidates. Review each migration by task. |
| Navigation | M103: root absolute layer does not account for native header; navigation success handoffs must remain supported. |
| Selection | N/A: no value selection; the optional action is a command. |
| Modality | Notice is nonmodal; explicit destructive/auth decisions use the separate native dialog surface. |
| Layout | M103: safe-area top alone cannot exclude navigation; no content-space reservation or native-header measurement. |
| Adaptation | Width follows root margins; source switches action to a column at fontScale 1.3. Phone/tablet/modal window behavior remains pending. |
| Typography | Existing source evidence retained; fixed title/message line heights and longest messages still need native assessment. Enlarged-text remediation follows normal-size findings. |
| Appearance | Semantic tone surfaces/text use the appearance palette. Current normal-text light/dark contrast remains unverified. |
| Localization | English dismiss/action strings; message trimming and joining are source-reviewed. Locale/RTL display needs native verification. |
| Imagery | Tone dot supplements written title/message; there is no image acquisition or essential icon-only action. |
| Targets | Existing 48-point body/action minima retained. Actual bounds and competing native chrome need verification. |
| Gestures | Tap-to-dismiss and upward swipe; explicit action remains available. Interrupted gesture restores position. Native hit competition pending. |
| Keyboard | Top overlay is independent of keyboard height, but that does not prove visible action/focus or correct modal placement. Pending native keyboard checks. |
| Accessibility | Polite live region, iOS announcement, labels, persistent actionable messages; existing source/runtime limits retained. No whole-surface VoiceOver/TalkBack certification. |
| Motion | Shared Reduce Motion subscription, still initial state, explicit dismissal/recovery branch; existing evidence retained. |
| Content | One active notice; replacement removes its predecessor. No collection or queue. Callers must preserve durable operation history. |
| Search | N/A: no search/refinement interaction in this feedback primitive. |
| Loading | Status is caller-owned; plain messages can expire, actions/errors/warnings persist. Does not itself control network loading. |
| Recovery | Earlier named fixture proved Retry persistence in one state; preserve that partial evidence, not a general placement pass. |
| Editing | Undo/View are injected commands, not fields. Dismissal occurs before callback; domain operation identity and failures remain caller responsibilities. |
| Privacy | M104: active notice/closure is not scoped to service/account transition. No authorization-bypass claim. |
| Notifications | N/A: this is in-app feedback, not OS push or notification settings. |
| Media | N/A: no capture/upload/playback here; callers may report media results. |
| Lifecycle | M104: timer/replacement cleanup exists, but authentication/connection transitions do not reset the provider. |

## Acceptance before closing findings

Exercise a normal-text notice under Home, Browse, a pushed detail screen, a form
sheet, and the keyboard. Back, primary commands, Retry/Undo and dismissal must stay
reachable without unintended navigation. Preserve separate native captures for
phone and iPad. Trigger sign-out/server change/session expiry with a persistent
notice and a delayed callback; old content/actions must not survive into the new
context. Retain legitimate same-session Add View and Edit Undo behavior.
