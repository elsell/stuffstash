# Persistent mobile tab navigation

Status: specified; implementation and native acceptance pending.

## Task and pattern

Home and Browse remain available while navigating ordinary app screens. The current
root stack presents asset details, expiration, notifications and Settings above
the tab navigator, so those destinations cover the bar. Put ordinary destination
stacks inside the native tabs instead. Use shared routes for screens reachable
from either tab; do not draw replacement tab buttons into individual screens.

Apple's Tab bars guidance recommends keeping the bar visible during navigation,
with temporary self-contained modals as the exception:
https://developer.apple.com/design/human-interface-guidelines/tab-bars
Expo supports shared routes that preserve the active tab during in-app navigation:
https://docs.expo.dev/router/advanced/shared-routes/

## Required behavior

- Preserve Home and Browse labels, ordering and native appearance. Ordinary asset,
  location, expiration, notification, history, sharing and Settings screens keep
  the tab bar visible. Existing regular Settings editors remain within their tab.
- Each tab owns its navigation history. Switching away and returning restores the
  same screen and existing query, filters and scroll state. Do not reset a tab
  merely to make the bar appear. Preserve unsaved editor drafts across tab switches;
  continue existing discard protection when leaving the editor's stack.
- Existing modal tasks, including Add, asset Edit, Move, filters, inventory switcher
  and voice, may cover the tabs. Dismissal returns to the same originating tab and
  underlying screen. Photo inspection may retain its immersive presentation.
- Keep onboarding/authentication outside the signed-in tab hierarchy. Preserve
  existing access checks, inventory isolation, notification and invitation entry.
- Preserve public route URLs. In-app shared destinations retain the active tab;
  cold links use a deliberate Home fallback with a valid back destination. Keep
  the existing Browse URL. Do not duplicate domain queries or screen implementations.
- Native safe areas own bar clearance; no per-screen guessed bottom padding or
  duplicate bar. Verify existing voice accessory context and avoid double insets.

## Acceptance

Walk Home → expiration → item → history, and Browse → item → parent location,
then switch tabs and return to each retained screen. Visit Settings and a regular
editor; switch tabs and verify its draft. Open/dismiss Move and Filters from both
origins. Verify notification cold/warm entry, normal Back and Android system Back,
list-bottom clearance, and no duplicate headers or tabs. Reuse existing workflow
checks and add route-ownership/retention coverage before implementation. Native
phone/iPad/Android evidence is required; route structure alone cannot prove layout
or restoration. This structural change does not silently expand the frozen batch.

## Returning from root modal filters

Expo resolves shared unqualified URLs using the current route segments. A root
modal has no tab segment, so Expiration Filters must carry its originating tab as
navigation metadata and return to that qualified shared destination. Validate the
metadata against Home/Browse; missing or malformed origin uses Home. Preserve all
filter parameters. Browse Filters targets the unique Browse URL and needs no
shared-destination disambiguation. Back/cancel continues to pop the modal.

Native35951026134 phone failure occurred before Browse navigation: fixture button
bounds y65–103 overlapped the transparent navigation bar y62–116. The fixture
must use automatic scroll insets like production Browse; assert full header
clearance before activation. Home/detail tab visibility and modal cancel passed
before this failure. Do not claim Browse history/draft acceptance until rerun.


Native35952669886 passes the Home header and tab-retention workflows on iPhone and
iPad at5760f037. This establishes the tested detail/Move cancel, independent Browse
Back history and Settings draft retention. Expiration Filters still needs a native
shared-route roundtrip: retain a different query in each tab, apply filters from
Browse, then switch to Home and back. The applied route must remain in Browse and
must not overwrite Home's query. The fixture may supply inventory/filter choices;
use production expiration screen, filter screen and return-path policy with only
the fixture route namespace substituted. This does not verify production service
loading or replace its existing boundary tests.

The final phone Settings-return capture in35952669886 lacks a visible tab bar,
despite earlier tab activation passing. The final draft assertion can settle
before navigation presentation does. Explicitly wait for both native tab controls
to be hittable after returning to the editor, then capture. A failure is a
navigation defect; an earlier successful tap does not establish final visibility.

## Accessory appearance after navigation

The persistent voice command must retain its visible microphone in ready state
when navigating into a destination and returning through tabs. An accessible name
and tappable blue background alone do not establish visual acceptance. Use one
controlled normal-text navigation observation: capture the full screen on the
list, on detail after a two-second settling interval, and after a tab round trip.
Keep the same ready fixture state and actual tab accessory.

Element-only XCTest screenshots are not authoritative for the layered native tab
accessory. Baseline36035027845 shows a visible glyph in full-screen captures while
`command.screenshot()` yields a blank blue button for the same journey. The native
symbol experiment36041316716 has the same discrepancy. This does not establish a
persistent product rendering defect or justify changing its renderer. Retain the
existing renderer; use full-screen captures for appearance review. Passing
navigation assertions establishes only that the observation completed. Physical
audio and transient animation behavior remain outside this scoped acceptance.
