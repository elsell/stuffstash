# CLI evaluation run cancellation

`evaluation runs cancel RUN_ID` requests cancellation in the selected household. It reuses household-only saved context and picker behavior; it does not require inventory selection. Cancellation is a state-changing administrator operation and requires confirmation.

Explicit input uses `--input FILE` or `--input -` with a JSON object containing a positive 64-bit integer `expectedVersion`. The optional API `$schema` field is accepted; unknown fields, missing/null/noninteger/out-of-range versions, and unsupported command flags fail before authentication or network access. Pagination, unrelated resource fields, and idempotency keys are not accepted.

Scripts and JSON mode require explicit input plus `--yes`; the CLI never silently fetches a newer version to replace their compare-and-swap value. Interactive callers without input can read the current run version, see the server, household, run ID, and version, and confirm or cancel. `--yes` without explicit version input is rejected, so an automatically fetched version always receives interactive review.

The generated HTTP SDK remains behind the evaluation port. The cancellation request includes exactly the selected expected version and optional schema, and preserves existing authentication, tenant scoping, and request correlation. The complete returned evaluation run and response envelope use the existing typed evaluation decoder and human/JSON renderers, including nested values, explicit nulls, and numeric precision.

Send no cancellation request before confirmation. Send at most one request per invocation. A conflict directs the user to inspect the current run before retrying; never refetch and resubmit automatically. Transport or protocol failures after submission report an unknown result and direct the caller to inspect the run. Do not expose private server response bodies.

Critical tests cover version parsing before authentication, explicit CAS without refetch, interactive current-version confirmation and cancellation, correct household routing, 401/403 and cross-household denial, conflicts without retries, and complete output. Shared scope and output machinery retains its existing tests. No API coverage counts or human documentation change in this isolated implementation.
