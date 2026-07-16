# Asset Checkout Spec

## Purpose

Stuff Stash must let users record that an asset has been taken out of its normal inventory place with the intent to return it.

Checkout is lightweight bookkeeping for real household use. It answers "why is this not where it belongs right now?" without changing the asset's home, containment, lifecycle, or ownership.

## Scope

This spec covers the first asset checkout and return slice:

- Checkout records for existing assets.
- Return records that close an open checkout.
- Current checkout state on asset reads and lists.
- Checkout history for an asset.
- Checked-out asset filtering and listing.
- REST, mobile, conversational voice, and MCP exposure requirements.
- Audit, authorization, observability, persistence, and tests.

This spec does not define reminders, due dates, recurring notifications, borrower accounts, reservations, partial quantity checkout, condition tracking, external borrower self-service, bulk checkout, or consumable usage.

## Domain Model

Checkout is an availability overlay on an asset.

Checkout must not:

- Move the asset.
- Change the asset's parent asset reference.
- Change the asset's home or normal location.
- Change the asset lifecycle state.
- Create a synthetic location, holder, borrower, or person resource.

The asset remains in its normal inventory structure. User interfaces must still show the asset in its current location or container, with a checked-out state.

The first domain concept is `AssetCheckout`.

Fields:

- `id`: ULID.
- `tenantId`: tenant security boundary.
- `inventoryId`: inventory scope.
- `assetId`: checked-out asset.
- `state`: checkout record state.
- `checkedOutAt`: application time when the checkout was created.
- `checkedOutByPrincipalId`: authenticated principal that performed checkout.
- `checkoutDetails`: optional freeform user note.
- `returnedAt`: optional application time when the checkout was closed.
- `returnedByPrincipalId`: optional authenticated principal that performed return.
- `returnDetails`: optional freeform user note.

`checkoutDetails` is intentionally broad. It may describe that the asset was loaned to someone, taken to a desk, moved temporarily into a work area, packed for a trip, or any other human context. The first slice must not create a separate holder model.

The first checkout states are:

- `open`: the asset is currently checked out.
- `returned`: the checkout was closed by a return operation.
- `undone`: the checkout operation was undone before a normal return.

## Invariants

- An asset may have at most one open checkout.
- An open checkout is one with state `open`.
- Checkout requires the asset to exist in the requested tenant and inventory.
- Checkout requires the asset to be active.
- Checkout requires a portable asset. Location assets are non-portable place
  concepts and checkout attempts for them must fail with a safe validation
  error at the application boundary.
- Checkout must fail when the asset already has an open checkout.
- Return requires an open checkout for the requested asset.
- Return must fail when the asset has no open checkout.
- Return may be performed by any authorized editor, not only the principal that checked out the asset.
- Returning an asset must close the existing checkout record. It must not create a second replacement checkout record.
- Checkout and return must use application time supplied by the injected clock port.
- Retrying checkout or return must not create duplicate open records. Until a general REST idempotency-key contract exists, duplicate checkout and already-returned return attempts may fail with safe conflict errors rather than pretending to be idempotent successes.

Archived assets:

- Archived assets must not be checked out.
- Returning an already checked-out asset after it has been archived is allowed only if the caller is authorized to edit the inventory and the asset still exists. Return closes availability history; it does not restore the asset.
- Asset archive behavior must not silently close an open checkout unless a future lifecycle spec explicitly defines that coupling.

Hard delete:

- Asset hard delete must fail for assets with open checkout records.
- Asset hard delete may remove terminal checkout records for the deleted asset only when the delete operation preserves audit records.
- The asset lifecycle spec must include these checkout constraints so hard-delete behavior has one source of truth.

## Application Operations

The asset application layer must expose checkout behavior through domain-oriented operations, not through persistence-specific update methods.

Required commands:

- `CheckoutAsset`
  - Inputs: tenant ID, inventory ID, asset ID, authenticated principal, optional checkout details, source, request ID when available.
  - Output: the checked-out asset or checkout record, the undoable operation ID, plus current checkout state needed by clients.
- `ReturnAsset`
  - Inputs: tenant ID, inventory ID, asset ID, authenticated principal, optional return details, source, request ID when available.
  - Output: the returned asset or checkout record, the undoable operation ID, plus current checkout state needed by clients.
- `UpdateReturnedCheckoutDetails`
  - Inputs: tenant ID, inventory ID, asset ID, checkout ID, authenticated principal, optional return details, source, request ID when available.
  - Output: the returned checkout record.
  - This operation may only update a checkout that is already in the `returned` state for the requested asset.
  - This operation must not reopen the checkout, move the asset, or change checkout/return principals or timestamps.
  - This operation must produce a distinct return-details-updated audit action, not a second asset-returned action.

Required queries:

- `GetCurrentAssetCheckout`: returns the open checkout for one asset, if present.
- `ListAssetCheckoutHistory`: returns `open`, `returned`, and `undone` checkout records for one asset.
- `ListCheckedOutAssets`: returns assets in one inventory that currently have open checkout records.

Asset detail, asset list, search result, location contents, root inventory contents, and mobile summary read models must include a compact current-checkout projection when an asset is checked out.

The compact projection must include:

- Checkout ID.
- Checkout state.
- Checked-out timestamp.
- Checked-out principal ID and safe resolved profile when available.

The compact projection must not include checkout details, return details, hidden tenant data, raw audit metadata, provider data, or fields from terminal checkout records.

Checkout details may be returned by asset detail and checkout history responses for principals with `inventory.view`. Asset list, search result, location contents, root inventory contents, and mobile summary projections must not include checkout details so broad browse responses do not leak freeform personal context.

## REST API

The first REST endpoints are:

- `POST /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/checkout`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/return`
- `PATCH /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/checkouts/{checkoutId}/return-details`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/checkouts`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/checked-out-assets`

Checkout request:

- `details`: optional string.

Return request:

- `details`: optional string.

Update returned checkout details request:

- `details`: optional string.

Details validation:

- Details are optional.
- Details must be trimmed.
- Empty details must be stored as absent.
- Details must have a bounded maximum length. The first implementation should use 1,000 Unicode scalar values unless another existing project limit is already used for comparable user notes.
- Details must be plain user text. They must not be interpreted as commands, borrower identity, JSON, markdown with special behavior, or authorization metadata.

Endpoint rules:

- All endpoints require authentication.
- Checkout and return require `inventory.edit_asset`.
- Updating returned checkout details requires `inventory.edit_asset`.
- Checkout history and checked-out listing require `inventory.view`.
- Endpoints must use standard Stuff Stash success and error envelopes.
- Endpoints must be represented in the generated OpenAPI contract.
- Endpoints must fail safely without revealing hidden cross-tenant, cross-inventory, or unauthorized asset existence.
- Read endpoints must be cursor paginated where they can return multiple records.
- Checkout history ordering defaults to newest checkout first using `(checkedOutAt desc, id desc)` or an equivalent deterministic order. History cursors must be scoped by tenant, inventory, asset, and ordering.
- Checked-out assets ordering defaults to newest checkout first using `(checkedOutAt desc, assetId desc)` or an equivalent deterministic order. Checked-out asset cursors must be scoped by tenant, inventory, lifecycle filter if present, and ordering.
- Checked-out asset listing must include assets with open checkout records regardless of asset lifecycle by default so archived checked-out assets do not disappear from return workflows. Responses must include lifecycle state. A future lifecycle filter may narrow this list only if the default remains safe for open-checkout recovery.

Asset read responses:

- Asset detail responses must include current checkout state when open.
- Asset list responses should include current checkout state when available without requiring clients to make per-asset checkout calls.
- Location/container contents must include current checkout state for contained assets.
- Existing clients that do not understand checkout state must still be able to render assets normally.
- Checkout and return mutation responses must include the undoable operation ID so clients can present immediate undo flows without guessing from audit history.

## Search

Search must support checkout-aware discovery.

The first search extension is a `checkoutState` filter on asset search:

- `any`: default; includes checked-out and not-checked-out assets.
- `checked_out`: returns only assets with an open checkout.
- `available`: returns only assets without an open checkout.

Search results must include current checkout state when present.

Search filtering must preserve tenant and inventory authorization rules. The search repository must receive authorized inventory candidates from the application service and must not decide authorization itself.

## Web UX

Checkout and return are primary web asset maintenance actions and must use the same domain semantics as mobile, REST, conversational flows, and MCP.

Web requirements:

- User-facing command labels must use the verb phrase `Check out`; the noun `checkout` remains appropriate for status and history labels.
- Checkout from asset detail must be one visible action when the asset is active and not currently checked out.
- Return from asset detail must be one visible action when the asset is currently checked out.
- Checkout and return details must be optional and must not block the fastest checkout or return path.
- Asset cards, recent assets, search results, location contents, checked-out browsing surfaces, archived asset rows, and asset detail must show a compact checked-out indicator when an asset has an open checkout.
- The checked-out asset projection must include the same safe primary-photo
  summary used by ordinary asset lists so Home and checked-out browsing can
  recognize the asset without issuing per-asset detail requests.
- The checked-out indicator must not hide the asset's normal location or imply that the asset moved.
- The selected-inventory browsing experience must provide a checked-out surface or filter that includes assets with open checkout records regardless of lifecycle.
- On a dedicated checked-out surface, the open checkout indicator is the primary status. An `active` lifecycle badge must be suppressed as redundant; an `archived` lifecycle badge must remain visible because it changes recovery expectations.
- The asset detail workspace must expose checkout history through the same safe-history design principles used for audit history: compact rows, safe actor labels, timestamps, and bounded details.
- Checkout-aware search filters must be available from the web search experience without requiring users to know API query parameters.
- Web Home must show each checked-out asset as a compact photo-first row. An
  editor sees a trailing `Return` action with an asset-specific accessible name;
  activating it returns the asset immediately with no confirmation or required
  details. Viewers do not see the mutation action. The row disappears after a
  successful return and the normal saved feedback remains perceivable.

## Mobile UX

Checkout and return are primary mobile asset maintenance actions.

Mobile requirements:

- A recently used, searched, scanned, or browsed asset must be check-outable from a fresh app launch in no more than three primary user actions after the app is ready.
- Checkout from asset detail should be one visible action when a portable asset is active and not currently checked out.
- Return from asset detail should be one visible action when the asset is currently checked out.
- Checkout details must be optional and must not block the fastest checkout path.
- After checkout, mobile may offer `Add details` as a secondary follow-up action.
- Return details must be optional and must not block the fastest return path.
- Asset cards, recent assets, search results, location lists, map rows, and asset detail must show a compact checked-out indicator when an asset has an open checkout.
- The checked-out indicator must not hide the asset's normal location or imply that the asset moved.
- The selected-inventory browsing experience must provide a checked-out filter or surface.
- The mobile Home screen must show a checked-out row directly under the recently changed row using the same card mechanics as recent assets. When no assets are checked out, this row must collapse into a compact, low-prominence empty state that does not consume recent-asset-card vertical space.
- Checked-out cards on the mobile Home screen must include a one-tap `Return` action. Pressing it must immediately return the asset, then open a native sheet for optional return details. Saving the sheet updates the returned checkout details. Canceling the sheet undoes the return through the returned undoable operation ID, restoring the asset to checked-out state.
- The mobile asset detail workspace must expose checkout history through the same safe-history design principles used for audit history: compact rows, safe actor labels, timestamps, and bounded details.

Permission behavior:

- Viewers may see checkout state and checkout history.
- Viewers must not see checkout or return mutation actions.
- Editors may check out active non-location assets when domain invariants allow it.
- Editors may return any existing asset with an open checkout, regardless of lifecycle, when domain invariants allow it.

## Conversational Voice And Text

Checkout and return must be first-class conversational inventory actions.

Supported intents include:

- "Check out the socket set."
- "I'm taking the drill out."
- "Mark the tent as checked out."
- "Return the Dune paperback."
- "What is checked out?"
- "Show checkout history for the ladder."

Conversational checkout and return are state-changing actions and must use structured action plans.

Rules:

- Voice or text may initiate checkout and return.
- Execution requires explicit client UI confirmation, such as a button or equivalent accessible control.
- Spoken "yes" must not be the approval path for checkout or return in the first slice.
- The confirmation control must describe the planned action and target asset in human language.
- The client must support cancellation from the confirmation point.
- The agent must ask a clarifying question when the asset reference is ambiguous.
- The agent must not invent asset IDs or checkout details.
- Optional details may be captured from the user's utterance only when they are clearly user-provided context, such as "Loaned to Alex" or "using at my desk".
- The action plan must preserve details as bounded structured command arguments, not as lossy natural language.
- Plan execution must call the same checkout and return application services used by REST and mobile UI.
- Execution must re-authorize the initiating principal at execution time.

The action-plan command enumeration must include:

- `checkout_asset`
- `return_asset`

`checkout_asset` arguments:

- `assetId`: required existing active asset ID in the same inventory.
- `details`: optional bounded user text.

`return_asset` arguments:

- `assetId`: required existing asset ID in the same inventory with an open checkout.
- `details`: optional bounded user text.

Realtime events must not expose raw transcripts, raw prompts, raw model responses, provider credentials, or hidden inventory data.

## MCP And Agent Tools

MCP must expose checkout state and history through the same application boundary as REST and conversational flows.

First read tools:

- Search authorized assets with checkout state.
- Get asset detail with current checkout state.
- List currently checked-out assets.
- List checkout history for an authorized asset.

Write-capable tools:

- Check out asset.
- Return asset.

Write-capable MCP tools must follow the MCP action-plan and confirmation rules. External MCP clients must not directly execute checkout or return unless a future spec defines an approved action-plan or confirmation-token contract for low-risk writes.

Internal conversational agent tools may propose checkout and return action plans, but they must not execute the mutation before explicit client UI approval.

## Audit

Checkout and return are audited state-changing actions.

Audit action types:

- `asset.checked_out`
- `asset.returned`

Audit source must reflect the initiating adapter, such as `api`, `conversation`, or `mcp`.

Checkout audit metadata may include:

- `checkout_id`
- `asset_id`
- `details_present`

Return audit metadata may include:

- `checkout_id`
- `asset_id`
- `details_present`
- `checked_out_by_principal_id`

Audit metadata must not include raw provider prompts, raw transcripts, raw model responses, bearer tokens, hidden inventory data, or unbounded details. Checkout and return details live on the checkout record, not in audit metadata.

Conversational audit records must reference the action plan and approval when available.

## Undo And Redo

Checkout and return must be undoable through the existing undoable-operation model.

Undoable operation creation must be atomic with checkout and return mutations.

Undo behavior:

- Undoing `asset.checked_out` must set the original checkout record state to `undone` without setting `returnedAt`, `returnedByPrincipalId`, or `returnDetails`.
- Undoing `asset.checked_out` must fail if the checkout was already returned by a later return operation.
- Undoing `asset.returned` must reopen the same checkout record by clearing `returnedAt`, `returnedByPrincipalId`, and `returnDetails`.
- Undoing `asset.returned` must fail if the asset has another open checkout, if the checkout record no longer exists, or if later checkout history makes the saved operation stale.

Redo behavior:

- Redoing an undone `asset.checked_out` must reopen the original checkout record when normal checkout validation still passes.
- Redoing an undone `asset.returned` must reapply the original return fields to the same checkout record.
- Redo must fail if the target asset no longer exists, if the checkout record no longer exists, or if later checkout or return operations make the saved before/after state stale.

Stale-operation predicates:

- Undoing `asset.checked_out` requires the saved checkout ID to exist, to belong to the same tenant, inventory, and asset, and to still be in state `open` with the saved checkout fields.
- Redoing `asset.checked_out` requires the saved checkout ID to exist, to still be in state `undone`, to have no return fields, and to have no later checkout record for the same asset.
- Undoing `asset.returned` requires the saved checkout ID to exist, to still be in state `returned`, and to have the same `returnedAt`, `returnedByPrincipalId`, and `returnDetails` captured by the undoable operation.
- Redoing `asset.returned` requires the saved checkout ID to exist, to still be in state `open`, and to have no later checkout record for the same asset.
- Any checkout record for the same tenant, inventory, and asset with `checkedOutAt` after the saved checkout record's `checkedOutAt`, or with the same timestamp and a greater checkout ID, counts as later checkout history.

Validation:

- Undo and redo must require `inventory.edit_asset` for the operation inventory.
- Undo and redo must not bypass checkout invariants, tenant isolation, inventory isolation, authorization, audit history, or observability.
- Undo and redo must produce their own audit records through the existing undo/redo action types.
- Undoable operation snapshots may include safe checkout fields needed for compensation: checkout ID, asset ID, timestamps, principal IDs, and bounded details. Snapshots must not include provider prompts, raw transcripts, credentials, bearer tokens, authorization internals, or hidden tenant data.

## Persistence

The first durable persistence shape must use an `asset_checkouts` table or equivalent repository-backed storage.

Required columns:

- `id`
- `tenant_id`
- `inventory_id`
- `asset_id`
- `state`
- `checked_out_at`
- `checked_out_by_principal_id`
- `checkout_details`
- `returned_at`
- `returned_by_principal_id`
- `return_details`
- created and updated timestamps as needed by the repository adapter.

Persistence rules:

- Rows must be scoped by tenant ID and inventory ID.
- The asset reference must be in the same tenant and inventory.
- The database shape must defensively prevent more than one open checkout per asset where the backend supports a partial unique index or equivalent constraint. For PostgreSQL, this should be a partial unique constraint on `(tenant_id, inventory_id, asset_id)` where `state = 'open'`.
- Repository adapters must also enforce the one-open-checkout invariant before commit so memory adapters and database adapters behave consistently.
- Repository reads for checkout state must require tenant ID and inventory ID.
- Repository reads for asset-specific checkout history must require tenant ID, inventory ID, and asset ID.
- Persistence models are infrastructure-only data mappers. They must not become domain entities.

Transactions:

- Checkout must atomically create the checkout record, write audit history, write undoable-operation metadata, and record the terminal application outcome.
- Return must atomically close the checkout record, write audit history, write undoable-operation metadata, and record the terminal application outcome.
- Action-plan execution for checkout and return must atomically apply the checkout mutation and transition the action plan to its terminal state.

## Authorization

Checkout and return require `inventory.edit_asset`.

Checkout state and history reads require `inventory.view`.

Authorization must be checked at the application boundary for every REST, mobile, conversational, MCP, CLI, worker, or future adapter path.

The checkout repository must not decide authorization. It must receive tenant, inventory, and asset scope from already-authorized application services.

Security-sensitive tests must cover:

- Unauthenticated checkout, return, asset read/search/list checkout projections, history, and checked-out list rejection.
- Viewer read success.
- Viewer checkout and return denial.
- Editor checkout and return success.
- Wrong-tenant and wrong-inventory denial.
- Hidden asset safe not-found behavior.
- Cross-tenant asset ID attempts.
- Attempts to return an asset checked out in another inventory.
- Attempts to smuggle principal IDs, roles, checkout actor IDs, return actor IDs, or authorization hints through request bodies.
- Unauthenticated checkout/return undo and redo rejection.
- Viewer checkout/return undo and redo denial.
- Wrong-tenant, wrong-inventory, hidden-operation, cross-tenant operation ID, and stale-operation ID undo and redo denial.
- Authorized editor checkout/return undo and redo success.

## Observability

Checkout and return must record domain-oriented observability through injected ports.

Events should cover:

- Checkout requested.
- Checkout succeeded.
- Checkout rejected because the asset is missing, archived, unauthorized, or already checked out.
- Return requested.
- Return succeeded.
- Return rejected because the asset is missing, unauthorized, or not checked out.
- Checkout history listed.
- Checked-out assets listed.

Observability metadata must be safe and bounded. It may include tenant ID, inventory ID, asset ID, checkout ID, source, outcome category, and latency. It must not include raw details, provider prompts, transcripts, credentials, tokens, or hidden resource data.

## Testing

Tests must use fakes rather than mocks.

Domain and application tests must cover:

- Successful checkout.
- Successful return by the same principal.
- Successful return by a different authorized editor.
- Duplicate open-checkout rejection.
- Return-without-open-checkout rejection.
- Checkout of archived asset rejection.
- Checkout and return preserving asset parent, location, lifecycle, title, description, custom fields, and custom asset type.
- Checkout details and return details trimming, empty normalization, and length validation.
- Current checkout projection on asset reads.
- Checkout history ordering and pagination.
- Checked-out asset listing ordering and pagination.
- Audit history for checkout and return.
- Undo and redo for checkout and return, including stale-operation rejection after later checkout history changes.
- Checkout history presentation distinguishing `undone` from `returned`.
- Safe observability outcomes.

REST boundary tests must cover:

- Endpoint success envelopes and safe error envelopes.
- Generated OpenAPI coverage.
- Adversarial authentication and authorization failures.
- Tenant and inventory isolation.
- Cross-tenant and cross-inventory asset ID attempts.
- Viewer read success and write denial.
- Editor write success.

Persistence tests must cover:

- One-open-checkout invariant in memory and GORM adapters.
- Defensive tenant and inventory scoping.
- Atomic checkout plus audit writes.
- Atomic return plus audit writes.
- Atomic checkout and return undoable-operation writes.
- Cursor pagination for asset checkout history and checked-out asset listing.

Conversational and MCP tests must cover:

- Voice/text plan proposal for checkout and return.
- Button-backed approval and cancellation.
- Rejection of spoken approval as execution authority in the first slice.
- Ambiguous asset clarification.
- Unauthorized execution rejection at action-plan execution time.
- MCP read tools for current checkout state and history.
- MCP write tool rejection before approved action-plan or confirmation-token support.
- No raw transcripts, prompts, provider responses, bearer tokens, or credentials in durable records, audit, observability, or client events.

## Open Questions

- Should future reminders be based on due dates, recurring prompts, or user-created follow-up tasks?
- Should checkout support quantities once consumables or multi-count assets are specified?
- Should an external borrower/contact model ever exist, or should checkout details remain the long-term holder mechanism?
- Should checkout history be included in export formats by default?
