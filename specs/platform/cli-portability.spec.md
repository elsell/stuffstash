# CLI archive and import workflows

## Commands and scope

Implement `archive-jobs list`, `show JOB_ID`, `preview JOB_ID`, `create`,
`upload`, `download JOB_ID`, `approve JOB_ID`, `retry JOB_ID`, and `delete JOB_ID`.
List accepts limit/cursor and defaults to the selected household without importing
its remembered inventory. An explicit inventory flag or environment override is
an optional filter on household job operations. Creation selects an inventory;
JSON inventoryId is authoritative and must agree with any explicit inventory.
Upload creates a household restore-review job. Approval names a new destination
inventory and does not replace an existing inventory. Preserve server decisions.

`import-jobs preview` and `start JOB_ID` use selected inventory scope and the
complete ImportSourceRequest, extending existing job inspection/cancellation.
All commands use generated SDK transport behind project-owned ports. Responses
retain every field, schema, metadata, exact integers, nulls and pagination.
Human output gives readable job state and restore counts/remappings.

## Input and mutation behavior

Create/approve/import preview/start accept exact `--input FILE|-` JSON. Reject
unknown fields, incorrect types and missing required values before authentication;
never echo input values in errors. Creation requires inventoryId, photos and
otherFiles (explicit false is meaningful), plus optional $schema. Approval requires
name plus optional $schema. Import source accepts sourceType, baseUrl, username,
password, includeImages, allowInsecureTLS, allowPrivateNetwork, fileName,
contentBase64 and $schema. Preserve omission/null and original request bytes.

Interactive creation asks whether to include photos/other files; approval asks
for the destination name. Guided imports choose live legacy Homebox or CSV,
collect live-source URL and masked credentials, or read a selected CSV file and
base64 encode it. Image inclusion and exceptional TLS/private-network permissions
are explicit choices; no permissive default. Scripts supply protected JSON input;
no credential argv flags. The CLI itself never contacts the import source.

Mutations show the effective server and scope, explain effects, and require a
keyboard confirmation or --yes. Preview creates server-side job state and also
requires confirmation. Create/upload accept or generate an idempotency key and
report it before the request; uncertain results require inspection and explicit
retry with the same key. No operation automatically retries or polls.

## Binary files

Archive upload uses --file PATH|- and streams application/zip without extraction
or local ZIP interpretation. Reject terminal stdin and nonregular paths safely.
Download requires --output PATH|- and uses the shared BinaryFiles publication port.
A file is private, created atomically without replacement; failed transfers do not
publish it. The caller closes server response streams. Stdout mode writes only
bytes, with notices on stderr. Reject --output - together with --json before
any request, matching media download commands. Never use a
server filename or content URL as a local destination or request target.

## Evidence

Critical CLI boundary tests cover exact JSON/false/null and full response fields,
scoped URLs, household-wide defaults, optional inventory filters, auth denials,
wrong scope, rejected confirmation, invalid input before auth, no automatic retry,
secret-safe diagnostics, and binary transfer routing/publication. Test guided
source choices through controlled ports, including protected credential entry.
Shared binary adapter tests own stream cancellation/privacy/no-overwrite details.
Run the full CLI suite and applicable hooks on Paul and obtain code-critic review.

Import JSON reads are bounded at 16 MiB to accommodate the API's 10 MiB decoded CSV
contract after base64 encoding; guided CSV reads stop at 10 MiB. Other JSON commands
retain the existing 1 MiB bound. These reads use the streaming input adapter and
never extract, interpret or execute CSV contents. The server remains authoritative
for source validation and deployment request-size constraints.
