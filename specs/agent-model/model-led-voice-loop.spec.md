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
