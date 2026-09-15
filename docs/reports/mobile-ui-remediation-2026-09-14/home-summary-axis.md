# Home summary sections review

S063–S064 at a67157f9. Inspected HomeScreen, ExpirationHomeContent/Section,
HomeDashboardQuery, ExpirationWorkspaceQuery and AssetCard. This covers the two
summary sections, not header actions, checkout commands or full destination screens.

| Axis | Source evidence and remaining acceptance |
| --- | --- |
| Task | Expiration highlights dated items needing attention; recent changes helps resume inventory work. Sections share rows and explicit full-list entry. |
| Navigation | Expiration count rows open their modes; See all opens all dates. Recent See all opens /assets; cards and breadcrumbs open shared asset routes. M128 guards expiration scope ownership. Destination ordering and actual return remain native checks. |
| Selection | Count rows navigate rather than toggle a filter locally. Asset rows have explicit open labels. No persistent selection in these sections. |
| Modality | Neither section opens its own modal. Destinations own their presentation. Back behavior needs native acceptance. |
| Layout | Both live in the Home scroll body; headers have section actions. Home uses left/right safe areas and native header arrangement. Last-row/accessory overlap remains a runtime gate. |
| Adaptation | Row cards use flexible title/body layout. Phone/tablet widths and long breadcrumb trails need native review. |
| Typography | Row titles wrap without fixed line count. Expiration and updated labels share card components. Normal-size long medication names remain a representative acceptance case. |
| Appearance | Semantic palette is passed into cards and section styles. Expiration uses text plus status imagery, not color alone. Real dark/light contrast over photos remains unverified. |
| Localization | Counts use String conversion and English labels; date formatting is shared. Singular count labels and translated lengths are not polished. RTL breadcrumb/section alignment remains pending. |
| Imagery | Photo or placeholder in each card; expiration status accompanies title. Missing-photo response behavior needs native checking. |
| Targets | Section action styles and count selection rows provide targets; card title is separately accessible from decorative photo region. Actual geometry/reachability remains unverified. |
| Gestures | Vertical scroll and explicit taps suffice; pull refresh has an explicit gesture owner. No hidden swipe command in these summaries. |
| Keyboard | No text input here. Home scroll supports dismissal if another interaction leaves a keyboard visible; voice/route return needs native checks. |
| Accessibility | Section titles are headers. Card labels include title and expiration status/date. Count rows describe destination. Reading order and duplicate status announcements require native review. |
| Motion | No custom section animation. Home scroll and native destination transitions require reduced-motion runtime checks. |
| Content | Recent renders first3 of up to10 active assets, relying on repository order. Expiration fetches2 expired and2 soon, interleaves and caps at3. Full destinations handle longer lists. Independent count/snapshot consistency needs API/runtime verification. |
| Search | Summaries have no search. Users navigate to full destinations. Existing destination search state and sort semantics need separate review. |
| Loading | Dashboard owns its load state; expiration has independent labeled loading. M128 prevents scope-less navigation. Background refresh must not own pull spinner. |
| Recovery | Expiration has local Retry and retains resource data for ordinary refresh errors; scope/access errors hide it. M128 covers failed scope retry/recovery. Recent uses dashboard recovery. Native errors and empty layout remain pending. |
| Editing | No edit draft in these sections. Opening an asset delegates editing; no save side effect occurs on summary selection. |
| Privacy | Dashboard uses scoped query adapter. Expiration keys include session/tenant/inventory and M128 gates cached presentation/navigation. These client checks do not replace API authorization. |
| Notifications | Bell/push are separate surfaces. Refresh after notification navigation and changed expiration counts need actual device evidence. |
| Media | Images are display-only. Camera/library/audio belong elsewhere; interruption by global voice remains a runtime scenario. |
| Lifecycle | Home dashboard key separates inventory instances. M128 rejects retained callbacks on scope loss. Pull indicator ownership is shared. Warm return, background date rollover and inventory replacement need native verification. |

Expiration hides itself only when all dated-item count is zero and there is no
error; future dated items keep See all reachable with “None expiring soon.”
Recent empty copy says “No assets yet.” Existing tests prove source behavior,
not screenshot/layout acceptance. No full Home pass is claimed.
