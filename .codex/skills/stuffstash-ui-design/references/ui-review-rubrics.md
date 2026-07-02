# UI Review Rubrics

Use these rubrics as independent adversarial review passes. Report concrete issues, user impact, and the smallest practical fix.

## Clarity And Cognitive Load

Inspired by "do not make users think" usability principles.

Check:

- The primary action is obvious within five seconds.
- Labels use user language from Stuff Stash specs.
- The page answers: where am I, what can I do, what changed, what happens next.
- Users can recover from a wrong click without reading instructions.
- Similar objects have visibly similar behavior.
- The interface avoids clever naming for ordinary actions.

Flag:

- Ambiguous labels.
- Hidden state.
- Competing primary actions.
- Navigation that requires memorizing where things live.
- Copy that explains the UI instead of making the UI obvious.
- Product-surface text that talks about the interface as if it were a pitch or instruction manual.

## Usability Heuristics

Check:

- System status is visible after loading, saving, failing, and completing.
- Controls and copy are consistent across states.
- State-changing actions prevent common mistakes.
- Users can cancel, undo, or back out where risk justifies it.
- Recognition beats recall.
- Error messages say what failed, what changed, and what to do next.
- Dense data stays scannable.

Flag:

- Missing loading, empty, error, denied, or saved states.
- Irreversible actions without adequate friction.
- Repeated patterns that behave differently.
- Controls that look available but are disabled without explanation.

## Accessibility And Inclusive Interaction

Use WCAG 2.2 as the baseline.

Check:

- Keyboard access reaches every interactive element in a sensible order.
- Focus indicators are visible and not color-only.
- Touch targets are at least 44 by 44 CSS pixels where practical.
- Contrast is sufficient for text, icons that convey meaning, focus rings, and boundaries.
- Forms have labels, descriptions, error messages, and programmatic associations.
- Dialogs and menus manage focus correctly.
- Motion is unnecessary for comprehension and respects reduced-motion preferences.
- Layout does not require hover.
- Screen-reader names match visible labels.

Flag:

- Color-only status.
- Placeholder-only labels.
- Truncated text that hides meaning.
- Small icon buttons without accessible names.
- Responsive layouts that reorder content incoherently.

## Mobile Ergonomics

Check:

- Primary actions are reachable and thumb-friendly.
- Navigation works without hover, side-by-side comparisons, or tiny targets.
- Forms minimize typing and expose good defaults.
- Dense lists remain scannable in one column.
- Critical context is not lost when controls collapse.
- Destructive actions are hard to trigger accidentally.
- Long titles, room names, and custom field values wrap cleanly.

Flag:

- Desktop-first tables squeezed onto mobile.
- Sticky elements that obscure content.
- Modals that feel cramped on small screens.
- Required sidebars for core workflows.

## Information Architecture

Check:

- Inventory, location, container, item, custom field, sharing, audit, and media language stays consistent with specs.
- Browse, search, and conversational entry points do not compete confusingly.
- Resource hierarchy is clear without exposing implementation details.
- Grouping matches how a homeowner thinks about finding and managing household items.
- Permissions and sharing concepts are visible when they affect available actions.

Flag:

- Backend terminology leaking into UI.
- Multiple names for the same concept.
- Navigation organized by implementation layer instead of user task.
- Search and browse results that obscure why an item is visible.

## Visual System

Check:

- The UI feels clean, cool, clever, fast, trustworthy, native, and useful.
- The palette uses system grays and whites with restrained brand accents.
- Primary actions are blue/system-like.
- Photos and user content carry warmth.
- Spacing, type scale, icon style, and component states are consistent.
- Cards are used for repeated items or framed tools, not nested section decoration.
- Text fits containers at mobile and desktop sizes.

Flag:

- Green SaaS styling.
- Purple-blue AI gradients.
- Beige, rustic, or warehouse palettes.
- Decorative blobs, sparkles, or mascot-like behavior.
- Oversized hero or marketing composition in product workflows.

## Implementation Feasibility

Check:

- Generic controls can map to shadcn-svelte primitives.
- Product-specific composition stays separate from UI primitives.
- API DTOs can remain behind adapters.
- Runtime configuration boundaries remain intact.
- State shape is testable with fakes.
- Large lists and media surfaces have a performance strategy.
- Candidate code does not imply unsupported backend behavior.

Flag:

- UI requiring domain behavior not present in specs.
- Global client state where route-local state is enough.
- Hard-coded URLs, tenant IDs, auth assumptions, or environment details.
- Patterns that make accessibility depend on future refactors.
- Route or component files that are absorbing unrelated adapter, mapper, config, fixture, observability, or primitive responsibilities.

## Frontend Engineering Discipline

Check:

- Domain concepts are represented with frontend types, enums, literal unions, or value objects where behavior depends on them.
- UI components do not depend directly on generated API DTOs when an adapter or frontend domain model should own translation.
- Auth, runtime configuration, API access, and observability remain behind explicit helpers or ports.
- Environment-specific URLs, tenant IDs, inventory IDs, OIDC values, feature flags, and provider settings are not hard-coded into UI logic.
- Production paths do not use raw `console.log`, `print`, `println`, or ad hoc diagnostics.
- Duplication is removed when it protects behavior, but shared abstractions are not created from accidental visual similarity.
- Files are cohesive and split by route, product component, UI primitive, adapter, mapper, domain type, config, mock-data, observability, or test-helper responsibility.

Flag:

- Magic strings for asset kinds, lifecycle states, roles, route modes, or invitation statuses.
- Candidate code that would require bypassing adapters when promoted.
- Mock IDs leaking into product logic.
- Observability expressed as implementation details instead of domain events.
- "God" files that mix unrelated concerns or keep growing because there is no clear ownership boundary.
