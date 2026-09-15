# Onboarding interaction and recovery review

Source revision45e37013, September15. This review covers S058 server entry, S059
browser sign-in return, S060 household setup, S061 first inventory, and S062
partial setup recovery. It is not whole-flow native acceptance.

## Task and interaction fit

These are prerequisite connection/account tasks, not a tutorial or a collection
of settings. Server entry leads to system-browser sign-in; new accounts then name
a household and its first inventory in one form. Existing access can skip that
creation. Keep the short flow and contextual connection help; do not introduce a
picker for arbitrary server addresses or user-created names.

Apple recommends brief prerequisite onboarding, contextual instruction, and
postponing nonessential customization. These are design criteria, not a reason
to make authentication optional. [Apple onboarding guidance](https://developer.apple.com/design/human-interface-guidelines/onboarding).

The URL field has a visible label, URL keyboard, no autocorrection/capitalization,
and a native iOS TextField adapter. Household/inventory names use ordinary text
entry; Home Inventory is a reasonable editable default. Browser sign-in keeps
password entry out of the inventory form. Copy/paste, autofill, external keyboard,
VoiceOver focus and native return still need named runtime evidence.

## State, editing and recovery

OnboardingScreen blocks duplicate submissions with an immediate pending guard;
fields and commands lock while submitting. Errors are displayed in the form and
preserve drafts. Heading focus follows step changes. Source traits alone do not
prove the announcement order or native focus landing.

OnboardingCommand delegates authorized work through authentication, profile and
API ports. HouseholdSetup distinguishes definitive failures from ambiguous writes:
retry discovers results before attempting another creation. A household that was
created successfully is retained while its first inventory is retried. The mounted
screen tests cover this transition, retained inventory name, and no duplicate
household write. This review does not replace adversarial API/authentication tests.

M75 fixes reset callbacks after unmount; Connect/Create already had the equivalent
generation guard. Command cleanup is allowed to finish. In-place replacement,
process termination and browser interruption remain separate acceptance scenarios.
AppServicesContext guards startup completion after effect cleanup and retains the
invitation notice while setup is necessary. M71 separately covers initial-link
lookup rejection and clear ownership.

## Open decisions and native gates

Required names and URL are currently validated on submission, with field-specific
messages, rather than disabling an incomplete primary action. Apple's entering-data
guidance favors clear labels/defaults, timely validation and enabling progress once
required data is present. Review consistent required-field behavior across setup
before changing it; avoid silently disabling an action without an understandable
reason. This is a documented design gap requiring follow-through, not proof of
native breakage. [Apple entering-data guidance](https://developer.apple.com/design/human-interface-guidelines/entering-data).

M20 remains open: pre-candidate iPad runs34929746647 and34932076384 pass inside-form
dragging and landscape but fail margin keyboard dismissal. The full-width content
candidate in4bc12e5e awaits native results. M35 phone keyboard action visibility
must remain a separate gate. Broad dynamic type, dark mode, RTL, reduced motion,
TalkBack and VoiceOver coverage remains pending. Do not certify those axes from
the default light iPad screenshots.

Source files: OnboardingScreen.tsx, OnboardingPresentation.ts, both
OnboardingAddressInput adapters, AppServicesContext.tsx, OnboardingCommand.ts and
HouseholdSetup.ts. Remote mounted checks and evidence are recorded with M75 in
findings.md; full native lineage is in native-evidence.md. Apple HTML required
JavaScript; current official DocC JSON supplied the readable guidance.

Critic review found no report overclaims and requested reconciling older S058
editing evidence with later native address/help passes. The matrix now retains
that history and names the later passes without claiming Go submission acceptance.
