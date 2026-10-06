# Workflow authoring and activation through the CLI

`workflows create --input FILE|-`, `workflows revisions create WORKFLOW_ID
--input FILE|-`, and `workflows activate WORKFLOW_ID --input FILE|-` expose the
three workflow write operations. These are household administration commands;
they use existing saved household selection and picker behavior, without requiring
an inventory or changing context persistence policy.

Nested definitions and activation evidence require structured JSON. Input errors
and missing input are reported before credentials are loaded. File/stdin handling
uses the existing bounded input port. Preserve the original JSON bytes, including
schema, nulls and exact integers; validate known request fields and required
structure without remarshal. Do not invent scalar defaults or open an editor.
Creation requires definition; revision creation additionally requires a positive
expectedRevision. Activation requires revisionId, runId and cases, with optional
expected selection. The server remains authoritative for evaluation gates and
semantic domain validation. Unknown fields or wrong JSON types fail locally.

Show the server, household and target on stderr before mutation. Require explicit
confirmation, or --yes in scripts/JSON/no-input mode. Activation explicitly warns
that it changes the selected workflow. Preserve caller-supplied concurrency and
evaluation evidence; never fetch replacement versions, retry writes, or override
conflicts. An uncertain result tells users to inspect workflow state before retry.

Generated SDK calls remain behind the workflow port. Return the complete existing
revision envelopes and human formatting, including metadata and schema.
No generic raw-request command, dependency, credential output or chat execution is
introduced. The help catalog documents structured input, confirmation and examples.

Critical boundary tests cover complete payload forwarding, exact numbers/schema,
full response, authorization/wrong-household denial, declined confirmation,
conflict without retry, and invalid input before authentication. Scope selection
uses the existing household picker. Required checks and code critic precede commit.
