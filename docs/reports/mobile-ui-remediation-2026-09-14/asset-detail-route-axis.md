# Asset Details routes — all 24 axes

R012 and R020 share AssetDetailRouteScreen. Both route adapters inject the same
services and assetId; the nested location route does not implement another Details
view. Reviewed at fa504041 with M195 candidate applied. Component evidence covers
the shared screen, not both native navigation stacks.

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | Inspect identity, location, availability, expiration, photos and contents; explicit native commands enter maintenance. No new page is needed for simple checkout/return. |
| Navigation | Both entry routes share Details/Place title logic. Edit, Move, contained assets, breadcrumbs, tag search and History navigate to their own tasks. Nested/direct Back and delete fallback require native acceptance. |
| Selection | Details itself has no editable field choice. Photo selection opens a viewer; location/tag taps navigate. Menu entries are commands, not selected values. |
| Modality | Photo viewing uses a sheet/viewer; lifecycle confirmations identify the asset. Photo source uses the platform chooser. Inspect actual iPad anchoring and dismissal. |
| Layout | One FlatList contains header, contents and maintenance footer, with left/right safe area. No fixed action footer is introduced. Header/search, viewer and notice placement need normal-size captures. |
| Adaptation | Maintenance commands wrap with140-point basis; photos use shared responsive gallery. iPad/narrow-window behavior and full footer visibility remain unverified. |
| Typography | Asset title, description and metadata wrap. Long names, long breadcrumbs and mixed date precision remain native checks. Enlarged-text failures stay recorded behind normal-size work. |
| Appearance | Shared appearance palette and native commands are used; archived/checked-out/expiration status has text. Light/dark contrast and native disabled rendering remain unmeasured. |
| Localization | Labels are English; dates use presentation helpers. User titles and descriptions are arbitrary content. RTL and long localized values remain unverified. |
| Imagery | Gallery supports missing photos, authenticated sources and selection by photo ID. Empty photos differ from unavailable/loading regions. Native image failure and transition appearance remain pending. |
| Targets | Native maintenance/availability commands replace earlier custom controls. Header overflow is a real iOS bar menu; Android uses NativeActionMenu. Target bounds, crumbs/tags and viewer controls require device checks. |
| Gestures | Pull refresh is explicit gesture-owned via usePullRefresh; navigation and commands have visible alternatives. Scroll, sheet dismissal and return behavior remain native acceptance work. |
| Keyboard | Contained-place search uses NativeNavigationSearch. FlatList dismisses keyboard on scrolling and preserves handled taps. Native search return, clearing and keyboard/footer coexistence remain open. |
| Accessibility | Title is a header; photo/contents progress is named; command labels identify actions. Full reading order, focus restoration and status announcements remain unverified. Initial Loading asset includes visible text but no explicit progress role. |
| Motion | No screen-specific animation added. Native transitions, viewer movement and Reduce Motion need runtime checks. |
| Content | Core data can render before contents/photos. Contained items are virtualized through the shared workspace list. Search/empty results and region failures are distinct. See contained-items-axis.md for details. |
| Search | Search is enabled only for appropriate contained-place workspaces, resets with asset identity and rejects old search-owner updates. Native search lifecycle remains covered by its separate runtime backlog. |
| Loading | Independent core/contents/photos queries prevent slow photos blocking identity. Pending mutations disable commands and hold synchronous locks. Background queries do not own pull indicators. |
| Recovery | Core error has conditional native Retry; independent failed regions have their own recovery. Partial uploads retain failed-only retry. M195 prevents departed picker/upload exception notices; M189 covers removal alerts. |
| Editing | Edit completion returns Saved/Undo through the service-scoped notice system. That intentionally supports notice handoff; it must not be replaced with a blanket focused-route rule. Undo repeat/session behavior needs integrated acceptance. |
| Privacy | Queries are service/tenant/inventory/resource scoped and use shared access-failure handling. Capability fields govern visible commands. These source checks do not replace backend adversarial tests. Header cleanup on core failure is an integration check still needed. |
| Notifications | No OS notification handler here; S130 resolves entry into Details separately. Arrival with missing/denied/deleted assets needs physical notification acceptance. |
| Media | Source chooser, viewer, removal and partial retry are composed here. Existing application/route tests cover duplicates and replacement assets. Physical camera/library permissions and upload remain separate. |
| Lifecycle | Mutation presentation uses focused/resource owners; photos retain per-asset operation state. M195 extends exception notices to visit ownership while keeping completion. Native picker focus, background and route return remain unverified. |

Evidence:112 related route/presentation cases, TypeScript and mobile structural
checks pass on paul (`/tmp/photo-acquisition-visit-green.log`). New picker/upload
failure cases reproduced departed notices before the two-line candidate fix.
Critic found no blocker. No new native pass is claimed for either route.
