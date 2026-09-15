# Browse filters review

R016 and S071–S073 reviewed at e303e4f7, September15. Sources: browse-filters
route, BrowseFiltersScreen, BrowseFilterRouteState, useBrowseFilterNavigation,
NativeFilterSheet, NativeNavigationSearch, SettingsPickerRow and sheet options.
The review covers overview, tag selection and expiration handoff, not results
ranking, the full Expiration workspace or API authorization certification.

| Axis | Source result and remaining acceptance |
| --- | --- |
| Task | Short type/status/availability/sort choices stay in place. Tags justify searchable multi-selection. Expiration is a separate date-review task with explicit explanatory context. |
| Navigation | Internal Back returns to overview; Cancel leaves without applying. Expiration replaces the filter route after verifying scope. M115 binds asynchronous handoff to the focused visit. Native Back/swipe return remains pending. |
| Selection | Native picker adapters own short exclusive choices; tag rows are independent checkboxes. Draft remains local until Show results or expiration handoff. Controlled selection/reset cases pass. |
| Modality | Native filter sheet has0.7/full detents and a title. Initial loading only renders progress in source; explicit cancellation during that state needs follow-up, distinct from ready-state footer actions. |
| Layout | M111 shared measured opaque footer reserves scrolling clearance. The production ready route had an unconditional outer View, unlike the direct fixture. The follow-up candidate removes that wrapper and puts errors inside the scroll body. Actual native geometry remains pending. |
| Adaptation | Choices and footer use shared layout; no fixed item count. Long tags, iPad, compact width and keyboard layouts remain pending. |
| Typography | Shared settings text styles and native picker text are used. Normal-size long names and wrapping remain pending; enlarged-text remediation follows normal-size fixes. |
| Appearance | Settings and footer use appearance palette. Native command materials are platform-owned. Verify light/dark empty states, disabled actions and theme changes. |
| Localization | Sorting/filter matching uses locale-aware string operations; labels remain English. RTL, translated labels and selected-count grammar remain pending. |
| Imagery | No photos are needed for selecting filters. Native picker disclosures/checkmarks and row checkboxes communicate choice. Native icon alignment remains pending. |
| Targets | Footer commands/native pickers and full choice rows provide actions. M111 last-tag reachability has controlled layout tests; runtime geometry is pending. |
| Gestures | Ready-state Back/Cancel provides an explicit alternative to sheet dismissal. Search clearing is native. Initial-loading dismissal needs a separate check; no gesture-only acceptance is claimed. |
| Keyboard | Native tag search is integrated-button style; shared footer handles measured keyboard overlap. Search cancel/clear preserves draft. Native focus and tab/sheet keyboard transitions remain pending. |
| Accessibility | Tag rows expose checkbox state; navigation and commands have labels. M114 empty copy is ordinary section text. Search-result announcement, reading order and native picker traversal remain unverified. |
| Motion | No custom filter animation is added. Native sheet/search transitions and Reduce Motion require runtime verification. |
| Content | Tag choices are alphabetically sorted, filtered locally and all rendered. No pagination is implemented for this option list; very large inventories need performance acceptance. |
| Search | Tag search affects options only, not selected IDs. M114 explains no tags versus no matches. Search or tag-filtered Browse uses relevance and suppresses an ineffective sort picker. |
| Loading | Tags/scope are loaded before mounting the draft editor. Busy scope verification disables primary application; internal Back cancels it. Spinner-only initial loading still requires explicit-dismissal review. |
| Recovery | Query/scope failures have Retry/Cancel. Verification errors stay with the sheet; M115 suppresses departed feedback. M114 clarifies empty results. Native error placement/retry focus remains pending. |
| Editing | Reset changes only the draft; Cancel does not apply. Show results captures the draft for verified navigation. Some internal controls remain available during verification; Back cancels pending work. No persistence mutation occurs here. |
| Privacy | Target carries tenant/inventory/session; tag loading verifies before and after read, and navigation verifies current scope. This is client scope protection, not a substitute for server authorization. Native account-change behavior remains pending. |
| Notifications | No push control is owned here. Notification-driven navigation during scope verification is represented by controlled blur tests; real notification interruption remains unverified. |
| Media | No recording/upload/file selection task belongs in filters. A covering global voice or media task still needs native interruption checks. |
| Lifecycle | Scope/unmount/cancel guards are supplemented by M115 focus abort/ownership. Late results cannot unlock a replacement request in controlled tests. Background, remount and actual native focus delivery remain pending. |

The full mobile suite passes1,610 tests/264files remotely at e303e4f7, including
new empty-search and late-verification cases. Native execution and screenshots
remain separate requirements. This table documents source review across all
four surfaces; it does not mark native acceptance complete.
