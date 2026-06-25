# Conversational Action Plan Spec

## Purpose

Stuff Stash needs a structured action plan format for language-model-assisted inventory work.

The action plan is the bridge between natural language and domain operations. It lets clients show what will happen, ask for approval, and execute only safe, authorized application commands.

## Scope

This spec covers the initial action plan contract for conversational inventory.

This spec does not define prompts, provider-specific structured output formats, UI layout, or every executable command.

## Requirements

- Conversational flows must produce a structured action plan before state-changing work is executed.
- The action plan must be represented with typed values and enumerations, not free-form magic strings.
- The action plan must be safe to render in web and mobile clients.
- The action plan must identify the tenant, inventory scope, initiating principal, intended action, matched resources, proposed resource creation, required confirmations, and executable commands.
- The action plan must distinguish between:
  - User intent.
  - Model interpretation.
  - Resolved existing resources.
  - Proposed new resources.
  - Clarifying questions.
  - Required confirmations.
  - Executable application commands.
- The action plan must not be treated as authorization.
- Every executable command in the plan must be authorized at execution time.
- The action plan must be auditable.
- The action plan must be reproducible in tests with deterministic fake model responses.
- The action plan must be independent of provider-specific model response formats, tool call formats, prompt templates, and streaming protocols.
- The action plan must be produced and executed by the core API conversational application services, not by clients or provider-hosted agents.

## Initial Shape

The exact serialized format remains open, but the initial structure must include these concepts:

- `planId`: stable ULID for the plan.
- `tenantId`: tenant security boundary.
- `inventoryIds`: inventories in scope.
- `principalId`: authenticated user who initiated the flow.
- `source`: interaction source, such as mobile voice, mobile text, web text, web voice, REST, or MCP.
- `transcript`: optional safe transcript artifact reference governed by a specified retention and redaction policy, not raw audio.
- `intent`: interpreted action intent.
- `confidence`: coarse confidence value if useful to the UI.
- `matches`: existing assets, locations, inventories, or custom fields matched by the flow.
- `proposedCreates`: assets, locations, or custom field definitions the flow proposes to create.
- `commands`: executable application commands.
- `confirmation`: confirmation requirement, human summary, and approval state.
- `clarification`: question and options when the system cannot safely proceed.
- `risks`: safe user-facing reasons why approval or clarification is needed.
- `auditMetadata`: safe metadata for audit and observability.

The action plan may reference the realtime session that produced it, but it must not contain provider credentials, raw provider prompts, raw provider responses, raw audio, raw transcripts, generated speech bytes, or provider-specific session identifiers.

Until a transcript retention and redaction policy is specified, action plans must not reference persisted transcript artifacts. They may use only ephemeral transcript text during planning and safe derived metadata in persisted records.

## Command Rules

- Commands must map to application operations, not database operations.
- Commands must not contain provider-specific model output.
- Commands must not contain raw prompts, raw provider responses, credentials, or secrets.
- Commands must be validated before execution.
- Commands must be authorized before execution.
- Commands must be executed atomically when the domain operation requires atomicity.
- Commands that cannot be executed safely must fail without applying partial state changes unless a future spec defines a compensating workflow.

## Approval Rules

- Clients must be able to display a concise human-readable summary of the plan.
- Clients must be able to display the specific operations that need approval.
- Users must be able to approve or cancel a plan.
- Users must be able to answer clarifying questions without restarting the conversation.
- Approval must be recorded with the audit history of executed actions.
- Approval of one plan must not authorize unrelated future plans.

## Testing

- Tests must verify plan creation, clarification, approval, cancellation, execution, malformed model output, and provider failure.
- Tests must verify that unauthorized commands in a plan are rejected at execution time.
- Tests must verify that model output cannot smuggle unapproved commands into execution.
- Tests must use fakes for model providers, authorization, repositories, observability, and audit history.

## Open Questions

- What exact command enumeration should exist for the first release?
- Should action plans expire if not approved within a configured time?
- Should users be able to edit a proposed plan before approval?
- Which action plan fields should be persisted long term versus only referenced by audit events?
