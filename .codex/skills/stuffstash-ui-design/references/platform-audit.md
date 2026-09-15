# Whole-product interaction audit

## Establish coverage

Record revision/date, shipped clients/OS scope, and available runtime access. Read
route registries and conditional route dispatchers, then shared controls and nested
states. Include onboarding/auth, deep links, sheets, notifications, media, voice,
settings/admin, empty/denied/error states, and documentation/navigation surfaces.
A catch-all web route is not one screen. A filter sheet with seven internal pages
is not one state. Separate entrypoint inventory from deeper interaction evidence.

Map the current Apple HIG hierarchy, including Getting Started/platform guidance,
Foundations, Patterns, Components, Inputs, and Technologies. Each topic needs:
applicable / conditional / not applicable; a product-specific reason; linked
surface families; and evidence or an explicit coverage gap. Follow nested category
pages so component groups are not mistaken for leaf guidance. Record inaccessible
sources and do not assert their contents. Use authoritative page links; Apple
DocC JSON can supply content when the HTML page requires JavaScript.

Do not add missing optional features to satisfy the index. Apple Maps guidance does
not automatically apply to Stuff Stash's containment “Map.” Do not apply visionOS,
watchOS, or macOS-only controls to iOS. Platform-specific guidance is not universal.

## Apply the skill

First evaluate task/pattern fit using `platform-interaction-decisions.md`. Then
inspect accessibility/state/layout implementation. Read relevant source context;
search matches nominate findings but do not establish them. Look for counterevidence
(existing native adapters, draft persistence, focus recovery, intentional specs).

For every surface family record:

- Entrypoints and nested tasks; shared consumers.
- Chosen interaction and whether it fits.
- Source inspection result and linked findings.
- Runtime evidence level and pending scenarios.

Evidence labels: **source-confirmed**, **runtime-observed** (name build/device),
**recommendation** (judgment, not a prohibition), **unverified risk**, **N/A**.
A source-confirmed pattern choice is not a source-confirmed rendering defect.
Historical screenshots do not establish current-build failures.

## Findings and acceptance

Use stable IDs. State priority, actual behavior, user cost, source location,
platform guidance, smallest appropriate correction, affected consumers, and a
reproducible acceptance scenario. Prioritize inability to complete tasks or loss
of work, then repeated friction and accessibility, then polish. Never manufacture
a precise compliance percentage from source-only inspection.

Test contrasting cases: short exclusive filter versus long searchable locations;
boolean preference versus destructive command; bounded sheet versus editor;
local save versus background refresh. A good audit preserves appropriate patterns
as well as replacing inappropriate ones.

Native acceptance matrix: phone narrow/standard, iPad supported window sizes,
light/dark, enlarged text, reduced motion/transparency, VoiceOver, keyboard,
RTL/localized long content where supported. Exercise entry/back/tab return,
scrolling, searching, dismissing, empty/error/loading/denied, and interrupted saves.
Web: keyboard/focus, browser Back/deep links, zoom/reflow, screen reader, dialogs,
responsive touch and pointer. Android: native controls, TalkBack, system Back and
keyboard/edge-to-edge behavior. Use synthetic data; avoid changing user inventories.

## Report and durable follow-through

Store the concise main report and linked coverage appendices under `docs/reports/`.
Keep source-grounded details out of product UI copy. Separate confirmed backlog,
intentional decisions, and required live checks. Record the scope actually audited,
what is only inventoried, and what could not be exercised.

Update roadmap when this becomes the product focus. Do not silently fix or release
the audited UI under an audit-only request. Validate skill metadata/references and
ask the required code critic to check findings and policy overreach. A skill update
needs a realistic application, not a wording snapshot test.

## Native accessibility instrumentation

Use [Apple's XCTest accessibility audits](https://developer.apple.com/documentation/accessibility/performing-accessibility-audits-for-your-app)
for named visible states when the runtime supports them. Check hit regions,
descriptions, traits, contrast, Dynamic Type support and clipping; retain issue
artifacts. Never blanket-ignore issues to obtain a green run. Record specific
false positives with evidence before filtering one. These checks supplement
pattern review and assistive-technology use; a viewport audit cannot certify
unvisited controls, other configurations, or the whole application.
