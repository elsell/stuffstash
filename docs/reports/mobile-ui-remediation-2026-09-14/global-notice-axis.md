# Global notice review

Reviewed at `ab1861d6`, normal-text remediation priority. The
[consumer inventory](global-notice-consumers.csv) enumerates all 36 production
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
| Task | Shared nonmodal status/Undo/View has a purpose; 36 call sites include local recovery candidates. Review each migration by task. |
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

## M104 implementation candidate

The production services gate owns the state and passes its ready serviceScopeId
(or disconnected state) into AppFeedbackProvider. A fresh notice-owner token per
context hides old content during render and invalidates old publishers and action
callbacks during layout cleanup. Returning to an earlier identity creates a new
owner; it cannot revive old publishers. Same-context navigation retains notices.
The passive state cleanup preserves a new notice published by a child's layout
effect. Native dialogs keep a stable separate callback so the existing services
initialization dependencies do not change with notice ownership.

Three mounted feedback regressions cover scope transitions, retained old callbacks,
same-context navigation, unmount and new-owner publication. The first two failed
before implementation. All 1,538 mobile tests, TypeScript and structural checks
passed remotely; the layout-effect variant also has a targeted rerun.

The actual gate and feedback wiring are now extracted behind injected onboarding,
profile-store and composition-factory dependencies. AppServicesProvider supplies a
stable production runtime. Five mounted integration cases use the real
OnboardingCommand with controlled ports, covering sign-out, server change, expiry,
reconnection, and rejected push cleanup. Old notices/actions disappear on completed
transitions; failed cleanup preserves the current session and action. Notice updates
do not repeat startup or composition construction. The targeted 18-test set,
TypeScript and structural checks passed remotely. Critic found no extraction
regression. Native sign-out/server-change visibility remains open before treating
M104 as verified. The change does not repair M103 placement or modify API
permissions.

M106 removes the two provider-editor failure publications in favor of field-local recovery; the current inventory contains 36 production sites. Native acceptance remains pending.

M103 now has focused native push/sheet regression journeys. Their whole-notice
bounds and native Back/Close assertions are pending execution. This preparation
does not claim a placement fix. Apple's [feedback guidance](https://developer.apple.com/design/human-interface-guidelines/feedback)
favors feedback integrated into the interface; its [alerts guidance](https://developer.apple.com/design/human-interface-guidelines/alerts)
helps distinguish interruptions from nonmodal status. React Navigation's
[native stack documentation](https://reactnavigation.org/docs/native-stack-navigator/)
distinguishes transparent headers that overlap content, so one root safe-area
offset cannot serve every presentation. These sources inform the candidate
design; they do not establish a specific implementation as correct.

## M103 active-screen placement candidate

Notice data, action ownership, expiry and one-time announcement stay in the shared
service-scoped provider. Ready-app native stacks render through a focused leaf
presenter; the root tab container delegates to Home/Browse stacks. Presenters are
siblings of existing scroll content, avoiding the extra flex wrapper implicated
by sheet diagnostics. Transparent headers use the navigator's measured height,
regular-header content uses its reserved area, and headerless content uses safe
area insets. Onboarding retains root presentation outside the ready-app stack.

Mounted regressions cover a single presenter, focus handoff, live header-height
changes, retained Undo, and disconnected-root/ready-screen transitions through
sign-out, expiry and reconnection. Review caught restarted plain-notice expiry
and repeated announcement/entry animation on handoff; a failing regression now
passes after keeping lifetime shared. Rehosting preserves the original deadline.
All 1,573 mobile tests across 260 files, TypeScript and mobile structural checks
passed on paul (`/tmp/pr150-notice-full.log`). Critic re-review found no remaining
confirmed source blocker.

This is not native geometry proof. Push/sheet containment, native header context,
Home/Browse transparent-header behavior, keyboard coexistence, direct-scroll
compatibility and Android live-region behavior remain runtime acceptance. M103
stays open; do not call every notice consumer visually verified from mounted tests.
