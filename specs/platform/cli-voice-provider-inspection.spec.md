# CLI voice provider configuration inspection

## Scope

`stuffstash voice-provider show` reads the authenticated household's voice provider
configuration. It uses explicit household flags, a remembered context, or the
authorized household picker, without requiring or prompting for an inventory.
Noninteractive requests without household scope fail before the resource request.
This is administration inspection only: it does not start chat, use a microphone,
change provider selection, or expose credential values.

## Contract and boundaries

A CLI-owned read port returns configuration models. The HTTP adapter invokes the
pinned generated GET voice-provider-configuration SDK operation and maps all
response fields: household, readiness, selected profile IDs, update timestamp,
slots, issues, selection source, recommended action, selected profile summaries,
and duplicate profile summaries. Summaries include credential purpose/status,
last test time, capability, name, model, provider kind, ID and lifecycle state.
Unknown response fields (including credentials) never enter output models.

JSON retains schema and response metadata, nullable configuration and nested
lists, empty lists, and optional fields. Human output presents household readiness,
configured profile IDs and labeled slot details, including issues and complete
selected/duplicate summaries. Empty or absent configurations receive clear text.
All server strings are safely escaped by the presentation adapter.

The command accepts only shared connection, context, display and authentication
options. Pagination and unrelated action flags fail before network access. Calls
use existing safe errors and the injected observer emits
`cli.voice_provider.read.completed` after successful inspection.

## Verification

Critical tests exercise the real CLI boundary with a controlled HTTP service:
route, authentication and correlation headers, every documented output field,
null versus empty nested data, safe human output, unknown credential exclusion,
401/403/404 errors, cross-household denial, saved household context and invalid
flags. A scope-selection test verifies no inventory is selected. Invalid command
shapes and missing scripted scope must fail without resource calls.
