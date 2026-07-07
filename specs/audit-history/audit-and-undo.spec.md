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
