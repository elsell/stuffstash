# Full unsigned print-worker transport ranges

Existing print-worker HTTP transport must preserve the server contract's complete
uint64 attempt revisions and binding/connector generations, and uint32 media and
adapter contract versions. This is a transport correctness fix, not a new command,
worker reporting policy, recovery policy, or API-parity completion claim.

Retain generated SDK routes behind the existing ports. Decode attempt and printer
responses through shared project-owned numeric models, with worker time fields
still parsed as timestamps. Encode proof, outcome, reconciliation, idle-confirmation
and heartbeat bodies with their declared unsigned widths via generated WithBody
methods. Artifact request headers must carry the exact decimal uint64 revision;
a generated per-request editor may correct that header's signed SDK parameter.
No generated files, proof identities, lease decisions, journal behavior, hardware
submission, redirects, retry policy, or credential handling change.

Tests use controlled HTTP peers to verify boundary-range claim/start/renew/outcome,
attempt/reconciliation/unsettled/idle/artifact flows and printer/heartbeat transport.
Include maximum uint64 revisions/generations, maximum uint32 media/contract versions,
exact proof fields and decimal headers, rejection of out-of-range wire values, and
existing authorization/redirect/settlement and worker safety tests. Native integer
hardware dimensions remain governed by the existing worker/platform contract.
