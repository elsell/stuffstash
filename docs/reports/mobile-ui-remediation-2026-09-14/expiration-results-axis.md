# Expiration results review

R018/S074 source review at50d48f3c. Inspected ExpirationRoute,
ExpirationWorkspaceScreen, useExpirationSearch, usePullRefresh, route parameters,
ExpirationSections and filter-header adapters. Results rendering and query wiring
are reviewed here; server ordering/authorization and shared AssetCard acceptance
are separate requirements.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | Date-oriented review groups active item results by month/status. Soon, expired and all modes are explicit. Real inventory density and discovery remain native acceptance work. |
| Navigation | Mode/search update route parameters; filters receive current pending search through flush. Asset and parent links open details. Return-position behavior remains unverified. |
| Selection | Native segmented mode selection uses a native menu at narrow widths or increased font scale. Filter-active state excludes text query intentionally. Native selected appearance remains pending. |
| Modality | Results is a route; refinements open the filter sheet. System Back and sheet return need runtime coverage. |
| Layout | A FlatList sits inside a flex wrapper with automatic insets and32 bottom padding. Transparent header/scroll geometry requires direct native evidence. M118 removes the persistent search field. |
| Adaptation | Width/font scale selects the alternative mode control. Phone/tablet content width and landscape remain pending. |
| Typography | Month headings use20pt and labels17pt; AssetCard supplies rows. Normal long titles and month labels require runtime checks before enlarged-text remediation. |
| Appearance | Palette supplies text, background and progress tint; native controls own materials. Contrast and appearance transitions remain pending. |
| Localization | Presentation formats month labels; grouping relies on supplied expiration state and date month. Locale/timezone boundary correctness needs repository/presentation coverage, not screen inspection alone. |
| Imagery | Shared AssetCard owns photos, placeholders and breadcrumb rendering. This review does not certify those shared consumers. |
| Targets | Retry/load-more custom text buttons have44 minimum height. Native pattern fit remains review work; runtime reachability is unverified. |
| Gestures | Explicit Retry and Load more supplement pull/scroll. M119 separates Retry from the pull-indicator handler; native acceptance remains pending. |
| Keyboard | M118 uses compact native search, preserves debounce and flushes before navigation. Native header focus and keyboard transitions remain pending. |
| Accessibility | Headings and error alerts are labeled; loading states have labels. Result changes and pagination announcements need VoiceOver acceptance. |
| Motion | No screen-specific custom animation. Native refresh/search transitions under Reduce Motion remain pending. |
| Content | Infinite query fetches30 items, deduplicates IDs, groups consecutive month/status records. Load more is explicit. Stable grouping depends on repository ordering; no ordering proof is claimed here. |
| Search | Search debounce is300ms, clear immediate, filters flush pending text. External route changes update native text. M120 pauses pending debounce on blur and resumes retained text on return. |
| Loading | Initial and pagination spinners are separate. Pull refresh uses the shared focused-gesture owner, and M119 gives Retry a separate direct-query command. |
| Recovery | Error copy has Retry, while inventory mismatch asks the user to return Home. M119 replaces the ineffective mismatch Retry with Return to Home. Native return remains pending. Pagination errors currently appear at the top; local footer recovery remains review work. |
| Editing | No persisted edits occur here; route state holds filters. Filter clear semantics are reviewed separately. |
| Privacy | Scope mismatch and access failure hide items. Query keys include scope/tenant/inventory. Retried access failure behavior and real API authorization require boundary verification. |
| Notifications | Screen owns no notification registration. Notification-driven departure needs actual native return/focus checks. |
| Media | Images are read through AssetCard; no recording/upload occurs here. Global voice interruption remains runtime work. |
| Lifecycle | Query interval uses expiration context and pauses background intervals. Shared pull owner clears on blur; M120 also cancels debounce on blur and ignores hidden callbacks. Cold/deep-linked entry and retained-screen blur require verification. |

Source review is not runtime acceptance. M118's three focused tests cover compact
configuration, pending-query handoff and clearing, not native layout. The earlier
full suite at62895a06 excludes that later change.
