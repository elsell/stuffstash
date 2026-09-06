# Voice Conversation Quality Spec

## Purpose and precedence

The replacement runtime is defined by [Model-led Voice Loop](model-led-voice-loop.spec.md). The model chooses its tools, questions, proposals and natural responses. Retired investigation, phrase-validation and deterministic response fallback mechanisms must remain deleted.

Make conversational inventory useful across ordinary household language, corrections, and follow-ups. This spec defines outcome and release acceptance for the mobile realtime voice flow, including ordinary answer follow-ups. Security, tenant/provider isolation, audited reads and approval-backed writes remain mandatory.

## Acceptance behavior

- “Where are my baby clothes?” must discover authorized assets such as “3–6 months clothes” tagged baby and clothes, even when no title contains the exact phrase. Search evidence must retain why the asset matched so the model can recognize a relevant metadata match despite a different title. Return grounded locations for the relevant results; do not invent a location or replace uncertainty with a confident claim.
- Resolve existing items before proposing creation. “Put my drill in the garage” must resolve the drill and propose a move, not create another drill. A clear request for an additional physical item may create one despite a matching title; a genuinely ambiguous duplicate must prompt a focused choice. Names are not unique identities.
- A valid factual answer must not become a failure because an asset title contains words such as resolution, candidate, or punctuation. Application validation enforces output shape and observed navigation references; outcome evaluations assess factual correctness and useful wording. The model authors the answer, with no deterministic factual fallback. Titles and navigation targets come from authorized metadata.
- Recoverable errors must retain bounded conversation history and useful results. Ask for missing user information specifically; do not ask users to add detail to repair an internal model/schema/provider error. Text results remain usable when speech fails.
- Ordinary answers support bounded same-conversation follow-ups as well as clarification replies. Scope context by principal, tenant and inventory; reauthorize each turn and approved command. Cancel, expire or invalidate stale context explicitly. A turn completion is distinct from conversation closure and approval state.
- Progress must reflect actual stages without exposing provider internals. Recording/upload, understanding, inventory lookup, review, answer and speech must update incrementally. Cancellation must stop pending work and audio without allowing stale events to revive a cancelled turn.

## Architecture and responsiveness

Keep the self-hosted API as the sole mobile orchestration endpoint. STT, reasoning, speech and inventory operations stay behind project-owned ports and configured adapters. No mandatory hosted orchestration service, model SDK in domain code, or direct mobile-provider connection is introduced. Local/self-hosted providers remain compatible through capability negotiation; this does not add offline queues.

The application owns a bounded model/tool loop and validates typed proposals through domain services. The model chooses useful reads, questions, answers and proposals. Each utterance has explicit model-call, tool-call and processing-time limits. Avoid redundant calls through useful tool metadata and retained context, without rejecting equivalent queries by semantic policy. Dependency ordering and authorization take precedence over speculative concurrency.

Measure stage durations, model/read calls, time to useful feedback, answer and first playable audio with safe correlation metadata. Treat full-file audio as batch processing, even when transported in chunks. Streaming is an optional adapter capability with a usable batch fallback, not a prerequisite for correctness.

## Verification and release gates

Write failing regression tests before each implementation change and preserve observed red/green evidence. Include metadata-only discovery, existing-item versus additional-item intent, valid unusual titles, provider failure without activating a legacy fallback, speech failure with retained text, follow-up reference resolution, cancellation and stale-context isolation. Changed authentication/authorization interactions require adversarial boundary tests for unauthenticated, cross-tenant, wrong-role and replay attempts as applicable.

Provider connectivity checks must not imply audio or structured-contract readiness. Evaluate real ADC-backed audio transcription, actual structured inference, grounded retrieval and speech synthesis using controlled data and retain safe traces. Keep credentials out of artifacts. Assess task outcomes and user-facing interaction as well as deterministic invariants; historical traces and fake audio do not establish device acceptance.

All builds run in CI on the current disk-constrained host. Code-critic review is required before finalization. Release requires green relevant checks, reviewed live evaluations, TestFlight availability, and healthy Kubernetes deployment through the infra GitOps repository. Record any unavailable device verification explicitly rather than claiming a full phone acceptance pass.

## Configurable workflows and evaluation workspace

The current product direction includes tenant-owned, versioned conversation workflow profiles and a web evaluation workspace. This extends the voice-quality and release scope above; the household acceptance cases remain required regression scenarios.

Administrators can select configured provider profiles and tune a single configured conversation model, optional prompt guidance, per-turn model/tool/time budgets and a session follow-up limit. Configuration must validate against provider capabilities and server-enforced limits. It cannot bypass authentication, tenant/inventory isolation, provenance, audit or explicit approval of writes. Support a simple default workflow and progressively disclose advanced settings. Do not require users to author code or install an agent framework.

The web workspace must let authorized administrators define reusable test cases, input utterances and expected outcomes, run them against a selected workflow revision and provider configuration, inspect safe conversation traces and compare results. Test cases must distinguish expected grounded answers, clarification and proposed changes; success must evaluate outcomes rather than exact wording alone. Text-input cases isolate reasoning from audio; audio cases exercise transcription and speech capabilities where available.

Evaluation must use the same application orchestration ports as production. Default evaluation uses isolated fixture inventory and captured proposals without command execution, never silent mutations to household inventory. Read-only evaluation against selected real inventory may be supported with explicit scope and authorization. Model evaluation incurs real provider usage; expose limits and run cancellation. Credentials, raw provider secrets and inaccessible resources must never appear in traces or exports.

Workflow edits create draft revisions. Compare candidates against a baseline, record provider/model configuration and workflow revision, and explicitly activate a validated revision. Retain the previous revision for rollback. In-flight conversations retain their selected revision. Configurable prompts are lower trust than application policy and cannot grant capabilities. Provider-specific adapters translate capability differences; the workflow and evaluation domain must not depend on Gemini, Claude or a hosted orchestration SDK.

Implementation must include persistent workflow/test/run models, authorized REST ports/adapters, background execution with cancellation and bounded retention, a usable web configuration/evaluation surface, production conversation integration, adversarial boundary tests and real provider evaluation. A configuration-only UI or a disconnected test harness does not fulfill this scope.

## Tag evidence preservation

Authorized voice search carries assigned tag display names into tool results and candidate observations so a differently named item can be recognized by its tags. For example, `3–6 months clothes` tagged `baby` and `clothes` must expose those values in authorized tool results. This reuses the authorized search response and adds no per-item tag request. Tag names are untrusted inventory data, never instructions. Each observation carries at most 32 nonblank tag names of at most 80 bytes, deduplicated in stable order. Later detail/history observations that omit tag data preserve the prior search tag evidence; they do not erase it. Limits and validation apply before sending evidence to a provider.

Tag evidence distinguishes unobserved (`null`) from an observed empty list (`[]`). A fresh search with no assigned tags clears prior tag evidence; a detail/history read that did not fetch tags preserves it.

## Failure and recovery

Tool argument errors return safe structured feedback to the model for correction within the remaining budget. Authorization failures terminate processing; provider failures remain explicit failures. Retain completed safe history for a permitted follow-up without claiming unfinished operations succeeded. An answer already delivered as text remains visible if speech synthesis fails. No response brief, deterministic renderer or retired investigation executor may answer in place of the model.

## Engineering references

[Anthropic's Building Effective Agents](https://www.anthropic.com/engineering/building-effective-agents) favors simple, composable orchestration with explicit stopping conditions and feedback from the environment. Stuff Stash applies that principle through a bounded model/tool conversation with application-owned retrieval, validation and approval. A visual arbitrary-code workflow engine is not required for this product direction.

[Anthropic's agent evaluation guidance](https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents) distinguishes observable outcomes from an agent's claims and recommends measuring conversational task completion alongside interaction quality. Stuff Stash therefore checks grounded references and complete proposed commands, preserves safe traces, and treats isolated text tests and live audio/device evidence separately. Repeat trials are needed to assess model variability; one passing suite is not evidence of universal reliability.

## Immediate live voice acceptance milestone

Pause further workspace features until recorded speech through the authenticated realtime WebSocket, configured STT, language and TTS providers produces a useful grounded answer. The baseline household question is “Where are my baby clothes?”; successful speech must state the observed locations, not merely list matching titles or leave locations only on cards. The model should distinguish locating a category from listing a container’s contents, without an application request taxonomy or hard-coded household terms.

Use read-only live requests and retain sanitized transcripts, model-selected tool calls, completion events, response artifacts, speech bytes and latency. A successful transport session with an unhelpful answer fails product acceptance. Repeat runs to expose variability; distinguish deployed behavior from branch behavior and recorded synthetic speech from device microphone/playback testing. Temporary changes to the existing tenant prompt may isolate interpretation faults; retain the previous value and restore it if the experiment does not establish a reliable improvement. This does not add configuration features or loosen validation.

A CI-built Linux HTTP adapter test executable may be run on the authorized ADC host to evaluate un-deployed branch code without local builds. The opt-in test uses controlled in-memory inventory through application ports, real STT/conversation/TTS adapters, and the same realtime WebSocket messages as mobile. Record its exact commit and distinguish it from deployed OIDC/provider-profile routing and physical device verification. Missing explicit live-test inputs must fail when the test is enabled, not silently skip. Keep credentials on the ADC host and limit published traces to fixture data; never upload credentials with binaries or evidence.

Live fixture diagnostics may record bounded successful provider response text and tool calls/results through a test-only HTTP transport wrapper. Never record thought content, opaque signatures, authorization headers, token exchanges or raw provider error bodies. Deterministic checks enforce command graphs and security invariants; human trace review assesses whether the answer or clarification is useful.

## Conversational existence answers

The device question “Do I have any chemicals?” exposed a product failure: successful speech consisting of repeated “Found <item>” rows is not conversational acceptance. An existence answer must answer the question directly. Spoken prose may summarize the grounded category and name a useful subset of matches; the display and cards retain the model-selected relevant findings with trusted navigation. Do not require every result title in speech or treat spoken summarization as incomplete retrieval. Retrieval truncation and uncertainty must still be disclosed, and unsupported existence, location, quantity or state claims remain invalid. Empty findings cannot support a positive answer.

Preserve model-authored natural wording without semantic prose filtering. Do not replace this behavior with a new canned success template. Regression coverage must exercise a category with multiple distinct item names, accept a concise supported spoken answer with relevant cards, and flag contradictory absence or unrelated answers as evaluation failures. This supersedes earlier requirements to repeat every existence finding in both channels. Validate the real provider response and audio as well as deterministic tests before claiming the device issue fixed.
