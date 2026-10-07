---
title: Administer Stuff Stash from the CLI
description: Manage inventory access, archives, imports, and household providers.
---

[Sign in and select your context](../cli/#work-with-assets) before using these
commands. Each operation checks your permissions. Commands that change access
or remove data ask for confirmation; scripts supply `--yes` where required.

- [Manage access and invitations](#inspect-inventory-access).
- [Back up or restore an inventory](#create-and-restore-archives), or [import records](#preview-and-start-imports).
- [Configure model providers](#set-up-a-provider) and [voice selections](#choose-household-voice-providers).
- [Review workflows and evaluations](#inspect-workflows-and-evaluations).

Use each command's `--help` for its input fields, scope, and examples. List
commands that support `--all` can [read every page](../cli/#read-every-page).

## Read audit history

```sh
stuffstash tenants audit
stuffstash inventories audit --limit 20
stuffstash assets activity ASSET_ID --view changes
stuffstash assets audit ASSET_ID
```

Household history uses the selected household; inventory and asset history also
use the selected inventory. Add `--json` for complete records and metadata.
Household and inventory history accept the returned `--cursor` for another page.
Asset history supports `--limit` but has no cursor in the current API.

Use `assets activity ASSET_ID --view all` to include technical events. Activity
shows changed values and available undo operation IDs. It does not undo changes.
Use `--cursor` with the returned cursor to read the next page.

## Inspect inventory access

```sh
stuffstash access-grants list
stuffstash access-grants show PRINCIPAL_ID editor
stuffstash access-grants remove PRINCIPAL_ID editor
```

Removal asks for confirmation. In scripts, review the target and add `--yes`.
It removes that relationship only; another grant can still provide access.

Use `stuffstash access-grants create` to choose a principal ID and access level.
For scripts, put `{"principalId":"USER_ID","relationship":"viewer"}` in a JSON
file and run `stuffstash access-grants create --input grant.json --yes`.
Use `editor` to allow changes. Check the displayed household and inventory
before you confirm.

## Manage invitations

```sh
stuffstash invitations create --email friend@example.test --role viewer --yes
stuffstash invitations list --status pending
stuffstash invitations show INVITATION_ID
stuffstash invitations cancel INVITATION_ID
stuffstash invitations delete INVITATION_ID
```

Creation shows a one-time invitation link. Save or share it before closing the
terminal; list and show cannot retrieve it later. Use `--input FILE|-` with
`email` and `relationship` for scripts, or omit missing fields on a terminal
to choose them interactively.

Cancel stops a pending invitation. Delete removes its stored metadata.
Both ask for confirmation; scripts require `--yes`. These actions do not remove
an accepted user's access grant. Use `access-grants` to manage that access.

Use `stuffstash invitations expiration INVITATION_ID` to change a pending
invitation's deadline. Enter a timestamp with a timezone, such as
`2030-01-01T12:00:00Z`. Scripts can supply a JSON file containing
`{"expiresAt":"2030-01-01T12:00:00Z"}` with `--input FILE --yes`.

## Accept an invitation

Use the household, inventory and invitation IDs from the invitation. The inventory
may not appear in your picker until you accept it.

```sh
stuffstash invitations preview INVITATION_ID --tenant HOUSEHOLD_ID --inventory INVENTORY_ID
stuffstash invitations accept INVITATION_ID --tenant HOUSEHOLD_ID --inventory INVENTORY_ID
```

The terminal asks for the acceptance token without showing it. Acceptance shows
the destination and role before confirmation. Scripts use `--input FILE --yes`,
with `{"acceptanceToken":"TOKEN"}` in the file. Use `--input -` for JSON on stdin.
Keep the token private. If the acceptance result is unknown, preview the same
invitation before retrying.

## Inspect your server

`stuffstash server show` displays the instance ID and protocol version.
`stuffstash server auth-config` shows the server's CLI login configuration.
Neither command requires a login or inventory selection. Add `--json` for the
complete API response, or `--server URL` to inspect another server.

## Inspect import jobs

```sh
stuffstash import-jobs list
stuffstash import-jobs show JOB_ID
stuffstash import-jobs show JOB_ID --json
stuffstash import-jobs delete JOB_ID
```

JSON includes the full preview, counts, messages, progress history and created
resources. Delete removes a job from history; it does not remove imported assets.
It asks for confirmation, or requires `--yes` in scripts. These commands do not
poll or retry automatically.

Use `stuffstash import-jobs cancel JOB_ID` to stop a job. Choose whether to keep
or discard partial progress, then confirm. Scripts use `--input FILE --yes`
with `{"mode":"keep_partial_progress"}` or
`{"mode":"discard_partial_progress"}`. Discarding can remove imported records.
Cancellation can continue in the background; use `import-jobs show JOB_ID` to
inspect its current state.

## Preview and start imports

Run `stuffstash import-jobs preview` for guided legacy Homebox or CSV input.
Credentials are masked. Private networks and untrusted TLS stay blocked unless
you explicitly enable them. The server reads the source; the CLI does not connect
to it. Preview creates a review job without importing inventory records.

Scripts can pass protected JSON with `--input source.json --yes`. See
`import-jobs preview --help` for all fields. Keep credentials out of arguments and
shell history. After reviewing `import-jobs show JOB_ID`, run
`import-jobs start JOB_ID --input source.json --yes` with the same source and
security choices. The server checks that they match the preview. CSV files may
be up to 10 MiB; JSON input may be up to 16 MiB. Server deployment limits also apply.

## Create and restore archives

Archive history is household-wide by default. Add `--inventory INVENTORY_ID`
to filter it; your saved inventory does not narrow the list automatically.

```sh
stuffstash archive-jobs list
stuffstash archive-jobs create
stuffstash archive-jobs show JOB_ID
stuffstash archive-jobs download JOB_ID --output backup.zip
stuffstash archive-jobs upload --file backup.zip
stuffstash archive-jobs preview RESTORE_JOB_ID
stuffstash archive-jobs approve RESTORE_JOB_ID
```

Creation asks which inventory and attachments to include. Upload sends ZIP bytes
without extracting them. Review the preview before approving a restore into a
new inventory. Scripts use `--input FILE --yes`: creation requires `inventoryId`,
`photos` and `otherFiles`; approval requires `name`. Save the reported request
key when creating or uploading. After an uncertain response, inspect job history
before repeating the same request with `--idempotency-key KEY`.

Downloads never overwrite a path. `--output -` writes only archive bytes to
stdout; omit `--json` in this mode. `--file -` uploads bytes from stdin. `archive-jobs retry JOB_ID` asks the
server to retry a job once. `archive-jobs delete JOB_ID` removes its retained
content. Mutations require confirmation; jobs are not polled automatically.

## Inspect model providers

Use `stuffstash provider-profiles list` to see model providers in the selected
household. Use `stuffstash provider-profiles show PROFILE_ID` to inspect a
provider's configuration, credential status and last test time. These commands
do not require an inventory and do not test or change the provider.

Use `--json --no-input` for scripts. JSON includes all profile configuration
and response metadata. Credential values are never returned.

Use `provider-profiles enable PROFILE_ID`, `disable PROFILE_ID`, or
`archive PROFILE_ID` to change a provider's state. Use
`provider-profiles test PROFILE_ID` to contact the configured provider and
record a test result. Each action asks for confirmation; scripts need `--yes`.

Inspect the test's `status` and `message`: a completed request can report a
failed provider test. The CLI does not retry these actions automatically. If a
request is interrupted, read the profile before repeating it.

Use `stuffstash voice-provider show` to inspect the household's voice provider
configuration and selected profiles. It does not start a voice session or change
settings. Use `--json` for all configuration fields and response metadata;
credential values are not returned.

## Set up a provider

Run `provider-profiles create` in a terminal to choose the provider and capability,
name the profile, and configure optional fields. The CLI does not guess an
endpoint or model. Use your provider's documented values.

`provider-profiles update PROFILE_ID` changes only the fields you choose.
`provider-profiles credential PROFILE_ID` asks for a credential purpose and reads
keys or tokens without displaying them. Server ADC uses the server's Google
credentials and does not ask for a secret.

For scripts, each command accepts `--input FILE|- --yes`. A create request can be:

```json
{
  "displayName": "Local inference",
  "providerKind": "local_http",
  "capability": "language_inference",
  "endpointUrl": "https://model.example",
  "modelName": "your-model",
  "enable": false
}
```

Update JSON can include displayName, endpointUrl, modelName, promptTemplate,
runtimeOptions, and capabilityMetadata. Omitted fields stay unchanged. JSON
values, including explicit null and false, are sent without rewriting them.
Credential JSON uses purpose (`api_key`, `oauth_bearer`, or `server_adc`) and
credential for keys or tokens. Keep credential files private, or pipe them through
stdin. Never put credentials in command arguments or runtimeOptions.

All three commands show the household and ask for confirmation; scripts use
`--yes`. They do not test providers or retry changes automatically. Use
`provider-profiles test PROFILE_ID` separately to check the configuration.

## Choose household voice providers

`stuffstash voice-provider update` offers keyboard choices for each capability:
keep the current explicit profile, use automatic selection, or choose an existing
compatible profile. Review all three choices before confirming. This does not
create provider profiles or start a voice session.

Scripts use `--input FILE|- --yes`. Supported fields are
`languageInferenceProfileId`, `speechToTextProfileId` and `textToSpeechProfileId`.
This API replaces all three selections: omitted, null or empty IDs reset that
capability to automatic selection. They do not disable voice. The command shows
each resulting choice before writing. There is no revision check on this API;
concurrent edits can replace one another. The CLI never retries the update.

## Inspect workflows and evaluations

These administration commands use the selected household without requiring an
inventory. They read configuration and evidence; they do not start conversations
or run evaluations.

| Task | Command |
| --- | --- |
| List workflows | `workflows list` |
| Read the latest workflow revision | `workflows show WORKFLOW_ID` |
| List workflow revisions | `workflows revisions list WORKFLOW_ID` |
| Read a workflow revision | `workflows revisions show WORKFLOW_ID REVISION_ID` |
| Read the selected workflow | `workflows selection show` |
| List evaluation cases | `evaluation cases list` |
| Read the latest case revision | `evaluation cases show CASE_ID` |
| List case revisions | `evaluation revisions list CASE_ID` |
| Read a case revision | `evaluation revisions show CASE_ID REVISION_ID` |
| List evaluation runs | `evaluation runs list` |
| Read run results | `evaluation runs show RUN_ID` |

Prefix each command with `stuffstash`. Lists support `--limit` and `--cursor`.
Use `--json --no-input` for complete configuration, evidence and response
metadata in scripts.

To stop an evaluation, run `stuffstash evaluation runs cancel RUN_ID` and
confirm the displayed version. Scripts must supply `--input FILE --yes` with
`{"expectedVersion":3}`, using the version from `evaluation runs show RUN_ID`.
A conflict requires a new review of the run. The CLI does not retry cancellation
or silently replace your version check.

## Create evaluation cases and queue runs

Use `stuffstash evaluation cases create --input FILE` with a `definition` that
contains `title`, `utterance` and `expectations`. Optional fixture `assets` describe
the inventory used by the case. To revise a case, use
`stuffstash evaluation revisions create CASE_ID --input FILE` and include its
current `expectedRevision`. Command help lists the supported expectation fields.

`stuffstash evaluation runs create --input FILE` queues a background text-only
evaluation. Supply `workflowId`, `revisionId` and one to 100 case/revision pairs in
`cases`. This can make provider calls; it does not activate the workflow. Review
the displayed household and target before confirming. Scripts must add `--yes`;
`--input -` reads standard input. Use `evaluation runs show RUN_ID` for results.
The CLI preserves your versions and does not retry writes automatically.

## Create and activate workflows

Use `stuffstash workflows create --input FILE` to create a workflow, or
`stuffstash workflows revisions create WORKFLOW_ID --input FILE` to add a revision.
The JSON must contain `definition`; a new revision also requires the current
`expectedRevision`. Use `--input -` to read JSON from standard input.

To select an evaluated revision, run
`stuffstash workflows activate WORKFLOW_ID --input FILE`. Supply `revisionId`,
`runId` and `cases`, plus an expected selection when needed. The server checks the
evaluation evidence before activation. These commands show the household and
ask for confirmation; scripts must add `--yes`.

Requests retain your exact values and concurrency checks. A conflict requires
review of current state. If a reply is lost, inspect the workflow before trying
again; the CLI does not retry writes automatically.

## Submit existing client measurements

`stuffstash telemetry submit --input measurements.json --yes` records an explicit
batch of one to 50 measurements from iOS, Android or web clients. Command help
lists the required fields. This account-level command does not collect CLI usage
automatically. Without `--yes`, a terminal asks for confirmation. Do not repeat an
uncertain submission without checking server telemetry; the batch can be counted
twice. Use `--json` for the accepted count and response metadata.
