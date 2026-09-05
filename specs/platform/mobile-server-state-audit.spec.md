# Mobile Server State Audit Evidence

## Completion Rule

This matrix tracks implementation against `mobile-server-state.spec.md`. A migrated query is not a completed surface until its mounted behavior, cancellation, request graph, mutation reconciliation, and scope isolation are verified. All build validation runs in PR CI: the development host has insufficient disk for builds.

## Surface Matrix

| Routes or entry | Read ports and current query ownership | Request budget and remaining evidence |
| --- | --- | --- |
| Session/connection onboarding, root layout | Authentication commands; composition-owned QueryClient | QueryClient cancellation/clear tested; mounted provider replacement and connectivity behavior still to audit. |
| `(tabs)/index`, `tenant-switcher` | HomeQuery, CurrentInventoryScopeQuery, tenant selection | Focused selected inventory; warm cache and return critical path tested. Recheck mounted switch behavior. |
| `assets/index`, `locations/[locationId]/index` | InventoryAssetsQuery, LocationAssetsQuery | Focused cached lists; mutation observer reconciles selected inventory only. |
| `(tabs)/search`, List and Places | SearchAssetsQuery, InventoryAssetTagsQuery, LocationsQuery | Infinite query owns pages/cursors; mounted switch, replacement failure, secondary summary retry and inactive-consumer cancellation tests pass. List header uses focused context; page breadcrumbs fetch only missing ancestors. Remaining budget work: avoid repeated ancestors across pages and server-side kind filtering scans. |
| `(tabs)/search`, Map and overlays | InventoryMapQuery; shared AssetCoreQuery/AssetContentsQuery/AssetPhotosQuery | One cancellable scoped containment query; mounted warm-revisit, overlay core reuse/cancellation, inventory replacement and embedded-navigation reset tests pass. Native gesture verification remains a device check. |
| `assets/[assetId]/index`, nested location asset detail | AssetCoreQuery, AssetContentsQuery, AssetPhotosQuery | One selected asset GET before core; secondary queries reuse core. Independent loading and mutation request deduplication tested. Mounted delayed-region and changed-parent refresh tests pass; old/new ancestor mutation impacts are covered. |
| `add` | AddAssetContextQuery, focused ParentLookupRepository | Context and bounded candidate queries cached; 250 ms debounce, cancellation and warm reuse tested. Draft/tag lifecycle audit remains. |
| Asset `edit`, `move`, `move-here` | Shared AssetCoreQuery, independent tags, AssetPlacementQuery, bounded parent candidates | No photos or containment traversal for forms. Mounted dirty Edit/Move drafts survive background refresh and changed parent. Placement adapter and picker request budgets verified. Mutation adapter still traverses inventory before create/update and needs reduction. |
| Asset `checkouts` | AssetCheckoutHistoryQuery and independent shared AssetCoreQuery | Focused scope discovery, cursor pages, retained rows on continuation failure; mounted test passes. |
| Asset `history/index`, `history/[activityId]` | Scoped infinite pages and activity detail query; reversal observer | Mounted warm-page reuse, retained rows on failed refresh, scoped detail seeding/cancellation tests pass. No secondary application cache. Reversal keeps unrelated core/photos fresh; containment invalidation remains conservative until full paths are returned. Uncached direct event links still scan all-event pages. |
| `settings/index`, `account`, `household/index`, `inventory/index` | Settings context/account ports | Audit coupled reads, share selected identity, preserve navigation during secondary loading. |
| Household/inventory `asset-types`, `fields`; inventory `tags` (index/new/resource) | Customization collections and editor ports | Manual collection/editor requests. Add resource-family keys, permission-aware failure handling and mutation reconciliation. |
| `settings/sharing`, `invitations/accept` | Invitation list/create/cancel/preview/accept ports | Cache lists only; one-time token/secret results stay workflow-owned. Verify selected scope refresh after acceptance. |
| `settings/voice` capability routes, profiles list/add/detail/prompt/credential | Provider profiles and voice readiness ports | Split readiness from navigation; retain no credential values; reconcile profile mutations with readiness. |
| `voice` realtime sheet | Session controller and approved action commands | Stream stays controller-owned. Audit all action mutation impacts against shared query keys. |
| `settings/appearance`, `about`, `connection`, `diagnostics` | Local preferences/connection commands | Verify which reads are local. Do not cache native preferences or connection/authentication commands as server state. |

## Verified Source Checks

- Asset forms working pass: mobile TypeScript check and 143 test files / 1010 tests pass; mobile structural check passes. This is source evidence, not a native build or end-to-end device check.
- `AssetCoreQuery.test.ts`, `AssetContentsQuery.test.ts`, `AssetPhotosQuery.test.ts`, and `ApiInventorySummaryRepository.test.ts` cover the focused detail graph.
- `fetchMobileInventoryServerQuery.test.ts` uses a real QueryObserver to prove one post-mutation request, joining active work and reusing completed fresh work.
- `QueryClientInventoryMutationObserver.test.ts` covers parent reconciliation without cached child core and unaffected inventory/composition caches.
- `AssetDetailView.test.tsx` covers independent photo and contents progress/action availability. Mounted route tests use native-component fakes and real application queries/TanStack observers.
