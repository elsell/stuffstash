# Browse containment Map review

S068–S069, source 608b1e54, September 15. Inspected InventoryMapScreen,
its styles, InventoryMapPresentation, useInventoryMapSearch and
useMobileInventoryServerQuery. This is source review, not native acceptance.
The containment columns are an existing project interaction, not an Apple map
control. Apple's [search guidance](https://developer.apple.com/design/human-interface-guidelines/searching)
supports explicit scope; its [search session](https://developer.apple.com/videos/play/wwdc2026/292/)
describes communicating no results. Applying that to path lookup is design judgment.

| Axis | Evidence and remaining work |
| --- | --- |
| Task | Browse containment by column; row opens branch or leaf details, separate Info opens any asset. Existing spec permits this custom spatial interaction. Native gesture fit remains open. |
| Navigation | Breadcrumbs restore ancestor path; Add here and Info enter shared routes. Path store is keyed by session, tenant and inventory. Actual return/scroll preservation needs native acceptance. |
| Selection | Expanded and highlighted states are represented in row accessibility state. Search selects the first title/kind/path substring match, not a selectable list of all matches. Matching ambiguity remains a product limitation. |
| Modality | Map owns no modal workspace; commands route to shared screens. Covering-sheet and dismissal behavior remains a runtime gate. |
| Layout | Horizontal animated columns clip within mapScroller. Bottom clearance uses safe area plus a fixed 150 rather than measured accessory height. Header/keyboard/last-row geometry remains unverified. |
| Adaptation | Column width clamps to 292–370 with window width minus 72; compact windows narrower than this minimum need checking. Tablet can show adjacent columns. |
| Typography | Titles, metadata, trails and column headings truncate at one line. Normal-size long-title distinguishability needs native review; enlarged-text work follows. |
| Appearance | Palette supplies surfaces, focus border and text; hard-coded shadow color is decorative. Contrast and returned theme changes require runtime evidence. |
| Localization | Labels/count grammar are English; matching uses lowercasing, not locale-aware collation. Paging/swipe coordinates are physical left/right. RTL is not established. |
| Imagery | Rows show remote photo or kind placeholder and child-count badge. Photo load failure has no explicit onError fallback; native failed-media presentation remains open. |
| Targets | Main rows are 72 points high and Info is at least 48 wide. Empty-column Add is a custom 40-point command; Retry lacks an explicit minimum target. M126 tracks native command correction. |
| Gestures | Tap and breadcrumbs provide alternatives to branch swiping; custom pan arbitration locks vertical scrolling during branch movement. Interrupted gestures and restored scroll lock need native acceptance. |
| Keyboard | Shared native search plus keyboard dismissal on list drag. M122 protects hidden callbacks and M124 pauses unfinished debounce on blur. Returned field text and keyboard insets remain open. |
| Accessibility | Rows have names, hints, expanded/selected state and separate Info labels. Offscreen/exiting columns are pointer-disabled but not explicitly accessibility-hidden. Reading order and hidden-column traversal need native testing. |
| Motion | Reduced-motion preference bypasses column and swipe animations. OS preference delivery and interrupted animations remain runtime gates. |
| Content | Each column virtualizes rows, but the view model contains the full inventory. No pagination here; large hierarchy loading and memory are not certified. |
| Search | Placeholder describes path expansion. No match only clears highlight and leaves old path without explanation: M125. Empty query clears highlight; manual navigation cancels search rerun. |
| Loading | Initial loading has text. Pull uses usePullRefresh, independent of background query activity. Retry incorrectly calls the pull handler: M126. |
| Recovery | Initial error offers Retry. Cached refresh failure retains data with a notice. Access errors remove cached data in the shared query adapter. Retry command and silent no-match need correction. |
| Editing | No asset draft is owned here. Search/path are transient; Add delegates editing to its route. Returning from those routes needs native state verification. |
| Privacy | Query keys and path-store keys carry session/tenant/inventory. Shared query hides resource data on access failure. This source review does not replace adversarial API tests. |
| Notifications | No notification settings are owned here. Notification navigation can interrupt search/gesture work; physical notification interruption remains unverified. |
| Media | Photos are displayed; capture/upload belongs to other routes. Global voice interruption and media load errors require runtime checks. |
| Lifecycle | M124 cancels debounce on blur, resumes unfinished query and preserves explicit cancellation. Background refresh may replace map/path/highlight; native background/return consistency is still open. |

No new runtime pass is claimed. M125 and M126 are normal-text findings requiring
spec-first remediation. Other qualifications above remain review/verification work.
