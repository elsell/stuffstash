# Audit And Undo Spec

## Purpose

Stuff Stash must keep an audited history of inventory actions.

Users should be able to understand what changed, who changed it, and undo supported actions when that is safe.

## Scope

This spec covers initial audit history and undo requirements.

This spec does not define the final event store, persistence schema, retention policy, UI timeline, or every undoable command.

## Requirements

- Every state-changing action must produce an audit record.
- Audit records must include the authenticated principal, tenant, inventory, action type, target resources, timestamp, source adapter, and request identifier where available.
- Audit records must distinguish between direct user actions, conversational actions, MCP actions, imports, background jobs, and system actions.
- Conversational audit records must reference the action plan and approval when available.
- Audit records must not store raw provider prompts, raw model responses, raw audio, secrets, tokens, or sensitive data beyond what the audit use case requires.
- Audit behavior must be behind ports and adapters.
- Audit writing must be part of the application operation boundary.
- Audit records must preserve tenant and inventory isolation.

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

The first audited action set should include:

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
- Tests must verify undo for supported commands.
- Tests must verify that unauthorized users cannot read audit records or undo actions they cannot perform.
- Tests must use fakes, not mocks.

## Open Questions

- Which first-release actions are undoable?
- How long should audit records be retained?
- Should audit history be exportable with tenant-level exports?
- Should audit reads be available to inventory viewers or only higher relationships?
