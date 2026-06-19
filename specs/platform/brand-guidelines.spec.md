# Brand Guidelines Spec

## Purpose

Stuff Stash needs brand guidance before client applications, design tokens, app icons, marketing pages, empty states, error states, and conversational inventory surfaces are implemented.

This spec defines the durable brand direction for Stuff Stash.

The brand must make the product feel clean, cool, clever, fast, trustworthy, native, and useful without drifting into generic AI-app styling, enterprise inventory software, or a high-school project.

## Scope

This spec covers:

- Brand principles.
- Brand voice.
- Visual direction.
- Color and typography direction.
- Accessibility expectations.
- Internationalization and localization expectations.
- Icon and imagery direction.
- Product-surface implications that should inform future client specs.

This spec does not define:

- Final web navigation.
- Final mobile navigation.
- Screen-level interaction design.
- Component APIs.
- Design token file formats.
- App store metadata.
- Marketing positioning outside the product experience.

When this spec and a future client UX spec overlap, this spec owns brand intent and the client UX spec owns screen structure and behavior.

## Source Research

The initial direction came from brand interviews and product mockup review.

The following external standards and practices inform this spec:

- WCAG 2.2 is the accessibility baseline for digital content and interactions.
- W3C internationalization guidance treats internationalization as design and development that allows later adaptation across language, culture, and region.
- Mature design systems generally define semantic color and typography roles before raw values.
- Platform-native client design should use platform conventions where they improve speed, accessibility, and user trust.

## Brand Position

Stuff Stash is a low-friction home inventory system for a tech-savvy homeowner.

The first loyal user is a homeowner with technology, clothes, medicine, documents, garage supplies, tools, and other household items scattered across real places.

The product should help that user:

- Add assets quickly.
- Find assets confidently.
- Move assets with little friction.
- Use voice interaction without losing control.
- Trust that edits are saved, reversible, and auditable.

The brand should feel like a sharp personal tool, not a warehouse system, novelty AI product, smart-home gimmick, or generic productivity template.

## Core Brand Adjectives

The brand must be:

- Clean.
- Cool.
- Clever.

Clean means visually orderly, uncluttered, and easy to scan.

Cool means modern, composed, native-feeling, and technically capable without showing off.

Clever means the product removes work through smart defaults and voice interaction without performing cleverness for its own sake.

The brand must not be:

- Messy.
- Confusing.
- Clunky.

The brand must also avoid:

- Enterprise inventory aesthetics.
- AI sparkle, magic gradients, or assistant gimmickry.
- Student-project inconsistency.
- Cute mascot-like behavior.
- Decorative friction.
- Copy that hides the real system state.

## Brand Principles

### Low-Friction Confidence

Every brand choice must help users act quickly while feeling sure they are acting on the right asset.

Brand expression must never slow down:

- Adding assets.
- Editing assets.
- Moving assets.
- Voice interaction.
- Reviewing and approving a proposed action.

Decorative motion, elaborate empty states, verbose confirmations, and clever names for common actions are not allowed when they increase task time.

### Preview Before Action

Conversational inventory earns trust by clearly showing what action it will take before it applies a meaningful change.

When speech or natural language produces a proposed state-changing action, the user-facing surface must show:

- The action.
- The target asset.
- The relevant source state, when applicable.
- The intended destination or new state, when applicable.
- Whether anything has been saved yet.
- A clear approval path.
- A clear cancellation path.

Meaningful state-changing voice actions must ask for approval before execution unless a future conversational UX spec explicitly defines a safe exception.

### Safe Saved State

After an edit, saved state must be obvious.

The product should make users feel that speed is safe by making undo easy to find and easy to use.

Saved-state communication may use different patterns by risk, such as:

- Inline saved status.
- A brief saved toast.
- A timestamp.
- A checkmark.
- An undo bar.
- A persistent history entry.

Future UX specs must define exact behavior by surface and action risk.

### Photos Carry Warmth

The user's own inventory photos are the primary expressive layer.

The UI should frame inventory photos clearly and avoid competing with them through decorative illustration, heavy color, or visual noise.

Photo-first browsing is a strong brand direction, but final screen layout belongs in future client UX specs.

### Native Calm

Mobile should feel almost fully iOS-native where platform and React Native support allow it.

Native platform conventions are preferred when they improve:

- Speed.
- Accessibility.
- Trust.
- Gesture familiarity.
- System integration.

Custom UI should be used only when it provides a clear Stuff Stash product advantage, such as the voice affordance, asset photo treatment, action preview, empty states, error states, or app icon.

### Care In Small Details

The brand earns maturity through consistency.

The following must be treated as brand-quality concerns, not incidental polish:

- Spacing rhythm.
- Typography scale.
- Icon style.
- Color roles.
- Touch target size.
- Focus states.
- Error language.
- Empty-state language.
- Motion timing.
- Saved-state feedback.
- Undo placement.
- Photo cropping behavior.

Inconsistent small details are a brand defect because they make the product feel clunky or unfinished.

## Voice And Tone

Stuff Stash copy should be GitHub-precise.

That means copy should be:

- Direct.
- Specific.
- Accurate.
- Calm.
- Actionable.
- Short unless more detail is needed.

Copy should not be:

- Vague.
- Cute during task flow.
- Over-coached.
- Marketing-heavy.
- AI-magical.
- Needlessly apologetic.
- Technically obscure by default.

### Error Voice

Error states must explain:

- What failed.
- Why it failed, if known.
- What changed, if anything.
- What did not change, if that matters.
- The next best action.

Example direction:

- Good: `Move not saved. Fertilizer is still on Garage shelf. Choose a confirmed destination and try again.`
- Bad: `Something went wrong.`
- Bad: `Oops, Stuff Stash got confused.`

### Empty-State Voice

Empty states may be lightly witty when they remain useful and actionable.

Empty states should include a next action.

Example direction:

- Good: `This shelf is suspiciously tidy. Add the first item here.`
- Good: `No stuff here yet. Add an item or move existing stuff into this location.`
- Bad: `Nothing to see here.`
- Bad: `Your stuff adventure begins!`

### Technical Detail

The default voice should be user-friendly.

Technical details should be available when useful, especially for tech-savvy users, but they should not dominate the default surface.

Good candidates for optional technical detail include:

- Authorization denial.
- Model uncertainty.
- Sync or save failure.
- Import/export results.
- Audit history.
- Data portability.

## Visual Direction

The visual system should be:

- System-native.
- Content-first.
- Restrained.
- Photo-forward.
- Spacious enough to feel calm.
- Dense enough to support repeated work.
- Bespoke in key details without fighting the platform.

The visual system should avoid:

- A green inventory/SaaS theme.
- Purple-blue AI-gradient dominance.
- Heavy beige, brown, or rustic palettes.
- Decorative blobs, orbs, or bokeh.
- Literal warehouse or logistics visuals.
- Literal Homebox-like iconography.

## Color Direction

Color must be defined by semantic roles before raw values.

The first color direction is:

- System grays and whites for most surfaces.
- A familiar system-like blue for primary actions, links, and interactive emphasis.
- Semantic status colors only when they communicate real status.
- Warmth used sparingly in details, not as the main theme.
- User inventory photos as the main source of visual variety.

The brand must avoid the earlier green direction because it reads as generic AI/SaaS styling.

Color roles to define in a future token spec include:

- Page background.
- Surface.
- Elevated surface.
- Text strong.
- Text muted.
- Border.
- Primary action.
- Primary action hover/pressed/disabled.
- Link.
- Focus ring.
- Success.
- Warning.
- Danger.
- Informational.
- Selected state.
- Voice active state.
- Saved state.
- Undo state.

Raw token values must be tested against real product screens, item photos, light mode, dark mode, and accessibility requirements before they are considered stable.

## Typography Direction

Client typography should prefer platform-native system fonts by default.

For Apple platforms and iOS-like mockups, the direction is SF-style system typography through the platform font stack.

Typography should optimize for:

- Scan speed.
- Legibility.
- Clear hierarchy.
- Native feel.
- Localized string expansion.
- Dynamic type or equivalent text scaling.

Typography must not rely on:

- Narrow containers that break with longer translations.
- Font sizes that only work in English.
- Negative letter spacing.
- Decorative display type in task surfaces.

Monospace typography should be limited to:

- Code.
- IDs.
- Logs.
- Optional technical details.

## Icon And App Icon Direction

The app icon should signal a container.

The direction is an abstract iOS-style cardboard box glyph.

The icon must:

- Read clearly at small iOS icon sizes.
- Avoid copying Homebox or other home-inventory products.
- Avoid literal house imagery.
- Avoid rustic brown dominance.
- Avoid detailed realistic cardboard.
- Feel clean, cool, and clever.

The product icon system should use a consistent stroke, optical size, and corner language.

Icons should clarify actions, not decorate labels.

## Motion Direction

Motion should be subtle, native-feeling, and task-supportive.

Motion may be used for:

- Confirming a saved state.
- Revealing an action preview.
- Showing undo availability.
- Moving between native-feeling surfaces.
- Supporting voice interaction state.

Motion must not:

- Delay adding, editing, moving, or approving assets.
- Make users wait for decorative transitions.
- Reduce clarity.
- Trigger discomfort when reduced motion is enabled.

All motion must respect reduced-motion settings.

## Accessibility Requirements

Accessibility is a brand requirement from day 0.

Stuff Stash must target WCAG 2.2 AA for web content and equivalent platform accessibility expectations for mobile applications.

Accessibility requirements include:

- Text and interactive controls must meet WCAG contrast requirements.
- Information must not be communicated by color alone.
- Touch targets must be large enough for reliable one-thumb mobile use.
- Keyboard access must be available for all web interactions.
- Focus order must be logical.
- Focus indicators must be visible and brand-consistent.
- Screen reader labels must describe action and state accurately.
- Voice interaction must have non-voice alternatives.
- Audio feedback must have visual equivalents.
- Reduced motion settings must be respected.
- Text must support user text scaling and dynamic type where the platform supports it.
- Saved, unsaved, and undo states must be perceivable by assistive technology.
- Error states must identify the field or action affected and provide a recovery path.
- Empty states must not be the only way to discover core actions.
- Photo-first layouts must provide useful text alternatives and accessible names.
- Drag, gesture, or speech interactions must have accessible alternatives.

Accessibility testing must be included before user-facing client surfaces are considered complete.

Future client specs must define concrete accessibility verification for:

- Web keyboard flows.
- Screen reader announcements.
- Mobile dynamic type.
- Reduced motion.
- High contrast modes.
- Voice interaction approval.
- Undo flows.
- Error recovery.

## Internationalization And Localization Requirements

Internationalization is a brand requirement from day 0.

Stuff Stash must be designed so future localization does not require rewriting core UI, brand copy, or layout assumptions.

Requirements:

- User-facing strings must not be hard-coded directly in reusable client components once client implementation begins.
- Text must allow translation expansion.
- UI must not depend on English word length.
- Layout patterns, spacing tokens, icon placement, and navigation concepts must not block future right-to-left language support.
- Dates, times, numbers, units, and list formatting must use locale-aware formatting.
- Sorting and search behavior must be designed with locale-aware collation in mind.
- Pluralization must use localization-aware plural rules.
- Gendered or culturally specific language should be avoided unless needed.
- Idioms and jokes must be rare and easy to localize.
- Empty-state wit must remain translatable and actionable.
- Error copy must be clear enough to translate without losing recovery instructions.
- Voice interaction copy must account for locale-specific grammar and phrasing.
- Images and icons must avoid culture-specific assumptions where possible.
- Pseudolocalization should be used before shipping broad client UI.

The brand should remain recognizable after localization through structure, clarity, icon quality, and interaction patterns, not through English-only wordplay.

## Design Token Expectations

Future design tokens should be shared where useful across web, mobile, and docs.

Tokens should be semantic and platform-adaptable.

Token categories should include:

- Color.
- Typography.
- Spacing.
- Radius.
- Elevation.
- Icon sizing.
- Motion.
- Focus.
- Touch target sizing.
- Photo treatment.

Tokens must support:

- Light mode.
- Dark mode.
- High contrast or increased contrast modes where supported.
- Reduced motion.
- Text scaling.
- Localization expansion.

Raw values should not leak into application UI outside token definitions except in prototypes or temporary experiments.

## Product-Surface Implications

The following decisions are brand-relevant but need future client UX specs before implementation:

- Mobile should prefer native tab bar conventions where supported.
- A central mobile voice affordance is a strong product direction.
- Web search should feel closer to Google Drive than Spotlight or Raycast.
- Search should show the closest asset match first and visually separate commands below.
- The main user-facing browse noun is provisionally `Stuff`.
- `Inventory` is reserved domain language because the domain hierarchy is `Tenant -> Inventory -> Asset`.
- A unified browsing surface is preferred for simplicity.
- Photo-first asset browsing is preferred, with row/list view as a possible option.
- Photo-first tiles should initially show photo, title, and location.
- Asset actions should live on the opened asset page or detail panel.
- Details, move, and edit are likely primary asset actions.
- Glass/material effects should be restrained and most appropriate for item view cards or transient approval surfaces.

These points must not be treated as final UI specifications until captured in client-specific specs.

## Verification

Brand implementation is not complete unless the affected surface is checked for:

- Alignment with the core adjectives: clean, cool, clever.
- Absence of forbidden qualities: messy, confusing, clunky.
- GitHub-precise copy.
- Accessibility against WCAG 2.2 AA or equivalent platform expectations.
- Localization readiness.
- Contrast in light and dark mode.
- Text scaling.
- Keyboard and screen reader behavior where applicable.
- Reduced-motion behavior.
- Photo-first clarity.
- Saved-state clarity.
- Undo discoverability.
- Voice action preview and approval behavior where applicable.

Mockups and prototypes should be used to test whether brand guidance transfers to real product surfaces before client implementation settles.

## Open Questions

- Is `Stuff` durable enough as the user-facing noun for assets, containers, places, documents, and shared household contexts?
- What exact design token format should the monorepo use?
- How should mobile implement the central voice affordance while staying native and accessible?
- What exact app icon composition best signals an abstract cardboard box without overlapping Homebox?
- What localization framework should each client use?
- What accessibility test stack should each client use?
