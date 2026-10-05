---
title: Use the CLI
description: Sign in and work with your inventory from a terminal.
---

The Stuff Stash CLI uses the same permissions as the web and mobile apps. You can
list inventories, find assets, and create, move, archive, or restore them.

See [versioned downloads](../cli-downloads/) for published binaries and checksum
verification.

## Install from source

With Go 1.25.8 installed, build from the repository root:

```sh
go build -buildvcs=false -o ./stuffstash ./apps/cli/cmd/stuffstash
./stuffstash version
```

## Help and shell completion

Use `stuffstash --help` for command groups, or `stuffstash assets create --help`
for the options, scope, confirmation behavior and examples for one command.
Help does not require sign-in.

With `stuffstash` on your `PATH`, enable completion for the current Bash session:

```sh
stuffstash completion bash > stuffstash.bash
source ./stuffstash.bash
```

For Zsh, use `completion zsh` and source the generated file after `compinit`.
For Fish, save `completion fish` output as
`~/.config/fish/completions/stuffstash.fish`. To keep Bash or Zsh completion,
source the saved file from your shell startup file.

Completion suggests commands and relevant options. It does not look up inventory
values, read credentials, or contact the server.

## Sign in

Your administrator must configure a public CLI client with your OIDC provider
and set `STUFF_STASH_OIDC_CLI_CLIENT_ID` on the API. See the
[configuration reference](../configuration/#api-authentication).

```sh
export STUFF_STASH_CLI_SERVER=https://api.example.com
./stuffstash login
```

Browser sign-in opens your provider's login page. On a computer without a browser,
use device-code sign-in if your provider and administrator enable it:

```sh
./stuffstash login --device-code
```

Open the displayed verification address on your phone or another computer, enter
the code, and approve sign-in. The CLI waits for approval.

Credentials use your operating system's keyring. On a Unix headless machine without a
keyring, explicitly choose a private credential file before signing in:

```sh
mkdir -m 700 -p "$HOME/.config/stuffstash"
export STUFF_STASH_CLI_CREDENTIAL_FILE="$HOME/.config/stuffstash/session.json"
```

Windows uses Credential Manager and does not support credential files.

The file must remain accessible only to your account. `./stuffstash logout`
removes the locally stored session for the selected server.

## Create and rename households or inventories

```sh
./stuffstash tenants create --name 'Home'
./stuffstash inventories create --tenant HOUSEHOLD_ID --name 'Garage'
./stuffstash inventories update --name 'Workshop'
```

Omit `--name` in an interactive terminal to enter the name in a prompt. Use
arrow keys to edit and Ctrl-C to cancel. Scripts must supply a name or JSON.

Use `--input request.json` instead of `--name` to send a JSON object, or pipe it
with `--input -`. The limit is 1 MiB. The CLI prints the target scope to stderr
before a write; `--json` keeps the result on stdout. New resources do not change
your current context.

These four create/update operations do not support idempotency keys. If a create
response is lost, check `tenants list` or `inventories list` before retrying.

## Archive, restore, or delete a household or inventory

Use `tenants archive`, `tenants restore`, or `tenants delete` for the selected
household. Use the equivalent `inventories` commands for the selected inventory.
Each command shows its target before it runs. Archive and delete ask for
confirmation with Cancel selected first. Scripts must supply `--yes`; restore
does not need confirmation. Delete permanently removes the selected resource.
After deletion, matching saved scope is cleared for your account.

```sh
./stuffstash inventories archive --tenant HOUSEHOLD_ID --inventory INVENTORY_ID
./stuffstash inventories restore --tenant HOUSEHOLD_ID --inventory INVENTORY_ID
```

## Manage tags

```sh
./stuffstash tags list
./stuffstash tags create --name 'Tools' --key tools --tag-color '#247BA0'
./stuffstash tags update TAG_ID --name 'Hand tools'
./stuffstash tags update TAG_ID --tag-color ''
./stuffstash tags delete TAG_ID
```

Tags use the current household and inventory. Omit a required name to enter it
interactively, or use `--input FILE` for JSON. A color-only update leaves the name
unchanged; an empty color value clears the color. The stable key is set only when
you create the tag. Delete asks for confirmation; scripts must add `--yes`.
`--tag-color` sets the tag color, while `--color` controls terminal presentation.

## Work with assets

Run `stuffstash account show` to check the signed-in account and
`stuffstash tenants list` to see your households. A tenant is a household's
security boundary; each household can contain multiple inventories. Use
`tenants show` or `inventories show` to inspect the selected scope and your access.

In an interactive terminal, commands ask you to select any missing household or
inventory. Type to search, use the arrow keys, and press Enter. Escape cancels.
Completed selections are saved for the server and signed-in account. Use
`context list`, `context current`, and `context use NAME` to inspect or switch
saved contexts. `context delete NAME` removes a saved context without deleting
server data. Sign-out clears saved account scope.

For scripts, use explicit scope or a saved context for the same signed-in account.
`--json`, `--no-input`, and redirected streams never open a picker. Set
`STUFF_STASH_CLI_CONFIG_FILE` to choose a private context file. Context files do
not contain credentials. Picker color follows terminal support and `NO_COLOR`;
`--color always` or `--color never` overrides automatic color selection.

Listing inventories requires a tenant; asset commands also require an inventory:

```sh
export STUFF_STASH_CLI_TENANT=YOUR_TENANT_ID
./stuffstash inventories list
export STUFF_STASH_CLI_INVENTORY=YOUR_INVENTORY_ID
./stuffstash assets list --limit 20
./stuffstash assets create --kind item --title 'Cordless drill'
./stuffstash assets show ASSET_ID
./stuffstash assets move ASSET_ID --parent CONTAINER_ID
./stuffstash assets archive ASSET_ID
./stuffstash assets restore ASSET_ID
```

Create or update an asset with a complete JSON request using `--input FILE` or
`--input -` for stdin. This supports descriptions, tags, custom fields, expiration,
and parent IDs. For example, `{"expiration":null,"tagIds":[]}` clears expiration
and removes all tags with `assets update ID --input FILE`. Omitted fields stay
unchanged. Do not combine JSON input with asset field flags or `--print-label`.
Interactive creation asks for a missing title and kind. Scripts must provide
these fields or a JSON request. Updates do not support `--idempotency-key`. Ordinary
creation also does not support this key. If its response is lost, list assets
before retrying to avoid a duplicate. Create-and-print supports the key: keep
the key shown on stderr and reuse it with the unchanged request if needed.
The CLI does not retry writes automatically.

Archive and delete require confirmation. Add `--yes` in scripts after checking
the target. `assets delete ID --yes` permanently deletes an asset; the server
checks whether deletion is allowed. Restore needs no confirmation. None of
these lifecycle commands supports `--idempotency-key`.

List archived assets with `assets list --lifecycle archived`, or include both states
with `--lifecycle all`. Use `--sort updated_desc` to put recently changed assets
first; `id_asc` is also supported.

`assets show` displays the location ID, tags, expiration, checkout, and custom
fields. Its JSON output retains the complete asset response, including photo
metadata, request metadata, and exact custom-field numbers.

Use `--parent root` to move an asset out of a container. Add `--json` for structured
output. Add `--request-id ID` to correlate API requests with server logs; this
does not prevent duplicate writes. Directory JSON includes the API metadata
and optional schema reference. Lists include a continuation cursor when more results are available;
pass it with `--cursor`. Flags can also set `--server`, `--tenant`, and
`--inventory` for one command.

### Check out and return assets

```sh
stuffstash assets checked-out --limit 20
stuffstash assets checkout ASSET_ID --details 'Lent to Sam'
stuffstash assets return ASSET_ID --details 'Returned with charger'
stuffstash assets checkouts ASSET_ID --limit 20
stuffstash assets return-details ASSET_ID CHECKOUT_ID --details ''
```

Checkout and return notes are optional. Use `--details ''` to clear returned
checkout notes. The write commands also accept `--input FILE|-` with a JSON
`details` field. Do not combine the two input methods. Use the history command
before retrying a write whose result is unknown. These commands do not support
retry keys. History includes pagination; pass `--cursor` for the next page.

### Browse expiration dates

```sh
stuffstash assets expiration --mode expired
stuffstash assets expiration --mode soon --location-id LOCATION_ID
stuffstash assets expiration --from-date 2026-01-01 --through-date 2026-12-31 --tag-id TAG_ID
```

Omit `--mode` to include all expiration dates. Narrow results with `--kind`,
`--checkout-state any|available|checked_out`, `--query`, or `--type-id`.
Repeat `--tag-id` for multiple tags. Lists include counts, the inventory timezone,
and item locations. Use `--limit` (1–100) and `--cursor` to page through results,
or `--json` for complete items and metadata.

## Release operations

The normal Release workflow attaches five portable CLI archives, individual
SHA-256 files, and `stuffstash-cli-release.json` to the project tag. The archive
includes third-party notices; `stuffstash version` reports its tag and source
commit. `version --json` reports whether that build includes a USB printer adapter;
this is separate from whether a printer is connected or ready. USB printer support
remains limited to Linux and the supported Brother
profile, even when ordinary inventory commands run on another platform.

If publication stops after staging its assets, run **Release → Run workflow**
on **main**, with `repair_run_id` set to the original Release workflow run ID.
Repair uses its retained `release-publication` artifact and exact tag/commit.
Draft-creation failures include the GitHub error status and validation details.
A lost response may still leave a draft behind; repair finds that draft and resumes
from the original assets.
It uploads missing assets and verifies existing bytes; it never overwrites a
mismatch. Investigate a mismatch or an expired artifact rather than rebuilding
an old tag with new source or dependencies.

Verified stable publication opens a maintenance PR for download links and
self-host image digests. The docs deploy after that PR merges. If its checks or
merge require attention, fix the PR; then run **Docs Pages** on **main** to refresh
the site. A draft, failed publication, or older repaired release cannot replace
newer stable download instructions.

## Register a printer connector

Connect the Brother QL-800 to a Linux computer by USB, then discover it:

```sh
./stuffstash printers discover
./stuffstash connectors print register --name 'Garage computer'
```

Open the displayed approval address on your phone or computer and enter the
short code. Confirm the inventory, printer, and **29 × 90 mm** label size. The
CLI waits for approval and saves a separate, restricted connector credential.
Keep the connector ID printed when registration completes.

The connector uses its own keyring entry. For a headless computer, explicitly
choose a different credential file from your human login session:

```sh
mkdir -m 700 -p "$HOME/.config/stuffstash"
export STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE="$HOME/.config/stuffstash/garage-connector.json"
```

Set this before registering. Keep the file private and persistent across restarts.
Use a separate file for each connector. Registration never prints the connector
credential. If approval expires or an exchange response is lost, register again.
If activation fails after saving, retain the saved credential; the print worker
can retry activation.

## Run the print worker

Use the same Linux account and connector credential store used during registration:

```sh
export STUFF_STASH_CLI_CONNECTOR_ID=YOUR_CONNECTOR_ID
export STUFF_STASH_CLI_PRINT_STATE_DIRECTORY="$HOME/.config/stuffstash/print-state"
./stuffstash connectors print run
```

Leave this command running. It handles every printer assigned to the connector;
one disconnected printer does not stop the others. Each device has one active
worker, even if another process uses a different journal directory.

Only the **Brother QL-800 over USB on Linux with 29 × 90 mm stock** is supported
initially. Keep Editor Lite off. The Linux `usblp` driver must expose a writable
bidirectional device, usually `/dev/usb/lp0`. Your administrator can grant a
printer group access using a persistent udev rule; run the worker as a regular
user with that group, not as root. Discovery reads device information but does
not prove that the account can print.

If your distribution uses the `lp` group, this is an example rule for
`/etc/udev/rules.d/70-stuffstash-ql800.rules`:

```text
SUBSYSTEM=="usb", KERNEL=="lp[0-9]*", ATTRS{idVendor}=="04f9", ATTRS{idProduct}=="209b", GROUP="lp", MODE="0660"
```

An administrator must load `usblp` if necessary, install the rule, and add the
worker account to the selected printer group. Reconnect the printer and start a
new login session after permission changes. Group names and driver packaging
vary by distribution.

### Check status and recover

Stuff Stash shows connector availability separately from printer readiness. A
powered-off printer needs attention even while the connector is online. Wake it
or resolve the paper/cover error, then leave the worker running. The first adapter
does not provide remote wake or change the printer's power settings.

Keep the private journal directory across restarts. The worker records output
before sending it and reconciles API status before taking another job, even if
the printer is now unplugged. A lost USB or API response may leave an **uncertain**
job. Check the physical label before explicitly resolving or reprinting it;
restarting the worker never automatically repeats uncertain output. Retiring a
printer stops new claims while an already-started attempt can finish.

Ctrl-C stops the worker and preserves recovery evidence. Invalid or revoked
connector credentials stop it with re-pair guidance. Revoke the old connector
before pairing a replacement, use a separate credential file, and retain the old
journal until its attempts are settled.

### Run as a service

A user service can run the same foreground command. Put the binary at
`~/.local/bin/stuffstash`, then create
`~/.config/systemd/user/stuffstash-print.service`:

```ini
[Unit]
Description=Stuff Stash print connector

[Service]
Type=simple
Environment=STUFF_STASH_CLI_SERVER=https://api.example.com
Environment=STUFF_STASH_CLI_CONNECTOR_ID=YOUR_CONNECTOR_ID
Environment=STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE=%h/.config/stuffstash/garage-connector.json
Environment=STUFF_STASH_CLI_PRINT_STATE_DIRECTORY=%h/.config/stuffstash/print-state
ExecStart=%h/.local/bin/stuffstash connectors print run
Restart=on-abnormal

[Install]
WantedBy=default.target
```

Replace the server and connector ID, then run:

```sh
systemctl --user daemon-reload
systemctl --user enable --now stuffstash-print.service
systemctl --user status stuffstash-print.service
journalctl --user -u stuffstash-print.service
```

A dedicated service account needs persistent configuration and USB group access.
An administrator can enable user lingering when the worker must survive logout.
To uninstall, stop and disable the service, remove its unit and binary, and revoke
the connector in Stuff Stash. Retain journal and credential files until pending
attempts have been reconciled.

The worker and recovery flows have stateful API and USB protocol checks, plus
[scoped physical-print evidence](../printing/setup/#what-is-available) for a
candidate USB status fix. The example udev rule and service operation still need
verification on your host.

## Request and inspect labels

Set your server, tenant, and inventory context, then use your human login:

```sh
stuffstash printers list
stuffstash labels print ASSET_ID --printer PRINTER_ID
stuffstash printers test PRINTER_ID
stuffstash print-jobs list --printer PRINTER_ID
stuffstash print-jobs show JOB_ID
stuffstash print-jobs cancel JOB_ID
stuffstash print-jobs reprint JOB_ID --printer PRINTER_ID
```

These commands enqueue work for the connector; they do not send USB output from
the computer running the command. A test prints one diagnostic label. A reprint
creates a new job linked to the original; unresolved uncertain jobs require
explicit resolution first. Use inventory printing settings in the web or mobile
app to resolve an uncertain outcome after checking the physical printer.

Omit `--printer` to use the inventory's configured default. Template selection
also uses inventory defaults; override it with `--template qr-title
--template-version 1 --show-reference=true`. The destination determines the size.
Offline destinations still accept queued work.

Scripts create assets without printing unless requested:

```sh
stuffstash assets create --kind item --title "Spare batteries" \
  --print-label --printer PRINTER_ID --idempotency-key batteries-label-1
```

A failed label validation leaves no new asset behind. The response includes the
label job ID. Each print request writes its request key to stderr before sending.
If the response is lost, retry the same command with that key using
`--idempotency-key`; keep its arguments unchanged. A changed request conflicts
instead of silently printing another label. Use `--json` for structured output.

To replace a connector credential, use the same server and connector credential
store as the worker:

```sh
stuffstash connectors print rotate --connector CONNECTOR_ID
```

Open the printed link and approve replacement for that existing connector. The
CLI keeps its current credential until a matching replacement is received and
saved. If activation fails after saving, run `connectors print run` with the same
connector to retry activation. A rotation does not create a new printer or change
its label size. Stop the old worker and restart it after rotation so it loads the
new credential; existing uncertain jobs still require their normal recovery.

### Save or resolve a label

You can save labels without running a USB connector:

```sh
stuffstash labels templates
stuffstash labels render ASSET_ID --printer PRINTER_ID --format png --output label.png
stuffstash labels render ASSET_ID --media-preset brother-ql800-29x90 --format pdf --output label.pdf
stuffstash labels resolve 'https://old.example.com/l/v1/INSTANCE_ID/LABEL_ID'
```

Rendering uses the inventory's template defaults, with the same template options
as `labels print`. Omit the media selector to use the default printer. Choose a
catalog media preset to render without registering a printer. Paired `--width-mm`
and `--height-mm` also select an exact supported profile; the nominal 29 × 90 mm
Brother preset has a physical profile of 29 × 89.8 mm. Unsupported sizes are
rejected instead of being resized. See [supported label sizes](../printing/label-sizes/)
for available templates and stock.

PNG and PDF downloads are checked against the server's checksum before saving.
The output path must be new: an existing file or symlink is never overwritten.
Saving label files currently requires Linux or macOS. Windows refuses file output
until a private Windows file adapter is available; listing templates, resolving
labels, and requesting prints still work.
`--json` reports the path, format, and checksum without mixing image bytes into
terminal output. Saving a label does not enqueue a print job.

Resolution uses your configured server and login, even when the scanned label
contains an old hostname. It checks the instance identity and access before
returning the asset, tenant, and inventory IDs. It never sends your credentials
to the hostname printed in the QR code. Unlike scoped rendering, resolution does
not require a selected tenant or inventory.

### Update the loaded label size

Use your human login with permission to configure the inventory:

```sh
stuffstash printers configure PRINTER_ID --label-size brother-ql800-29x90
```

The command selects a supported preset for the registered printer. Initially,
only the Brother QL-800's 29 × 90 mm stock is supported. It preserves the printer's
name and other settings and works while the printer is offline. The connector
reads the updated setting from the server; its restricted credential cannot edit it.

If another person changes the registration meanwhile, the command reports a
conflict. Inspect the current printer before trying again. Updating stock never
resizes labels already queued: jobs keep their original media requirements.

## Manage attachments

Use the saved inventory, or pass `--tenant` and `--inventory`:

```sh
stuffstash attachments list ASSET_ID
stuffstash attachments show ASSET_ID ATTACHMENT_ID
stuffstash attachments archive ASSET_ID ATTACHMENT_ID
stuffstash attachments restore ASSET_ID ATTACHMENT_ID
stuffstash attachments delete ASSET_ID ATTACHMENT_ID --yes
```

Archive and delete ask for confirmation. Scripts must pass `--yes`.
Detail output includes the file size and SHA-256 digest. Add `--json` for the
complete API result. List supports `--limit` and `--cursor`.

If a direct upload has transferred its bytes but still needs completion, run
`stuffstash attachments complete-upload ASSET_ID UPLOAD_ID`. Keep the upload ID
private. This command does not resend the file. If completion returns an
uncertain result, use `attachments list ASSET_ID` to check before retrying.

## Review notifications

```sh
stuffstash notifications list --unread-only
stuffstash notifications show NOTIFICATION_ID
stuffstash notifications read NOTIFICATION_ID
stuffstash notifications unread NOTIFICATION_ID
stuffstash notifications unread-count
stuffstash notifications read-all
```

Commands use your selected inventory. List supports `--limit` and `--cursor`.
Unread count and read-all also accept a cursor. If read-all reports
`Complete: false`, use the returned cursor to continue. Add `--json` to retain
all notification fields and response metadata in scripts.

Look up a push registration with
`stuffstash notification-devices show INSTALLATION_ID`. To stop that registration,
run `stuffstash notification-devices remove DEVICE_ID --revision N` using the
returned device ID and revision. Removal asks for confirmation; scripts must
add `--yes`. A revision conflict requires reviewing the current registration
before retrying. The CLI does not replace your supplied revision automatically.

Use `stuffstash notification-preferences show` to inspect notification defaults,
type overrides, timezone, and revision. Initialize preferences with
`stuffstash notification-preferences initialize --timezone America/New_York`.
An interactive terminal can ask for the timezone when omitted. Scripts can also
supply `--input FILE` or pipe a JSON object through `--input -`.

Run `stuffstash notification-preferences update` in a terminal to edit current
settings with keyboard choices. Use `override TYPE_ID` to edit a type-specific
policy. Both preserve the revision loaded at the start and stop if it changes.
For scripts, pass `--input FILE` with the complete API request, including its
revision. Update requires `defaults`, `timezone`, and `pushEnabled`; override
requires `settings`. Policies contain `enabled`, `upcoming`, `expired`, and
`advanceDays`.

To return a type to the default policy, use
`stuffstash notification-preferences remove-override TYPE_ID --revision N`.
Removal asks for confirmation; scripts add `--yes`.

Register a push device with `stuffstash notification-devices register`. A terminal
asks for the installation ID, push service, revision, and a hidden token. Use
revision `0` for a first registration. For scripts, supply `--input FILE` or pipe
JSON with `--input -`; the body contains `installationId`, `transport` (`apns` or
`fcm`), `token`, and `revision`. Treat this input as secret. The CLI does not
include the token in results or error messages.

## Undo and redo

Use the `undoableOperationId` returned by a mutation:

```sh
stuffstash operations undo OPERATION_ID
stuffstash operations redo OPERATION_ID
```

These commands require confirmation; scripts add `--yes`. Use an operation ID,
not an asset ID. The server checks whether the change can still be reversed or
reapplied. If the result is uncertain, inspect the affected asset before retrying.

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

### Inspect inventory access

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

### Manage invitations

```sh
stuffstash invitations list --status pending
stuffstash invitations show INVITATION_ID
stuffstash invitations cancel INVITATION_ID
stuffstash invitations delete INVITATION_ID
```

Cancel stops a pending invitation. Delete removes its stored metadata.
Both ask for confirmation; scripts require `--yes`. These actions do not remove
an accepted user's access grant. Use `access-grants` to manage that access.

Use `stuffstash invitations expiration INVITATION_ID` to change a pending
invitation's deadline. Enter a timestamp with a timezone, such as
`2030-01-01T12:00:00Z`. Scripts can supply a JSON file containing
`{"expiresAt":"2030-01-01T12:00:00Z"}` with `--input FILE --yes`.

### Accept an invitation

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

### Inspect your server

`stuffstash server show` displays the instance ID and protocol version.
`stuffstash server auth-config` shows the server's CLI login configuration.
Neither command requires a login or inventory selection. Add `--json` for the
complete API response, or `--server URL` to inspect another server.

### Inspect import jobs

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

### Inspect model providers

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

### Investigate a print job

Use `stuffstash print-jobs list` to find a job, then
`stuffstash print-jobs show JOB_ID` to inspect its status, attempts and resolution.
The detail view includes attempt timing and the number of completed copies.
Inspect uncertain outcomes before printing again to avoid duplicate labels.
Add `--json` to retain the full job and response metadata in a script.

Use `stuffstash print-settings show` to inspect the selected inventory's default
printer, print-on-create setting and label template options. An unset printer is
shown as **Not set**. Use `--json` for the complete settings and revision.

Use `stuffstash printers show PRINTER_ID` for a printer's readiness reason,
last report time and complete media configuration. `printers list` gives a
compact overview; both commands retain all printer fields with `--json`.

Use `stuffstash labels templates` to see supported label options, defaults and
font coverage. Use `stuffstash printers profiles` for the server's printer
adapters, supported platforms and media presets. This differs from
`printers catalog`, which inspects the local CLI catalog without login.
Both server catalogs support `--json` for their complete configuration.

`stuffstash labels resolve LABEL_URL` shows the label's full identity, canonical
URL and inventory destination. It checks the instance against your configured
server and does not send your credentials to the link's host.

Use `stuffstash labels show ASSET_ID` to read an asset's existing label.
Use `stuffstash labels assign ASSET_ID` to obtain its stable label identity
without rendering or printing. Assignment asks for confirmation; scripts use
`--yes`. Both commands use the selected household and inventory.

Use `stuffstash connectors print list` or
`stuffstash connectors print show CONNECTOR_ID` to inspect connector availability,
assigned printers, heartbeat times and reported capabilities. These commands use
your normal sign-in and saved inventory context; they do not read connector secrets.

On a registered connector computer, use `stuffstash connectors print printers
--connector CONNECTOR_ID` to inspect its printer bindings. Use
`stuffstash connectors print attempts list --connector CONNECTOR_ID` or
`stuffstash connectors print attempts show ATTEMPT_ID --connector CONNECTOR_ID`
to inspect delivery attempts. These commands use the stored connector credential
and its registered scope. They do not access printer hardware or change attempts.
The list accepts `--printer`, `--status unsettled`, `--limit` and `--cursor`;
`--json` preserves the complete declared response and pagination metadata.

`stuffstash print-jobs cancel JOB_ID` asks for confirmation before reading and
cancelling the job. Scripts must add `--yes`. Cancellation uses the current job
revision and does not retry conflicts. A label might already have printed;
inspect the returned job before submitting another print.

Use `stuffstash print-jobs resolve JOB_ID` to record what you observed after an
uncertain print: printed, not printed, or unknown. This does not print again.
The interactive command asks for an outcome and confirmation. For scripts, pass
`--input FILE|- --yes` with `reportedOutcome` (`printed`, `not_printed`, or
`unknown`), the current positive `revision`, and
`acknowledgeUncertainty: true`. A stale revision fails without retrying.

### Inspect workflows and evaluations

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

### Create evaluation cases and queue runs

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

### Create and activate workflows

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

### Submit a reviewed print request

`labels print ASSET_ID`, `printers test PRINTER_ID`, and
`print-jobs reprint JOB_ID` accept `--input FILE` or `--input -`. The JSON supplies
the full selection: printer, media fingerprint, template and version, template
options, copies, and optional preview fingerprint. This form preserves your
reviewed request without looking up or replacing its values. Do not combine it
with selection flags. A test request's printer ID must match the command target.

For the usual flag-based workflow, `--expected-media-fingerprint` and
`--preview-fingerprint` let you retain values from a prior review. The server
rejects changed media or previews. It also enforces template and copy limits;
a printer test permits one copy.

Keep the printed request key and the exact request. If a reply is lost, inspect
the print job before repeating anything. Reuse `--idempotency-key` with the same
request for a retry; a new key can print another label. The CLI does not retry
a submission automatically.

### Manage printer configuration

`printers create` accepts `--input FILE|-` or `--name`, `--adapter`, `--label-size`
and `--preset-version`. A terminal can ask for missing fields. Supply an explicit
`--idempotency-key`; retain it with the unchanged request if a reply is lost.

Use `printers update PRINTER_ID`, `connectors print update CONNECTOR_ID`, or
`print-settings update` with `--input FILE|-` for configuration changes. Printer
and settings updates require the current `revision`; connector updates require
`generation`. Inspect the resource first and retain those exact values. Command
help lists supported fields. False, null and omitted fields remain distinct.

These commands show the selected household, inventory and effect before asking
for confirmation. Scripts add `--yes`. Connector changes can revoke access or
change printer bindings; print settings can change automatic label printing.
The CLI does not retry updates or replace a stale revision automatically.

### Submit existing client measurements

`stuffstash telemetry submit --input measurements.json --yes` records an explicit
batch of one to 50 measurements from iOS, Android or web clients. Command help
lists the required fields. This account-level command does not collect CLI usage
automatically. Without `--yes`, a terminal asks for confirmation. Do not repeat an
uncertain submission without checking server telemetry; the batch can be counted
twice. Use `--json` for the accepted count and response metadata.
