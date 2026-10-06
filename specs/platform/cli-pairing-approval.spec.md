# CLI pairing review and approval

Human administrators can review and approve a printer connector from the CLI.
The existing browser approval and machine registration flows remain supported.
Use named `connectors print pairings review PAIRING_ID` and
`connectors print pairings approve PAIRING_ID` commands with the selected
household and inventory. These commands use human authentication, never the
connector credential or a polling token. Missing scope uses the existing saved
context and keyboard picker behavior without changing persistence policy.

The user code comes from an interactive prompt or structured `--input FILE|-`.
Do not introduce a code argument that exposes it in process listings. Scope in
structured input must agree with the effective command scope. Review forwards
the exact pairing ID, code and scope through the generated SDK and returns the
complete public review envelope. It never exposes private device IDs, a poll
token, a signature, a private key or a connector credential.

Interactive approval first reviews the pairing. Show its name, public-key
fingerprint, rotation status and candidate names. Let the administrator choose
printer bindings with the existing keyboard picker. Preserve candidate and
printer IDs; names are labels, not identity. Show the full binding summary and
effective server/household/inventory before confirmation. Never choose bindings
solely by similar names or approve an undisplayed candidate. Page through
available printers using the existing authorized inventory list port. Offer only
non-retired printers whose adapter ID matches the candidate.

For scripts, accept explicit bindings and user code in structured input and
require `--yes`. Review still validates the intended pairing before approval;
confirm all supplied candidate IDs appear in that review. Preserve the original
request's schema and exact binding choices. Validate one to sixteen bindings,
reject duplicate candidates or printers, and reject unknown fields locally.
Review does not reserve or bind scope. Approval must send the same reviewed
scope. The server remains authoritative for authorization and binding eligibility.
Do not retry an approval automatically or replace supplied values after failure.
For an uncertain response, instruct the operator to inspect the connector state
before attempting approval again. Ordinary approval returns the complete non-secret connector
envelope, preserving unsigned generation values. Rotation approval instead returns
the complete public pairing-status envelope (pairing ID, state and expiry), not
a connector or credential.

Rotation is a separate action on an existing connector, not ordinary pairing
approval. Ordinary approval must reject a rotation pairing and direct the
operator to the rotation approval workflow. That workflow requires the existing
connector ID and exact generation, reviews the pairing fingerprint and warns
that approval replaces the connector credential. It must not silently read a
new generation after a conflict. Keep the stored machine-credential exchange and
activation-deadline safeguards intact.

Critical tests cover scope mismatch before network access; missing and invalid
input before credentials; human authentication and authorization denial;
review-to-binding consistency; declined approval; exact payload and unsigned
response fidelity; no retries; and exclusion of secrets from output. Tests must
use controlled HTTP endpoints and existing ports, not physical printers.

Rotation approval uses `connectors print rotations approve CONNECTOR_ID` with
JSON containing generation (positive uint64), pairingId and userCode; optional
$schema is preserved. Interactive use prompts for pairing ID, hidden user code,
and the exact generation; it does not fetch a substitute generation. Ordinary
review input contains userCode, tenantId and inventoryId; ordinary approval adds
bindings (candidateId/printerId pairs). Both accept optional $schema. The code
uses the existing hidden secret-input port, never a command-line value.
