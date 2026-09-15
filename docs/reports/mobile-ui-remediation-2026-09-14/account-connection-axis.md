# Account and connection recovery audit

September 15 source/controlled-render review covers root Settings (R035), Account
(R024) and Connection (R026), with partial comparison of About (R023) and Diagnostics
(R027). Native phone/iPad, accessibility and Android acceptance remain pending.

## Task and pattern fit

Account and server switching are commands with different consequences, not a
single value picker. Each belongs on its own settings destination. Existing native
confirmation alerts explain the retained server on sign-out and the local reset
on changing servers; neither claims to delete inventory data. Cancel is explicit.
Pending guards prevent repeated execution. Keep this separation and confirmation.

Root Settings is navigation. Appearance is already an in-place choice. Grouped
Account/Connection rows lead to tasks, while household/inventory rows depend on
scoped access. About and Connection read local diagnostics. Diagnostics combines
identity and scope and remains dependent on those reads; this is not needed for
sign-out recovery.

## M73 — inventory failure prevents account recovery (P1)

Account reused SettingsModelScreen, which waits for selected-inventory data. A
missing, denied or offline inventory therefore replaced Sign Out with a loading
or error view. Root Settings similarly withheld both Account and Connection until
scope was ready. Two mounted regressions reproduced unavailable recovery links
and blocked sign-out.

Root's loading/error states now retain Account and Stuff Stash server navigation.
Account reads only the principal label, falls back to Current account, and always
exposes confirmation for Sign Out. Principal failure offers retry without removing
the command. Existing authorization for inventory administration is unchanged.
This is task-dependency correction: leaving an account must not require its current
inventory to be readable. It is a project usability inference, not a quoted Apple
requirement about this architecture.

Evidence: 27 remote settings/cache checks, TypeScript and structural checks pass on
paul. Critic found no blocker and requested pending/failed identity cases; both now
verify fallback sign-out. Existing mounted checks retain scoped-denial behavior,
confirmation, pending/error recovery and cache ownership. This is not new proof of
server authorization or native confirmation rendering.

## Other axes and remaining scenarios

Settings rows use shared scalable text, minimum targets and stacked large-text
layout; inspect native narrow/iPad/long identity and URL values before accepting
geometry. Loading/error content scrolls. These screens have no media, search or
notification task; sharing native adapters elsewhere does not make those features
required here. Light/dark, contrast, RTL, VoiceOver/TalkBack order and keyboard
navigation still need runtime coverage. Native screen return after failed teardown
and focus ownership of delayed error notices remain open lifecycle checks.

The account command still delegates to the existing push-session disconnect and
onboarding teardown. Source review found performance disposal precedes asynchronous
profile reset; failure recovery of that resource lifetime needs a separate
engineering review. No implementation change or defect claim is made for it here.
