# Full REST CLI Parity

## Status and approved scope

Approved October 5, 2026. This supersedes the limited command scope and
explicit-only context selection in `cli.spec.md`. Existing commands remain
compatible unless a security requirement prevents it. Implementation is not yet
complete; an SDK method or a command name alone does not establish parity.

Provide a discoverable named command for every REST operation. The initial
OpenAPI inventory at `a8183b048` has 192 operations on 145 paths. Track changes
against the current contract, not a frozen count. Include tenant and inventory
administration, sharing, customization, assets, search, media, audit/undo,
notifications, expiration, portability jobs, conversation administration,
provider administration, printing and connector REST operations. Chat, microphone
voice, WebSocket sessions and MCP protocol interaction are excluded. Public
health and mobile-auth discovery routes outside OpenAPI must also be inventoried;
the HTML API index and schema documents are documentation, not resource commands.

Use the generated SDK behind adapters. Do not replace named commands with an
arbitrary method/path request command and call that parity. A low-level escape
hatch does not close an operation's acceptance. Connector-only endpoints retain
the connector credential boundary; human login never supplies their authority.

## Command organization and coverage

- Preserve existing `assets`, `inventories`, `labels`, `printers`, `print-jobs`,
  and `connectors` command vocabulary. Group new commands by product domain.
  Household is user-facing language; `tenants` remains the stable API-facing
  command name and may have a `households` alias.
- List/show/create/update/delete and explicit lifecycle verbs have consistent
  meanings. Child-resource commands make their parent IDs and scope apparent.
- `--help` is local, fast and free of network/authentication side effects. Each
  command documents input, scope, examples, output, pagination and confirmation.
  Provide shell completion without loading secrets or issuing mutations.
- Maintain a reviewed operation-to-command manifest. Generated SDK dispatch is
  separate from reviewed command metadata and human presentation. The manifest
  records path/query/header parameters, body/media type, response handling,
  security identity, confirmation and workflow dependencies. Validate missing,
  duplicate and stale mappings against OpenAPI and explicit non-OpenAPI routes.
- Every supported request field must be expressible. Common scalar values have
  named flags; nested input supports `--body-file PATH` and `--body-file -`.
  Omission, explicit null, false, zero, empty list and empty string remain distinct.
  Do not silently drop fields, truncate lists, or convert patch into replacement.
- Binary upload/download, multipart bodies, pagination, asynchronous jobs,
  conditional revisions, idempotency and signed upload targets need actual
  workflow handling. JSON-only dispatch is not full parity.
- Generated dispatch must be split by domain and use generated SDK types and
  methods. Do not embed the full OpenAPI document into release binaries or use
  runtime reflection merely to avoid maintaining reviewed command definitions.

## Automatically remembered context

Use named per-server contexts comparable to kubeconfig. Context contains server,
selected tenant and inventory IDs plus non-authoritative display labels. It is
not a credential store or a permission cache.

Resolution is explicit flags, then corresponding environment variables, then the
selected saved context. An explicit server override can use saved scope only for
that exact canonical server. An explicit tenant change clears an inherited
inventory from another tenant. Never combine IDs across server or tenant scope.

With a usable terminal, a missing required tenant or inventory opens a searchable
arrow-key picker of authorized choices. Enter accepts, Escape/Ctrl-C cancels;
selection by typing a numeric row is not required. Fetch further pages so choices
are not limited to the first result page. Search and labels must not leak hidden
resources. Empty lists give the next permitted command; access denial does not
silently fall back to a different context. Show names and enough ID/context to
resolve duplicate names.

Accepted picker selections are remembered automatically. Explicit one-command
flags and environment overrides do not silently replace saved defaults.
`context list`, `context current`, `context use`, and `context delete` provide
inspection, selection and removal. Switching requires an existing context name.
Deleting the current context clears the current selection; it must not select
another context. Logout removes saved scope for every context on that server,
but keeps context names and server addresses. Remembering a picker selection
requires a verified principal and cannot overwrite a context bound to a different
server or account. A server's remembered context must not leak
one signed-in principal's resource names into another principal's picker.
Store configuration at the OS user configuration directory under
`stuffstash/contexts.json`; `STUFF_STASH_CLI_CONFIG_FILE` may override that path.
Use a stable verified issuer/subject identity for account binding, not the raw
token or a token hash that changes on refresh. Existing sessions without a stored
verified subject must refresh or sign in before saved account scope is reused.
Logout clears or invalidates principal-bound scope; unauthorized saved IDs require
explicit reselection rather than silent fallback. Write versioned owner-only
configuration atomically; reject unsafe files and avoid lost concurrent updates.
Validate updates before replacing the saved file. Cancellation before the atomic
rename must preserve the previous file; rename is the commit point. Lock waits
must honor cancellation across separate CLI processes.

Non-interactive commands may use saved context. If still incomplete, fail with
an actionable message naming the missing flags/context command. Never select the
first resource merely because it exists. Show resolved scope on stderr before
interactive mutations; machine output is unaffected.

## Terminal experience and scripting

Use a command-first interface, not a mandatory full-screen application. Human
output has concise tables, aligned detail labels, clear empty states, bounded
wrapping and terminal-width adaptation. IDs remain available without dominating
every human view. Do not print arbitrary terminal control sequences from API
content; sanitize human text while preserving data in machine output.

Guided prompts fill missing required values only on a terminal. Searchable
pickers, sensible defaults and inline validation minimize typing. Complex drafts
may use an explicitly invoked editor through a temporary private file. Execute
editors without shell interpolation of user data; remove temporary secret drafts.
Do not launch an editor, browser, pager or prompt unexpectedly in a script.

Interactivity requires terminal stdin and stdout and is disabled by `--no-input`.
JSON output also disables prompts. Piped body input and redirected output never
hang awaiting a selection. Existing explicit login/browser commands retain their
documented behavior; device-code login supports headless authentication.

- `--json` produces stable data on stdout, without spinners, color or diagnostics.
  Error objects go to stderr with stable code/category and actionable message.
  Preserve existing JSON contracts; new fields may be additive.
- Finite list commands retain explicit limit/cursor semantics. `--all` follows
  pages with cancellation and does not imply an atomic snapshot. Machine output
  must not imply completeness when more pages remain.
- Progress and notices go to stderr; they disappear cleanly on cancellation.
  Never report queued work as completed. Job commands return the ID immediately;
  an explicit bounded wait retains that ID on timeout without repeating creation.
- Honor `NO_COLOR`, terminal capability and `--color=auto|always|never`. Color
  reinforces meaning; labels/symbols carry it without color. Respect reduced
  terminal capability and provide plain text fallback. No dependency on emoji.
- Keep stable exit codes: 0 success, 2 usage/configuration, 1 operation failure;
  cancellation uses 130. Stable error codes distinguish authentication, denial,
  not-found, validation, conflict, timeout and unknown mutation outcome.

## Mutations, secrets and errors

Destructive operations require a clear scoped confirmation on a terminal and
`--yes` otherwise. `--yes` bypasses confirmation only, not validation, permission
checks or missing required input. Read actions and non-destructive operations do
not acquire unnecessary confirmation prompts. Unknown mutation outcomes never
trigger a fresh idempotency key or a guessed retry. Preserve concurrency headers,
existing server-side authorization and approved-plan boundaries.

Provider secrets and connector credentials must not appear in ordinary output,
logs, command history guidance or context files. Use protected file/stdin inputs
and masked prompts for secret values. Credential exchange commands explicitly
store or hand off the one-time secret through the appropriate protected adapter.
Do not print raw provider errors or credential-bearing response bodies.

All authored CLI error messages follow ASD-STE100 Issue 9 writing rules and
approved meanings, with a reviewed project technical vocabulary. Explain the
failure, then a concrete next action when known. Do not invent a cause from an
HTTP status. Map server errors through safe categories and allowlisted details;
retain field/resource context without exposing internals. Commands, IDs and
user-authored data are literal values, not text to rewrite. A vocabulary check
supports review but cannot prove full linguistic conformance by itself.

## Architecture and binary budget

Reuse the existing Go CLI and generated SDK. Command parsing, terminal rendering,
context storage, credentials and HTTP remain adapters to application-owned ports.
Split files by domain/responsibility. Avoid catch-all command runners and copies
of the same input/output/error logic in every domain.

Prefer standard library and already-reviewed dependencies. A small terminal
library is allowed only with documented need, pinned version, license and
transitive/security review. Compare stripped release binary size against the
same baseline toolchain/target/flags, separately for Linux USB and other targets.
Do not use executable packers. Avoid a second CLI framework or large TUI stack
when narrow terminal primitives suffice. Record measured growth with each
terminal/dispatch dependency decision; do not invent an arbitrary size cap.

## Critical verification and delivery

Write only tests that protect meaningful risks: command/request mapping including
null/false and binary data, forgotten scope and cross-server context, no prompts
in pipelines, destructive confirmation, idempotency/unknown outcomes, safe errors
and stdout/stderr contracts, terminal restoration, and API coverage drift.
Exercise shared mechanisms once, then representative consumers and exceptional
protocols; do not create 192 copies of the same dispatch test. Security-sensitive
boundaries require adversarial end-to-end authorization/isolation checks and
legitimate-success cases. Use fakes, not mocks. Preserve required CI and critic
review; don't turn unrelated historical acceptance into a release gate.

Frozen batches: (1) specification/coverage inventory and context/error foundation;
(2) terminal interaction and everyday inventory commands; (3) administration,
sharing and configuration; (4) media/portability and asynchronous workflows;
(5) remaining operational commands and parity/release acceptance. Each batch must
state exactly what remains. Full parity is complete only when every mapped
operation and input/output form works, contextual and script modes pass, helpful
error review passes, binary growth is measured, and delivery is verified.

Sources: [ASD-STE100 Issue 9](https://www.asd-ste100.org/assets/files/ASD-STE100_ISSUE9.pdf).

Verified identity binding: sessions retain the verified OIDC subject. Reject ID
tokens with an empty subject. A refresh must retain the same issuer and subject;
a changed identity requires a new sign-in and must not replace stored credentials.
Legacy sessions without a subject can acquire it through a verified refresh.

Explicit empty scope flags are usage errors, including an empty shell variable;
they must never fall back to environment or saved scope. Deleting a local context
only removes preferences and does not require destructive-resource confirmation.
It never deletes credentials, inventories or server data.

Windows context storage must use a private DACL for the current user (and the
trusted Windows SYSTEM account, when present), not Unix permission bits.
Create the dedicated context directory with private inherited access; reject
existing directories or files with broader access instead of changing their ACLs.
Validate ownership and permissions on opened handles. Use cancellable OS file
locks and atomic replacement. Reuse the pinned x/sys package for these OS calls.

Native Windows CI verifies private context ACLs and separate-process lock
cancellation. Cross-compilation and Wine are supplementary evidence only.

Windows file creation explicitly assigns the current user as owner and installs
the private DACL atomically, including under elevated accounts. Creation stays
relative to the opened directory handle and rejects reparse points.

Scope picker implementation uses `golang.org/x/term` v0.38.0 for terminal
detection, raw mode and restoration. It reuses the existing pinned x/sys
dependency; no full-screen TUI framework is required. Prompts require terminal
stdin, stdout and stderr and are disabled by JSON output or --no-input. Catalog
queries use the generated SDK, include all authorized pages, and reject missing
or repeated continuation cursors. IDs disambiguate duplicate display names.

Only a completed scope picker flow persists the complete selected scope; explicit
flags or environment overrides alone never replace saved defaults. Reuse an
existing matching account/server context. If none exists, use the endpoint as
its default name; append a stable account-key suffix only to avoid a name conflict.

Picker rows reserve space for identifiers independently of display names. On
Windows, enable virtual-terminal output for the prompt and restore the previous
console mode when it ends.

`apps/cli/api-coverage.json` records each OpenAPI operation, its contract
fingerprint, named commands, implementation status and remaining gaps. The
fingerprint includes referenced schemas so nested field changes cannot pass
unnoticed. `scripts/check-cli-api-coverage.py` checks drift; --update records a
new baseline without claiming implementation, and --require-complete rejects
pending or partial entries. This inventory is traceability, not runtime proof.

### Household discovery command

`tenants list` lists the signed-in account's households without requiring or
selecting a tenant or inventory. It supports `--limit` and `--cursor`, preserves
access relationship and permissions in JSON, and displays ID, name, and lifecycle
in terminal output. This command uses the same generated-SDK catalog adapter as
the scope picker. An authorization denial must not expose server error details.

### Account and selected-scope details

`account show` returns the signed-in principal without requiring resource scope.
`tenants show` returns the selected household and requires only tenant scope.
`inventories show` returns the selected inventory and requires both scopes.
Each accepts explicit scope options or saved scope; interactive selection asks
only for the levels the command needs. JSON includes every resource field,
including access permissions and inventory tenant ID. These read commands use a
directory port implemented by the generated SDK adapter. Account-wide reads run
before saved resource-scope resolution so unrelated context ambiguity cannot
prevent account discovery.

### Household and inventory writes

`tenants create/update` and `inventories create/update` accept either `--name`
or `--input FILE` (`--input -` reads stdin), never both. Create requires a name;
update preserves an omitted name and accepts an empty JSON object. JSON input
must be one object, at most 1 MiB, with no trailing value. The source adapter
accepts regular files only; it rejects terminal stdin instead of waiting for
manual JSON entry. Cancellation releases an idle stdin read. On Unix, open file
paths without blocking before checking the handle, and make inherited stdin
pollable through a private duplicate. JSON bytes reach generated SDK body methods unchanged, so
future nullable fields retain their wire meaning. The API remains responsible
for schema and domain validation. Input errors occur before sign-in or writes.
The CLI must reject --input on commands that do not support a request body.

Household creation needs no existing resource scope. Inventory creation needs
only tenant scope. Updates use the selected tenant/inventory. Creating a resource
does not silently change the current context. Before a scoped write, print the
effective server and resource IDs to stderr. These four REST operations do not declare idempotency support. Reject
--idempotency-key instead of suggesting it can prevent duplicate creation. Do
not retry writes automatically. If a create loses its response, instruct the
user to list resources before retrying; the result can be unknown. These creates and name updates are not destructive operations and
do not require confirmation. Guided missing-name prompts remain part of the full
interactive delivery; scripts must provide --name or --input.

### Request correlation and response metadata

Finite SDK-backed user commands accept `--request-id` for the API's
`X-Request-ID` header. Reject control and non-ASCII characters before any request.
The option does not change authentication and is not sent to the OIDC provider.
Directory and inventory-list JSON responses retain the API `meta` object,
including request ID, tenant ID, and pagination, plus optional `$schema`. Preserve
null collections versus empty arrays and empty strings versus
null cursors. Keep the existing top-level `pagination` field for compatibility;
new metadata is additive. Header correlation is diagnostic, not an idempotency
mechanism. Commands must never promise duplicate-write protection from it.

### Guided directory names

When a directory create/update command omits both --name and --input, a capable
interactive terminal asks for the household or inventory name. Use the existing
pinned terminal library for familiar cursor movement, deletion, and Unicode
entry. Ctrl-C cancels without a request. Trim surrounding whitespace, reject an
empty name or a name over the API's 120-character limit, and let the user correct
it in the same prompt. JSON, --no-input, redirected streams, and unsupported
terminals fail with actionable explicit-input guidance instead of prompting.
The prompt completes before authentication or scope selection; no mutation occurs
until input and target scope are ready. Prompt diagnostics use stderr only.

### Directory lifecycle commands

`tenants archive/restore/delete` operates on the selected household;
`inventories archive/restore/delete` operates on the selected inventory. They
send no JSON body and do not support idempotency keys. Print the exact server,
household ID, and applicable inventory ID before any lifecycle mutation.

Archive and delete require confirmation. Interactive confirmation is a keyboard
choice with Cancel selected by default and an explicit action choice. Delete
warns that it permanently removes the selected resource. `--yes` skips this
prompt; JSON, --no-input, or redirected streams require --yes and never wait.
Declining or canceling performs no lifecycle request. Restore needs no
confirmation. Authorization denial leaves local contexts unchanged.

A confirmed HTTP 204 delete returns a CLI result with status `deleted` and the
selected resource IDs. Never decode 204 as JSON or claim failure because it has
no body. Other successful lifecycle responses retain full resource metadata.
Do not automatically retry uncertain mutations. After a confirmed delete, clear matching saved resource scope for the same
server and verified account. Household deletion clears tenant and inventory;
inventory deletion clears only the matching inventory. Preserve context names,
other accounts, and other servers. If local cleanup fails after server success,
return the successful delete result with an actionable stderr warning; never
report the server delete as failed or retry it.

### Tag commands

`tags list` lists inventory tags with complete metadata and all tag fields. It
accepts limit 0 for the API default, positive limits, and cursors. `tags create`
accepts --name (display name), optional --key, and --tag-color; `tags update ID`
accepts --name and --tag-color. The color flag is separate from terminal --color.
An explicit empty --tag-color sends an empty string, not an omitted field. The
stable key cannot be updated. Either write also accepts --input FILE or stdin,
without mixing JSON input and field flags. A missing required name is prompted
only in interactive mode, with the API's 80-character limit. Color-only updates
do not ask for a name. JSON input retains omitted/null/empty distinctions.

`tags delete ID` requires cancel-default confirmation or --yes and returns the
API tag response, including its resulting lifecycle state. All commands require
household and inventory scope. All mutations display their scope and target,
use generated SDK methods, preserve API metadata, reject unsupported idempotency
keys, and never retry automatically. Diagnostics do not leak response bodies.

Command dispatch must fail closed: an unhandled command family must never fall
through to another domain's execution path. The legacy asset executor accepts
only asset commands and its explicitly supported inventory-list route. This
check protects new command integrations from accidentally mutating assets.

### Complete asset reads

`assets list` accepts --lifecycle active|archived|all and --sort id_asc|updated_desc
in addition to pagination. Omitted filters retain API defaults. Reject unknown
values and use of these options outside asset lists. Asset list and detail JSON
preserve every asset field: scope, description, timestamps, nullable expiration,
expiration context, tags, custom fields/type, checkout state/principal, primary
photo/thumbnails, parent, print job, and undoable operation. Preserve absent
optional strings versus present empty strings and null versus empty tags.
Arbitrary custom-field numbers must retain their JSON precision rather than
passing through float64. JSON includes response schema and metadata.

Human asset details show the title, identity/scope, kind, lifecycle, description,
location parent, expiration and checkout state, assigned tags, and custom fields.
List output remains compact. Generated SDK models remain confined to the HTTP
adapter; project-owned response models cross the port.

The HTTP adapter must bypass the generated nullable list decoder for asset
lists because it resets JSON number handling. Embed the generated envelope and
decode its data directly into generated asset models with `UseNumber`, preserving
null versus empty lists without changing the SDK.

### Complete asset write input

`assets create` and `assets update ID` accept `--input FILE|-` for the complete
API request object, preserving null, empty arrays, omitted fields, and exact
custom-field numbers. The generated SDK's body methods send the validated object
without a lossy decode/encode cycle. Do not combine input with asset field flags
or `--print-label`. The existing create-and-print shortcut remains available.
Interactive flag-based creation asks for a missing title (160 characters) and
uses a keyboard picker for kind (item, container, location). Update asks for a
missing title only when no JSON input is supplied. Scripts must supply the
required fields or JSON input; they never prompt. Keep scope selection and
server authorization unchanged. Asset updates do not declare idempotency support;
reject explicit retry keys for updates and moves before authentication instead
of claiming safe replay.

### Asset creation recovery

The create endpoint honors an idempotency key only when `printLabel` is present.
The CLI rejects an explicit key for ordinary creation and does not generate one.
For create-and-print, including JSON input, generate a missing key and display it
before sending the request. On an uncertain response, direct the user to retry
with the same key and unchanged request. For ordinary creates, direct the user
to list assets before retrying. Never retry automatically. Preserve actionable
authorization and validation failures instead of replacing them with uncertainty.

### Asset lifecycle commands

`assets delete ID` uses the generated DELETE operation and reports the deleted
asset, household, and inventory IDs after a 204 response. Preserve the selected
inventory. Archive and delete show the effective server/scope/asset and require
the shared cancel-default confirmation or `--yes`. Restore needs no destructive
confirmation. Reject retry keys for all three lifecycle operations because these
contracts do not support them. Preserve full archive/restore response fields.

### Checkout workflows

Add `assets checkout ID`, `assets return ID`, `assets checkouts ID`, and
`assets return-details ASSET_ID CHECKOUT_ID`. The three writes accept optional
`--details TEXT` or `--input FILE|-`, preserving explicit empty details. Updating
return details requires either option so an omitted value cannot silently clear
notes. Checkout and return can omit details and send an empty object. History
supports limit/cursor pagination. All commands use household/inventory scope,
complete checkout models and metadata, and generated SDK operations. Writes show
effective scope, reject unsupported retry keys, and never retry automatically.
Uncertain write results instruct users to inspect checkout history. Preserve
server authorization and validation failures. Human output presents checkout
state, borrower, dates, details, and IDs; JSON preserves the complete response.

Checkout writes reject unrelated name/title/kind/parent flags before login.
Human checkout history includes checkout and return notes and return dates so
users can inspect the outcome of an uncertain write without switching formats.

### Inventory-wide checked-out assets

`assets checked-out` lists the inventory's currently checked-out assets with
limit/cursor pagination. Preserve each complete asset and current checkout,
response metadata, schema reference, null/empty list distinctions, and exact
custom-field numbers. Use the same number-preserving envelope override as asset
lists. Human rows identify the asset and borrower with the checkout date.

### Expiration browsing

`assets expiration` exposes the complete expiration workspace: `--mode`
(all/soon/expired), `--kind`, `--checkout-state`, `--query`, `--type-id`, repeated
`--tag-id`, `--location-id`, `--from-date`, `--through-date`, limit and cursor.
Omitted filters retain server defaults. Validate enums, ISO date values, date
order, and a page size of 1–100 before requests. Reject expiration-only filters
on unrelated commands. Keep query text and repeated tag IDs intact through the
SDK. Preserve complete items, ancestor paths, counts, timezone, metadata, and
exact custom-field numbers. Human output shows counts, timezone, expiration
state/date, item titles, paths, and pagination. JSON retains the API shape.

### Asset search transport

The search adapter supports every GET /tenants/{tenantId}/search/assets filter:
optional inventory scope, query, fuzzy/exact mode, repeated tag IDs, custom type,
lifecycle, checkout state, limit, and cursor. It preserves full asset summaries,
inventory names, match explanations, ancestor paths, metadata, exact custom-field
numbers, and null/empty distinctions. Generated SDK methods remain the transport
boundary. Search default scope and saved-default behavior after a one-time scope
override remain pending user decisions; transport support does not imply a
complete CLI search command.

### Custom asset type transport

Support all seven custom asset type operations at both household and inventory
scope: list, show, create, update, archive, restore, and delete. The transport
requires an explicit typed scope level and IDs; an unknown level or missing ID
must fail before any request, never fall back to household scope. List supports
lifecycle and pagination. Writes preserve the complete supplied JSON object;
results retain every type field, optional inventory ID, metadata and schema.
Delete succeeds only on the documented 204 response. Use generated SDK routes
for every operation. CLI scope syntax remains a pending user decision; these
transport operations alone do not count as completed CLI workflows.

### Custom field definition transport

Support list, show, create, update, archive, restore and delete for custom field
definitions at both household and inventory scope. Reuse explicit definition
scope validation; never infer or downgrade scope. List includes lifecycle and
pagination. Keep complete JSON bodies, including empty option/target arrays,
and retain every response field, optional inventory ID, null/empty arrays,
metadata and schema. The server remains responsible for immutable keys/types
and append-only option/target policies. Delete requires 204. CLI workflows stay
incomplete until their scope syntax, inputs and confirmations are implemented.

### Attachment metadata and lifecycle

Provide inventory-scoped attachment list and detail reads, archive, restore and
permanent deletion through the generated SDK. Each operation requires an asset
ID; detail and lifecycle operations also require an attachment ID. Preserve all
attachment fields, including the 64-bit size, digest, lifecycle, timestamps and
scope IDs, plus response metadata and pagination. Delete requires HTTP 204.
Archive and delete use the shared destructive confirmation policy. These
metadata operations do not imply upload or download completion. Keep command
coverage partial until dispatch, human output and confirmation are verified.

Attachment command syntax is `attachments list ASSET_ID`, `attachments show
ASSET_ID ATTACHMENT_ID`, and `attachments archive|restore|delete ASSET_ID
ATTACHMENT_ID`. List accepts limit/cursor. Human lists show ID, name, media type,
size and state; detail includes digest, creation time and ownership IDs. JSON
retains the full envelope. Mutation notices identify server, household,
inventory, asset and attachment. Archive/delete require confirmation; restore
runs directly. Unsupported retry keys and unrelated write fields fail before
network access. Successful deletion returns a scoped status result.

### Attachment transfer boundary

Content and thumbnail reads must use the generated SDK's raw HTTP response.
The server sends binary bytes, not a JSON envelope. Return a closable stream,
media type, content length and content disposition through a transfer port;
never infer a local destination from server headers in the transport adapter.
Only HTTP 200 is a successful full download. Denials and redirects close the
body and return a safe error; API credentials must not follow redirects.
Thumbnail variants are small, medium, large, or omitted for the server default.
Reject unknown variants before network access. File destination policy is a
pending user decision; transport coverage alone remains partial.

Upload transport supports the JSON attachment-create operation, direct-upload
initiation, and completion. Accept JSON readers so the adapter does not impose
the text-input limit on encoded file content or duplicate large bodies. Preserve
all direct-upload response fields: upload/attachment IDs, method, URL, headers,
form fields and expiry, plus envelope metadata. These instructions are data,
not permission for the authenticated API client to visit the storage URL.
Never retry a create or completion automatically; an uncertain result must be
resolved by the calling workflow. File upload guidance, storage transfer and
recovery commands remain required before these operations count as complete.

### Direct storage upload adapter

A separate storage-transfer port streams a known-length file to the issued
storage destination. Support multipart POST policies (all form fields before the
file) and raw PUT. Compute the request length without buffering the file. The
caller owns the file reader. Use HTTPS; explicit local-development configuration
may permit HTTP only to a literal loopback address or localhost. Reject userinfo,
fragments, unsupported methods, credential headers, and conflicting framing
headers before sending data. The storage client has no cookie jar, API request
editor or automatic redirect following. Do not retry requests. Do not include
signed URLs or storage response bodies in errors. A canceled transfer returns
cancellation; short input and non-success responses must never trigger completion.
The CLI workflow must verify success before calling the completion API.

### Upload file source

The file source opens a readable, non-empty regular file and returns an owned,
closable reader, basename, byte length and detected media type. Detect the type
from at most 512 bytes without consuming the upload stream. Support the API's
JPEG, PNG, WebP and PDF types. Reject other types with an actionable error.
Use the opened handle's metadata; never rely only on a pre-open path check.
Unix opens must be nonblocking before regular-file validation to prevent FIFO
replacement from hanging. Cancellation closes the owned handle. Do not buffer
the entire file or expose the full local path as the attachment filename.

`attachments complete-upload ASSET_ID UPLOAD_ID` calls the completion API for
an existing direct upload in the effective inventory scope. It does not resend
file bytes or automatically retry. Show the mutation scope, but do not print the
signed upload token in notices or errors. Return the complete attachment result.
For uncertain completion (network/protocol/unavailable/generic API failure),
direct the user to inspect `attachments list ASSET_ID` before another attempt.
Preserve authentication, authorization and validation errors. This recovery
command is independent of the default upload-method decision.

### Notification inbox

Provide `notifications list [--unread-only]`, `show ID`, `unread-count`, `read ID`,
`unread ID`, and `read-all` in the effective inventory. List uses pagination;
list, unread-count and read-all preserve the API cursor. List limit is 1–100 and
cursor is at most 128 characters. Keep all notification fields, nullable ancestor
trails, optional read timestamps, response metadata and pagination. Read actions
are reversible inbox-state changes and do not require destructive confirmation.
Show their effective scope on stderr. Read-all must preserve the API's complete
flag and pagination; do not claim the whole inbox was marked when complete is
false. Do not automatically retry mutations. Human output must expose unread
state, expiration, asset reference and location; quote untrusted text.

### Notification settings and devices

Support the complete preference and device contracts through generated SDK
adapters. Preference updates replace defaults, timezone and push-enabled state
and require the caller's revision; do not synthesize omitted booleans or silently
retry revision conflicts. Preserve raw JSON for initialization, replacement,
type override and device registration. Override removal and device removal must
send the explicit 64-bit revision query value. Preserve preference policy fields,
nullable override arrays, device state and revision, schema and metadata. Device
tokens are write-only and must not be copied into results or diagnostics. Keep
these operations partial until named commands, guided inputs and confirmations
are implemented.

Device lookup is `notification-devices show INSTALLATION_ID`; removal is
`notification-devices remove DEVICE_ID --revision N`. These IDs are different
and help must name them explicitly. Removal requires a positive revision and the
shared confirmation policy; scripts pass `--yes`. A terminal may prompt for a
missing revision, but must never fetch or substitute a newer one automatically.
A conflict must instruct the user to review the current registration. Human
output includes ID, installation ID, transport, revision and active state; JSON
retains the response envelope. Registration remains a separate unfinished flow.

`notification-preferences show` displays current defaults, timezone, push state,
revision and all type overrides. `notification-preferences initialize --timezone
ZONE` initializes preferences with an explicit timezone; a terminal prompts for
it when omitted. Also accept `--input FILE|-` for the complete JSON request,
mutually exclusive with --timezone. Never infer timezone from the workstation.
The server validates supported timezone identifiers. Noninteractive calls with
missing input fail without issuing the mutation. Both commands retain full
JSON envelopes; initialization shows effective mutation scope on stderr.

`notification-preferences update` replaces defaults, timezone and push state.
`notification-preferences override TYPE_ID` replaces one type policy. Scripts
must provide `--input FILE|-` containing every required field and revision.
Interactive calls without input load current preferences once, retain that
revision, and offer keyboard choices with the current boolean first. Timezone and
advance days offer Keep current or Change; Change opens text input. Override editing starts from
its current override, or defaults if no override exists. No implicit defaults
may reset omitted fields. A conflict never fetches a newer revision and retries.
`remove-override TYPE_ID --revision N` removes an override with confirmation;
terminals may prompt for a missing positive revision. JSON bodies and revision
flags are mutually exclusive. Show the target type and effective scope before
mutation. Validation errors and cancellation issue no mutation.

`notification-devices register --input FILE|-` accepts the full registration
body without putting a token in command arguments. A terminal without input
prompts for installation ID, APNs/FCM via keyboard selection, nonnegative
revision (0 for initial registration), and token through a non-echoing secret
input port. Never use the ordinary echoed text input for a token. Interactive
tokens are limited to 4095 input characters; JSON input supports larger values within
the shared input-file bound. Detect excess input before the terminal library can
truncate it, and fail without submitting a partial token. Cancellation restores terminal state and sends no
mutation. The result and errors contain no token. No automatic retry occurs;
uncertain results direct the user to lookup by installation ID before retrying.

### Undo and redo

`operations undo OPERATION_ID` and `operations redo OPERATION_ID` call the
inventory-scoped compensating-operation endpoints. These arguments are operation
IDs returned by prior mutations, not asset IDs. Display the effective server,
household, inventory and operation before shared confirmation; scripts require
`--yes`. The server determines whether the operation can be undone or redone.
Return the complete asset response, including the next undoable operation ID.
Do not automatically retry or accept unsupported idempotency keys. If the result
is uncertain, direct the user to inspect the affected asset before retrying.
