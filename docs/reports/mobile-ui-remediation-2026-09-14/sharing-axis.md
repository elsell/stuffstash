# Sharing interaction and recovery review

## Android direct-command verification — September16

APK SHA256 `ab28e3ef2f0d1e13053ba9b0a109e264cf9102c5e7eab7ba11395ec50cb19015`
on paul's Pixel6 Android16/API36 emulator, font scale1, light theme, includes the
1a373d0b Sharing change on the existing selectively patched audit build. It is not
a whole-HEAD or production-service acceptance run.

The production Sharing screen with controlled invitation ports verifies exact email
entry, unavailable-link recovery, a visible native Cancel invitation command,
recipient-named confirmation, Keep Invitation retaining the pending row, intentional
cancellation failure and a subsequent successful retry. The final row reads Cancelled
and has no cancellation command. Captures show no keyboard over these actions.
An initial rapid coordinate sequence acted before modal presentation settled; it
was discarded and the fixture restarted. The retained failure/retry sequence waits
for native hierarchy observations between taps.

- [Failure with reachable retry](evidence/android-sharing-direct-failure.png)
- [Cancelled row and stale guidance](evidence/android-sharing-direct-complete.png)

This confirms Android normal-size behavior only. M193's phone/iPad keyboard issue
remains open pending the queued iOS revision. TalkBack, enlarged text, landscape,
multi-row native disambiguation and real server cancellation are not established.

M239 is newly runtime-confirmed: after cancellation succeeds, the retained creation
notice still tells the user to cancel the invitation below. The same source stores
creationError independently of cancellation. Recovery guidance must reflect the
completed prerequisite without clearing unrelated or newer failures. Fix and native
reverification remain outstanding.

## Native follow-up: M193/M194

Phone run35042066124 shows the keyboard overlapping invitation cancellation
recovery and an ellipsis trigger measuring 21.7 by 6.7 points. Retained evidence:
[screenshot](sharing-keyboard-350420.png) and [hierarchy](sharing-keyboard-350420.txt).
Creation now dismisses editing before sending, preserves failed email drafts,
and the iOS ellipsis label owns a 44-point native frame/content shape. The shared
Asset overflow consumer is included in review; label/sort-icon triggers are
unchanged. Twenty-eight Sharing/menu checks plus TypeScript and structural checks
pass remotely on paul. An additional 18 Asset detail/overflow checks pass there.
Critic found no blocker. Native assertions now require
44-point button bounds and keyboard absence before cancellation. Actual iPhone
and iPad acceptance remains pending; do not infer it from these source checks.

Reviewed September 15 against source after PR138, covering R048 Sharing and a
partial comparison with R019/S131 invitation acceptance/link entry. This is source
and controlled-render evidence, not native accessibility or security certification.

Sharing's two-option Access menu fits an in-place exclusive choice. The email
editor and access choice lock during creation, and a failed request retains the
draft. Destructive cancellation uses an explicit confirmation naming the invitee.
Copy and system Share are distinct commands; the one-time link explains its lifetime.
List pagination remains explicit. These choices should be preserved.

Cached invitation metadata is hidden on access denial. The separate route guard
explains missing permission, while the list's own denial fallback is more generic;
copy consistency remains a review candidate, not a reason to weaken the guard.
Created secrets are component-local and scope-bound, and are absent from query
cache in existing behavior checks. This inspection does not replace API adversarial
coverage or prove all lifecycle paths secure.

## M69 — late Sharing feedback outlives its screen

Source-confirmed P2: creation and cancellation catch blocks and link-action
completion displayed global feedback without verifying their initiating screen.
Existing scope checks protected the successful creation result but not errors.
A delayed failure could therefore describe the previous inventory after switching
inventory, leaving, or returning to the route. Three rendered regressions reproduced
these cases before the correction.

Feedback now captures both scope and focused-session identity. Returning creates a
new identity, so old feedback stays suppressed. Normal focused errors still show
and retain the invitation draft. The authorized command and cache invalidation
still complete; this is presentation ownership, not permission enforcement.

Apple's [Feedback guidance](https://developer.apple.com/design/human-interface-guidelines/feedback)
describes communicating status through the interface. Binding a notice to its
initiating context is this project's engineering interpretation of relevant,
understandable feedback; Apple does not prescribe this React hook implementation.

Critic review found no blocker; requested copy/share/cancel departure coverage and
normal focused error coverage were added before final verification.

Remote evidence: 16 sharing, guard and application checks, TypeScript and structural
checks passed on paul. Native route blur/return and system-share cancellation remain
pending, as do enlarged text, keyboard reachability, VoiceOver/TalkBack and RTL.
The acceptance screen uses request generations and separates successful acceptance
from failure to open an inventory. Its remaining actions and cold/warm link races
need their own complete audit; they are not cleared by this Sharing correction.

## M70 — old invitation recovery overwrites a replacement

Source and rendered failure confirmed (P2), covering R019/S131 recovery, editing
and lifecycle. Preview and acceptance use a request generation, but opening and
start-over recovery originally did not. A late open failure could replace a newer
invitation with the old inventory's opening error. A pending start-over also left
the new invitation disabled. Two regression scenarios reproduced these failures.

Opening and start-over now use the existing generation guard; a replacement resets
its start-over availability. All 13 invitation screen/progress/onboarding checks,
TypeScript and structural checks pass remotely on paul. The existing current-open
failure scenario still exposes retry and explains that acceptance succeeded.
Native simultaneous link delivery/navigation remains pending. This correction
covers screen state only: it does not cancel an authorized callback, roll back
account changes, or certify the route callback's later navigation side effects.

Further entry review: the link provider protects a foreground link from a late
initial URL result. Its initial lookup has no rejection handler; failure readiness
needs a controlled test before choosing recovery semantics. Existing provider tests
replace React hooks and therefore do not establish real mounted lifecycle behavior.
These are explicit outstanding review gaps, not passing coverage.

## M71 — initial-link lookup rejection leaves initialization unfinished

The preceding initial-lookup gap is now reproduced and corrected. Extracting the
existing lifecycle into a mounted hook produced one failed readiness case and two
unhandled rejection errors. The hook accepts a link-source interface, and the
production provider retains Expo Linking plus runtime origin configuration.

Rejected lookup now finishes readiness unless a foreground invitation already won.
Disposal removes the listener and ignores late lookup/event completions. Existing
route filtering and trusted parser remain unchanged. The old test that replaced
React hooks was replaced by mounted tests with a controlled source fake, preserving
foreground-versus-initial and null-result coverage. All ten selected hook/domain
checks, TypeScript and structural checks pass remotely on paul. This verifies
controlled lifecycle behavior, not OS universal-link delivery or complete account
isolation. Native cold/warm link and account-transition scenarios remain pending.

Critic follow-through: clearing now invalidates a pending initial lookup, preventing
a dismissed invitation from reappearing. The added regression failed before the
generation guard and passes afterward; later foreground links still work.

## M72 — route completion clears a replacement invitation

The route-side completion gap from M70 is now addressed: successful selection or
start-over callbacks previously cleared the link and navigated Home unconditionally.
A newer invitation could therefore disappear even though screen-state errors were
guarded. The extracted original callbacks failed both mounted replacement tests.

The route now delegates to a focused invitation-session action hook. A late
completion cannot clear/navigate after replacement, blur/refocus or unmount.
Current-session success still clears and returns Home. Rejections propagate to the
screen recovery state; authorized account/inventory side effects are not reversed.
Critic found no blocker and requested unmount/rejection cases, which were added.
Seven selected route/progress/selection checks, typecheck and structural checks
pass remotely. Native link delivery during selection and navigation transitions
still require verification; this is controlled lifecycle evidence for R019/S131.

## Complete R048 source pass at0041dfad

Reviewed the route, access guard, complete Sharing screen and current22 behavior
cases. The table covers outgoing invitation management only; incoming acceptance
and universal links above remain separate surfaces and acceptance requirements.

| Axis | Source result and remaining evidence |
| --- | --- |
| Task | Create by email with viewer/editor access, then copy or share the one-time link. Current context names the inventory. Invitation receipt/delivery expectations need end-to-end user verification. |
| Navigation | Settings route wraps an access guard and retains native parent navigation. Leaving during requests is covered by controlled ownership tests; actual Back/system-share return is pending. |
| Selection | Two access values use the shared native in-place picker. Creation freezes access/email. Native selection appearance and traversal remain pending. |
| Modality | Cancellation uses a native destructive confirmation with the invitee named. Copy and system Share stay separate. Native sheet anchoring/dismissal remains pending. |
| Layout | Settings sections scroll with shared padding; link actions stack. No fixed footer is introduced. Header/inset and keyboard clearance need native confirmation. |
| Adaptation | Row text flexes beside menus. iPad screenshots show the email above the keyboard; that single state does not establish all width/rotation combinations. |
| Typography | Shared styles plus13–17pt field/link/metadata sizes. Long emails and links need normal-size wrapping verification before enlarged-text work. |
| Appearance | Palette supplies foreground, background and separators. System commands own materials. Dark/light disabled command contrast remains pending. |
| Localization | Expiry dates use device locale; labels/status capitalization remain English. Invalid date text is shown unchanged. RTL and translated long strings remain pending. |
| Imagery | No asset imagery is part of invitation management. Native menu icon identifies cancellation. Actual symbol alignment remains pending. |
| Targets | Email minimum48pt; rows minimum68pt; native commands/menu supply actions. Native target and footer/keyboard reachability is not proved by source sizes. |
| Gestures | Native Back, explicit cancellation choices and named commands avoid gesture-only completion. System Share cancellation and return remain pending. |
| Keyboard | M121 gives iOS email a stable native seed with explicit resets and recovery. Android remains controlled. Observed truncation is still open until a native rerun passes. |
| Accessibility | Email has a label; operation errors/progress use live regions and alerts; complete link is selectable/labeled. Screen-reader ordering, duplicate announcements and native picker states require runtime checks. |
| Motion | No new custom animation. System Share/alert transitions and reduced motion remain pending. |
| Content | Safe invitation pages are retained; explicit Load older is disabled while fetching. Pending nonexpired entries offer cancellation. Large histories and repeated page boundaries need performance acceptance. |
| Search | No search task is exposed here. Current invitation list is paginated; whether large histories require search is a product-discovery question, not a missing button finding. |
| Loading | Guard explains access checking; list initial state shows a spinner. Create/copy/share/cancel/pagination expose busy state. Initial list spinner labeling and native announcement remain review work. |
| Recovery | Guard distinguishes permission loss from failed verification. Screen-level denial uses generic loading-error copy. M121 preserves email through denied-form recovery; link-unavailable recovery refreshes safe metadata. Native retry focus remains pending. |
| Editing | Failed creation retains email/access; success clears email; scope changes clear both. M121 tests native seed/reset contracts and same-scope recovery. Native autofill/composition and interrupted editing remain pending. |
| Privacy | Access denial hides cached metadata and one-time link. Created links stay component-local and scope-bound, outside retained query cache. System clipboard/share handoff and API authorization need their own runtime/security evidence. |
| Notifications | No push delivery is promised or configured by this screen. Receiving an invitation/universal link is a distinct surface; notification interruption during editing remains pending. |
| Media | No camera/file/audio capture task lives here. System sharing exports the invitation link; destination behavior is platform-owned and unverified here. |
| Lifecycle | Feedback and confirmation ownership use scope/focused visit; pending command locks persist until settlement. Success data remains scope-bound. Native background/foreground and account changes need verification. |

Source coverage does not mean acceptance. Full1,618-test validation atcd245a20
includes the current Sharing candidate; no screenshot yet proves its typing fix.
