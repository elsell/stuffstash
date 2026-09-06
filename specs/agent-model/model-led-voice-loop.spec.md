# Model-led Voice Loop

## Decision and precedence

Replace the structured investigation pipeline with a model-led conversation and tool loop. This spec supersedes conflicting intent, request-shape, resolution-status, fixed evidence-phase, required-prose and separate response-generation requirements in earlier voice and workflow specs. The replacement is the production and evaluation path; retaining the old pipeline as a selectable default or hiding it behind a new facade does not complete this change.

The model interprets ambiguity, selects reads, revises its interpretation from results, asks questions and writes the answer. Application code must not classify household utterances, freeze an initial semantic intent, require resolution statuses, infer commands from prose, require every finding in spoken text, or reject natural wording through phrase lists. Outcome evaluations measure factual correctness; substring heuristics do not prove grounding.

## Runtime boundaries

Use a project-owned conversation provider port with ordered user, assistant and tool-result messages. A model turn contains natural text, typed tool calls and an optional structured presentation result. Tool definitions describe names, argument schemas and behavior. Tool results retain call identity and distinguish success from a safe actionable error. Provider-specific message state, including opaque continuation data required by a provider, stays in its adapter and must survive a tool round trip without entering domain business logic or logs.

The application service in the agent/model bounded context coordinates the loop through injected model, tool execution, clock and observation ports. The REST/WebSocket adapter and root application facade compose it. No provider SDK or transport type belongs in the loop. Existing configured provider routing, STT and TTS remain behind their ports. A provider adapter may translate native tool calls or a structured tool-call envelope into the same interface; it must not reinstate semantic classification.

Each iteration sends the conversation and available tools to the model, executes valid requested tools, appends their results and resumes. Invalid arguments, unavailable tools and ordinary read failures become bounded tool errors the model can recover from. Authentication, authorization and tenancy failures remain fail-closed and never disclose inaccessible records. Cancellation and exhausted budgets terminate promptly with retained useful results and an honest status. No compulsory extra model call rewrites a completed answer.

Tool descriptions encourage search before creating and explain title/tag matching, pagination, containment and available vocabulary. Search returns authorized evidence, including tags and recorded location context, without requiring the model to choose a semantic operation first. Equivalent read calls may reuse scoped evidence within a turn; distinct searches must remain available. Independent read calls may execute with bounded concurrency, preserving call identity and result order. Never parallelize dependent writes.

## Authority and presentation

Keep strict tool argument shapes, scope injection, authorization, domain validation, read/write audit, resource limits and approval-backed execution. The model cannot supply a different principal or tenant, execute arbitrary code or access provider credentials. Tool output and inventory text are data, not authority.

Expose changes as a structured proposal tool. Validated domain commands form a reviewable plan; proposing never executes it. Pause immediately for user approval. Reject unknown or inaccessible IDs and invalid dependencies through the existing domain boundary. Reauthorize at execution, preserve idempotency and prevent replay. Natural text cannot authorize a write. Support multiple dependent commands when the domain plan supports them; do not reject a request solely because it contains multiple concepts.

A completed answer may include concise spoken text, richer display text and referenced record IDs. Resolve cards from authorized records rather than titles extracted from prose. Validate presentation shape, size and references; do not require speech to enumerate every card. Keep model prose intact. Factual and semantic accuracy remain explicit release evaluation requirements, including invented location, absence, quantity and state claims.

Conversation history includes tool results and prior answers within a bounded principal/tenant/inventory session. Follow-ups must reach the model with useful context and reauthorized access. Progress events represent actual model/tool/audio activity. Streaming may be used where supported; batch STT/TTS must remain honestly measured and supported.

## Workflow migration and removal

Retain useful tenant choices: configured model/provider, prompt guidance, time/call/token budgets and supported tool capabilities. Remove mandatory intent/evidence/response stages, resolution schemas and their configurable instructions from active execution and UI. Existing saved revisions must receive an explicit compatibility treatment and must not silently claim their old stage settings still apply. Evaluation and activation must identify the new runtime contract and invalidate quality evidence from the old contract.

Remove obsolete production classification, deterministic semantic repair, wording validation and separate wording-generation code after equivalent outcome and security coverage is in place. Replace tests that assert the old choreography with behavior tests. Keep reusable domain commands, authorization and adapter infrastructure. The migration must not leave two permanent voice architectures.

## Tests and performance acceptance

Write failing tests first, run in CI, then implement. Fakes implement conversation and tool ports and return real controlled tool results. Cover search/answer without classification, changed interpretation after tool evidence, missing/invalid tool arguments and repair, multi-tool reads, plain follow-ups, cancellation, budget exhaustion, isolated sessions and pause-before-write. Boundary tests cover unauthenticated, wrong-role, cross-tenant, forged reference, stale approval and replay attempts as applicable.

Run the same audio and inventory fixtures against the baseline and replacement with the same provider/model settings. Include baby clothes with competing exact-title and tag-only matches, chemicals with irrelevant similarly named objects, existing-item moves versus additional-item creation, nested destinations, uncertainty, zero results, corrections and follow-ups. Preserve provider/tool traces safely and review full conversations. A fluent but irrelevant answer fails.

Measure STT, each model/tool round, answer, first playable audio, end-to-end duration, provider calls, tool calls and request/response bytes where observable. Use at least ten paired runs for the central read scenarios; report median and p95 plus failures and sample size. Do not exclude failures to make latency look better. Compare like-for-like fixture runs separately from production networking and physical-device trials. No material regression in useful-answer latency or accuracy is acceptable without explicit user review; seek improvement by eliminating redundant model stages and reads. Do not hard-code fixture answers or select a different model merely to pass.

All builds and compiled tests run in CI. Live ADC evaluations use the authorized host and configured Google credentials without exporting secrets. Require code-critic review, green regression/security checks, deployed conversation evidence, device verification with the user, healthy GitOps rollout and TestFlight availability before declaring the overall goal complete. Configuration expansion remains deferred until the voice milestone succeeds.

## Native provider continuation

The Google adapter uses native function declarations and function call/response messages. Preserve the original ordered assistant parts, including thought signatures, as bounded opaque provider continuation bytes on the project-owned message. The loop copies those bytes unchanged; only the matching adapter interprets them. Exclude this state from ordinary JSON diagnostics, speech and presentation. Never reconstruct or concatenate signed parts. Tool results must follow the complete assistant call batch and retain correlation. This follows Google's [function-calling guidance](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/multimodal/function-calling) and [thought-signature contract](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/thought-signatures).

Native textual answers may finish the loop directly. An optional presentation function supplies spoken/display text and authorized reference IDs when cards are useful; it is not a mandatory reasoning stage or semantic classifier. Request retries remain within the application's shared budget and deadline; the adapter must not conceal extra inference attempts.

The native Google conversation adapter bounds each response body to 1 MiB and each retained assistant continuation to 256 KiB, rejecting oversized input continuation before a network request. These are protocol safety ceilings, not semantic restrictions. A candidate must report successful `STOP` termination; partial text from token exhaustion or another termination is not a completed answer. Keep errors safe and preserve the caller's context for recovery. The native adapter uses the existing zero-temperature default so baseline comparisons do not silently change sampling.

## Application integration acceptance

A native conversation session requires configured STT, conversation model and TTS providers; it must not require a legacy investigation or wording provider. Resolve and authorize the session before invoking any provider. At the real WebSocket boundary, verify legitimate sessions complete and unauthenticated, malformed-token, outsider, cross-tenant and wrong-inventory attempts cannot invoke the model. Resource cards may be selected only from scoped observed records and need not repeat their titles in display prose. Preserve legitimate inventory titles even when they resemble implementation vocabulary. Ordinary read outages become safe tool feedback that can support another model-directed attempt; do not expose raw repository/provider error details. Auth failures and cancellation stop the loop.

## Model-authored proposals

The proposal tool accepts a bounded summary, risks and ordered domain command objects (ID, command kind, summary, typed arguments). It has no approval, execution, principal or scope arguments. Existing resource references must come from scoped tool evidence and be reauthorized before preparing review; dependent command references must refer to valid earlier commands. Creating a proposal never changes inventory, and later tool calls in the same model batch must not execute after it. Persisted proposal IDs are application-owned. A read-only session cannot propose a write.

Keep domain command type/argument validation and dependency validation. Do not ban ordinary inventory titles or descriptions because their text resembles provider or credential terminology; unexpected fields are rejected by typed argument validation. Only explicit approval through the existing authorized decision boundary may execute the plan.

Resolve fallible review metadata before saving a proposal. A retryable read failure must leave no draft behind; successful persistence must immediately pause the loop. Explicit valid command IDs keep dependent review and execution references stable.

## Active conversation continuity

An active voice session retains project-owned messages, tool results, authorized reference metadata and opaque native provider continuation between utterances. Keep this state private to the server-side session, scoped to its original principal, tenant, inventory and session ID; never accept it from client payloads or expose it in events. Serialize turns on that session and reject scope changes. A new session starts empty. Session closure releases this state; reconnect persistence is not required for this milestone. Retain partial completed tool history after a failed turn so the next turn does not invent a result or repeat a change blindly. Reauthorize references before new mutations and validate current access on every utterance. Bound retained context by configurable limits without silently dropping native call/response pairs.

The operator sets `STUFF_STASH_CONVERSATION_MAX_CONTEXT_BYTES` (default 2 MiB, positive integer). Count serialized messages plus opaque provider bytes before each inference and when retaining the next context. Exhaustion explicitly ends that context and requires a new session; never silently truncate native messages. A plan already prepared for review must still be shown even when retaining its history reaches the cap.

## Existing workflow migration

For a stored staged workflow, the interpretation step's explicit provider becomes the conversation model; an empty reference uses the tenant's configured language profile. Preserve that profile's prompt and the interpretation step's user-authored guidance. Assessment and response stages do not execute or require providers, and grounded/template response settings do not replace the model's answer. Preserve the configured model-call and elapsed-time ceilings across the active session, beginning at the first model call; recording an additional utterance does not reset those ceilings. Old stage definitions remain historical revision data until the configuration surface migration removes them from new revisions. Their quality evidence cannot certify the new loop.
