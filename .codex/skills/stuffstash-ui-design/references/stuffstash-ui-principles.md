# Stuff Stash UI Principles

Use these constraints when shaping web UI direction.

## Product Feel

Stuff Stash should feel like a sharp personal tool for a technically adept homeowner. It should not feel like warehouse software, a generic admin panel, a novelty AI product, or a marketing site.

Prefer:

- Clean, calm, scan-friendly layouts.
- Fast task flow.
- Direct copy.
- Obvious saved state.
- Easy undo.
- Photos as the warm visual layer.
- Restrained brand color.

Avoid:

- Decorative AI styling.
- Mascot-like behavior.
- Beige, rustic, green SaaS, or purple-blue gradient palettes.
- Oversized hero sections for authenticated product workflows.
- Cute names for common actions.
- Copy that pitches the product, narrates the interface, or explains what the user can already see.

## Mental Model

Users think in household concepts:

- Inventories organize a household or collection boundary.
- Locations are user-facing places backed by location-like assets.
- Containers hold things.
- Items are the things people want to find, move, document, or share.
- Photos help recognition.
- Search, browse, and voice should all resolve to the same authorized domain actions.

Do not expose persistence, transport, SpiceDB, OIDC, generated DTOs, or audit implementation details in UI language.

Keep UI implementation aligned with the same domain boundaries:

- Product components use frontend domain language.
- API transport details stay behind adapters.
- Auth provider details stay behind auth helpers.
- Runtime values stay behind configuration helpers.
- Observability uses domain events instead of raw diagnostics.

## Interaction Defaults

Use these defaults unless the relevant spec says otherwise:

- Make browse and search available without forcing a setup explanation.
- Prefer affordances, hierarchy, labels, and state over explanatory product copy.
- Show state-changing conversational actions as previews before execution.
- Make destructive actions explicit and recoverable where possible.
- Prefer inline editing only when saved state and undo remain obvious.
- Prefer route-local state until a cross-route workflow proves shared state is necessary.
- Treat denied states as normal product states, not generic errors.

## Responsive Behavior

Design mobile and desktop together.

Mobile:

- Use one-column layouts for core task flow.
- Keep primary actions reachable.
- Avoid compressed desktop tables.
- Keep filters and secondary actions available without taking over the page.

Desktop:

- Use extra width for comparison, preview, filtering, and faster scanning.
- Do not fill wide screens with low-density decorative cards.
- Keep dense operational surfaces calm and organized.

## State Coverage

Design these states before implementation:

- Signed out.
- No inventory.
- Empty inventory.
- Active asset list.
- Archived asset list.
- Asset detail.
- Search with no results.
- Loading.
- Save success.
- Validation error.
- API failure.
- Permission denied.
- Undo available.
- Long or missing photos.
- Long titles and locations.
