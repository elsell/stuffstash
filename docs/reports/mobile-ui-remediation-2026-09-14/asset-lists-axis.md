# Inventory and location asset lists — 24 source axes

R015 and R021 atfc3bb7a3 plus M161. Reviewed route wrappers, both list screens,
queries, ApiInventorySummaryRepository snapshot methods, shared query/pull hooks
and tag navigation. These are the Recently changed and location-content grids,
not Browse's searchable list or Map. LocationsScreen is a retained component with
test consumers but no current route import; its shared refresh correction is not
counted as another shipped surface or native acceptance.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | R015 expands Home's recent summary; R021 shows descendants of a location. Card collections fit visual identification. They are not alternative editors. |
| Navigation | Cards open detail, parent breadcrumbs open their location asset, tags push scoped Browse search. Location cards retain their location route ancestry. Mounted tag-navigation tests cover destinations; native Back remains pending. |
| Selection | No bulk or persistent selection. Tags navigate to refinement, rather than selecting the card. No flat value choice requires a picker here. |
| Modality | No local sheet/dialog. Detail destinations own their subsequent tasks. |
| Layout | Safe area owns left/right edges; FlatList owns scrolling with content padding. Automatic content inset adjustment is not explicitly set here. Top/bottom reachability under actual root chrome needs native evidence before changing inset ownership. |
| Adaptation | Both grids use two columns regardless of window width. Card wrapping is shared; narrow/landscape/iPad density and normal-size long titles remain native checks. Enlarged-text work remains queued. |
| Typography | Content headings use30-point text and card text uses shared styles. Location title also appears in native navigation title. Readability and duplicate-heading impact need rendered review. |
| Appearance | Both use appearance palette for shell and cards. Native recovery controls replace custom buttons under M13. Real contrast/material acceptance remains open. |
| Localization | Shared card dates are formatted elsewhere; route copy remains English. Long location/inventory names, RTL card/breadcrumb order and timezone changes need runtime checks. |
| Imagery | Shared cards provide photo/placeholder and preserve photo headers. Adapter maps primary photos for each result. Slow/missing images need runtime verification. |
| Targets | Card and breadcrumb targets come from shared AssetCard. Source dimensions are not measured touch reachability; native card/tag/header tests remain required. |
| Gestures | Vertical scrolling and pull refresh use native lists; commands have explicit tap targets. No hidden swipe task is introduced. |
| Keyboard | No local text entry. Tag search enters Browse's native search flow; keyboard restoration belongs to that destination. Hardware focus remains unverified. |
| Accessibility | Cards expose shared action names; initial errors have alert roles and named native Retry. Content headings do not explicitly declare header roles, a remaining accessibility review point. Native reading order/announcements remain pending. |
| Motion | No custom route animation. Spinner and navigation are native; reduced-motion behavior still needs runtime review. |
| Content | R015 adapter loads recent inventory results plus active ancestry; R021 traverses active descendants, excluding the selected location. FlatList virtualizes rendering but data/photo preparation is eager. Very large inventory performance is not established by source review. |
| Search | These bounded-purpose routes have no local search/filter control. Tags enter Browse search. Broader inventory discovery remains in Browse rather than duplicating its controls here. |
| Loading | Initial load has labeled progress; background reads retain ordinary cached content without a pull indicator. Explicit pulls own the spinner. M161 adds failure feedback for an explicit failed pull. |
| Recovery | Initial errors use disabled-while-fetching Retry. M161 retains loaded cards after ordinary pull failures and explains retry. Access failures still suppress cached data. Background failures remain query state without an unsolicited notice. |
| Editing | No drafts/mutations in either route. Detail and Browse destinations own edits, filter state and recovery. |
| Privacy | Session/tenant/inventory and location keys scope reads. The shared adapter rejects obsolete scope publication and hides cached access failures. M161 changes presentation only, not server permissions. |
| Notifications | No notification delivery/tap handler here. External navigation into detail is audited separately. |
| Media | Display-only primary photos. Camera/library/audio belong to other surfaces. |
| Lifecycle | Shared pull hook retires spinner on blur. M161 also retires feedback after leaving/returning and keys ownership to the resource. Native scroll restoration, deep links and replacement-resource journeys remain pending. |

Three current-visit tests failed before M161 because no notice appeared. The
nine-case mounted matrix covers all three component consumers and current,
departed, returned visits. Visible synthetic cards survive the failed read, and a
subsequent successful pull remains possible. Existing initial-failure/scope retry,
background refresh and tag navigation cases remain green. This is source/mounted
evidence only; native geometry, assistive technology and Android remain open.
