# Sharing interaction and recovery review

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
