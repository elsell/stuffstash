# Conversation Configuration Workspace

## Product scope

This workspace implements the tenant-owned workflow and evaluation behavior in `conversation-workflows.spec.md`. Its user is a household administrator tuning conversations for their inventory and configured models. Success means saving a revision, exercising realistic cases, understanding failures and latency, and activating a proven revision without editing code. The default conversation remains usable without opening this workspace.

The user authorized design and implementation of this feature. Integrate with the established settings shell and visual primitives; do not introduce a separate administration application or redesign global navigation. A working SvelteKit candidate, responsive/accessibility review and real API integration are still required evidence before completion. Local builds are prohibited on the current host; use CI artifacts for compiled verification.

## Navigation and presentation

Add Conversations to tenant settings at `/settings/tenants/{tenantId}/conversations`. Keep tenant context visible. The overview shows the active workflow and revision, provider compatibility, recent evaluation result and clear Edit workflow, Test cases and Runs destinations. Tenant configure permission controls the surface. A directly opened denied route shows a denial state and a route back to settings, with no cached fixture text.

Workflow editing shows the fixed sequence Understand, Look up and assess, Respond. Each model step exposes configured profile selection and optional instructions. Retrieval policy, retries and shared budgets belong in an Advanced section. Grounded response mode makes its unused model selection visibly inapplicable. Saving creates a draft revision; it does not activate it. Keep unsaved input on recoverable errors. A stale-save conflict offers reloading the latest revision and preserves the user's text for comparison.

Test cases provide title, utterance, fixture assets and expected result. Fixtures use familiar item/container/location names, tags and parent selection rather than live inventory IDs. Expected answers can identify items and their locations; expected changes name target, destination and exact additional fields. Fixture and expectation validation is inline, with a summary linking to invalid fields. Case revision history is read-only. Audio and follow-up cases must be clearly distinguished from text-only coverage when those execution capabilities are added.

Running selected cases requires a saved workflow revision. Show the configured models, case count and maximum usage before dispatch. The run page displays Queued or current case progress immediately, completed results as they arrive, elapsed time, model calls and Cancel. Each failure expands into expected versus observed facts and a bounded safe stage trace. A failed assertion differs from a provider/configuration or worker failure. Cancelled and unrun cases cannot appear passed.

Revision comparison uses the same pinned case versions and compatible provider configuration, with both quality and latency visible. Activation is a separate action available only when the server's current quality gate passes. A changed provider or case suite invalidates old passing evidence. Show the server's reason and a Run tests action. Rollback chooses a historical revision through the same checked activation command. Text-only results never imply microphone or speech quality verification.

## Responsive and accessible behavior

Use existing local shadcn-style controls and semantic tokens. Desktop may place fixture editing beside expectations; narrow screens stack them in reading order. Lists remain bounded and use Load more, without requiring a wide table. Long titles and provider names wrap. Form controls have persistent labels, associated errors and keyboard access. Announce saved/error/terminal state through a polite live region without announcing every polling refresh. Maintain focus through saves, expansion and cancellation, and move focus to the validation summary only when submission fails. Status uses text as well as color. No sticky controls may obscure fields or results.

## Frontend ownership and network behavior

Introduce focused frontend domain types and separate workflow, case and run repository ports. Generated DTOs stay inside API adapters and mappers. Routes compose editors and result components; they do not perform transport mapping or store credentials. Reuse runtime API/auth configuration and injectable workspace observability.

TanStack Query owns remote revisions, heads and runs. Query keys include API identity, signed-in principal and tenant, plus resource/revision identity; sign-out or lost permission clears sensitive cached data. Draft fields stay local to the editor. Saving reconciles the returned immutable revision and invalidates only affected heads and lists. Never optimistically mark a revision active or a run passed. Only nonterminal runs poll, initially every two seconds with bounded error backoff; pause polling in hidden tabs and stop on terminal state. Cancellation uses the repository port and reconciles the server result. Query cancellation reaches the transport's AbortSignal. Do not repeatedly fetch provider credentials or unbounded history.

The web package does not currently depend on TanStack. Add an exact reviewed Svelte-compatible version and lockfile change through the client tooling specification before implementation; do not assume the mobile query integration supplies the web dependency.

## Verification

Use behavior tests with fake repositories for draft preservation, conflict recovery, scoped caches, cancellation and terminal polling. Adapter tests cover full fixture/expectation mapping and server error translation. Browser tests cover the complete save-case/run/review flow, direct-route denial, revoked permission, long Unicode labels, keyboard navigation and narrow layout. Real provider runs and audio/device gates remain separate from fake UI evidence. Passing fake UI tests alone does not fulfill the activation or release evidence gates.

The conversation query client uses 30-second default freshness and five-minute inactive retention, with no focus-triggered refetch. It shares in-flight reads and retries a transient read once. Authentication/authorization failures are never retried and cancel/clear the conversation client before notifying its owner to enter a denied state. Keys contain separate API/principal/tenant segments, then resource/revision segments. Create the client per authenticated workspace, never module-global or shared across SSR requests. Run polling starts at two seconds, doubles on read failures up to 30 seconds, and stops for hidden pages or terminal runs. Sign-out/context replacement cancels and clears the client. Immutable revision queries may use infinite freshness; activation and progress never use optimistic success.
