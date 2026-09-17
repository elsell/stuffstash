# Phone search placement — native350465

Atb6321dcb, the phone Place and settings-collection journeys fail before typing.
Inspected captures show a persistent bottom SearchField, not the expected header
Search button. Both hierarchies place it at33,803 with size336×38 in the402-point
wide phone coordinate space. Search is present, but its placement conflicts with
the compact header requirement. Do not weaken the original tests to accept it.

- [Settings screenshot](phone-settings-search-350465.png) and
  [hierarchy](phone-settings-search-350465.txt): Search tags at the bottom; Add in header.
- [Place screenshot](phone-place-search-350465.png) and
  [hierarchy](phone-place-search-350465.txt): Search this place at the bottom; More in header.

Source inspection confirms the tested revision supplied integratedButton and
allowToolbarIntegration=false. The pinned React Navigation native stack spreads
headerSearchBarOptions to SearchBar; react-native-screens4.23.0 forwards the values
and RNSScreenStackHeaderConfig assigns preferredSearchBarPlacement and
searchBarPlacementAllowsToolbarIntegration to UINavigationItem. No dropped prop
was established. A render/update ordering problem or OS behavior remains a
hypothesis, not a confirmed root cause.

The shared adapter also serves Browse list/map, filters, timezone and voice
location. Those are affected-consumer investigation targets, not additional
confirmed bottom-placement failures from these two screenshots. Expiration
results configure native search separately and need comparison too.

## Next diagnostic

The runner-only static placement route supplies the same placement options at
route registration before presentation, without production query/ref/callback
composition. XCTest captures idle state, requires a hittable Search button within
the native header, then verifies the expanded field. Original production search,
result, reset and navigation assertions are unchanged.

The fixture installer test failed first because the comparison route was absent;
both installer tests and TypeScript/mobile structural checks now pass on paul.
Swift compilation/execution is pending. Critic found no confirmed blocker.
This comparison differs in refs, callbacks, content and timing; a different outcome
would narrow investigation rather than identify the cause. No production fix is
claimed and no live native run was canceled/restarted.

## Run350549 comparison and M215

The phone static comparison passes. Its inspected idle capture shows the icon in
the navigation bar; Place's failed production journey instead shows a bottom
“Search this place” field. Artifact10431223485 retains both. See
`evidence/place-bottom-search-350549.png` and
`evidence/static-header-search-350549.png`. This confirms a rendered difference,
but the comparison differs in more than registration timing.

Independent mounted feedback testing reproduces a shared NativeNavigationSearch
options loop: every render creates another headerSearchBarOptions object. The
candidate memoizes presentation by enabled state/placeholder and reads committed
current callbacks from a ref. Query synchronization and focus/removal guards remain.
The test fails at the bounded25-update guard before the fix and passes afterward,
including replacement callbacks, disabled events and changed placeholders.

Eight consumers were inspected: Browse List, Map, Browse tags, Expiration filter
selection, timezone search, settings collections, Place contents and voice location.
All1,869 tests across288 files, TypeScript and structural checks pass on paul.
Four Browse tests were corrected to read the latest explicit search-options update
rather than assuming unrelated title/action updates always precede search; explicit
removal still fails the test. Critic found no blocker. Source checks prove stable
configuration and current callbacks, not corrected native placement. M207 stays open.

## Controlled delayed-registration comparison

At d1f5a4ba, production Place search is disabled while loading and recreated when
contents become searchable; settings collections also enable it only after ready
and authorized. The static passing fixture starts with search already registered.
The existing preconfigured Place comparison still uses the production component,
which can explicitly remove search while loading, so it does not isolate that
transition.

The pinned RNSSearchBar initializes automatic placement/toolbar integration, then
updates its own placement fields from props. RNSScreenStackHeaderConfig later
copies those fields to UINavigationItem. This confirms separate update steps,
not a proven ordering bug.

A new isolated fixture starts with the shared adapter disabled, explicitly enables
it, records placement, and changes only the route title before recording again.
Both snapshots precede placement assertions, preserving evidence if either fails.
Original Place/settings tests remain unchanged. No title-toggle workaround or
native dependency patch is applied to production. The installer test failed first
for the absent route; all6 installer tests, TypeScript and structural checks pass
on paul. Swift/native execution remains pending.

## Run351318 comparison and upstream cross-check

The managed enable/title-change/native-action probe passes on both devices, while
phone Settings and preconfigured Place still render bottom search fields. iPad
passes both production journeys. Thus generic enablement/header mutation alone
is not a sufficient reproducer. Place additionally keys its search component on
asset/enabled state, but Settings does not; that key cannot by itself explain both
failures. Preserve both production acceptance cases.

The pinned react-native-screens4.23.0 adapter forwards integrated-button placement
and disabled toolbar integration to UINavigationItem. Its native header code
contains an iOS26 repeated-configuration workaround specifically for stacked
placement, but the current evidence does not establish the same cause here.
Upstream [issue4381](https://github.com/software-mansion/react-native-screens/issues/4381)
reports missing stacked search with scrolling collapse; our configuration already
sets hideWhenScrolling=false and fails with a visible bottom field.
[Issue3935](https://github.com/software-mansion/react-native-screens/issues/3935)
reports title overlap for integratedCentered placement after navigation. Neither
report proves a fix for this integratedButton failure. No dependency upgrade or
blanket custom-search replacement is justified by those reports alone.
