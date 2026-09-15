# Contained items: source review

Reviewed at 2db6080c on September 15. Surface S102 is shared by the normal
asset route and the map detail route through AssetDetailRouteScreen. This is
source evidence, not a native interaction pass.

Current search applies only to locations. Containers return immediate children
without filtering; extending search to containers would be a separate explicit
product decision, not an implicit consequence of swapping the control.

| Axis | Evidence and remaining acceptance |
| --- | --- |
| Task | This is navigation within a known place plus scoped search and spatial commands. Existing custom search and command controls need native-pattern remediation (M88). |
| Navigation | Child and parent callbacks open asset workspaces; Add preselects the current parent and Move items here opens the placement workflow. Native back/return preservation on both detail entry points remains pending. |
| Selection | Contents rows navigate; they are not value selectors. This does not certify the downstream parent or move selection surfaces. |
| Modality | Shared detail is also presented from Map. The native search integration must target that sheet's navigation owner, preserve Close, and avoid changing the underlying map path. Not runtime verified. |
| Layout | Virtualized FlatList owns rows, header and maintenance footer. Insets, overlay occlusion, search expansion and footer reachability remain pending. |
| Adaptation | Rows use flexible text containers and spatial labels shrink rather than fixed one-line labels. Tablet/window changes remain uninspected. |
| Typography | Titles and supporting text have no explicit line limit. Largest text and combined action heights still require inspection. |
| Appearance | Controls consume the appearance palette, but custom filled/bordered commands bypass native styling. Contrast and Increase Contrast remain pending. |
| Localization | Copy is English and title/path matching uses toLocaleLowerCase. Diacritic matching, RTL, translated counts and long names remain pending; locale folding alone does not establish correct search collation. |
| Imagery | Rows provide a photo or decorative fallback and exclude the thumbnail from the accessibility name. Native image decode failure has no row-specific recovery treatment; inspect separately from the full-screen viewer. |
| Targets | Search input minimum height44, Clear minimum44×60, spatial commands minimum50 and child rows minimum88 are source facts. They do not prove unobstructed native hit targets. |
| Gestures | Child opening and spatial commands have explicit press controls. Pull-to-refresh is currently the only suggested contents-query recovery (M89). |
| Keyboard | FlatList uses interactive dismissal and handled taps; search is a controlled AppTextInput. Focus, clear/cancel and native keyboard appearance need runtime evidence. |
| Accessibility | Rows combine title, eyebrow and supporting path in the accessible name; headings have header role. Spatial controls declare disabled state. VoiceOver/TalkBack traversal and recovery announcements remain pending. |
| Motion | No contained-workspace animation is defined here. Native scrolling/navigation and Reduce Motion remain pending. |
| Content | Locations separate spaces and descendant items; containers show immediate children. Query matching preserves headings/counts. Pagination/completeness must be traced through the repository before treating the list as exhaustive. |
| Search | Existing spec requires an inline field only at20 combined rows. Implementation matches that old requirement, but conflicts with the newer native/search-on-demand direction. NativeNavigationSearch is the existing candidate adapter (M88). |
| Loading | Core data renders independently of contents and photos. The loading label does not suppress empty contents rows; source permits false empty-state messaging while content is unknown (M89). |
| Recovery | Contents and photo query failures emit a root notice recommending pull-to-refresh. A map sheet can cover that notice; there is no persistent region-specific retry. Contents can still show empty-state copy after failure (M89). |
| Editing | Spatial commands retain existing create/edit capability decisions and callback navigation. Downstream Add draft/Move commit behavior is outside this source pass. |
| Privacy | AssetContentsQuery uses core tenant/inventory and permissions when building its view model. Actual repository authorization boundaries were not exercised here. |
| Notifications | No contained-list-owned notification control. Notification interruption and return to this screen remain pending. |
| Media | Child images and gallery are separate paths. Thumbnail failure, authenticated retrieval and gallery return need runtime evidence. |
| Lifecycle | contentsQuery is local component state. Scope replacement and list-size changes must be tested when introducing native search; no confirmed reset regression is claimed from useState alone. |

Sources: AssetContainedWorkspace.tsx, AssetDetailView.tsx,
AssetDetailIdentitySection.tsx, AssetDetailRouteScreen.tsx, AssetContentsQuery.ts,
AssetDetailQuery.ts and NativeNavigationSearch.tsx under apps/mobile/src.

Apple's [Searching guidance](https://developer.apple.com/design/human-interface-guidelines/searching)
asks apps to communicate search scope. The proposed native field should say
Search this place. Choosing an integrated search button and retaining the current
20-row discovery threshold are project choices, not universal Apple requirements.
A separate search route is unnecessary for this local filter.

Next implementation acceptance: preserve both sections, title/path matching,
counts, clear/no-match recovery and scoped state; verify navigation search in both
normal detail and map sheet. Use native spatial commands with primary Add
prominence and quiet maintenance. For query failure, retain available content,
replace unknown-data empty claims with persistent inline status and independent
native Retry contents/Retry photos actions. Test failures and retries through real
query fakes before implementation, then verify on iPhone and iPad.


M89 implementation follow-up: AssetDetailView now receives region availability
and recovery content from the route. Unknown collections are not rendered as
empty; cached results remain. AssetRegionRecovery uses NativeCommandButton and
region-specific Retry labels. Two query-fake scenarios cover unknown/failed data,
independent retries and retained contents after a failed refresh.35 remote checks,
TypeScript and structural validation pass; no native acceptance is claimed.
The table above records the discovery baseline. M88 remains unimplemented.
