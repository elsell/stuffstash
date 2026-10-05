# Explicit client telemetry submission

`telemetry submit --input FILE|-` sends an explicit batch of existing client
measurements through the authenticated REST telemetry API. It does not collect
CLI usage automatically or label CLI activity as a supported mobile/web platform.
The operator supplies one to fifty measurements with the API's platform,
operation, surface, variant, outcome and duration fields. Preserve the exact JSON
request, including optional schema and numeric values; reject unknown fields,
invalid enumerations and durations outside zero to 60000 milliseconds before
loading credentials. All required fields must be present, including zero duration.

This command is account-scoped, with no household/inventory selection. Show the
server and measurement count, and require confirmation or --yes for scripts.
Use normal human authentication, generated SDK behind a telemetry port, and the
existing safe error/JSON output path. Preserve the complete accepted-count
response envelope. Never retry automatically: a lost reply can mean the batch
was already counted. Emit an injected completion event without measurement data.
Do not introduce tracking configuration, background collection or dependencies.

Critical tests exercise real command authentication, exact request and complete
response, input rejection before credentials, refusal without confirmation,
401/403 and no retry, and safe error output. This is an explicit API parity tool,
not a change to application telemetry policy.
