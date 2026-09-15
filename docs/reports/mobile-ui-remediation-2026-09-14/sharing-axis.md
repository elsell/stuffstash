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
