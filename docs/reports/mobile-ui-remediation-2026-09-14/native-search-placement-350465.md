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
