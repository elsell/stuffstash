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
