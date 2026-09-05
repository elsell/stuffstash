# Mobile Server State And Responsiveness Spec

## Purpose

Stuff Stash mobile must feel immediate on real devices even when a self-hosted API has noticeable network latency. The client must request only the server state needed by the active surface, reuse trustworthy state during a session, and make successful actions visible without waiting for unrelated workspace hydration.

## Scope

This specification applies to every mobile surface and every API-backed interaction under `apps/mobile`. It defines the shared server-state architecture, port boundaries, request-shape expectations, freshness rules, mutation coordination, loading presentation, and verification required across:

- connection onboarding and authenticated session establishment;
- Home and tenant/inventory switching;
- Browse list, search, filters, Places, and Map;
- asset detail, contained assets, checkout history, activity history, edit, move, lifecycle, undo, and photo workflows;
- Add, parent lookup, draft tag resolution, and attachment upload;
- invitation preview/acceptance and inventory sharing;
- Settings account and selected-scope reads;
- tenant/inventory customization collections and editors;
- provider-profile and voice settings;
- realtime voice actions that mutate or reveal cached inventory state.

Device-local preferences, secure credentials, in-progress Add drafts, native picker state, and live realtime voice/audio streams are client state rather than server state. They remain behind their existing application ports and must not be forced into the server-state cache.

## Decisions

### Ports And Adapters

- Domain and application packages remain independent of React, React Native, Expo, and TanStack Query.
- Generated API DTOs and the generated API client remain confined to API adapters. UI modules consume application view models and commands only.
- Each application read port and command port must express one cohesive product responsibility. A screen-specific read must not obtain its identity by hydrating unrelated tenants, inventories, assets, locations, tags, attachments, or thumbnails.
- The broad `ApiInventorySummaryRepository` must be decomposed behind focused application ports. Compatibility facades may exist during migration, but migrated operations must not route back through broad workspace hydration.
- Selected tenant and inventory identity is session scope. API adapters may reuse an already-resolved selected scope, but the source of that identity must remain replaceable and must reset when the principal, connection profile, tenant, inventory, or mobile composition changes.
- Every inventory-selection path, including invitation acceptance, must use one application command that refreshes stale inventory discovery when necessary and resets the composition-scoped query cache only after selection succeeds.
- TanStack React Query is a UI-side server-state adapter. Query functions call application queries and commands; they do not call generated clients or API repositories directly.
- The mobile composition owns one Query Client for one authenticated service-scope identity. Sign-out, session expiry, profile replacement, or composition replacement must cancel active work and clear state so data cannot cross principals or Stuff Stash instances.
- Query cancellation must reach cancellable API work. A superseded route, search, scope, or composition must not populate a cache entry after it is no longer authoritative.
- Caller cancellation and the mobile network timeout must remain independently effective. A caller cancellation must remain distinguishable from a timeout, and supplying a query cancellation signal must never disable the timeout.

### Query Identity And Freshness

- Query keys are centralized, typed, and begin with the mobile composition scope. Relevant tenant ID, inventory ID, asset ID, filter state, sort, pagination cursor, and resource ID must be represented explicitly.
- Query keys must normalize semantically equivalent inputs such as empty search text, duplicate tag IDs, and omitted default filters.
- Initial reads may use cached data while a stale refresh runs. A warm revisit must not replace usable content with a full-screen loading state.
- Pull-to-refresh explicitly refetches the active surface and exposes native refresh state without blanking its last successful content.
- App foreground and network-reconnect behavior may refetch active stale queries. It must not refetch every inactive query or create duplicate focus requests.
- Concurrent consumers of the same key share one request. Route focus effects must not independently duplicate a query already owned by the server-state coordinator.
- Foreground mutation reconciliation joins an active invalidation request or reuses its fresh result if it already finished. Explicit user refresh may force a read; mutation completion must not force a second read.
- Pagination caches preserve page identity and merge only compatible scopes. A result from an obsolete search, filter, tenant, inventory, or cursor must never overwrite the current surface.
- Browse uses an infinite query per normalized applied filter set. Query-owned cursors and cancellation replace manual page arrays and request-sequence arbitration. Replacing filters preserves the last successful same-inventory result with progress/error feedback; switching inventories immediately removes prior results and picker metadata. Pagination cannot append while showing a previous filter result.
- Browse header identity and create permission come from a focused selected-inventory context read; List and Map entry must not fetch the Places catalog for header metadata. Places summaries may load independently of paginated location results.
- Browse/search breadcrumb resolution loads only missing ancestors of the returned page, deduplicating shared ancestors. It must not traverse the complete inventory. A current result's null parent is authoritative inventory-root placement and must not be replaced by linkage from an unrelated or older tree projection.
- Inactive Browse surfaces unsubscribe from their queries, cancelling in-flight work only when no other consumer needs it. Secondary Places summary failures keep result rows available with a non-blocking error and targeted retry.
- Map owns one selected-inventory containment query, forwards cancellation through its traversal, and preserves usable columns during stale refresh. Initial mounting and route focus must not create consecutive duplicate traversals. Map overlays share the same progressive asset core/contents/photos keys as ordinary details; changing an overlay selection cancels obsolete secondary work. Touch and voice mutations reconcile the Map query alongside affected detail regions.
- Asset create/update load only the selected asset (for updates) and old/destination ancestor chains before writing, preserving full mutation impact without inventory traversal. Creating under an item also reports the parent promotion so its core and containment refresh together. Mutation responses map directly without a follow-up attachment listing that could make a successful write appear failed. Home checkout cards use API primary-photo metadata, as recent cards do; attachment enumeration belongs to asset photos.
- Add and Move share a cancellable, inventory-scoped parent-candidate query through a focused parent lookup port. Empty input reads five recent active assets; nonempty input searches at most fifty candidates before exact-title prioritization and six-result display. Requests hydrate only candidate ancestors, never the full inventory graph. Typing is debounced for 250 ms; superseded and hidden pickers unsubscribe and cancel unused work. Candidate keys normalize whitespace, and successful asset mutations invalidate relevant picker results.
- Edit and Move reuse the selected asset core. Edit tag options load independently with visible retry; draft values survive background refresh and initialize only once per asset/scope. Move loads placement separately when current ancestry is needed; photos and unrelated contents must not block editing. Lookup loading/error state must not present stale results as matches for new input.
- Asset activity and checkout history use independent scoped infinite queries and cancellable cursor reads. History detail reuses sanitized entries from the same scoped cached page; application queries must not own a second unbounded entry cache. A direct activity link without a cached page may scan the scoped all-events cursor endpoint, cancelling when superseded; a future direct-entry endpoint can eliminate that fallback scan.
- Undo/revert adapters report successful mutation impacts through the shared observer. History, checkout history, relevant detail regions and lists reconcile after reversals; failed reversals do not invalidate authoritative data. Reversal responses identify the affected asset, preserving unrelated core, photo and History caches. Because reversal does not return both complete containment paths, cached containment regions conservatively reconcile across that inventory until the API supplies those paths. Checkout history discovers only selected scope, never a full inventory summary.
- Thumbnail and attachment-resource caching remains identity-specific and bounded. Metadata may render before media materialization; invisible media must not block core content.

### Mutations And Visible Feedback

- A mutation must use the selected scope already known to the active composition. It must not perform full workspace hydration merely to discover tenant or inventory IDs.
- Mutation responses update every directly represented cache entry that can be made authoritative from that response. Related entries that cannot be updated safely are invalidated narrowly by resource and scope.
- Mutation completion must not await broad invalidation refetches. Background reconciliation may continue after the user sees success.
- Safe optimistic updates are required where the inverse is deterministic and rollback can restore the exact previous cache state. Destructive actions, permission-sensitive transitions, or operations whose server result supplies important identity may wait for the mutation response while still showing immediate pending feedback.
- Optimistic updates must cancel conflicting reads, snapshot affected entries, roll back on failure, and reconcile with the server response. They must not manufacture permissions, audit identifiers, checkout identifiers, or undo identifiers.
- Successful checkout, return, edit, move, lifecycle, photo, customization, sharing, invitation, provider-profile, and approved voice mutations must update or invalidate every visible affected surface, including Home, Browse, Map, detail, history, and Settings where applicable.
- Return must issue the scoped return mutation without preceding workspace hydration. After success, the return-details sheet must open before any dashboard reconciliation. Updating details or cancelling the return must preserve truthful pending, success, error, and undo behavior.

### Loading And Native Presentation

- Initial full-screen loading is reserved for a surface with no usable cached or route-provided content.
- A surface with usable cached data keeps it visible during refetch and uses native pull-to-refresh, a compact inline progress treatment, or a subtle stale indicator as appropriate.
- Skeletons are appropriate for stable content geometry such as Home rows, Browse cards, asset identity, Settings rows, and editor field groups. Skeletons must be accessible as one polite busy status, must not expose decorative elements to assistive technology, and must honor reduced-motion settings.
- Native activity indicators remain appropriate for indeterminate commands, compact modal work, authentication handoff, pagination footers, and layouts whose final geometry is not predictable.
- Independent secondary regions load independently. Asset identity must not wait for attachments or checkout history; Settings navigation must not wait for voice readiness; attachment metadata must not wait for thumbnails.
- Asset contents and photos consume the selected core snapshot without repeating its GET. Contents identity includes kind, lifecycle, and parent; unrelated core changes do not discard usable secondary data. Photo loading gates photo additions independently of placement/contents loading.
- Contained rows retain primary-photo references from containment metadata; constructing a thumbnail URL is not a media download. Native visible rows load those images without attachment-list requests. Explicit refresh retries photos even if placement changed.
- Containment reconciliation follows cached ancestor trails and old/new parent dependencies, including ancestor recursive contents and descendant breadcrumbs. A newly created child must reconcile cached ancestor contents even before that child's core is cached.
- Every action provides visible pressed or pending feedback immediately. Network completion must not be the first visible acknowledgement of a tap.
- Errors shown while cached content remains usable are non-blocking and use the shared feedback surface. Initial-load errors retain a direct retry path and safe, curated copy.

## Surface Audit Contract

The migration is complete only when each surface family below has an explicit application port, centralized query identity, loading/freshness behavior, mutation impact mapping, and focused tests:

| Surface family | Required server-state shape |
| --- | --- |
| Onboarding and session | Uncached reachability/authentication commands plus narrowly cached authenticated metadata; composition replacement clears all prior server state. |
| Home and switcher | Selected context, bounded recent assets, bounded checked-out assets, and required tag metadata; no full attachment or full-map hydration. |
| Browse List and Search | Keyed paginated results for the applied route filters; last successful results remain visible during replacement and superseded searches cancel. |
| Browse Map and Places | Complete active containment metadata only when Map requires it; Places and list browsing do not inherit the Map fetch. Visible media hydrates separately. |
| Asset workspace | Core asset first; attachments, thumbnails, checkout history, activity, and contained children as independent queries according to the active region. |
| Add and movement | Cached selected scope plus focused parent/tag choices; drafts and native photo selections stay local; success updates affected lists and detail entries. |
| Checkout and return | One scoped mutation after scope resolution, immediate pending feedback, response-driven checkout/undo identity, and narrow Home/detail/history reconciliation. |
| Activity and undo | Asset-scoped paginated history; reversal updates the affected asset and invalidates only relevant history/list entries. |
| Sharing and invitations | Inventory-scoped invitation lists and mutations; invitation secrets are never persisted in a general query cache beyond the one-time workflow that owns them. |
| Settings and customization | Selected-scope identity shared across screens; each collection/editor loads only its resource family and mutations reconcile that family plus dependent pickers. |
| Provider and voice settings | Tenant-scoped profiles/readiness split so readiness cannot block unrelated Settings; secret credential values never enter the query cache. |
| Realtime voice | Live stream state remains controller-owned; completed approved actions reconcile the same typed resource keys used by touch workflows. |

## Request Budgets

- A return tap with selected scope already resolved performs no discovery, asset-list, tag, attachment, thumbnail, or Map request before the return mutation.
- Showing the return-details sheet waits for at most the return mutation. Dashboard reconciliation is non-blocking.
- A warm Home revisit with fresh cache performs no blocking API request. A stale Home revisit preserves content while deduplicated component queries refresh.
- Asset core presentation waits for one asset-detail request after selected scope is known. Attachments, history, contained children, and thumbnails do not extend that critical path.
- A repeated Browse query with an identical normalized key reuses its fresh result. Changing search or filters performs only the new result request and any separately required visible-media work.
- Selecting a different inventory resolves that inventory's focused surface data without hydrating full data for every other visible inventory.
- Tests must assert request categories and ordering rather than accepting broad total-count regressions hidden by concurrent execution.

## Dependency And Lifecycle Configuration

- Mobile uses exactly pinned `@tanstack/react-query` `5.101.4` for server-state coordination.
- The Query Client default policy must be declared in one bootstrap-owned module and covered by tests. Individual surfaces may override freshness only with a product-specific reason represented in tests.
- Automatic retries must not multiply mutations. Read retries must be bounded and must not retry authentication, authorization, validation, or missing-resource failures.
- Garbage-collection and stale durations must be finite. Secret-bearing one-time results must use zero retention or remain outside the general cache.
- React Native app-focus and connectivity signals must be bridged to TanStack Query without adding an unpinned third-party connectivity dependency merely for cache coordination.

## Verification

- Every migrated query and mutation has application-layer tests using fakes and UI/query-adapter tests using controlled clients rather than mocks of internal implementation details.
- Request-spy fakes prove required call categories, ordering, deduplication, cancellation, and the budgets in this spec.
- Delayed controlled fakes prove that core content and successful mutation feedback are not blocked by secondary refreshes.
- Cache tests prove tenant, inventory, principal, connection-profile, route-filter, and composition isolation.
- Mutation tests prove optimistic rollback where used and complete affected-key reconciliation for successful touch and voice workflows.
- Loading tests cover initial, cached-refetch, pull-to-refresh, independent-region, empty, and safe error states, including accessibility and reduced motion where skeletons animate.
- `pnpm --dir apps/mobile test` and `pnpm --dir apps/mobile check` must pass without requiring a native build.
- The structural mobile check must reject generated API client imports in UI and direct React/TanStack imports in domain or application packages.
- Completion evidence includes a maintained surface audit table in the test suite or a generated verification report showing every route family mapped to its query/mutation contract.

## Settings and administration read ownership

Settings scope and principal identity have separate query lifetimes. Shared navigation renders when scope is ready while account identity continues loading; diagnostics and appearance stay local. Settings authorization scope refreshes the selected tenant and inventory from scoped API reads, rather than trusting session-long cached permissions. Customization context reads scope only and must not request principal identity. Every server read accepts the shared cancellation request through its application port and transport adapter. Refresh failures retain usable cached data with an explicit retry notice. Authorization denials remain authoritative and remove controls that require the denied permissions.

Provider profile summaries and voice configuration/readiness use shared tenant-scoped server state. Credentials remain local ephemeral form values and are never cached. Profile/credential/test/configuration mutation success invalidates the affected tenant's provider summary and voice readiness queries through an injected observer; mutation failures do not pretend the server changed.

Customization collections and editor initialization share scoped query-owned definition lists, including lifecycle and inheritance in their identity. Editors keep an editable snapshot as workflow state; background read updates must not overwrite dirty values. Cached collections update after successful definition mutations. Lifecycle switching keeps the previous committed tab and rows until the destination query succeeds, and rolls back on failure. The existing bounded collection loader remains capped at fifty pages with incomplete-list feedback, and forwards cancellation for each page. Context reuses the Settings scope cache. Permission failures hide collections and preserve populated editor drafts read-only until access is explicitly refreshed.

## Invitation Read Ownership

Sharing lists use scoped TanStack infinite queries with 50 safe summaries per page. Continuation loads only on request; cancellation reaches the HTTP adapter. Repeated cursors fail safely. Successful create/cancel commands notify an injected invitation mutation observer, which invalidates only the affected inventory list. One-time links remain local to the mounted, permission-qualified scope and never enter retained queries. Permission denial hides previously cached rows and links; transient refresh failures retain rows with retry feedback. Accept/preview are token-bearing workflows and remain outside retained query state. Opening an accepted inventory refreshes discovery and resets the selected-scope cache.

## Voice Context and Mutation Reconciliation

Voice preview and recording startup use a focused inventory-context port (tenant/inventory identity and names), never asset, attachment or tag enumeration. Preview is a scoped shared query; realtime audio and approval state remain controller-owned. Inventory replacement cancels and resets the active voice workflow. The voice parent picker shares the debounced, cancellable parent-candidate query.

Only an executed plan matching the active reviewed plan reports an injected voice mutation impact, before optional photo uploads complete. One batched impact invalidates inventory projections and affected asset core/history/checkout reads. Core and containment/placement refresh conservatively because execution results omit automatically promoted parents and old/destination ancestry; incomplete result identifiers conservatively refresh asset projections. Unrelated inventories, configuration and photos remain fresh. Photo uploads report their own successful attachment impacts.
