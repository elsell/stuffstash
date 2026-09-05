# Mobile Server State Audit Evidence

## Completion Rule

This matrix tracks implementation against `mobile-server-state.spec.md`. A migrated query is not a completed surface until its mounted behavior, cancellation, request graph, mutation reconciliation, and scope isolation are verified. All build validation runs in PR CI: the development host has insufficient disk for builds.

## Surface Matrix

| Routes or entry | Read ports and current query ownership | Request budget and remaining evidence |
| --- | --- | --- |
| Session/connection onboarding, root layout | Authentication commands; composition-owned QueryClient | Mounted selection reset/composition replacement cancels old reads, clears retired data and rebinds observers. Foreground/reconnect revalidation and injected native connectivity ordering/cleanup are tested. |
| `(tabs)/index`, `tenant-switcher` | HomeQuery, CurrentInventoryScopeQuery, tenant selection | Focused selected inventory and ancestor-only card hydration; checkout/recent cards use primary-photo metadata without attachment lists. Warm cache/return critical path and mounted inventory replacement pass. Discovery expires after five minutes without putting discovery back on the warm command path. |
| `assets/index`, `locations/[locationId]/index` | InventoryAssetsQuery, LocationAssetsQuery | Focused cached lists; mutation observer reconciles selected inventory only. |
| `(tabs)/search`, List and Places | SearchAssetsQuery, InventoryAssetTagsQuery, LocationsQuery | Infinite query owns pages/cursors; mounted switch, replacement failure, secondary summary retry and inactive-consumer cancellation tests pass. List header uses focused context; page breadcrumbs fetch only missing ancestors. Kind filtering scans at most five transport pages per UI continuation, with truthful sparse-page feedback. Repeated cross-page ancestors remain bounded per page; a server-side kind/breadcrumb contract is future work. |
| `(tabs)/search`, Map and overlays | InventoryMapQuery; shared AssetCoreQuery/AssetContentsQuery/AssetPhotosQuery | One cancellable scoped containment query; mounted warm-revisit, overlay core reuse/cancellation, inventory replacement and embedded-navigation reset tests pass. Native gesture verification remains a device check. |
| `assets/[assetId]/index`, nested location asset detail | AssetCoreQuery, AssetContentsQuery, AssetPhotosQuery | One selected asset GET before core; secondary queries reuse core. Independent loading and mutation request deduplication tested. Mounted delayed-region and changed-parent refresh tests pass; old/new ancestor mutation impacts are covered. |
| `add` | AddAssetContextQuery, focused ParentLookupRepository | Context and bounded candidate queries cached; 250 ms debounce, cancellation and warm reuse tested. Mounted draft regression verifies metadata refresh preserves edits and reads principal once; scope-keyed form owns draft restoration. |
| Asset `edit`, `move`, `move-here` | Shared AssetCoreQuery, independent tags, AssetPlacementQuery, bounded parent candidates | No photos or containment traversal for forms. Mounted dirty Edit/Move drafts survive background refresh and changed parent. Placement adapter and picker request budgets verified. Create/update now read only selected/old/destination ancestry before mutation; no post-write attachment reads. Parent promotion invalidation has regression coverage. |
| Asset `checkouts` | AssetCheckoutHistoryQuery and independent shared AssetCoreQuery | Focused scope discovery, cursor pages and retained rows on transient failure; mounted denial/failed-retry tests discard inaccessible rows. |
| Asset `history/index`, `history/[activityId]` | Scoped infinite pages and activity detail query; reversal observer | Mounted warm-page reuse, retained rows on failed refresh, scoped detail seeding/cancellation and denial/failed-retry tests pass. No secondary application cache. Reversal keeps unrelated core/photos fresh; containment invalidation remains conservative until full paths are returned. Uncached direct event links still scan all-event pages. |
| `settings/index`, `account`, `household/index`, `inventory/index` | Independent principal and scoped Settings queries | Mounted shared-read/progressive identity tests pass. Current scoped authorization discovery replaces stale permissions; denial and missing-inventory regressions hide cached controls. Customization context no longer reads principal. |
| Household/inventory `asset-types`, `fields`; inventory `tags` (index/new/resource) | Customization collections and editor ports | Scoped query-owned collections and independent editor choices; successful mutations reconcile affected resource families. Mounted denial/late-response, lifecycle rollback, dirty-draft and refreshed choice tests pass. Pages are cancellable and bounded with explicit incomplete feedback. |
| `settings/sharing`, `invitations/accept` | Invitation list/create/cancel/preview/accept ports | Sharing uses cancellable 50-row infinite pages and confirmed cancellation cache updates. Mounted warm-page, denial, failed-refresh and composition secret isolation pass. Acceptance stays workflow-owned; opening uses SelectInventoryCommand discovery refresh/cache reset. |
| `settings/voice` capability routes, profiles list/add/detail/prompt/credential | Independent tenant profile/configuration queries | Profile list/detail/edit do not request readiness. Successful mutations report tenant impacts; failure and unrelated-tenant tests pass. Mounted secure credential/duplicate submission tests use fakes and real queries; secret values stay out of cache. |
| `voice` realtime sheet | Session controller and approved action commands | Focused identity port removes startup workspace traversal. Preview shares scoped query; mounted inventory-switch recording disposal passes. Matching executed plans emit one scope impact; core/containment refresh conservatively for omitted promoted parents, photos/config stay fresh. Follow-up startup disposal regressions pass. |
| `settings/appearance`, `about`, `connection`, `diagnostics` | Local preferences/runtime diagnostics; shared identity for Diagnostics | Appearance/About/Connection need no account or inventory requests. Mounted About/Connection request-budget test passes. Diagnostics shares principal/scope to display IDs. Connection/auth commands remain workflow-owned. |

## Verified Source Checks

- Final source pass: mobile/API-client TypeScript checks, 160 mobile test files / 1055 tests and 34 API-client tests pass; mobile structural, script and dependency-age checks pass. This is source evidence, not a native build or end-to-end device check.
- `AssetCoreQuery.test.ts`, `AssetContentsQuery.test.ts`, `AssetPhotosQuery.test.ts`, and `ApiInventoryDetail.test.ts` cover the focused detail graph.
- `fetchMobileInventoryServerQuery.test.ts` uses a real QueryObserver to prove one post-mutation request, joining active work and reusing completed fresh work.
- `QueryClientInventoryMutationObserver.test.ts` covers parent reconciliation without cached child core and unaffected inventory/composition caches.
- `AssetDetailView.test.tsx` covers independent photo and contents progress/action availability. Mounted route tests use native-component fakes and real application queries/TanStack observers.

## Remaining Validation and Deliberate Limits

PR #44 owns build validation; no local application or native builds are permitted. A device run remains necessary to measure end-to-end latency, gestures, recording and native connectivity with the added Expo module. Source request-count tests establish reduced work, not measured wall-clock speedups.

Map still needs the complete active containment tree. Uncached direct activity links still scan cursor pages because the current API has no direct event lookup. Cross-page Browse breadcrumbs can repeat ancestor GETs; filtering and traversal have scan/cycle/cancellation guards, not an invented server contract. Legacy workspace compatibility methods remain outside the live focused screen graph.

Authoritative access failures discard the failed query's data; later failed retries cannot reveal it again. Mounted tests cover shared resources, Browse and History, recovery of failed scope discovery, and explicit context retry. Secrets and unsaved drafts remain owned by their workflows.

## Native Runtime Follow-up

The 0.16.3 device run exposed an unsupported `AbortSignal.throwIfAborted` call
that Node-based source tests had accepted. This blocked shared scope discovery,
Home and voice context. The fix uses a portable cancellation helper, runs the
mobile suite with React Native's actual bundled AbortController, and adds a
mounted native discovery/read regression plus a structural prohibition on the
unsupported method. All 1055 mobile tests pass; replacement native release
validation remains in CI.
