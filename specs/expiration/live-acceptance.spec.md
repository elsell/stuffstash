# Expiration live acceptance evidence

## Scope

The expiration release requires realistic conversation acceptance in addition to deterministic application and WebSocket security tests. These runs use isolated memory-backed inventories, real Google language inference, and typed WebSocket input. They do not establish microphone, speech-recognition, physical-device or push-delivery acceptance. No approval is submitted. The fixture contains Medicine with expiration enabled, Hall closet / Bin 8 / Tylenol, and an exact expiration seven days after the test's UTC date.

## 2026-09-11 baseline and guidance revision

Harness: `TestGoogleLiveExpirationConversationCorpus`. Provider: `gemini-2.5-flash-lite`, `us-central1`. Primary-agent human review; no independent model judge. Raw logs on the remote validation host: `/tmp/expiration-live-corpus-final.log` (baseline) and `/tmp/expiration-live-corpus-v2.log` (guidance revision). Structured events: `/tmp/expiration-live-evidence` and `/tmp/expiration-live-evidence-v2`. These are local diagnostic artifacts, not permanent public URLs.

The initial attempt omitted the evidence-directory environment variable. Its failures are harness failures and not acceptance results. The harness now checks that variable before making provider calls.

| Scenario / request | Guidance-revision deterministic result | Human verdict and evidence |
| --- | --- | --- |
| Add aspirin to Bin 8, Medicine, February 29, 2028 | Fail | Fail: vocabulary lookup guessed display-name key `Medicine`, then could not propose the requested item. The baseline exhausted its tool budget. |
| Add aspirin to Bin 8, Medicine, February 2028 | Fail | Fail: session failed; baseline incorrectly claimed Medicine was unavailable despite the seeded type. |
| Add aspirin to Bin 8, Medicine, next month | Fail | Fail: session failed. Baseline did obtain the calendar and propose October 2026, showing that this is inconsistent rather than an unsupported date format. |
| Add aspirin expiring 03/04 | Fail | Fail: did not ask a date clarification. Baseline proposed March 2026, inventing both interpretation and year. Guidance revision incorrectly claimed the type was unavailable. |
| What medicine expires soon? | Fail | Fail: claimed no matching medicine despite the fixture. Baseline correctly found Tylenol and September 18, 2026. |
| Where is my Tylenol? | Pass | Fail: correctly identified Bin 8 / Hall closet but said September 2026, losing the recorded exact day. Baseline omitted expiration entirely. Added a stronger date-precision assertion after reviewing this trace. |
| Change Tylenol expiration to February 2028 | Pass | Pass for this isolated proposal: one update targeted the existing Tylenol with month precision; no write before approval. |
| Remove Tylenol expiration | Pass | Pass for this isolated proposal: one update targeted the existing Tylenol with explicit null expiration; no write before approval. |

The guidance revision reinforces manifest-first type lookup, clarification for incomplete absolute dates, verified calendar resolution for relative dates, and expiration context in location answers. It is not a proven fix. Target/type/parent and response-reference assertions were strengthened; semantic trace review remains required. The month-only location answer motivated an additional exact-date assertion.

## Required follow-up

- Make vocabulary discovery recoverable without allowing fabricated type IDs or treating malformed lookup as proof of absence.
- Ensure ambiguous absolute dates require clarification rather than guessed proposals.
- Preserve the full recorded date in both spoken and written expiration context.
- Re-run queries, creates and follow-up scenarios until the actual traces satisfy the request; a valid tool envelope is insufficient.
- Compare stronger model behavior before changing provider defaults. The completed `gemini-2.5-flash` comparison is under `/tmp/expiration-live-corpus-flash.log` with events in `/tmp/expiration-live-evidence-flash`. All three creates timed out at 60 seconds after authorized destination/manifest discovery. The other five cases passed deterministic checks and human trace review: ambiguity asked both year and March 4 versus April 3; query/location named Tylenol, Bin 8, Hall closet and September 18, 2026; correction/removal proposed only the targeted update. The exact-date assertion was added after this binary started, so that assertion was not executed in this run; its date requirement was verified by reading the responses. The comparison does not establish release readiness or justify a silent model switch.

The expiration feature is not ready for release based on these results.

## Vocabulary recovery run

After returning the scoped manifest for bounded mistyped definition keys, `gemini-2.5-flash-lite` completed the month-date create, targeted correction and removal. Log: `/tmp/expiration-live-corpus-vocab.log`; events: `/tmp/expiration-live-evidence-vocab`. Exact-date create, relative create and ambiguous-date cases still failed with session failure. The query said “about a month” instead of the actual date; the location response retained only September 2026. Both now fail the stronger date assertion. This is evidence that lookup recovery works, not that overall voice acceptance is complete. Full deterministic API and structural checks passed after this change.

## Discriminator ordering experiment

A Google-only `propertyOrdering` hint placed tool/command discriminators before arguments. The three Flash create cases still failed (`/tmp/expiration-live-corpus-order.log`, 11.00/11.79/21.14 seconds). Traces showed repeated unrelated command variants instead of the requested create. The hint was removed; it is not shipped as a fix. Next investigate nested command-alternative decoding while preserving strict application command validation. Google's structured-output guidance describes schema complexity limits and property ordering, but these traces—not the documentation alone—determine acceptance: https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/control-generated-output.

## Command-union projection experiment

A provider-only flattened command schema was also tested and removed. The initial projection did not recognize grouped archive/restore and checkout/return enums; a new production-catalog boundary regression exposed this before any success claim. After supporting grouped kinds, the real schema was projected, but Flash exact-day and month creates still timed out at 60 seconds. Relative create completed in 10.77 seconds; ambiguity, location, correction and removal completed, while the query failed its date assertion. Full run: `/tmp/expiration-live-corpus-flat-v2.log` (154.55 seconds). These deterministic passes have not been promoted to comprehensive semantic acceptance.

Retaining the ten-command array bound on the flattened schema produced `provider_http_status_400` before the first tool call (`/tmp/expiration-live-corpus-flat-bounded.log`). The earlier ineffective projection run is `/tmp/expiration-live-corpus-flat.log`; it is not evidence about flattened decoding. All experimental production changes and their projection-specific tests were removed before the native function-calling comparison below.

## Native function-calling comparison

Native required function calling with the original bounded union-array schema was rejected before tool execution with static diagnostic categories `schema complex states`. Reusing the existing Google schema translation for that bound allowed the native protocol to execute. Strict application validation still enforces the command limit.

With `gemini-2.5-flash`, all eight cases passed in 33.07 seconds (`/tmp/expiration-live-corpus-native-v2.log`; events `/tmp/expiration-live-evidence-native-v2`). Primary-agent trace review confirmed exact-day, month-only and calendar-resolved relative creates with the intended type and destination; the ambiguous date requested March 4 versus April 3 and the year. Query and location answers named Tylenol, Bin 8, Hall closet and September 18, 2026. Correction and removal proposed only the intended existing-item change. No writes were approved and typed requests produced no audio.

The same native protocol with `gemini-2.5-flash-lite` passed seven cases in 26.56 seconds, but proposed a change for the ambiguous date instead of asking for clarification (`/tmp/expiration-live-corpus-native-lite.log`; events `/tmp/expiration-live-evidence-native-lite`). This remains a release concern; no model default is changed by the protocol fix.

The full remote API suite and targeted diagnostic-redaction test passed, as did structural checks. Code critic review found no confirmed actionable issues. The native protocol replaces the JSON response envelope based on this evidence. These results cover the isolated typed corpus only; broader conversation, microphone, physical-device and push-delivery acceptance remain required.
