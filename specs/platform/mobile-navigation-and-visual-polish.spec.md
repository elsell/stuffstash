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
- The Add form's parent picker must keep its search field visible while typing
  and provide an independently scrollable results area when the keyboard
  reduces the available sheet height. Creating a new parent must remain
  reachable without relying on content hidden behind the keyboard.
- The Add form must use the available form-sheet height intentionally: its
  header, context, photo affordance, name, parent, details, and commit action
  should read as one compact task rather than a sparse stack surrounded by
  unexplained blank space.
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

## Browse native search (2026-09-13)

- Browse has a nested native stack inside its existing tab, retaining the `/search`
  route and tab identity. A system navigation search bar serves both list and map;
  remove custom in-content search boxes and duplicate Browse titles.
- Use compact native navigation search (updated 2026-09-14), system
  clear/cancel and keyboard Search, no focus on entry. Keep list debounce and
  authorized result/filter/pagination semantics. Map search still finds an item
  and opens its containment path; update it after a short typing pause as well as
  explicit keyboard submission. Native callbacks must submit their event text.
- Carry current text across list/map switches, settling pending list search before
  the switch so delayed callbacks cannot restore the old surface. Existing list
  refinements and map path state remain intact. Clearing search clears map highlight
  without discarding the current map path. Restore route text in the native field.
- A shared native navigation-search adapter owns native clear/cancel/field sync;
  domain search behavior remains in the corresponding screen/application query.
  Do not add dependencies or change backend/security contracts.
- Validate field restoration, clear/submit semantics, switch races, map path
  behavior and existing Browse regressions remotely/CI; run critic before merge
  and follow the signed TestFlight workflow through upload.

### Compact Browse controls (2026-09-14)

- Browse list and map use native integrated-button search at the trailing edge
  of the navigation bar, expanding on interaction. Disable toolbar integration
  so iPhone does not move search into the bottom toolbar. iOS before 26 uses the
  library's native inline fallback; Android retains its native search action.
- Keep existing query restoration, clear/cancel, submit and debounce semantics.
  Search results remain described by the result summary when search is inactive.
- Remove the standalone Expiration pill from the result header: it is navigation,
  not an applied filter. Within expanded Browse Filters, expose standard disclosure
  rows for Expiring soon, Expired and All dates in an Expiration section. Explain
  that these review active items by expiration date. Preserve existing query and
  applied refinement handoff to the expiration workspace.
- Entering expiration from Filters uses the currently displayed draft selections
  and latest search text; cancel pending debounce before navigation. Browse's
  applied filters remain unchanged when returning, while its latest query is
  settled into the route and result state.

### Map navigation-header clearance (2026-09-14)

- Native search makes the iOS header translucent. Browse List continues using
  automatic scroll insets; Map must instead offset its entire fixed header and
  column layout by the native navigation header's measured height.
- Read the reactive header height through @react-navigation/elements 2.9.20,
  promoting the already locked navigation dependency to an explicit dependency.
  Do not hard-code status/search-bar heights. Android's opaque header already
  positions content below itself and receives no additional top padding.
- Map columns and horizontal breadcrumbs must not each apply automatic navigation
  insets. Keep existing bottom dock clearance and horizontal gestures unchanged.

## Native Home and Browse actions (2026-09-14)

- Browse moves Add out of its List/Map content rows into the native trailing
  navigation bar beside compact system search. Keep the List/Map control in
  content. Hide Add when inventory permissions disallow creation; preserve the
  existing Add route and cancel any pending map search before opening it.
- Home gains its own native stack within the existing Home tab, preserving its
  root route through a pathless (home) route group. Keep the inventory dropdown's title/subtitle and switch behavior
  as a custom leading header item. Bound its width and truncate visually while
  keeping full accessible context. Do not add a duplicate Home heading.
- Notifications, Add and Account become trailing native bar items. UIKit owns
  symbols, grouping, backgrounds and hit targets. Use the real iOS 26 bar-item
  badge for unread counts; retain full spoken count and loading/error labels.
  Pre-26 systems retain the accessible count even when native badges are absent.
- Preserve notification registration, inventory-scoped query identity, polling
  and retry-on-open behavior. Present native toolbar data without duplicating
  queries or introducing notification state into domain services.
- Reuse a focused platform header-action adapter. iOS uses UIBarButtonItem via
  native-stack items. Android's native-stack binding lacks that item API, so
  place native Compose IconButtons in its headerRight slot. Non-mobile renderers
  use accessible React Native controls for previews and tests.
- Home content uses automatic scroll insets with no duplicate top safe-area
  padding. Preserve Browse's measured map-header inset. Native bar options must
  update on permission, notification count and scope changes.

## Browse filter sheet (2026-09-14)

- Replace Browse's inline expanded filter panel with a native form sheet,
  initially about 70% height and expandable to full height. Keep a scrollable
  overview of Type, Status, Availability, Tags, Sort and Expiration summary rows.
  Use existing settings disclosure/checkmark patterns; show selections as row
  context and drill into one choice category at a time.
- Reset all changes only the sheet draft. Show results is a prominent full-width
  native bottom action; Cancel/dismiss discards the draft and Back returns from a
  choice page without losing choices. Share the native sheet-action adapter with
  Expiration. Sort lives in this sheet instead of a second Browse toolbar icon;
  indicate Relevance and disable explicit ordering while text/tags drive search.
- Tags use native navigation search and multi-select checkmarks; other categories
  use single-selection rows. Expiration offers its existing date-mode destinations
  using current draft and query. Keep the explanation that expiration reviews
  active assets, even when Browse lifecycle includes archived assets.
- Pass canonical Browse route state plus tenant/inventory identity into the sheet.
  Settle pending search before opening. Applying explicitly replaces every filter
  parameter, including empty/default values, so Reset cannot retain old URL state.
- Resolve tag choices through scoped server-query adapters. Reject wrong inventory,
  tenant or session targets before and after loading, and revalidate scope before
  Apply or expiration navigation. Cross-scope sheets must not display stale choices
  or apply to a newly selected inventory.
- Cancel, native dismissal and scope replacement abort pending Apply verification;
  a late response must not navigate after the sheet closes. Serialize submissions
  synchronously so repeated taps cannot produce duplicate navigation.

## Native tab-header scroll appearance (2026-09-14)

- Home and Browse navigation bars float over scrolling content on iOS. Remove
  opaque custom bar backgrounds and separators: on iOS 26 use the native soft top
  scroll-edge effect, with no additional blur layer; older iOS uses native system
  material blur. Do not simulate scrolling transparency with JS opacity or gradients.
- Preserve system automatic content insets on Home and Browse List, and measured
  header clearance around Browse Map's fixed controls. Android keeps its opaque
  native navigation bar and existing layout.
- This corrects scroll appearance, not navigation placement. Preserve current
  Home inventory selector, permission-aware actions, notifications and compact search.
- Follow Apple's TN3106 navigation-bar appearance guidance and the pinned native
  stack adapter's scrollEdgeEffects API. Do not combine headerBlurEffect with the
  iOS 26 native scroll-edge material.

## Inventory switching remains on Home (2026-09-14)

- Keep the inventory switcher only on Home. Browse returns to its native Browse
  title, Add and compact Search, without a leading inventory control or repeated
  inventory row in content. Clear any prior leading header item explicitly.
- Preserve the native transparent scroll-edge appearance and compact Browse
  controls. Existing scoped queries, permission checks and result error recovery
  remain responsible for loading the selected inventory.

## Home refresh and header fit (2026-09-14)

- Home's native pull-to-refresh control represents only a user-started pull.
  Background dashboard/expiration refetches and post-edit refreshes must not open
  or hold the pull control. Blur/unmount clears its presentation; late completion
  cannot clear a newer pull. Ignore late native pull events while Home is not
  focused. Keep ordinary scroll position when returning.
- Keep refreshing dashboard and expiration data together for an explicit pull;
  retain existing query/error recovery and do not cancel shared background reads
  just to reset the control. Suppress duplicate pulls while one is active.
- Reserve space for Home's native action buttons before sizing its inventory
  selector, including native group/margin allowance. Use explicit 44-point iOS
  bar-item widths. Arrange Home actions in this order: Add, Notifications,
  Profile (Account). Preserve this order when optional actions are absent.
- Bound selector width rather than crowding buttons; truncate its text while
  retaining full accessible inventory/tenant names. Add symmetric 12-point inner
  horizontal padding and compact vertical padding. Use a single label at enlarged
  text sizes instead of compressing two lines into the navigation bar.

Move destination creation chooses Location or Container through the shared native
value picker labeled Kind. It is a form value, not a tab or navigation target.
Keep the selected kind's explanation and proposed placement visible, and use that
kind in the create command. Pending creation disables the choice.

Asset Edit, Move, and Move here sheets own one pending mutation at a time. A
synchronous guard rejects duplicate save/create callbacks and freezes all draft
changes, destination selections, Cancel, and sheet gestures until completion.
Failure restores editing with the submitted draft intact. Destination creation
shares the same operation lock as Move. Completion after unmount must not navigate
or display an alert in a different screen. Keep disabled controls visibly present
and expose their disabled state to assistive technology.

The pending sheet lock also prevents system Back/navigation removal. Successful
submission explicitly permits its own return navigation; unavoidable teardown
still suppresses late navigation and alerts.

Asset Edit and Move text inputs expose stable accessible names matching their
visible purpose: Asset name, Description, Put in, and Find item, box, or place.
A preceding Text label alone does not label a native text field. Retain input
names while fields contain values and while submission disables editing.

The shared mobile minimum touch target is 48 logical points, covering both
platforms. Add/Edit tag choices and inline tag-creation controls use that minimum
without overlapping hit regions. Preserve tag selection and creation behavior.
Voice photo controls use native command buttons; each draft photo has a separately
reachable numbered Remove command below its preview, rather than an 18-point
overlaid close target. Photo rails can scroll and commands can grow with text.
Tag creation uses a native Add tag command below its fields/color choice, avoiding
an undersized inline action squeezed beside text entry.

### Asset gallery photo command

The gallery keeps one Add photos command below its empty or populated image area.
Use the shared native command adapter on iOS and Android; the image itself remains
an inspection action. Hide Add photos when permission or its callback is absent.
Preserve the selected photo ID, authenticated image headers and numbered image
accessibility labels. The custom full-screen viewer exception does not extend to
this simple gallery command.

## Dynamic native header update ownership

Native header options must settle after a navigation-context update. Changing only
an action callback's closure must not publish fresh toolbar option factories and
trigger another navigation update. Add uses stable native header presentation
options while handlers read the latest committed draft and dismissal callback.
Visible labels, badge counts, enabled state, order and platform presentation still
update when their inputs change; never freeze the header to hide an update loop.
Each action kind identifies one command within a header side; kinds must be unique
within that side. Stale native actions must respect current disabled state, removed
actions and screen teardown.

Verify a navigation fake that notifies consumers of changed options: initial Add
entry and unrelated draft edits converge, Save submits the latest draft, busy
Save/Close remain unavailable, and editing recovers after failure. Keep native
launch, typing and rejected-save recovery as the release acceptance scenario.

Apply the same stable presentation boundary to Home, Browse, notification inbox,
inventory switcher, checkout-history dismissal and reminder timing headers. Home's
inventory label still updates for inventory/tenant, width, appearance and text
size changes. Notification badges/commands and Browse's Add permission remain live.
These are preventive consumer fixes for the Add-discovered update pattern, not
claims that every consumer has exhibited a native crash.

## Expiration sheet scroll ownership

The expiration filter sheet exposes its ScrollView directly to the native screen
content wrapper, matching the direct-root layout that survives native detent
changes. Do not put a generic flex container above it. Keep the native action
footer as a bottom sibling and measure its actual height to reserve scroll-content
space, including safe-area and text-size changes. Preserve staged selections,
search, validation and native keyboard insets. Header presentation options remain
stable across unrelated draft changes.

This is a candidate for the observed M19 failure, not native acceptance. Existing
medium-to-expanded, long choice list, keyboard search and date-page scenarios must
verify that content and Apply/Back stay reachable on phone and iPad. If the native
sheet does not keep the footer above the keyboard, resolve its actual coordinate
behavior rather than adding a guessed fixed keyboard offset.

## Shared Reduce Motion preference ownership

Custom map, voice-result and notice motion must remain disabled while the native
preference is pending or unavailable. Subscribe before reading the initial value.
A live preference event supersedes that initial snapshot, including when the
snapshot resolves late. Remove subscriptions on unmount and ignore late reads.
Reuse one UI-level preference hook; keep screen-reader notice timing independent.
Native runtime tests must verify map navigation, voice rail movement and notice
entry/dismissal with Reduce Motion, including live preference changes.

### Measured expiration footer keyboard clearance

Keep the direct-root ScrollView and bottom sibling footer. Measure an unshifted
zero-height boundary at the sheet bottom in window coordinates; move the footer
only by the overlap between that boundary and the reported keyboard frame. React
Native0.83 converts iOS keyboard notification frames into the key window before
emitting them, matching measureInWindow. Re-measure on sheet layout/frame changes;
ignore superseded asynchronous measurements and clear on hide/unmount. A sheet
already resized above the keyboard needs no extra movement. Floating keyboards
that do not intersect the bottom boundary and off-window frames need no offset.
Do not reintroduce a wrapping view around native scroll content. Existing phone
and iPad keyboard/expansion tests are required acceptance of this candidate.

### Accessible native choice layout

At iOS accessibility text categories, put the native choice label above its menu
value instead of forcing both into narrow columns. Keep menu selection, disabled
state and accessibility naming. Use Expo's native VStack and hidden internal
picker label; its pinned55.0.17 adapter does not expose SwiftUI ViewThatFits.
The React Native0.83 default scale for AccessibilityMedium is1.786; use that named
threshold for the iOS adapter, retaining LabeledContent below it. No text-size
cap or custom menu is introduced. This applies to all NativeChoicePicker consumers
(filters, appearance, customization, expiration month, invitations, move and voice
settings). Recheck landscape/narrow layouts and all accessibility sizes on native.
See [Apple Dynamic Type](https://developer.apple.com/videos/play/wwdc2024/10074/).

### Refinement count badge contrast

The small count badge on a refinement command must use a paired semantic
foreground/background with at least 4.5:1 text contrast in light, dark and
increased-contrast appearances. Use the existing `onAction`/`action` pair;
`accent` is not a text-bearing background token. iOS, Android and fallback
refinement controls share badge presentation so their contrast and count
formatting cannot drift. The count remains supplementary visual information:
the parent command names the applied count for assistive technology.

This correction does not establish native target, Dynamic Type, badge placement
or material acceptance; those require rendered checks on the affected clients.

### Single keyboard-avoidance owner for measured sheet actions

When the expiration filter container measures and applies keyboard overlap,
its hosted SwiftUI actions must ignore the keyboard safe-area region. Otherwise
SwiftUI can move the buttons outside the React Native host's hit-test bounds.
NativeSheetActions exposes a fixed-per-mount keyboard-avoidance owner: native
by default, container only for the measured expiration footer. Other safe areas
remain active. Browse's unmeasured footer retains its existing native behavior.
Verify actual button hit-testing and navigation after keyboard entry; screenshots
or host props alone do not prove the correction.

The Edit asset form's body title must scroll with its metadata feedback and fields.
Do not reserve a fixed, scaling title above a small sheet scroll viewport. Keep
completion actions available separately, and verify actual large-text sheet
geometry on phone and iPad. Native search acceptance must address accessible
asset-result buttons, rather than requiring their text children to be separate
accessibility nodes; retain positive matching and negative nonmatching assertions.

Native place-search acceptance must demonstrate the matched result is reachable
and fully visible, not merely present in the accessibility tree behind the
keyboard. Dismiss the keyboard through the provided control, reveal the matching
row in the detail scroll, and check its bounds before capture. Retain exact query,
nonmatch exclusion, clear, cancel and navigation return checks.

Add quick parent creation requires a settled, available candidate result, including
a known empty result. Debouncing, initial loading or failed lookup without cached
results must not be treated as no duplicate. Retain the query and allow lookup
retry. Apply the same availability/known-match guard to the creation command and
its visible offer. Cached results may retain the existing name-based duplicate
heuristic; this does not assert global uniqueness.

Add parent lookup retry and quick creation are in-place commands. They use the
existing native command button adapter, with an explicit label and disabled state
while the draft is busy. Quick creation retains a readable “Creating place…”
label while pending instead of replacing the command with only a spinner. Search,
selection and the item draft remain in the current Add form; neither command
creates a new navigation destination. Retry respects draft-operation ownership.
Native large-text, keyboard and scroll reachability require runtime acceptance.

Move uses a short task heading and a separate, wrapping asset name at body emphasis;
a long name must not become the oversized sheet heading. Show the current location
once in quiet form context. Show “Move to” only after selection differs from the
current parent; identical From/To summaries add no information. The context and
query stay in the same scrolling form. Existing valid-change, pending-operation,
cancellation and destination-creation rules remain unchanged. This context repair
does not establish native disabled-button contrast; verify that separately.

Native command buttons must propose a finite available width while allowing the
outer SwiftUI Button to take its ideal vertical size. Applying vertical ideal size
only to its Text can leave the React Native host at the minimum height while text
renders outside it. The large-text native comparison showed a 48-point button
frame for three lines in the shipping control and 187.3 points with outer vertical
ideal sizing, with following content moved below the label. Apply that outer
measurement rule to the shared command adapter; preserve prominence and disabled
semantics. The comparison also confirmed Retry received after a native tap. Its final return
assertion must target the observed native BackButton identifier rather than assume
the localized/contextual label is Back, then verify the audit menu returns. Shared
consumer layouts, primary-button sizing and sheet-footer contrast require further
acceptance.
The diagnostic sizing fixture retains the old inner-only measurement as a named
baseline and compares it against the shipping adapter, so future runs exercise
production sizing rather than a copied candidate implementation.

A runner-only footer appearance diagnostic must use the shipping NativeSheetActions
inside the Move form-sheet container, with the app's real AppearanceProvider.
Capture light and dark resolved appearance with Move disabled and enabled; assert
Cancel remains reachable and an enabled Move invokes its callback. Use default
text and the largest accessibility text size on phone and tablet. A capture is
inspection evidence, not an automatic contrast pass. The diagnostic must not
load production sessions or perform inventory mutations, and must restore its
starting appearance on explicit cancellation. This investigates M100 without
speculatively replacing native disabled styling.

### Filter list and persistent action separation

Browse and Expiration filter pages must share the measured native-action footer
layout. Long tag lists must not remain visible through the footer or leave their
last choices underneath Show results/Back. Give the fixed footer an opaque theme
surface and reserve its measured height, including its safe-area padding, at the
end of scroll content and in the scroll indicator. Re-measure when button height,
keyboard state or sheet size changes; do not reserve a guessed button height.

Reuse the existing Expiration direct-scroll/native-footer arrangement because
nested sheet bodies have produced missing native content in the audit. The body
must remain a direct native screen child. Share its measured keyboard boundary
handling rather than letting the native action host and container both move it.
Keep native search, staged selections, Back, Cancel and Apply semantics unchanged.
Verify long lists at normal text size: scroll the final tag fully above the footer,
select it, return to the overview, search with the keyboard, dismiss the keyboard
and apply the retained draft. Include iPhone and iPad light/dark native checks.

Checkout-history native acceptance must exercise a failed independent asset-name
read while records remain visible, then retry the name without replacing the
history. The controlled native fixture fails its first name read and succeeds on
explicit retry; its existing pagination and dismissal journey also verifies the
name error and retry control are fully reachable before recovery.
