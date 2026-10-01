# Spec–implementation audit — October 1, 2026

This report preserves the October 1 baseline. Current delivery status lives in [the roadmap](../../specs/platform/roadmap.spec.md).

Baseline: `origin/main` at `dabe2839599c80c5db0ed22c8792adef1265ce95`, including PR209. This is a source and recorded-evidence audit, not a new device, production, security-penetration, or performance acceptance run. No product code or specs were changed. Only this report was added.

The inventory covers all 76 specification files and their product/engineering families. Source inspection was targeted at the corresponding entrypoints, ports, adapters, client surfaces, tests and existing evidence. The inventory at the end makes scope explicit; it is not a claim that every normative sentence has been independently proved. Absence claims below combine source searches with composition/route inspection, rather than treating an old TODO as evidence.

## Confirmed implementation gaps

### G1 — Inventory export is missing

**Impact:** users cannot export their inventory through the promised JSON/CSV product workflow. Priority: high product-completeness work.

[Import/export requirements](../../specs/data-portability/import-export.spec.md) explicitly require both JSON and CSV export. The implementation contains Homebox live/CSV imports and durable job infrastructure, but no export application port, export use case, route or client export workflow was found. The checked-in [OpenAPI contract](../../packages/api-client/openapi.json) has 99 paths and no export route; [HTTP composition](../../apps/api/internal/adapters/httpserver/api.go) registers imports only.

The export schema and media packaging remain undecided. Close the gap by specifying the smallest useful inventory export schema, then implementing authorization-preserving JSON/CSV exports and adversarial boundary coverage. Native JSON/CSV **import** is a separate future adapter; Homebox CSV import does not satisfy native round-trip portability.

### G2 — External MCP is specified but absent

**Impact:** external agents cannot use a Stuff Stash MCP endpoint. Priority: high if external-agent integration is a current product commitment.

The [MCP spec](../../specs/agent-model/mcp-agent-tools.spec.md) defines authenticated Streamable HTTP, discovery/read tools, authorization and adversarial tests. No MCP transport/adapter or bootstrap registration exists. The internal [conversation tool executor](../../apps/api/internal/app/realtime_voice_conversation_tools.go) is not an external MCP server.

Implement the first read-only transport over existing application services. External writes are deliberately gated on an approval contract; their absence is not an invitation to expose direct mutation tools.

### G3 — Provider-profile choices exceed executable provider support

**Impact:** representing a local/OpenAI-compatible profile does not make it usable for conversation. Priority: high for self-hosted/local-model positioning.

[Conversational inventory](../../specs/agent-model/conversational-inventory.spec.md) requires local HTTP and remote profiles and OpenAI-compatible design support; [realtime compatibility](../../specs/agent-model/realtime-interaction.spec.md) calls for an OpenAI-compatible adapter. The [provider model](../../apps/api/internal/domain/agentmodel/provider_profile.go) accepts `gemini`, `openai_compatible` and `local_http`, but [runtime composition](../../apps/api/cmd/stuff-stash/internal/bootstrap/voice.go) injects [GoogleProviderProfileFactory](../../apps/api/internal/adapters/voice/google_profile_factory.go), which rejects non-Gemini kinds.

The ports are present; executable alternative adapters are missing. Implement and verify a supported adapter family, or explicitly narrow supported runtime choices while tracking the larger commitment. The later Google-first mobile slice explains the current implementation, but does not fulfill local-provider support.

### G4 — Web conversational input is missing

**Impact:** mobile has typed/voice inventory conversation; the web workspace does not have the corresponding user interaction. Priority: high for client parity.

[Conversational inventory](../../specs/agent-model/conversational-inventory.spec.md) requires typed commands from web and mobile; [realtime interaction](../../specs/agent-model/realtime-interaction.spec.md) describes both clients connecting to the API. Web has an administrative conversation workflow/case/run workspace, not a user conversation session. No product WebSocket/realtime-session transport or user typed-turn client was found in `apps/web/src`.

Do not count workflow evaluation text inputs as the missing inventory conversation experience. A minimal typed web conversation with approval/recovery semantics would close the first part; browser microphone/speech behavior is a separate acceptance scope.

### G5 — Conversational edits do not cover the broader custom-field promise

**Impact:** users can manually edit data that they cannot change conversationally. Priority: medium; specify the additional command contract before implementation.

[Flexible fields](../../specs/assets/flexible-asset-fields.spec.md) says conversation must read/update custom fields and describes confirmed definition/type creation. Current [create argument mapping](../../apps/api/internal/app/action_plans_create_arguments.go) supplies an empty `CustomFields` map. The [command validator](../../apps/api/internal/app/action_plans_validation.go) sends `update_asset` to the [expiration-only parser](../../apps/api/internal/app/action_plans_expiration_arguments.go); it requires asset ID and expiration, not arbitrary title, description or custom-field changes. No definition/type-creation command exists in the [command enumeration](../../apps/api/internal/domain/actionplan/action_plan.go).

This is incomplete broad scope, not proof that the narrower approved expiration command is defective. Define field resolution, allowed patches, type validation and review/approval behavior before expanding execution. General asset-detail updates are likewise not fulfilled merely by having a command named `update_asset`.

### G6 — Localization readiness is not implemented across the clients

**Impact:** adding another language requires editing components and presentation logic instead of supplying translations. Priority: medium, foundational.

[Brand requirements](../../specs/platform/brand-guidelines.spec.md#internationalization-and-localization-requirements) prohibit embedded user-facing strings in reusable client components, require localization-aware pluralization and recommend pseudolocalization before broad UI release. Mobile and web still embed English labels, statuses and messages directly; there is no shared translation/catalog/pluralization system in the reviewed source. Examples include [Browse header](../../apps/mobile/src/ui/screens/BrowseAddHeader.tsx) and [web Home](../../apps/web/src/lib/components/workspace/HomeWorkspace.svelte).

Some dates already use `Intl`; this does not close the string/pluralization gap. Extract messages and establish plural rules first, then verify expansion and RTL. This finding does not assert a reproduced RTL or large-text defect.

### G7 — Visible-image telemetry is not connected to image rendering

**Impact:** enabling client telemetry can measure requests, but cannot explain the complete visible-image loading experience promised by the performance spec. Priority: medium; useful before more image optimization.

[Observability](../../specs/platform/observability.spec.md) requires visible-image duration/outcome by surface and variant. Both clients define an image-capable observer: [mobile](../../apps/mobile/src/application/observability/PerformanceObserver.ts), [web](../../apps/web/src/lib/ports/performanceObserver.ts). Mobile composition exposes it; web creates a context. No production UI/component consumers of those observers were found. Request transport instrumentation does exist.

Wire actual image lifecycle events to these ports, including completion, failure, replacement and cancellation. Then measure rendering on a named client/runtime. Backend thumbnail timing or HTTP request timing alone is not equivalent.

### G8 — Application package migration remains incomplete

**Impact:** domain responsibilities remain coupled through the root application package; this is maintainability/spec drift, not a demonstrated user-visible failure. Priority: medium/low, preferably addressed as affected domains change.

[Monorepo rules](../../specs/platform/monorepo-layout.spec.md#go-workspace) allow a compatibility facade while requiring domain behavior to move into domain packages. The root `internal/app` still contains 72 non-test Go files, including substantive tenant/inventory creation and authorization-outbox behavior in [inventory.go](../../apps/api/internal/app/inventory.go), import orchestration in [imports.go](../../apps/api/internal/app/imports.go), and action-plan/conversation orchestration. These are not all forwarding facades.

Assets, search, notifications, expiration, media and other responsibilities already have dedicated packages. Complete remaining migrations incrementally; do not turn this into a rewrite or claim hexagonal architecture is wholly absent.

## Spec and audit-record drift

The [baseline roadmap](https://github.com/elsell/stuffstash/blob/dabe2839599c80c5db0ed22c8792adef1265ce95/specs/platform/roadmap.spec.md#known-gaps) is not a reliable current implementation checklist. Confirmed counterexamples:

| Roadmap claim | Current source evidence |
| --- | --- |
| Rate limiting unimplemented | Default-enabled configuration, limiter port, token-bucket implementation and server middleware exist. |
| SpiceDB checks search inventories individually | `authorizer.go` uses `LookupResources` with candidate intersection. |
| Production direct-upload adapter absent | `s3_direct_upload.go`, completion validation and blob adapter tests exist. |
| Invitations are bare tokens with no acceptance journey | Canonical-link/API behavior, web `/invitations/accept`, and mobile acceptance flows exist. |
| Web media management and detail edit/move tests missing | AssetDetail, route-backed overlays, media actions and focused component/browser tests exist. Some live acceptance remains distinct. |
| Import/export is all later work | Durable Homebox live and CSV import are implemented; export is the missing side. |
| Workflow editor not wired into settings | The web conversation workspace includes Workflow, Case and Run screens under settings. |

The current delivery summary still names 0.25.0 while later changes and tag `v0.25.3` exist. This is a status-reconciliation question, not proof of Apple processing; this audit did not recheck TestFlight.

Other historical notes mix obsolete failing tests/pending runs with subsequent successes. `image-performance-evidence.spec.md` still contains earlier missing-profiling/type-check statements despite later profiling and client-session implementation. Preserve historical evidence, but explicitly date/supersede it rather than leaving it indistinguishable from current requirements.

Some limitations in the roadmap are correct but are **not violations of the scoped specs**: non-asset undo, destructive field-schema changes, model use of photos, and API-key TTS. The current provider spec requires ADC/OAuth for Google Cloud TTS; it does not require API-key TTS. The media spec explicitly defers model-provider image use beyond preparation.

Recommended documentation correction: one current capability/status index linked to durable evidence, with a clear distinction between binding requirements, intentionally staged scope, historical observations and future proposals. Do not rewrite product requirements to disguise missing implementation.

## Acceptance and verification gaps

These are missing proof, not newly confirmed product defects.

1. **Comprehensive mobile design acceptance remains open.** The committed ledger contains 147 surfaces ×24 axes =3,528 cells: 2,672 source-reviewed, 576 marked finding, 57 runtime-partial, 12 unverified and 211 not applicable. These labels are historical tracking data, not 576 distinct unfixed defects. The current report itself says the audit is incomplete, and later fixes have not been consistently reconciled into one current coverage picture. Validate shared controls plus representative connected workflows; do not run 3,528 independent tests.
2. **Real authenticated browser journeys are not equivalent to fixture E2E tests.** The Playwright workspace fixture seeds a stored session and intercepts API calls. It is valuable UI coverage, including viewer-denied cases, but does not prove the full browser→Dex/OIDC→API→SpiceDB journey. Existing API/Dex scripts cover different boundaries. The connected production-shaped journey still needs explicit evidence against the relevant requirement.
3. **Physical/device and assistive acceptance is narrower than source coverage.** No new VoiceOver/TalkBack, RTL, broad accessibility-size, physical Android, microphone/camera or production-provider tests were performed for this audit. Previously confirmed notification delivery must not be reopened as an assumed gap. Complete normal-text workflow review first, per the user's priority.
4. **Performance evidence remains bounded.** The existing thumbnail report records improved median but a worse slowest sample and unresolved storage stalls; source request budgets do not establish end-to-end production latency. Missing visible-image instrumentation limits diagnosis. No new benchmark or live deployment/configuration verification was performed here.

PR209's recent native phone/iPad checks passed, as established in the preceding task. That scoped result does not certify the whole mobile platform. Likewise, the prior notice-navigation test's foreground-app anomaly is unresolved evidence, not a confirmed new navigation defect.

## Deliberate limits — do not misclassify as defects

- Offline writes, queues and sync: explicitly excluded initially.
- Cross-inventory movement and multi-inventory action plans: deferred.
- Full calendar-grid expiration UI: deferred; list/workspace views are intended.
- Typo-tolerant semantic search: the current first-slice fuzzy contract explicitly uses all-term substring matching. Advanced ranking is not required to satisfy that contract.
- Non-asset undo, attachment/hard-delete reversal and whole-asset time travel: outside the current reversal slice.
- Destructive field schema changes, changing field types, option removal/reordering and conversational schema creation execution: require further contract work; do not bypass current validation.
- Native Stuff Stash JSON/CSV import adapters and media backup packaging: future definitions; distinct from required JSON/CSV export.
- General arbitrary-host browser→mobile invitation handoff: explicitly blocked until server binding is safe. Existing same-server/app acceptance is implemented.
- Model inspection of photos: preparation exists; provider consumption is future scope.
- Android store distribution: outside the first iOS/TestFlight distribution slice, despite Android being a client target.
- Server-side Browse kind filtering, complete-tree Map loading and direct event lookup: documented API/performance limitations, not independently established correctness failures.

## Priority recommendation

1. Confirm which missing capabilities are current commitments: export, local/OpenAI-compatible execution, web conversation, MCP. Export has the clearest unconditional portability requirement.
2. Finish the already authorized normal-text everyday-workflow acceptance and reconcile its evidence. Do not use unrelated feature gaps to block bounded release batches.
3. Define and add richer conversational edits; connect image telemetry before further performance claims.
4. Establish localization infrastructure and migrate remaining application packages in coherent increments.
5. Repair stale status documentation once, preserving durable evidence without repeating historical checkpoint churn.

No broad security failure was established by this review. Existing tenant scoping, authorization adapters, rate limiting, lifecycle/audit machinery and adversarial tests are substantial. Their existence is not a security certification; new export/MCP/provider work still needs its own boundary tests.

## Specification coverage inventory

Every spec below was included in the scope inventory. Family-level implementation mappings and exceptions are summarized here; entries without a confirmed gap are not assertions of exhaustive compliance.


### agent-model

- [conversation-workflows.spec.md](../../specs/agent-model/conversation-workflows.spec.md)
- [conversation-workspace.spec.md](../../specs/agent-model/conversation-workspace.spec.md)
- [conversational-action-plan.spec.md](../../specs/agent-model/conversational-action-plan.spec.md)
- [conversational-inventory.spec.md](../../specs/agent-model/conversational-inventory.spec.md)
- [mcp-agent-tools.spec.md](../../specs/agent-model/mcp-agent-tools.spec.md)
- [mobile-conversation-interface.spec.md](../../specs/agent-model/mobile-conversation-interface.spec.md)
- [mobile-conversation-lifecycle.spec.md](../../specs/agent-model/mobile-conversation-lifecycle.spec.md)
- [mobile-realtime-voice-query.spec.md](../../specs/agent-model/mobile-realtime-voice-query.spec.md)
- [model-led-voice-loop.spec.md](../../specs/agent-model/model-led-voice-loop.spec.md)
- [provider-profiles.spec.md](../../specs/agent-model/provider-profiles.spec.md)
- [realtime-interaction.spec.md](../../specs/agent-model/realtime-interaction.spec.md)
- [voice-conversation-quality.spec.md](../../specs/agent-model/voice-conversation-quality.spec.md)

### assets

- [asset-checkout.spec.md](../../specs/assets/asset-checkout.spec.md)
- [asset-model.spec.md](../../specs/assets/asset-model.spec.md)
- [asset-tags.spec.md](../../specs/assets/asset-tags.spec.md)
- [containment-model.spec.md](../../specs/assets/containment-model.spec.md)
- [custom-asset-types.spec.md](../../specs/assets/custom-asset-types.spec.md)
- [flexible-asset-fields.spec.md](../../specs/assets/flexible-asset-fields.spec.md)

### audit-history

- [audit-and-undo.spec.md](../../specs/audit-history/audit-and-undo.spec.md)

### data-portability

- [import-export.spec.md](../../specs/data-portability/import-export.spec.md)

### expiration

- [expiration-tracking.spec.md](../../specs/expiration/expiration-tracking.spec.md)
- [expiration-workspace.spec.md](../../specs/expiration/expiration-workspace.spec.md)
- [live-acceptance.spec.md](../../specs/expiration/live-acceptance.spec.md)

### identity-access

- [authentication-flow.spec.md](../../specs/identity-access/authentication-flow.spec.md)
- [mobile-oidc-authentication.spec.md](../../specs/identity-access/mobile-oidc-authentication.spec.md)
- [spicedb-model.spec.md](../../specs/identity-access/spicedb-model.spec.md)
- [spicedb-schema.spec.md](../../specs/identity-access/spicedb-schema.spec.md)
- [tenant-inventory-access.spec.md](../../specs/identity-access/tenant-inventory-access.spec.md)

### inventories

- [inventory-model.spec.md](../../specs/inventories/inventory-model.spec.md)

### locations

- [location-model.spec.md](../../specs/locations/location-model.spec.md)

### media

- [background-thumbnails.spec.md](../../specs/media/background-thumbnails.spec.md)
- [image-performance-evidence.spec.md](../../specs/media/image-performance-evidence.spec.md)
- [media-attachments.spec.md](../../specs/media/media-attachments.spec.md)

### notifications

- [expiration-notifications.spec.md](../../specs/notifications/expiration-notifications.spec.md)

### platform

- [api-contract.spec.md](../../specs/platform/api-contract.spec.md)
- [bounded-contexts.spec.md](../../specs/platform/bounded-contexts.spec.md)
- [brand-guidelines.spec.md](../../specs/platform/brand-guidelines.spec.md)
- [client-settings-management.spec.md](../../specs/platform/client-settings-management.spec.md)
- [client-technology.spec.md](../../specs/platform/client-technology.spec.md)
- [client-telemetry.spec.md](../../specs/platform/client-telemetry.spec.md)
- [github-pages-docs.spec.md](../../specs/platform/github-pages-docs.spec.md)
- [local-development-topology.spec.md](../../specs/platform/local-development-topology.spec.md)
- [local-scaffold.spec.md](../../specs/platform/local-scaffold.spec.md)
- [migrations.spec.md](../../specs/platform/migrations.spec.md)
- [mobile-add-location-selection.spec.md](../../specs/platform/mobile-add-location-selection.spec.md)
- [mobile-app-tracer-bullet.spec.md](../../specs/platform/mobile-app-tracer-bullet.spec.md)
- [mobile-comprehensive-ui-audit.spec.md](../../specs/platform/mobile-comprehensive-ui-audit.spec.md)
- [mobile-distribution.spec.md](../../specs/platform/mobile-distribution.spec.md)
- [mobile-move-destination-selection.spec.md](../../specs/platform/mobile-move-destination-selection.spec.md)
- [mobile-move-here-selection.spec.md](../../specs/platform/mobile-move-here-selection.spec.md)
- [mobile-navigation-and-visual-polish.spec.md](../../specs/platform/mobile-navigation-and-visual-polish.spec.md)
- [mobile-persistent-tabs.spec.md](../../specs/platform/mobile-persistent-tabs.spec.md)
- [mobile-selection-batch.spec.md](../../specs/platform/mobile-selection-batch.spec.md)
- [mobile-selection-lifecycle.spec.md](../../specs/platform/mobile-selection-lifecycle.spec.md)
- [mobile-server-state-audit.spec.md](../../specs/platform/mobile-server-state-audit.spec.md)
- [mobile-server-state.spec.md](../../specs/platform/mobile-server-state.spec.md)
- [mobile-settings-readback.spec.md](../../specs/platform/mobile-settings-readback.spec.md)
- [mobile-tag-selection.spec.md](../../specs/platform/mobile-tag-selection.spec.md)
- [monorepo-layout.spec.md](../../specs/platform/monorepo-layout.spec.md)
- [observability.spec.md](../../specs/platform/observability.spec.md)
- [offline-and-sync.spec.md](../../specs/platform/offline-and-sync.spec.md)
- [persistence.spec.md](../../specs/platform/persistence.spec.md)
- [platform-interaction-review.spec.md](../../specs/platform/platform-interaction-review.spec.md)
- [resource-lifecycle.spec.md](../../specs/platform/resource-lifecycle.spec.md)
- [rest-api-initial-slice.spec.md](../../specs/platform/rest-api-initial-slice.spec.md)
- [roadmap.spec.md](../../specs/platform/roadmap.spec.md)
- [secure-inventory-tracer-bullet.spec.md](../../specs/platform/secure-inventory-tracer-bullet.spec.md)
- [self-hosting-audit-skill.spec.md](../../specs/platform/self-hosting-audit-skill.spec.md)
- [self-hosting.spec.md](../../specs/platform/self-hosting.spec.md)
- [time.spec.md](../../specs/platform/time.spec.md)
- [tooling-versions.spec.md](../../specs/platform/tooling-versions.spec.md)
- [ui-design-workshop.spec.md](../../specs/platform/ui-design-workshop.spec.md)
- [web-frontend-tracer-bullet.spec.md](../../specs/platform/web-frontend-tracer-bullet.spec.md)
- [web-inventory-workspace.spec.md](../../specs/platform/web-inventory-workspace.spec.md)

### search

- [search-latency.spec.md](../../specs/search/search-latency.spec.md)
- [search.spec.md](../../specs/search/search.spec.md)
