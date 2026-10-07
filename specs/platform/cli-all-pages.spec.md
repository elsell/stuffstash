# CLI complete cursor traversal

`--all` follows the cursor of supported finite, read-only lists until the server
reports `hasMore: false`. It does not create a snapshot or retry a request.
`--limit` remains the per-request page size; an explicit `--cursor` is the start
position, so completion means all remaining pages from that position. Preserve
all other query filters, scope, credentials, and request options for every page.
The same invocation resolves credentials and interactive scope once. Traversal
itself never prompts, changes remembered scope, or invokes a mutation.

Support cursor-backed households, inventories, assets (list, search, expiration,
checked-out, checkouts, activity), household/inventory audit, attachments, tags,
access grants, invitations, notifications, archive jobs, custom asset types and
field definitions, workflows and revisions, evaluation cases/revisions/runs,
printers, print jobs, connectors, and consumer attempts. Reject `--all` on other
commands before authentication or network access, including asset audit (the
API has no cursor), notification counters/read-all, and non-paginated catalogs.
Explicit `--all=false` is still an option and must be rejected on unsupported
commands. Help and shell completion advertise it only for supported commands.

Use one typed application traversal helper behind existing domain ports. Append
list entries in server order without deduplication or numeric conversion. For
expiration workspaces, append `items` and retain final-page counts and timezone;
these summaries describe the last read, not an atomic snapshot of all entries.
A successful traversal emits one existing-shaped result with combined data and
final-page schema, metadata and pagination. It adds no response fields. Only
final-page request metadata remains in the combined response; help explains that
the result is not an atomic snapshot.
Human output uses the existing combined list/detail presentation and does not
show a misleading next-page hint after completion.

Accumulate results in memory (space proportional to returned data). Do not emit a partial successful list on request failure or
cancellation. Check cancellation before and after every request. Reject missing
pagination and any empty/missing/repeated next cursor when `hasMore` is true;
include the initial cursor in cycle detection. Errors keep their existing
categories; malformed pagination is a protocol error with actionable guidance.
No automatic retry or mutation is introduced.

Critical tests cover real HTTP inventory-list traversal with exact query/scope,
combined JSON and final-page metadata, malformed cursor failure without partial
stdout, cancellation without further requests, filter preservation, and nested
expiration items. Shared helper tests cover page-state boundaries. Existing
security tests continue to apply to each generated SDK request.

Pagination diagnostics use approved ordinary vocabulary and active sentences.
Describe incorrect page information without the unapproved adjective “usable”;
name an unavailable option without using “support” as an ordinary verb. Recovery
uses explicit single-page flags. This wording change does not change categories,
exit status, traversal, or output shape.
