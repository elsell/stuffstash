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

## Operator workflow limits

Runtime configuration captures operator limits at startup. `STUFF_STASH_WORKFLOW_MAX_EVIDENCE_ROUNDS`, `STUFF_STASH_WORKFLOW_MAX_MODEL_CALLS`, `STUFF_STASH_WORKFLOW_MAX_ELAPSED_SECONDS`, `STUFF_STASH_WORKFLOW_MAX_FOLLOW_UP_TURNS`, `STUFF_STASH_WORKFLOW_MAX_STEP_ATTEMPTS`, `STUFF_STASH_WORKFLOW_MAX_NAME_RUNES`, and `STUFF_STASH_WORKFLOW_MAX_INSTRUCTION_RUNES` default respectively to 4, 12, 60, 8, 2, 100 and 4000. Unset, empty and whitespace-only values use defaults. A nonblank supplied value must be a positive integer; zero, negatives, malformed values and integer overflow prevent application startup with the variable name, without echoing its value. Configuration is captured once rather than read from the environment by services or domain code. Programmatic zero-value configuration uses these same defaults. These are operator ceilings; saved workflows still specify their own budgets within them.

The in-memory runtime adapter must preserve the same scoped immutable revisions, compare-and-swap conflicts and atomic audit semantics as persistent adapters. Concurrent appends permit exactly one winner for an expected sequence. Duplicate audit identities and invalid cross-tenant audit records must fail before changing any workflow state.

## Fixture case definitions

A fixture case definition has a display title, one text utterance, at most 100 fixture assets, and structured expectations. This text definition is the first case input; audio case transport remains a separate required follow-on. Fixture asset identities are local opaque IDs (the workflow identifier grammar), never live inventory references. Each fixture has item/container/location kind, title, optional description, assigned tag display names and an optional parent fixture ID. IDs are unique; parents must exist; containment must be acyclic. Titles and case titles are bounded to 160 runes, utterances to 4000 runes, descriptions to 2000 runes, and tags use the established 32-name/80-byte voice evidence bounds.

Expected outcomes are answer, clarification, proposal or failure. Expectations may require references to fixture assets and containment facts expressed as asset/ancestor pairs. Such references must resolve inside the fixture; expected containment must agree with its parent chain. Proposed operation expectations use existing inventory operation vocabulary. A proposal expectation requires at least one proposed operation; other outcome kinds cannot require a proposal. Forbidden operations cannot also be required. No expectation executes a mutation. Definitions own defensive copies of nested collections so queued case snapshots cannot change under a running evaluation.

Fixture containment paths are limited to 32 assets to remain executable within voice evidence bounds. Proposed and forbidden operation expectations use mutation operations only; duplicate expectation entries and contradictory requirements are rejected.

Fixture parents must be containers or locations, matching production containment. An item cannot be a parent, and evaluation setup must not silently convert its kind.

## Outcome evaluation

The deterministic case evaluator compares application-observed outcomes, not model self-reports or exact prose. It checks outcome kind, required fixture references, required containment facts and required proposed mutation operations. Forbidden operations fail if proposed or executed. Any executed mutation fails the default fixture evaluation contract, which observes approval proposals without granting approval. Additional valid references are allowed unless separately prohibited by a future expectation. Unknown fixture references, invalid outcome kinds and malformed operations fail as invalid observations rather than counting as a passing expected failure. Results contain typed failure codes and safe fixture/operation identifiers; raw model text is not needed to judge these expectations.

Proposed mutation operation types form a closed expected set: every required type must occur and no unrequested type may be proposed. This prevents a requested move from passing alongside an unrequested archive. Exact proposal targets and arguments must additionally be checked by the evaluation runner's command assertions before this can serve as a complete action-quality gate.

## Exact proposal assertions

Proposal expectations use complete structured commands rather than operation names alone. Each command specifies mutation operation, existing target fixture ID where applicable, destination fixture ID where applicable, and new title/kind for creation. Existing targets and destinations must belong to the fixture; destinations must be containers or locations. Create commands use a new title/kind and cannot target an existing fixture ID. Non-create commands cannot carry creation fields; only create/move may carry a destination.

The judge compares the entire unordered collection, including command count. Wrong targets, wrong destinations, changed creation fields, extra commands and duplicate commands fail even when the operation type matches. Identical expected commands are rejected as an invalid case. Raw runtime IDs must be mapped to fixture IDs by the runner before judging. This replaces the preliminary operation-only proposal assertions; forbidden mutation rules remain operation-based.

Checkout and return assertions include the exact trimmed details text (at most the existing 500-rune voice detail bound). Other operations reject this field. The runner must fail observation mapping for semantic command arguments it cannot represent, including new-destination command references until graph assertions are implemented; it must never discard unsupported arguments and report an exact match. The current create compiler emits title/kind/parent only.

## Tenant workflow selection

A tenant has at most one selected workflow revision for production voice. Selection is an explicit tenant-scoped record containing workflow and revision identity; it is never inferred from draft update timestamps. Activation atomically compares the caller's expected tenant selection (both workflow and revision, or empty for none), updates that selection and the selected workflow's activation pointer, and writes audit. A stale activation conflicts even when switching between different workflows. Saving a draft cannot change selection. Sessions read and pin the selected immutable revision at start. An absent selection uses the documented default workflow. Selection must reference a revision belonging to that tenant and workflow.

## Model execution policy

A conversation execution owns a shared model-call budget across interpretation, repeated assessments and response wording. Each provider attempt consumes one call before dispatch, including failed attempts. Each step invocation can retry provider errors up to its configured attempts, but the shared call/elapsed budget and caller cancellation take precedence. A cancelled attempt is not retried. Execution deadlines use remaining elapsed allowance and propagate cancellation to providers. Execution requires an injected clock. Configured grounded response mode renders the brief without calling a wording provider. Step provider bindings are checked against explicit profile IDs and copied at construction; no fallback to a different provider is implicit. This policy component must be composed into both production and evaluation execution before workflow tuning is considered complete.

A provider result returned after its attempt deadline is rejected even if the provider reports success. An expired attempt exhausts model execution for that conversation. Workflow step instructions use a separate project-owned input field from provider prompt guidance and reach the provider as their own sanitized section; accepted Unicode instructions must not be silently byte-truncated or crowded out by long profile guidance. Existing credential redaction still applies.
