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

Mobile realtime approval and cancellation must use the same action-plan application service methods as any future REST, web, or MCP review surface. The realtime adapter may translate `action.plan.approve` and `action.plan.cancel` WebSocket messages into application-service calls, but it must not update action-plan persistence directly.

The first execution slices support single-command asset lifecycle work, asset checkout work, multi-command create plans, and multi-command create-then-move plans. `create_location` is executed as asset creation with kind `location`; `create_asset` is executed as asset creation with kind `item` unless the validated command arguments explicitly provide `item`, `container`, or `location`. `move_asset` is executed as an asset update that changes containment. `archive_asset` is executed as an asset lifecycle transition from active to archived. `restore_asset` is executed as an asset lifecycle transition from archived to active. `checkout_asset` creates an open asset checkout record without moving the asset. `return_asset` closes the open checkout record without moving the asset. Execution must use the existing asset application service boundary so tenant and inventory authorization, domain validation, audit history, undoable operations where supported, observability, and persistence continue to behave exactly like equivalent REST or UI commands.

Multi-command execution is initially limited to ordered plans containing `create_asset`, `create_location`, and `move_asset` commands. Each dependent command may depend only on an earlier create command in the same plan. The agent should prefer a multi-command plan for a clear request that needs missing parent containers or locations created, such as creating `Kitchen` and then moving an existing water bottle there. Clarification is appropriate when the intended destination is ambiguous, conflicts with visible inventory, or appears likely to be a speech-to-text mistranscription. Unsupported command mixes, forward references, cycles, missing references, or unsupported write commands in a multi-command plan must transition the plan to `failed` without applying inventory changes.

Execution must require the plan to be in `approved` state, must be scoped by tenant ID, inventory ID, principal ID, and plan ID, and must re-authorize the initiating principal at execution time before exposing plan existence or state. Successful execution transitions the plan to `executed`; failed execution transitions it to `failed`. For every executable command plan, applying all inventory changes, audit history, undoable operations, and the terminal `executed` plan transition must be one repository unit of work so a failed terminal transition cannot leave duplicate-prone inventory mutations or partial dependent hierarchies behind. Execution must not trust prior approval as authorization and must not execute commands from `proposed`, `cancelled`, `executed`, or `failed` plans.

The first persisted plan must not store raw transcript text. A safe `userIntentSummary` and `modelInterpretationSummary` may be stored only when they are bounded, user-renderable, and free of provider-specific raw output.

## Realtime Voice Proposal Tool

The mobile realtime voice loop may expose a single proposal tool named `propose_action_plan`. This tool creates a persisted `proposed` action plan and returns a safe review payload to the realtime client. It is not a domain write command and must not mutate inventory resources.

In the bounded production voice loop, the language model must not author executable action-plan commands. It may return only typed intent, ordered semantic references, bounded search hypotheses, and evidence resolutions through the project-owned investigation contract. After the application validates authorized evidence and chooses a terminal product outcome, an operation-specific application compiler must construct the existing typed command records for `create_asset`, `create_location`, `move_asset`, `archive_asset`, `restore_asset`, `checkout_asset`, and `return_asset`.

The compiler must be deterministic for the same canonical intent and authorized evidence. It must use only reference-scoped opaque IDs returned by authorized reads, preserve operation anchors, enforce lifecycle and custody preconditions, create a clear missing destination path as an ordered command dependency graph, and stop at the persisted `proposed` state. A provider-produced command, command ID, executable argument object, approval claim, or alternate operation must be rejected by the investigation boundary rather than normalized into an executable plan.

The existing `propose_action_plan` argument contract remains the internal application input used to persist and present a compiled plan and may remain available to other separately specified adapters. It is not a provider-callable tool in the bounded production voice loop. The realtime wire events, persisted action-plan shape, explicit mobile approval/cancellation messages, review edits, authorization, audit, and atomic execution semantics remain unchanged.

The proposal tool must accept only:

- either a single command kind from the initial command enumeration or an ordered bounded `commands` array,
- per-command IDs, command kinds, safe command summaries, and bounded structured command arguments when using `commands`,
- a safe user-facing intent summary,
- a safe model interpretation summary,
- a safe confirmation summary,
- a safe command summary for single-command compatibility,
- bounded command arguments validated by the action-plan application service, and
- bounded safe risk summaries.

The proposal tool must reject unknown arguments, unknown command kinds, empty confirmation summaries, unsafe command arguments, approval claims, provider credentials, raw prompts, raw provider responses, raw transcripts, bearer tokens, provider session IDs, and hidden resource data. The persisted plan must be scoped to the active tenant, inventory, principal, and realtime session.

The proposal tool must preserve executable command arguments exactly as bounded structured JSON, not lossy natural language. Provider adapters should expose command arguments as an object parameter whenever the provider supports native object parameters. String-encoded JSON is allowed only as a compatibility fallback and must be parsed by the application boundary before persistence.

The first realtime proposal slice must not approve, execute, or cancel the proposed plan automatically. Approval and cancellation require explicit later client messages and application-service handling. Execution may happen only after approval through the action-plan execution service.

## Initial Command Enumeration

The first command enumeration is:

- `create_asset`
- `create_location`
- `move_asset`
- `update_asset`
- `archive_asset`
- `restore_asset`
- `checkout_asset`
- `return_asset`

Commands must be stored as project-owned typed command records with a command ID, command kind, safe human summary, and bounded JSON arguments. The first persistence slice may store command arguments as reviewed JSON while application services still validate the command kind and safe summary. Command arguments must not contain provider-specific model output, raw prompts, credentials, bearer tokens, hidden resource data, or approval claims.

The first executable `create_asset` and `create_location` argument shape is:

- `title` or `name`: required non-empty user-facing asset title.
- `kind`: optional for `create_asset`, restricted to `item`, `container`, or `location`; ignored for `create_location`, which always creates a location.
- `description`: optional safe user-facing description.
- `parentAssetId`: optional existing parent asset ID in the same inventory.
- `parentCommandId`: optional command ID of an earlier `create_asset` or `create_location` command in the same plan. This creates the asset inside that newly created parent.

The execution service must reject command arguments outside this shape for executable create commands until richer command schemas are specified.

When a create command places a new asset inside an existing location or container, the proposal must use the existing parent's tool-derived `assetId` as `parentAssetId`. When a create or move command places an asset inside another resource being created by the same plan, it must use `parentCommandId` pointing to the earlier create command. Human titles such as `parentTitle`, `locationTitle`, or raw location names are read filters only and are not executable create arguments. If the parent cannot be found or safely created unambiguously, the agent must ask for clarification rather than inventing an ID.

Realtime proposal payloads must include enough safe command detail for clients to present the plan clearly. For each command, the payload should include command ID, command kind, summary, title when present, asset kind when present, operation category such as create/use/move/archive/restore, and a safe parent reference summary when present. The payload must not expose raw provider data, prompts, transcripts, credentials, or hidden resource data. Existing asset IDs and internal command IDs may be sent only as opaque plan references needed for accurate approval display; clients must render user-facing titles and summaries rather than raw IDs.

Proposed `create_asset` and `create_location` commands must remain editable during review. A review client must allow the user to edit the proposed title inline and choose the containing parent without leaving the review flow. Parent choices may be inventory root, an authorized visible asset in the same inventory, or an earlier create command in the same plan. In the current authorization model, inventory `view` permission grants visibility to the inventory's tenant-scoped assets; there is no separate per-asset visibility relationship. The review UI must distinguish an existing parent from a parent that the same plan will create.

Edits are draft review state until approval. Approval may include a bounded list of command edits keyed by command ID, containing only an edited title and one mutually exclusive parent reference: `parentAssetId`, `parentCommandId`, or explicit inventory root. The application boundary must reject edits for unknown commands, duplicate command IDs, non-create commands, blank or overlong titles, unknown fields, both parent reference kinds at once, forward or cyclic command references, hidden or cross-inventory existing parents, and edits to a plan that is no longer `proposed`. It must not accept replacement command kinds, arbitrary command arguments, summaries, risks, approval claims, tenant or inventory identifiers, provider data, prompts, transcripts, credentials, or hidden resource data.

Applying reviewed edits and transitioning the plan from `proposed` to `approved` must be one repository operation scoped to the initiating principal, tenant, and inventory. Execution must use the persisted edited command arguments. A failed edit or transition must leave the original proposed plan unchanged and executable only after a later valid approval.

The first executable `move_asset` argument shape is:

- `assetId`: required existing active asset ID in the same inventory.
- `parentAssetId`: optional existing active container or location ID in the same inventory. When omitted, empty, or null, the asset is moved to the inventory root.
- `parentCommandId`: optional command ID of an earlier `create_asset` or `create_location` command in the same plan. This moves the asset inside that newly created parent.

The execution service must reject command arguments outside this shape for executable move commands until richer command schemas are specified.

The first executable `archive_asset` argument shape is:

- `assetId`: required existing active asset ID in the same inventory.

The execution service must reject command arguments outside this shape for executable archive commands until richer command schemas are specified. `archive_asset` must use the same lifecycle rules as the asset application service, including rejecting archive when the asset has active children.

The first executable `restore_asset` argument shape is:

- `assetId`: required existing archived asset ID in the same inventory.

The execution service must reject command arguments outside this shape for executable restore commands until richer command schemas are specified. `restore_asset` must use the same lifecycle rules as the asset application service, including rejecting restore when the asset has an archived or unavailable parent.

The first executable `checkout_asset` argument shape is:

- `assetId`: required existing active asset ID in the same inventory.
- `details`: optional bounded user text.

The execution service must reject command arguments outside this shape for executable checkout commands until richer command schemas are specified. `checkout_asset` must use the asset checkout application service specified in `specs/assets/asset-checkout.spec.md`, including rejecting checkout when the asset already has an open checkout.

The first executable `return_asset` argument shape is:

- `assetId`: required existing asset ID in the same inventory with an open checkout.
- `details`: optional bounded user text.

The execution service must reject command arguments outside this shape for executable return commands until richer command schemas are specified. `return_asset` must use the asset checkout application service specified in `specs/assets/asset-checkout.spec.md`, including allowing any authorized editor to return the asset.

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
- State-changing voice plans, including checkout and return, require explicit client UI approval through a button or equivalent accessible control in the first slice. Spoken approval must not execute the plan.
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
- Which action plan fields should be persisted long term versus only referenced by audit events?
