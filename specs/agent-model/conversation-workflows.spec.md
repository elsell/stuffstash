# Conversation Workflows and Evaluation

## Domain and ownership

Conversation workflows belong to the agent/model bounded context. A tenant owns each workflow, its immutable revisions, reusable evaluation cases and evaluation runs. Production chooses the active workflow revision for the tenant; each conversation snapshots that revision. Changes to a draft never modify a running conversation. The application validates provider profile ownership and capabilities before saving or activating a revision.

## Workflow definition

A workflow definition contains a display name, bounded step instructions and an execution policy. The first supported step kinds are interpretation, evidence assessment and response realization. Each step may select a configured language-inference profile; an omitted profile uses the tenant's selected language profile. Transcription and speech continue to use the tenant's configured capability profiles. Inventory retrieval, validation, approval and execution remain application-owned transitions between model steps.

Interpretation produces a typed intent and authorized-read requests. Evidence assessment may request another materially different read, resolve a target, request clarification or finish. Response realization can use generated wording with grounded fallback or deterministic grounded wording. Users can tune each model step's instructions and attempts independently, choose response realization mode, select retrieval strategy (precise-first or expanded), and set total evidence rounds, model calls, elapsed time and follow-up turns. They cannot reorder approval before grounding, remove authorization or install arbitrary scripts. This structured configuration is the first workflow editor; it must clearly show the execution sequence and adjustable steps.

Configuration validation receives operator limits from environment-backed bootstrap configuration. Limits must be positive. The global model-call/time budget takes precedence over individual step retry and evidence-round maxima; these maxima are ceilings, not promises that every allowed attempt can run. Budget exhaustion returns a typed recoverable outcome, never an unbounded retry. Each workflow budget must be positive and no greater than its operator limit. Unknown modes, duplicate/missing step kinds, blank names, oversized names/instructions, malformed IDs and unsupported capability selections are rejected. Definition values are immutable snapshots: callers cannot mutate a saved definition by changing input or returned slices.

Steps and budgets must actually drive the shared production/evaluation executor. A cosmetic setting is not an implemented feature. Capability differences are surfaced before activation; missing capability is not silently routed to another external provider. Gemini, Claude and compatible local model integrations remain behind adapters with normalized project-owned contracts.

## Revision and activation

Workflow identity and revision identity are distinct. Saving creates a new numbered revision, with author and injected-clock timestamps, using optimistic concurrency. Activation checks ownership, compatibility and that the candidate has a completed passing evaluation for the selected test suite and current provider configuration. The active pointer changes atomically with audit. A previous revision can be activated for rollback with the same authorization checks. Deleting or disabling a provider used by an active revision must surface a clear precondition or unavailable state, never silently switch models.

## Test cases and runs

Cases contain a title, text utterance or audio attachment, isolated inventory fixtures and outcome expectations. Expectations include result kind, referenced fixture assets, location/containment facts, proposed commands, clarification requirements and forbidden mutations. Avoid exact prose comparisons as the default judge. Include default examples for tagged baby clothes, existing-item moves, an explicitly additional item, ambiguous duplicates, speech failure and follow-up references.

Runs snapshot revision, provider configuration identity (never credentials), case versions and operator budgets. A leased background worker executes cases through the same conversation service, records bounded safe stage events, outcomes, call counts and timings, and honors cancellation between and during provider calls. Run states are queued, running, succeeded, failed, cancelled. Recovery after worker loss must not repeat real inventory mutations; evaluation uses fixture repositories and simulated command execution by default. The API exposes queued/running state without holding a request open.

Tenant configure permission owns workflow/case/run mutation and trace access. Real inventory read evaluations require the selected inventory's view permission as well. Every mutation is audited; read audit follows existing lifecycle contracts. Trace records redact credentials, provider transport payloads and inaccessible data. Operator configuration controls retention and maximum concurrent runs. Evaluation instructions cannot grant tools or disable policy.

## REST and web surfaces

Tenant-scoped collections: conversation workflows, workflow revisions, evaluation cases and evaluation runs. Explicit activation and cancellation operations are audited application commands. List/read/create/update requests use consistent envelopes, generated contracts and cursor conventions; revision updates reject stale expected versions.

The web workspace offers a default workflow summary, draft editor with step controls, test-case editor with fixture assets/tags/containment, run progress, side-by-side revision comparison and activation/rollback. Surface provider compatibility, estimated limits and actual usage without showing secrets. Runtime behavior is previewed through real test execution rather than a separate mock UI. Use existing SvelteKit, TanStack ownership and frontend ports/adapters.

## Evidence gates

Domain validation and immutability tests precede implementation. Persistence/migration, application command/query, adversarial REST, worker cancellation/recovery, web adapter/UI and shared executor tests follow their responsibilities. Live ADC-backed cases exercise real audio, schema requests and speech alongside controlled fixture data. Run the same cases against a compatible local model where available; unavailable credentials/models are reported as verification limits rather than passed support. Release and device gates remain defined in voice-conversation-quality.spec.md.

## Persistence contract

Store workflow identity separately from revision snapshots. Every lookup requires tenant ID; revision lookups additionally require workflow and revision IDs. Workflow IDs and revision IDs are nonempty bounded opaque identifiers using ASCII letters, digits, underscore or hyphen. Author references preserve the established principal format: Unicode letters/digits, period, underscore or hyphen. Tenant references preserve the established tenant identity contract and are required to be nonblank; workflow identifier syntax does not redefine tenant IDs. A revision has a positive sequence number, immutable validated definition, author and nonzero creation timestamp supplied by the injected application clock. Rehydration must validate definitions against the recorded revision limits rather than reinterpret old revisions using changed defaults; activation separately checks current operator limits.

Appending a revision compares the expected latest sequence and atomically stores the revision, updates the workflow head and records audit. Concurrent saves must produce one winner and a conflict, never overwrite immutable history. Activating compares expected active revision and atomically updates the active pointer with audit. An active revision must belong to the same tenant and workflow. Persistence adapters must implement these operations transactionally, with production PostgreSQL migrations and SQLite parity tests. No direct application SQL is permitted.

Workflow revision appends emit `conversation_workflow.revision_created`; activation/rollback emit `conversation_workflow.activated`. Both actions are valid persisted audit actions. Safe metadata includes workflow/revision identity and sequence, not prompt bodies or credentials. Transactional repository commands reject audit records whose tenant differs from the workflow tenant. A missing workflow/revision is a normal scoped read miss; stale writes return a conflict sentinel that application adapters map consistently.

Migration rollback removes workflow tables but retains the expanded audit-action allowance, preserving historical audit records. It must not delete audit history to downgrade the schema.

Saving a draft validates all explicit step provider references for tenant ownership and language-inference capability. Empty references retain the tenant-default selection intent; the current default need not be configured to save a draft. Activation and execution must resolve defaults and require usable compatible profiles. Draft saving is authorized before ID generation, provider lookup or repository mutation. Returned revisions contain no credentials. A stale expected sequence maps to the standard application conflict error.

## Draft REST boundary

`POST /tenants/{tenantId}/conversation-workflows` creates the first immutable draft revision. `POST /tenants/{tenantId}/conversation-workflows/{workflowId}/revisions` appends a revision using a required positive `expectedRevision`. Both return 201 with a revision envelope containing workflow/revision identity, sequence, author, timestamp and the normalized definition. They require an authenticated principal with tenant configure permission; tenant membership or inventory ownership alone is insufficient. Missing workflows return 404, stale expected revisions return 409, and incompatible or inaccessible provider references return validation errors without disclosing provider details. Draft creation does not activate a workflow. Definition fields use camelCase transport names and remain separate from domain types.
