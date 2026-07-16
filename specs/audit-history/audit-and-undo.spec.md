# Audit And Undo Spec

## Purpose

Stuff Stash must keep an audited history of inventory actions.

Users should be able to understand what changed, who changed it, and undo supported actions when that is safe.

## Scope

This spec covers audit history, the first undo/redo slice, and requirements for future undoable work.

This spec does not define the final event store, retention policy, UI timeline, or every undoable command.

The first implementation slice defined durable, read-only audit records.

The second implementation slice defines undo and redo for a narrow set of asset actions.

## Requirements

- Every state-changing action must produce an audit record.
- List, detail, and content read endpoints must produce safe read audit records where specified by the REST and lifecycle specs.
- Audit records must include the authenticated principal, tenant, inventory, action type, target resources, timestamp, source adapter, and request identifier where available.
- Audit records must store only the stable principal ID for the actor. Audit read responses may resolve that ID to a safe user profile from the identity store, while preserving `principalId` as the durable audit identity and fallback. User profile resolution must be best-effort and must not make audit history unreadable.
- Audit records must distinguish between direct user actions, conversational actions, MCP actions, imports, background jobs, and system actions.
- Conversational audit records must reference the action plan and approval when available.
- Audit records must not store raw provider prompts, raw model responses, raw audio, secrets, tokens, or sensitive data beyond what the audit use case requires.
- Audit behavior must be behind ports and adapters.
- Audit writing must be part of the application operation boundary.
- Audit records must preserve tenant and inventory isolation.
- Audit records must be append-only through application ports. The first slice must not expose update or delete behavior for audit records.
- Append-only audit persistence must reject duplicate audit record IDs instead of overwriting existing records. This applies to durable adapters and in-memory adapters used for local runs and tests.
- Audit records must be cursor paginated.
- Audit pagination must be ordered by `(occurredAt, id)`, not by ID alone, so same-millisecond records remain deterministic without relying on monotonic ID generation.
- Inventory audit reads must require `inventory.view` for the target inventory.
- Tenant audit reads without an inventory scope must require `tenant.configure`.
- Asset audit history reads must require `inventory.view` for the asset's inventory and should expose a bounded asset-scoped read endpoint so clients do not need to scan broader inventory audit pages.
- The append-only audit stream and the homeowner-facing asset activity projection are separate read models. Filtering activity must never delete, rewrite, or suppress records in the raw audit stream.
- Asset activity must default to meaningful state changes so routine read audit records cannot bury a recent edit, move, lifecycle action, checkout action, return action, undo, or redo.
- Asset activity must support an explicit `all` view for authorized technical review of every asset-targeted audit record.
- Asset activity must be cursor-paginated newest-first by `(occurredAt, id)`. Cursor scope must include tenant ID, inventory ID, asset ID, and activity view.
- Asset activity reads must require `inventory.view`, verify that the asset belongs to the requested tenant and inventory, and return the same safe not-found behavior used by asset detail reads.
- Audit read responses must use the standard API success and error envelopes.
- HTTP adapters must capture `X-Request-ID` for audited requests when the client supplies it and store it on emitted audit records.

## First Slice Record Shape

The first durable record must include:

- `id`: generated application ID.
- `tenantId`: tenant security boundary.
- `inventoryId`: inventory scope when the action is inventory-scoped. Tenant-scoped records leave this empty.
- `principalId`: authenticated principal or system principal responsible for the action.
- `action`: typed action name.
- `source`: typed action source.
- `targetType`: typed target resource type.
- `targetId`: target resource ID.
- `occurredAt`: server timestamp.
- `requestId`: optional request identifier when a transport provides one, such as the REST `X-Request-ID` header.
- `metadata`: safe JSON object for small, non-secret context.

Metadata is for human context and filtering hints. It must not be treated as an authorization source.

## Asset Activity Projection

The first product-facing asset History surface uses a typed activity projection over durable audit records. The projection is an application query; it is not a second event store and must not become an authorization source.

The first activity response includes:

- `id`: the durable audit record ID.
- `action`: the typed audit action.
- `category`: `change` or `read`.
- `principalId` and the same best-effort safe resolved principal used by audit reads.
- `source`.
- `occurredAt`.
- `requestId` when present.
- `changes`: a bounded ordered list of safe structured change summaries.
- `undo`: optional target-scoped undo information containing the operation ID and status when the audit record references an existing operation in the same tenant and inventory.
- `technicalMetadata`: a safe allowlisted subset for an explicit technical-details disclosure. Unknown metadata remains omitted.

The first structured change fields are:

- `title`, with previous and current values when recorded.
- `description`, represented only as changed without storing or returning description contents.
- `tags`, with previous and current counts when recorded.
- `parent`, with previous and current parent asset IDs when recorded.
- `lifecycle_state`, with previous and current typed states when recorded.
- `checkout_state`, with previous and current typed states when recorded.

Change values must remain small and safe. The projection must never expose custom-field contents, descriptions, prompts, transcripts, credentials, tokens, blob keys, storage paths, provider internals, authorization internals, or undo snapshots. Operation IDs are behavior identifiers consumed by the authorized undo command; mobile must not render them as ordinary change metadata.

Action categorization must live with the typed audit action vocabulary or an application-owned policy. Persistence adapters and clients must not maintain unrelated hard-coded lists that can drift from the domain action set.

The first activity projection may include only audit records whose target is the requested asset. Attachment-targeted aggregation requires a future specified projection because attachment records use attachment identity as their durable target.

## First Slice Action Types

The first implementation wrote audit records for:

- `tenant.created`.
- `inventory.created`.
- `inventory_access.granted`.
- `custom_field_definition.created`.
- `asset.created`.
- `asset.updated`.
- `asset.moved`.
- Asset checkout actions are specified in `specs/assets/asset-checkout.spec.md` and extend the action set with `asset.checked_out` and `asset.returned`.

Asset update and asset movement may produce separate audit records from one request when both non-location fields and `parentAssetId` change.

One direct asset edit request must otherwise produce one coherent `asset.updated` record and one undoable operation. Title, description, custom fields, and complete tag assignments must be compared with current state and committed through one asset update unit of work. Re-submitting an unchanged tag set must not create another audit record or advance asset state. Safe update metadata must identify changed fields and may include previous/current title and tag counts under the activity projection rules above.

Authorization denied audit records remain required, but are not part of the first durable audit slice. They must be specified before implementation because denial auditing needs careful noise and sensitivity rules.

The current REST lifecycle action set is extended by `specs/platform/resource-lifecycle.spec.md`, including explicit read audit actions and lifecycle write actions.

## First Slice Sources

The first implementation must support these typed sources:

- `api`.
- `conversation`.
- `mcp`.
- `import`.
- `background_job`.
- `system`.

The current REST API must write `api`.

## Undo

- Users should be able to undo supported actions.
- Product language must distinguish an immediate `Undo` affordance from a historical reversal. A time-bounded action offered directly after a mutation may be labeled `Undo`; an action selected from durable History must be labeled `Revert change` because it applies a new compensating command to one specific historical operation.
- Reverting a History entry reverses only the selected operation. It is not time travel, does not restore the whole asset to that entry's point in time, and must not overwrite unrelated later changes.
- A future whole-asset `Restore to this version` capability would be a separate command with its own preview, conflict policy, authorization, validation, and audit behavior. It is outside this slice.
- Undo must be implemented as domain behavior or compensating application commands, not direct database reversal.
- Redo must be implemented as domain behavior or compensating application commands, not by deleting the undo audit record or mutating history.
- Undo must be authorized.
- Redo must be authorized.
- Undo and redo must produce their own audit records.
- Undo must not bypass validation, tenancy, authorization, or observability.
- Redo must not bypass validation, tenancy, authorization, or observability.
- Not every action must be undoable.
- Actions that are destructive, external, or ambiguous may be marked non-undoable.
- The system must tell the user when an action cannot be undone.
- History clients must explain stale reversal failures in user language: the item changed afterward, so the selected change cannot be safely reverted. They must not imply that retrying will overwrite newer work.
- Undo and redo must be target-scoped. The first API must operate on a specific undoable operation ID, not on a global implicit "last action" stack.
- Authorization must be checked when undo or redo is requested, not only when the original operation occurred.
- Later state-changing operations on the same target must invalidate redo when replaying the old action would no longer apply to the current target state.
- Undo and redo must not mutate or delete audit records.
- Undoable operation records must be durable and stored separately from audit records.
- The original audit record for an undoable action must include `operation_id` metadata so API clients can discover the operation through the normal audit trail.
- Creating an undoable operation must be atomic with the state-changing operation that produced it.
- Applying undo or redo must be atomic with the compensating state change and the audit record for that undo or redo.
- Undoable operation records may contain structured before/after snapshots of safe domain state needed to apply compensating commands.
- Undoable operation snapshots must not contain secrets, acceptance tokens, blob bytes, provider prompts, raw model responses, raw audio, or authorization internals.

## First Undo/Redo Slice

The first undo/redo slice supports assets only.

Supported original actions:

- `asset.created`.
- `asset.updated`.
- `asset.moved`.
- `asset.archived`.
- `asset.restored`.
- `asset.checked_out`.
- `asset.returned`.

The first slice does not support undo or redo for:

- Hard delete operations.
- Tenant operations.
- Inventory operations.
- Sharing, access grants, or invitations.
- Attachment upload, archive, restore, download, or delete.
- Asset checkout history reads.
- Custom asset type changes.
- Custom field definition changes.
- Search.
- Audit reads.

Undo behavior:

- Undoing `asset.created` must archive the created asset if it still exists, is active, and has no active children.
- Undoing `asset.updated` must restore the prior title, description, custom fields, and custom asset type-compatible field values.
- Undoing `asset.moved` must restore the prior parent asset.
- Undoing `asset.archived` must restore the asset.
- Undoing `asset.restored` must archive the asset.
- Undoing `asset.checked_out` must mark the original checkout record as undone without setting return fields.
- Undoing `asset.returned` must reopen the same checkout record by clearing the return fields when the checkout history has not changed in a conflicting way.

Redo behavior:

- Redoing an undone `asset.created` must restore the archived asset if normal restore validation passes.
- Redoing an undone `asset.updated` must reapply the original updated title, description, custom fields, and compatible custom asset type values.
- Redoing an undone `asset.moved` must reapply the original parent asset.
- Redoing an undone `asset.archived` must archive the asset.
- Redoing an undone `asset.restored` must restore the asset.
- Redoing an undone `asset.checked_out` must reopen the original checkout record when normal checkout validation still passes.
- Redoing an undone `asset.returned` must reapply the original return fields to the same checkout record.

Validation and invalidation:

- Undo or redo must fail if the target asset no longer exists.
- Undo or redo must fail if the current asset state does not match the expected state for that operation direction.
- Undo or redo must fail if normal asset validation would reject the compensating state, including containment, archived-parent, active-child archive, custom field, checkout, and authorization rules.
- Undo or redo must fail after hard delete because the target no longer exists.
- Undo or redo must fail after later changes make the saved before/after state stale.

Initial operation shape:

- `id`: generated application ID.
- `tenantId`: tenant security boundary.
- `inventoryId`: inventory scope.
- `principalId`: principal that performed the original operation.
- `source`: source adapter for the original operation.
- `targetType`: first slice uses `asset`.
- `targetId`: target asset ID.
- `originalAction`: audit action that produced the operation.
- `status`: `available`, `undone`, or `redone`.
- `createdAt`: server timestamp.
- `lastAppliedAt`: timestamp of the last undo or redo, if any.
- `beforeState`: safe structured state before the original action, when applicable.
- `afterState`: safe structured state after the original action.
- `undoAuditRecordId`: audit record ID for the most recent undo, if any.
- `redoAuditRecordId`: audit record ID for the most recent redo, if any.

The first slice may expose only `undo` and `redo` endpoints. Listing undoable operations can follow once the UI timeline is specified.

### Web experience

- A successful supported asset action whose response includes an undoable operation ID must show a time-bounded success notification with an `Undo` action. Supported web actions are create, edit, move, archive, restore, checkout, and return, including one-click return from Home.
- The notification action must apply the exact operation ID returned by that mutation; the web client must not infer a global "last action" or discover an operation by timing.
- While Undo or Redo is being applied, repeated activation for the same operation must be suppressed and the action must remain screen-reader announced as in progress.
- Successful Undo must reconcile the affected asset, checked-out collection, selected detail, and active lifecycle collection from the API result, then offer a time-bounded `Redo` action for the same operation ID. Successful Redo must perform the symmetric reconciliation and offer `Undo` again.
- Undo/Redo must stay behind the web inventory repository port. The Svelte component must not call the generated API client directly.
- An expired, stale, denied, invalid, or otherwise failed Undo/Redo must not announce success or remove the current page context. The notification must become a persistent error with the server-safe reason and a dismiss control; stale state is then refreshed from the selected inventory so the page does not continue presenting an invalid action state.
- Hard delete and attachment operations must not offer Undo in the first slice. Their confirmations and completion messages must continue to state their actual recovery semantics.
- If a successful mutation response does not include an operation ID, the web app must preserve the normal success confirmation without presenting a nonfunctional Undo action.

## REST API

The first undo/redo endpoints are:

- `POST /tenants/{tenantId}/inventories/{inventoryId}/undoable-operations/{operationId}/undo`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/undoable-operations/{operationId}/redo`

Both endpoints:

- Require authentication.
- Require `inventory.edit_asset` for the operation inventory.
- Use the standard success envelope.
- Return the affected asset in the first slice.
- Use safe error envelopes.
- Must not reveal whether an operation exists outside the caller's authorized tenant and inventory.
- Must emit domain-oriented observability.
- Must be represented in the generated OpenAPI contract.

## Undo/Redo Audit Actions

The first undo/redo audit actions are:

- `undoable_operation.undone`.
- `undoable_operation.redone`.

Audit metadata must include:

- `operation_id`.
- `original_action`.
- `target_type`.
- `target_id`.

Audit metadata may include compact safe state hints such as lifecycle state or parent asset IDs. It must not include full custom field values unless a future spec defines redaction and retention rules.

## Initial Audited Actions

The full audited action set should eventually include:

- Asset created.
- Asset updated.
- Asset moved.
- Asset archived or removed.
- Asset checked out.
- Asset returned.
- Location created.
- Location updated.
- Location moved.
- Inventory created or updated.
- Custom field definition created or updated.
- Custom field value changed.
- Conversational plan created.
- Conversational plan approved.
- Conversational plan cancelled.
- Conversational command executed.
- Authorization denied.

## Testing

- Tests must verify that every state-changing application operation writes audit history.
- Tests must verify that audited read endpoints write safe read history without storing response bodies, secrets, tokens, blob contents, or authorization internals.
- Tests must verify tenant and inventory isolation for audit reads.
- Tests must verify pagination for audit reads, including `(occurredAt, id)` ordering.
- Tests must verify newest-first cursor pagination for asset activity and reject cursors reused across tenants, inventories, assets, or activity views.
- Tests must verify that `change` activity still returns a recent edit after more than one page of read audit records, while `all` activity preserves authorized access to those reads.
- Tests must verify that activity returns safe structured title, description-changed, tag-count, parent, lifecycle, and checkout summaries without leaking unsafe metadata or undo snapshots.
- Tests must verify activity undo status only resolves operations from the same tenant and inventory.
- Tests must verify one direct edit request produces one coherent update record and operation, unchanged fields and tag assignments do not create duplicate history, and the asset, assignments, audit record, and operation commit atomically.
- Tests must verify duplicate audit record IDs are rejected instead of replacing existing records.
- Tests must verify undo for supported commands once undo is implemented.
- Tests must verify redo for supported commands once redo is implemented.
- Tests must verify that unauthorized users cannot read audit records or undo actions they cannot perform.
- Tests must verify unauthorized users cannot redo actions they cannot perform.
- Tests must verify undo and redo fail when later target changes make the saved operation stale.
- Tests must verify undo and redo fail across tenant and inventory boundaries.
- Tests must verify undoable operation creation is atomic with the original state change where the repository supports transactions.
- Tests must use fakes, not mocks.

## Open Questions

- How long should audit records be retained?
- Should audit history be exportable with tenant-level exports?
- Should authorization denied records be visible in normal user audit history, security-only history, or both?
- Should undoable operation history be listable before the product UI timeline is designed?
