# Local command help

## Behavior

`stuffstash --help` lists supported command groups and local entrypoints.
`stuffstash GROUP --help` lists that group's immediate actions or subgroups.
`stuffstash COMMAND --help` displays that implemented command's usage, purpose,
relevant options, required scope, input, output, confirmation behavior and example.
Nested groups such as `workflows revisions` and `connectors print` are supported.
Supplied resource arguments select the same command help without being echoed.
Unknown command paths produce a concise usage error, without inventing commands.

Help is plain text on stdout even with `--json`, redirected output or `--no-input`.
It runs before runtime configuration, context stores, credentials, authentication,
network calls, device discovery, input files, stdin and command execution. Invalid
environment configuration cannot prevent local help. Existing flag tokenization
is shared with execution, so option values containing `--help` and arguments after
`--` do not accidentally request help. A true help flag bypasses command required
arguments and execution-only validation; malformed option syntax still fails.

## Structure and content

A compact, reviewed command catalog owns help content, organized by domain. It
covers every implemented command without embedding OpenAPI or advertising pending
operations. Shared option descriptions and arity come from the execution parser's
registered flags. Command-specific metadata selects relevant options and records
scope, body requirements and confirmation semantics rather than guessing from
verb names. Examples use placeholders and only supported flags. Root and group
help are short navigation aids; command detail stays relevant to that command.

Help must distinguish local commands, public server discovery, account scope,
household scope, inventory scope and connector credentials. Scripted writes that
need JSON input or `--yes` state that explicitly. Render commands explain file
output and workers explain their long-running behavior. New commands must update
the catalog; a test cross-checks command coverage against the reviewed CLI API
coverage manifest without changing that manifest.

## Verification

Critical tests exercise representative read, write, nested and local help through
the bootstrap entrypoint with environment access forbidden. They verify relevant
options, scope, body and confirmation guidance; plain redirected output; concise
group navigation; no echo of resource arguments; and invalid command behavior.
Parser tests cover flag values, equals syntax and `--` boundaries. Catalog tests
check metadata completeness, known option names and implemented API coverage.
Existing CLI tests must continue to pass after sharing flag registration.

The catalog includes evaluation-run cancellation from its preceding stack change:
`evaluation runs cancel RUN_ID`, expectedVersion JSON input, and required scripted
confirmation. It does not add that operation or alter its behavior.
