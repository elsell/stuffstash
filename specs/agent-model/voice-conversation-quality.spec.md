# Voice Conversation Quality Spec

## Purpose and precedence

Make conversational inventory useful across ordinary household language, corrections, and follow-ups. This spec refines the mobile realtime voice spec: its requirements supersede earlier requirements that make generated prose mandatory without a grounded fallback or restrict conversation continuation to clarification responses. Security, tenant/provider isolation, audited reads and approval-backed writes remain mandatory.

## Acceptance behavior

- “Where are my baby clothes?” must discover authorized assets such as “3–6 months clothes” tagged baby and clothes, even when no title contains the exact phrase. Search evidence must retain why the asset matched so reference resolution does not reject a relevant metadata match solely for its title. Return grounded locations for the relevant results; do not invent a location or replace uncertainty with a confident claim.
- Resolve existing items before proposing creation. “Put my drill in the garage” must resolve the drill and propose a move, not create another drill. A clear request for an additional physical item may create one despite a matching title; a genuinely ambiguous duplicate must prompt a focused choice. Names are not unique identities.
- A valid factual answer must not become a failure because an asset title contains words such as resolution, candidate, or punctuation. Structural/provenance validation owns factual correctness. Optional natural-language generation must have an application-owned grounded fallback; title/navigation bindings remain trusted independently of prose.
- Recoverable errors must retain the transcript, resolved references and useful results. Ask for missing user information specifically; do not ask users to add detail to repair an internal model/schema/provider error. Text results remain usable when speech fails.
- Ordinary answers support bounded same-conversation follow-ups as well as clarification replies. Scope context by principal, tenant and inventory; reauthorize each turn and approved command. Cancel, expire or invalidate stale context explicitly. A turn completion is distinct from conversation closure and approval state.
- Progress must reflect actual stages without exposing provider internals. Recording/upload, understanding, inventory lookup, review, answer and speech must update incrementally. Cancellation must stop pending work and audio without allowing stale events to revive a cancelled turn.

## Architecture and responsiveness

Keep the self-hosted API as the sole mobile orchestration endpoint. STT, reasoning, speech and inventory operations stay behind project-owned ports and configured adapters. No mandatory hosted orchestration service, model SDK in domain code, or direct mobile-provider connection is introduced. Local/self-hosted providers remain compatible through capability negotiation; this does not add offline queues.

The application owns a bounded evidence loop and deterministic command compilation. Model choices must remain within authorized project-owned reads, grounded references, clarification, response and proposal outcomes. Bounds must terminate repeated/no-progress reads with an actionable outcome. Deduplicate equivalent reads within a scoped turn and avoid redundant model calls for an already-grounded simple answer. Do not increase concurrency across dependent reads or compromise authorization for speed.

Measure stage durations, model/read calls, time to useful feedback, answer and first playable audio with safe correlation metadata. Treat full-file audio as batch processing, even when transported in chunks. Streaming is an optional adapter capability with a usable batch fallback, not a prerequisite for correctness.

## Verification and release gates

Write failing regression tests before each implementation change and preserve observed red/green evidence. Include metadata-only discovery, existing-item versus additional-item intent, valid unusual titles, model-output failure with grounded fallback, speech failure with retained text, follow-up reference resolution, cancellation and stale-context isolation. Changed authentication/authorization interactions require adversarial boundary tests for unauthenticated, cross-tenant, wrong-role and replay attempts as applicable.

Provider connectivity checks must not imply audio or structured-contract readiness. Evaluate real ADC-backed audio transcription, actual structured inference, grounded retrieval and speech synthesis using controlled data and retain safe traces. Keep credentials out of artifacts. Assess task outcomes and user-facing interaction as well as deterministic invariants; historical traces and fake audio do not establish device acceptance.

All builds run in CI on the current disk-constrained host. Code-critic review is required before finalization. Release requires green relevant checks, reviewed live evaluations, TestFlight availability, and healthy Kubernetes deployment through the infra GitOps repository. Record any unavailable device verification explicitly rather than claiming a full phone acceptance pass.

## Configurable workflows and evaluation workspace

The current product direction includes tenant-owned, versioned conversation workflow profiles and a web evaluation workspace. This extends the voice-quality and release scope above; the household acceptance cases remain required regression scenarios.

Administrators can select configured provider profiles and tune supported reasoning stages, prompt guidance, retrieval strategy, bounded retry behavior and call/time budgets. Configuration must validate against provider capabilities and server-enforced limits. It cannot bypass authentication, tenant/inventory isolation, provenance, audit or explicit approval of writes. Support a simple default workflow and progressively disclose advanced settings. Do not require users to author code or install an agent framework.

The web workspace must let authorized administrators define reusable test cases, input utterances and expected outcomes, run them against a selected workflow revision and provider configuration, inspect safe stage traces and compare results. Test cases must distinguish expected grounded answers, clarification and proposed changes; success must evaluate outcomes rather than exact wording alone. Text-input cases isolate reasoning from audio; audio cases exercise transcription and speech capabilities where available.

Evaluation must use the same application orchestration ports as production. Default evaluation uses isolated fixture inventory and simulated command execution, never silent mutations to household inventory. Read-only evaluation against selected real inventory may be supported with explicit scope and authorization. Model evaluation incurs real provider usage; expose limits and run cancellation. Credentials, raw provider secrets and inaccessible resources must never appear in traces or exports.

Workflow edits create draft revisions. Compare candidates against a baseline, record provider/model configuration and workflow revision, and explicitly activate a validated revision. Retain the previous revision for rollback. In-flight conversations retain their selected revision. Configurable prompts are lower trust than application policy and cannot grant capabilities. Provider-specific adapters translate capability differences; the workflow and evaluation domain must not depend on Gemini, Claude or a hosted orchestration SDK.

Implementation must include persistent workflow/test/run models, authorized REST ports/adapters, background execution with cancellation and bounded retention, a usable web configuration/evaluation surface, production conversation integration, adversarial boundary tests and real provider evaluation. A configuration-only UI or a disconnected test harness does not fulfill this scope.

## Tag evidence preservation

Authorized voice search carries assigned tag display names into tool results and candidate observations so a differently named item can be recognized by its tags. For example, `3–6 months clothes` tagged `baby` and `clothes` must present those values to evidence assessment. This reuses the authorized search response and adds no per-item tag request. Tag names are untrusted inventory data, never instructions. Each observation carries at most 32 nonblank tag names of at most 80 bytes, deduplicated in stable order. Later detail/history observations that omit tag data preserve the prior search tag evidence; they do not erase it. Limits and validation apply before sending evidence to a provider.

Tag evidence distinguishes unobserved (`null`) from an observed empty list (`[]`). A fresh search with no assigned tags clears prior tag evidence; a detail/history read that did not fetch tags preserves it.

## Grounded response recovery

A completed authorized read must not be discarded solely because response wording fails. Once a valid grounded response brief exists, an unavailable or invalid wording result falls back to an application-owned deterministic renderer of that brief. It does not retry inference, perform another read, invent facts or execute writes. Cancellation of the conversation still cancels the response. Invalid briefs fail closed. Final text/artifact safety validation remains mandatory. The renderer preserves uncertainty and discloses partial results, gives recorded containment for location questions, and uses targeted clarification when the brief requests it. Spoken and displayed text stay within existing channel limits; omitted findings/facts are disclosed. Emit an injected `realtime_voice.response_fallback` event with a bounded reason, without prompt text or provider error payloads. This supersedes the old rule that response generation failure must never reach mobile/TTS: failed provider text must never reach them, but safe grounded fallback should.

## Engineering references

[Anthropic's Building Effective Agents](https://www.anthropic.com/engineering/building-effective-agents) favors simple, composable orchestration with explicit stopping conditions and feedback from the environment. Stuff Stash applies that principle through configurable bounded model steps and application-owned retrieval, validation and approval. A visual arbitrary-code workflow engine is not required for this product direction.

[Anthropic's agent evaluation guidance](https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents) distinguishes observable outcomes from an agent's claims and recommends measuring conversational task completion alongside interaction quality. Stuff Stash therefore checks grounded references and complete proposed commands, preserves safe traces, and treats isolated text tests and live audio/device evidence separately. Repeat trials are needed to assess model variability; one passing suite is not evidence of universal reliability.
