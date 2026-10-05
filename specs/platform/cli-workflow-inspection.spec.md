# CLI workflow inspection

## Scope and commands

The CLI exposes household administration reads for conversational workflows:

- `workflows list [--limit N] [--cursor CURSOR]` lists workflow heads.
- `workflows show WORKFLOW_ID` reads the latest workflow revision.
- `workflows revisions list WORKFLOW_ID [--limit N] [--cursor CURSOR]` lists revisions.
- `workflows revisions show WORKFLOW_ID REVISION_ID` reads one revision.
- `workflows selection show` reads the current workflow selection.

These commands use the authenticated household from explicit flags, saved context,
or the authorized household picker. They never require or prompt for an inventory.
They do not execute chat or mutate workflow configuration. Only list operations
accept pagination flags; unrelated command flags fail before any network request.

## Boundaries and output

Application commands depend on a workflow read port. Its HTTP adapter calls the
pinned generated SDK and maps response values into CLI-owned models. JSON retains
all documented fields, exact integer values, nullable data and list distinctions,
optional properties, schema links, and response metadata including pagination.
Workflow definitions include instructions, provider profile, and all budget values;
revision output includes author, timestamps, revision number and settings migration.
Human output uses labeled details and compact lists with pagination guidance. An
unset selection and empty lists receive clear messages. Server strings are escaped
by the presentation adapter. Errors use existing safe diagnostics, without bodies.
Completion observability uses the injected observer and a workflow read event.

## Validation

Critical CLI boundary tests exercise all five commands against a controlled HTTP
server, preserving routes, bearer auth, request correlation, pagination, JSON and
human output, nullable selection/list data, and exact large integer budgets. Access
denials and cross-household requests must fail without exposing response bodies.
Invalid shapes, unsupported flags and missing noninteractive household scope fail
before a resource request. Saved household contexts work without inventory scope.
