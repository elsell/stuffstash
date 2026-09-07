# Mobile Navigation And Visual Polish Spec

## Purpose

Define the native iOS interaction contract for the Stuff Stash mobile product
after the initial tracer-bullet and voice slices. This spec resolves the
boundary between navigation, temporary tasks, global Voice interaction, and
gesture-driven containment exploration.

## Decisions

- Home and Browse remain the only primary tab destinations.
- Voice remains a prominent persistent bottom accessory. Its visual prominence
  is intentional, but it is a global interaction layer rather than a product
  destination.
- Voice may present one native detent sheet over the current surface. Voice
  must never push an asset, Settings, provider, or other product route while
  its sheet is presented. A linked result must dismiss Voice first and then
  navigate from the underlying surface.
- Map remains a Browse sub-surface with horizontal containment columns,
  breadcrumbs, branch gestures, and progressive path navigation. The gesture
  model is product behavior, not disposable prototype behavior.
- Map must not embed a full asset workspace in a modal. Selecting an asset's
  information action dismisses any transient Map selection surface and opens
  the shared asset workspace as a native stack destination.
- Asset detail, child assets, parent assets, Places, History, Settings, and
  provider-profile detail are hierarchical content and must use native stack
  destinations with standard back navigation.
- Add, Edit, Move, Move Here, Filters, Tenant Switcher, and other bounded
  commit-or-cancel tasks may use native `formSheet` routes. They must not use a
  custom React Native `Modal` for their primary presentation.
- Full-screen photo viewing is the sole full-screen modal exception. It must
  not contain product navigation or another sheet.
- Native platform menus are required for overflow, sort, and short contextual
  action lists. Alerts are reserved for destructive confirmation, permission
  blockers, and authentication/session loss.
- In-component `Modal` is prohibited for product navigation and bounded task
  surfaces. Any exception must be named in this spec and covered by a focused
  structural test.

## Presentation matrix

| Surface | Presentation | May push product routes? | Exception |
| --- | --- | --- | --- |
| Voice | Native detent sheet | No | None |
| Map asset selection | Shared asset stack route | Yes, through normal stack | No embedded workspace |
| Asset detail | Native stack route | Yes | None |
| Add/Edit/Move | Native form sheet | No; finish or dismiss first | None |
| Filters/Switcher | Native form sheet | No | None |
| History/Settings | Native stack route | Yes | None |
| Photo viewer | Full-screen modal | No | Image viewing only |

## Voice behavior

- The Voice accessory remains the strongest secondary control in the bottom
  navigation area and may use a large action target, contextual state, and
  progress indicator.
- Prominence must come from size, placement, and state—not from making Voice a
  competing tab or leaving the tab bar visually ambiguous.
- Opening a grounded asset reference from Voice must preserve the completed
  session state for return, dismiss the Voice sheet, and open the shared asset
  route from the underlying navigation context.
- Opening Voice setup from a recovery state must dismiss Voice before opening
  Settings.
- Parent selection during Voice review must be a native task surface that
  returns to the same review state. It must not create a modal stack.

## Map behavior

- Branch swipes, horizontal column paging, breadcrumbs, and reduced-motion
  behavior remain supported.
- The selected branch and search result must be visually distinct from an
  asset selected for detail.
- Map asset detail uses the shared asset workspace. Map may provide a compact
  native preview only if it contains no edit, move, history, checkout, photo,
  or nested asset navigation controls.
- Map gestures must not steal vertical scrolling from the active column and
  must provide an accessible button path for every gesture-only action.
- Search may expand to a matched path, but must also expose the matched asset
  and its current containment path in user-facing copy.

## Visual and accessibility contract

- Native system typography and Dynamic Type remain the default. Fixed widths,
  forced one-line truncation, and fixed card heights must not hide title,
  placement, status, or primary action information.
- The primary hierarchy is title, placement, state, then supporting metadata.
  Tags and technical classification are subordinate.
- A location path must use a compact readable summary at the current surface;
  it must not auto-scroll away the leading context in a way that hides the
  location root.
- Add and edit forms must keep their primary commit action visible above the
  keyboard and must use keyboard-aware scrolling to the focused field.
- Every interactive control must have a 44-point minimum target and a
  non-color selected, disabled, loading, and error state.
- Every transient surface must provide an explicit title, dismissal action,
  safe-area handling, focus transfer, and restoration of focus/context on
  dismissal.
- User-facing copy must use domain language (`Name`, `Location`, `Container`,
  `Item`, and `Place`) and must not expose future-work or implementation
  limitation messages.

## Enforcement

- Structural tests must reject product `Modal` usage in navigation and screen
  files except the named full-screen photo viewer and the explicitly reviewed
  system picker exceptions.
- Structural tests must reject Voice route pushes to product destinations and
  Map detail modals containing `AssetDetailView`.
- Mounted behavior tests must prove Voice dismissal before linked navigation,
  Map-to-asset stack navigation, keyboard-safe Add parent selection, and
  preserved Map gesture path state.
- Visual verification must cover regular and inline Voice accessory placement,
  compact and expanded sheets, keyboard-open forms, large Dynamic Type, dark
  appearance, and the Map gesture states on a real native build in CI or on a
  device. This repository must not require a local native build on a
  disk-constrained development host.
