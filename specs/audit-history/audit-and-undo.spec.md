# Audit And Undo Spec

## Purpose

Stuff Stash must keep an audited history of inventory actions.

Users should be able to understand what changed, who changed it, and undo supported actions when that is safe.

## Scope

This spec covers initial audit history and undo requirements.

This spec does not define the final event store, retention policy, UI timeline, or every undoable command.

The first implementation slice defines durable, read-only audit records. Undo is specified as a future behavior and must not be partially implemented until the first undoable commands are specified.

## Requirements

- Every state-changing action must produce an audit record.
- Audit records must include the authenticated principal, tenant, inventory, action type, target resources, timestamp, source adapter, and request identifier where available.
- Audit records must distinguish between direct user actions, conversational actions, MCP actions, imports, background jobs, and system actions.
- Conversational audit records must reference the action plan and approval when available.
- Audit records must not store raw provider prompts, raw model responses, raw audio, secrets, tokens, or sensitive data beyond what the audit use case requires.
- Audit behavior must be behind ports and adapters.
- Audit writing must be part of the application operation boundary.
- Audit records must preserve tenant and inventory isolation.
- Audit records must be append-only through application ports. The first slice must not expose update or delete behavior for audit records.
- Audit records must be cursor paginated.
- Inventory audit reads must require `inventory.view` for the target inventory.
- Tenant audit reads without an inventory scope must require `tenant.configure`.
- Audit read responses must use the standard API success and error envelopes.

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
- `requestId`: optional request identifier when a transport provides one.
- `metadata`: safe JSON object for small, non-secret context.

Metadata is for human context and filtering hints. It must not be treated as an authorization source.

## First Slice Action Types

The first implementation must write audit records for:

- `tenant.created`.
- `inventory.created`.
- `inventory_access.granted`.
- `custom_field_definition.created`.
- `asset.created`.
- `asset.updated`.
- `asset.moved`.

Asset update and asset movement may produce separate audit records from one request when both non-location fields and `parentAssetId` change.

Authorization denied audit records remain required, but are not part of the first durable audit slice. They must be specified before implementation because denial auditing needs careful noise and sensitivity rules.

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
- Undo must be authorized.
- Undo must produce its own audit record.
- Undo must not bypass validation, tenancy, authorization, or observability.
- Not every action must be undoable.
- Actions that are destructive, external, or ambiguous may be marked non-undoable.
- The system must tell the user when an action cannot be undone.

## Initial Audited Actions

The full audited action set should eventually include:

- Asset created.
- Asset updated.
- Asset moved.
- Asset archived or removed.
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
- Tests must verify tenant and inventory isolation for audit reads.
- Tests must verify pagination for audit reads.
- Tests must verify undo for supported commands once undo is implemented.
- Tests must verify that unauthorized users cannot read audit records or undo actions they cannot perform.
- Tests must use fakes, not mocks.

## Open Questions

- Which first-release actions are undoable?
- How long should audit records be retained?
- Should audit history be exportable with tenant-level exports?
- Should authorization denied records be visible in normal user audit history, security-only history, or both?
