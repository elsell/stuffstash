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

The complete action-plan model is expected to include:

- `planId`: stable ULID for the plan.
- `tenantId`: tenant security boundary.
- `inventoryId`: first inventory in scope. Multi-inventory plans are deferred until cross-inventory conversational behavior is specified.
- `principalId`: authenticated user who initiated the flow.
- `source`: interaction source, such as mobile voice, mobile text, web text, web voice, REST, or MCP.
- `realtimeSessionId`: optional safe Stuff Stash realtime session ID that produced the plan.
- `intent`: interpreted action intent.
- `confidence`: coarse confidence value if useful to the UI.
- `matches`: existing assets, locations, inventories, or custom fields matched by the flow.
- `proposedCreates`: assets, locations, or custom field definitions the flow proposes to create.
- `commands`: executable application commands.
- `confirmation`: confirmation requirement, human summary, and approval state.
- `clarification`: question and options when the system cannot safely proceed.
- `risks`: safe user-facing reasons why approval or clarification is needed.
- `auditMetadata`: safe metadata for audit and observability.
- `createdAt`, `updatedAt`, and optional `approvedAt`, `cancelledAt`, `executedAt`, or `failedAt` timestamps.

The first persisted foundation may store the safe subset needed for durable approval boundaries: plan ID, tenant ID, inventory ID, principal ID, source, optional realtime session ID, state, bounded safe intent and interpretation summaries, confirmation summary, typed command records, bounded risks, and lifecycle timestamps. The deferred complete shape must be added before clarification UI, rich plan display, match visualization, confidence display, proposed-create previews, or audit metadata depend on those fields.

The action plan may reference the realtime session that produced it, but it must not contain provider credentials, raw provider prompts, raw provider responses, raw audio, raw transcripts, generated speech bytes, or provider-specific session identifiers.

Until a transcript retention and redaction policy is specified, action plans must not reference persisted transcript artifacts. They may use only ephemeral transcript text during planning and safe derived metadata in persisted records.

## First Persisted Lifecycle

The first implementation must persist action plans behind a project-owned repository port with memory and database-backed adapters.

The first lifecycle states are:

- `proposed`: the plan has been created and is awaiting user approval, cancellation, or clarification.
- `approved`: the initiating principal approved this specific plan, but commands have not necessarily executed yet.
- `cancelled`: the user cancelled the plan before execution.
- `executed`: all executable commands in the plan completed successfully.
- `failed`: the plan could not be executed safely.

The first slice may implement creation, approval, and cancellation before command execution. Approval must be explicit, tied to the initiating principal and plan ID, and must not execute commands until an execution service is implemented. Cancelling or approving a terminal plan must fail safely. Repository reads and state transitions must be scoped by tenant ID and inventory ID.

The first persisted plan must not store raw transcript text. A safe `userIntentSummary` and `modelInterpretationSummary` may be stored only when they are bounded, user-renderable, and free of provider-specific raw output.

## Realtime Voice Proposal Tool

The mobile realtime voice loop may expose a single proposal tool named `propose_action_plan`. This tool creates a persisted `proposed` action plan and returns a safe review payload to the realtime client. It is not a domain write command and must not mutate inventory resources.

The proposal tool must accept only:

- a command kind from the initial command enumeration,
- a safe user-facing intent summary,
- a safe model interpretation summary,
- a safe confirmation summary,
- a safe command summary,
- bounded command arguments validated by the action-plan application service, and
- bounded safe risk summaries.

The proposal tool must reject unknown arguments, unknown command kinds, empty confirmation summaries, unsafe command arguments, approval claims, provider credentials, raw prompts, raw provider responses, raw transcripts, bearer tokens, provider session IDs, and hidden resource data. The persisted plan must be scoped to the active tenant, inventory, principal, and realtime session.

The first realtime proposal slice must not approve, execute, or cancel the proposed plan automatically. Approval and execution require explicit later client messages and application-service handling.

## Initial Command Enumeration

The first command enumeration is:

- `create_asset`
- `create_location`
- `move_asset`
- `update_asset`
- `archive_asset`
- `restore_asset`

Commands must be stored as project-owned typed command records with a command ID, command kind, safe human summary, and bounded JSON arguments. The first persistence slice may store command arguments as reviewed JSON while application services still validate the command kind and safe summary. Command arguments must not contain provider-specific model output, raw prompts, credentials, bearer tokens, hidden resource data, or approval claims.

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
