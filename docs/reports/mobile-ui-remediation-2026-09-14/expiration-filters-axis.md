# Expiration filters review

## Current native review, September 16

[Run351404 filter evidence](native-filters-351404.md) verifies scoped long-list,
selection and calendar-dismissal behavior, but visual review finds M249: the phone
keyboard accessory overlaps Back despite a hittability-only test pass. The native
clearance candidate remains unverified. Earlier source observations follow.


Source review at b15a38d5, September 15: R017 and S075–S079. Inspected the
route, ExpirationFiltersScreen, ExpirationDateRange, route serialization,
NativeFilterSheet and existing component tests. This covers filter tasks only;
expiration query correctness and the results workspace are separate surfaces.

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | Kind and availability use short in-place choices. Tags support multi-selection; locations carry path labels. Type selection always uses a searchable page; suitability for small type sets needs review with real inventory sizes. |
| Navigation | Single type/location selection returns to overview; tags remain open for multiple choices. Back preserves the draft and Cancel dismisses. Apply dismisses to the expiration route with serialized parameters. Actual native return behavior remains pending. |
| Selection | Any clears single choices; tags toggle independently with checkbox state. Clear filters retains mode but clears query too. Whether clearing a search query matches the intended task needs product-contract review. |
| Modality | The route uses shared filter sheet options. Loading now has Cancel (M116), error has Retry/Cancel. Native detent and dismissal acceptance remains pending. |
| Layout | Shared measured opaque footer (M111) reserves content clearance. Expiration returns its native body directly. Long-list, keyboard and last-row geometry require native verification. |
| Adaptation | Footer height is measured; content scrolls. Compact date popovers and long path labels need phone/tablet/window checks. |
| Typography | Settings styles and native date controls supply text. No enlarged-text acceptance is claimed; normal-size long-label wrapping remains pending. |
| Appearance | Palette supplies footer and error text; native pickers own their appearance. Dark disabled actions and light/dark transitions remain pending. |
| Localization | Dates serialize local calendar components, parsing uses local noon for picker values. Search is locale-lowercased; copy is English. Timezone boundaries, RTL and long localized labels remain unverified. |
| Imagery | No photos belong to these choices; native picker indicators and checkboxes express state. Native alignment remains pending. |
| Targets | Whole settings rows and native commands are interactive. Measured footer does not itself prove targets or final choice reachability. |
| Gestures | Back/Cancel are explicit alternatives to sheet gestures. Native date dialogs support dismissal. Swipe, keyboard and gesture arbitration remain pending. |
| Keyboard | M117 found permanently stacked search. The candidate reuses Browse's integrated search adapter, keyed by page to reject departed callbacks. Native focus/close behavior also needs acceptance. |
| Accessibility | Tags expose checkbox state; dates have first/last labels. Range error is an alert. VoiceOver order, picker values, and no-match announcements remain unverified. |
| Motion | No custom animation is introduced here. System sheet/search/date transitions under Reduce Motion remain pending. |
| Content | All choice rows render eagerly in supplied order. Location labels use paths, without an expanding tree. Large option counts need performance and findability acceptance. |
| Search | Local substring matching does not clear selected IDs. Opening a page clears the previous search. No matches appears outside the section; empty options use the same copy. M117 implements compact search consistency; native verification remains pending. |
| Loading | All three choice queries resolve before the editor mounts. M116 supplies explicit loading Cancel. Native dismissal and transport abort are not proved by the component test. |
| Recovery | Route load failures show generic Retry/Cancel. Invalid date order disables Apply and shows explanatory alert. Native alert reachability and retry focus remain pending. |
| Editing | State is local until Apply; Back preserves it, Cancel does not apply it. Date toggles initialize today; Android dismissal leaves the current value. Component tests cover selected draft paths, not every interruption. |
| Privacy | Choice loading checks inventory scope before and after reads, with session-scoped query keys. Apply is synchronous route navigation. Actual account changes while the sheet is mounted and server authorization remain separate acceptance requirements. |
| Notifications | No notification controls live here. Notification-driven departure during loading or editing requires native interruption checks. |
| Media | No media capture or upload task lives here. Global covering media/voice tasks still need return checks. |
| Lifecycle | Editor key includes session/tenant/inventory, resetting drafts on identity change. Background/resume and retained screen behavior require runtime verification. |

This is source coverage, not a native pass. Existing tests cover staged choices,
checkbox semantics, availability selection, and Android date dialog dismissal.
They do not establish rendered geometry, focus, full route scope behavior, or
physical-device interruption handling.
