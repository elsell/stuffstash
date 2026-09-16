# Browse route and list search — 24-axis source review

R005/S067/S070 at2f861b25 plus M185. Inspected SearchRoute, SearchScreen,
BrowseHeader, BrowseAddHeader, BrowseSurfaceControl, BrowseResultStates and their
mounted behavior cases. Nested [Map](map-axis.md), [filters](browse-filters-axis.md)
and [shared cards](asset-lists-axis.md) retain their own review and native gaps.

| Axis | Source decision and outstanding acceptance |
| --- | --- |
| Task | Browse discovers inventory through visual list or containment Map. Search/refinement share this context; Add is a distinct native header command gated by create permission. |
| Navigation | Cards and parent paths open detail, tags navigate to scoped search. Filter route includes session/tenant/inventory; Add navigates to its editor. Native return and root scroll-edge integration remain open. |
| Selection | List/Map uses the existing native segmented control. Filter choices belong to the scoped filter sheet, not list cards. No persistent card selection. |
| Modality | Filter editing opens a native sheet; list and map remain tab content. No extra modal for simple search. |
| Layout | Left/right safe area, native navigation search/header, automatically adjusted FlatList insets. Map wrapper uses measured iOS header padding. Prior Map/header regressions require runtime verification; source is not proof of transparent scroll edges. |
| Adaptation | Column count depends on width/font scale/result kind, card width derives from width. Places use rows. Fixed-width List/Map control and adjacent summary/filter need normal narrow-screen review. |
| Typography | Shared card and header styles; result summary is single-line and may truncate. Review long queries/titles and ordinary-size density before enlarged-text work. |
| Appearance | Semantic palette, native headers/search/segments/refinement. Applied-filter tokens remain custom removable summaries; their contrast and geometry need captures. |
| Localization | English labels, summary grammar and filter text. Shared results contain dates/path labels. RTL and long translated copy remain unverified. |
| Imagery | Cards use shared photo/fallback and path/tag controls; place rows differ deliberately. Missing/slow photos and image crop require native review. |
| Targets | Native commands and shared card/tag targets; M185 replaces bespoke continuation Pressable. No rendered hit-area claim. |
| Gestures | Native list scrolling, explicit pull refresh and tap commands. Map's horizontal/path gestures belong to its review. |
| Keyboard | Native navigation search, interactive iOS dismissal and handled list taps. Focused debounce stops on blur and resumes latest draft on return; native typing fidelity remains unresolved separately. |
| Accessibility | Search/segment/filter names, live result summary and card semantics are declared. Native traversal across chips/cards/toolbar and no-match announcements remain open. |
| Motion | No custom route animation. Header transitions, spinners and map movement require reduced-motion review. |
| Content | Virtualized result rows, compatible page accumulation and explicit sparse-page continuation. M185 keeps the native command consistent with Retry. No false empty state until cursor exhaustion. |
| Search | Draft query and submitted criteria are distinct. Debounce300ms, trim on submit; route updates consume local echoes. External criteria replace paused drafts. Map/list switch settles pending query; mounted cases cover these paths. |
| Loading | Initial search, loading more and explicit pull are separate. Same-inventory previous results can remain during replacement; current result criteria label them. Background work does not drive the pull indicator. |
| Recovery | Distinct initial, replacement and pagination errors with native Retry. Missing place summaries do not hide places. Access/context failure suppresses previous results; mounted denial then transport-failure cases verify non-resurfacing. |
| Editing | No item draft. Applied filters are committed through their own sheet and route state. Query clear, individual token removal and filter reset preserve their distinct scope. |
| Privacy | Session/tenant/inventory-keyed data and context checks; prior results only reusable within identical scope. Map paths clear on scope change. This is frontend isolation evidence, not backend authorization certification. |
| Notifications | N/A: this route owns no OS notification task. |
| Media | N/A for acquisition/playback: photos are passive card content; Add owns picking/capture. |
| Lifecycle | Debounce ownership, route replacement and per-scope map reset are source/mounted verified. Native cold/deep-link entry, tab return and backgrounding remain open. |

M185 is a project native-control consistency finding: a continuation command needs
no custom substitute when the same native command adapter already handles Retry.
Its existing sparse-page mounted test verifies continued loading reaches the
matching item. All13 focused checks, TypeScript and structural checks pass remotely
(`/tmp/browse-continuation-native.log`). Native footer geometry is still pending.

Acceptance still required includes list/map switch during search, retained query
after detail return, filters Apply/Cancel, empty and sparse results, failed page
retry, inventory replacement and transparent header behavior at normal text size.
