# CLI evaluation inspection

The CLI exposes household administrator evaluation inspection through six read operations:

- `evaluation cases list` and `evaluation cases show CASE_ID` (latest revision).
- `evaluation revisions list CASE_ID` and `evaluation revisions show CASE_ID REVISION_ID`.
- `evaluation runs list` and `evaluation runs show RUN_ID`.

These commands resolve only household scope from explicit options, saved contexts, or the interactive household picker. They never require an inventory. Authentication and server authorization remain mandatory. Missing scope in JSON or noninteractive mode fails without guessing. The commands use a project-owned evaluation port with generated REST requests behind the HTTP adapter.

All three list commands accept `--limit` and `--cursor`, the complete filter set offered by the API, and retain pagination metadata. Detail commands reject pagination options. Unsupported mutation or unrelated filter options fail before requests.

JSON output retains the complete declared API data and response envelope, including schema, metadata, optional field presence, explicit nulls, empty lists, nested fixture/expectation/observation/verdict/provider fields, and numeric precision. CLI-owned typed models prevent transport SDK types from leaking into application code. Human output presents case and run summaries, complete revision/run details, and pagination continuation. Terminal text is escaped by existing presentation helpers. No evaluation writes or conversation/chat commands are added.

Boundary tests cover all six routes, authenticated household scope, pagination, full nested payloads, null/empty distinction, numeric fidelity, readable output, safe errors, and rejection of invalid commands before resource requests. Saved scope and household-only picker behavior reuse the shared scope-selection contract.
