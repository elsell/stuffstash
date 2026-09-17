# Asset History list — all 24 source axes

R011 at153a6be3 plus M162. Reviewed route parameters, AssetHistoryRouteScreen,
AssetActivityQuery, history grouping/filter/error presentation and mounted query
journeys. Entry details/revert are separate surfaces. These observations apply
the existing platform review standard; they do not establish native acceptance.

| Axis | Source evidence and acceptance gap |
| --- | --- |
| Task | Read an item's chronological activity. Changes is the default; All events includes reads. Full entry details justify navigation. |
| Navigation | Row pushes scoped asset/activity identifiers and title to detail. Native title is History. Back and scroll restoration remain native checks. |
| Selection | Two flat choices use NativeActionMenu with selected state, not another stack screen. Choosing the current value is inert. |
| Modality | No local modal. Native filter menu dismisses through its platform adapter; entry detail owns any confirmation. |
| Layout | Heading/filter and refresh-error panel sit above the SectionList. Bottom padding is present; full safe-area/short-window/long-error reachability remains unverified. |
| Adaptation | One flexible column suits phone and tablet. No explicit maximum text width; iPad density and narrow normal-size long titles need captures. |
| Typography | Item/date headings have header semantics; row summary and metadata wrap without line caps. Enlarged-text remediation remains lower priority per user instruction. |
| Appearance | Semantic surface/text/error colors and native command controls. Actual dark/light contrast and native material are not proven by source. |
| Localization | Groups by local calendar date and formats using locale; invalid timestamps get Date unavailable. Shared timestamp formatting supplies row labels. RTL, clock changes and translated lengths remain acceptance work. |
| Imagery | Text-only collection with no media thumbnails. No missing-image state is required here. |
| Targets | Row main targets have minimum76-point height; command targets use shared native adapter. Native hit geometry and menu reachability remain open. |
| Gestures | Vertical scroll and explicit row/menu/buttons. Pull refresh is optional; M162 makes inline Retry independent of its spinner. |
| Keyboard | No text entry or local search. Hardware keyboard and menu/row focus order remain runtime checks. |
| Accessibility | Headings, button roles, action hint and error alerts exist. Row names rely on native aggregation of title/summary/metadata; actual verbosity/order needs VoiceOver/TalkBack review. |
| Motion | No custom animation; native progress and navigation remain subject to reduced-motion runtime checks. |
| Content | Infinite query requests20 entries per page, groups loaded records by local date and offers explicit older-page loading. Failure retains loaded pages and provides a separate older-page retry. Large-history scroll performance remains unverified. |
| Search | No free-text search required by current History spec. Show menu changes query identity; histories do not mix between views in source. |
| Loading | First read shows labeled progress; cached pages remain during ordinary refresh. Older-page progress is local. Only the actual pull gesture now starts its native indicator. |
| Recovery | Friendly first-load classifications distinguish denied/missing from transient failure. Cached refresh errors remain inline with Retry. M162 retires old-visit notices and disables retry during fetching. |
| Editing | No editable draft. Revert belongs to the entry-detail surface; selecting/filtering History does not mutate an asset. |
| Privacy | Keys include session, tenant, inventory, asset and view. Repository inputs are scoped and access failures hide pages. Technical metadata is allowlisted in the query. No server authorization behavior changed. |
| Notifications | No system notification handling in this route. External asset entry remains separately audited. |
| Media | No camera/library/audio operation. Global interruptions and return remain native lifecycle cases. |
| Lifecycle | Warm pages reuse the cache. M162 scopes notice ownership to visit and query identity. Mounted departed/returned/current cases pass; full native lifecycle acceptance is pending. |

Two late-notice cases failed before correction; a separate cached Retry test
failed because the button started the pull indicator. All12 focused History
tests pass remotely, including warm pagination and access-failure suppression.
This does not newly verify API authorization, system interaction geometry or
physical-device behavior.
