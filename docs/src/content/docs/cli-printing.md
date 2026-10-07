---
title: Print labels from the CLI
description: Register a printer connector, request labels, and recover uncertain print jobs.
---

Use a human [CLI sign-in](../cli/#sign-in) to configure printers and request
labels. A registered Linux connector runs the USB print worker with its own
restricted credential. You can render label files without a USB connector.

- [Register a connector](#register-a-printer-connector) and [run the worker](#run-the-print-worker).
- [Request labels](#request-and-inspect-labels) or [save a label file](#save-or-resolve-a-label).
- [Inspect and resolve a print job](#investigate-a-print-job).

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

Pairing registration and rotation emit safe protocol receipts as they progress.
With `--json`, each receipt is a JSON line containing its operation name and
response envelope. Receipts include pairing expiry and the new credential's
activation deadline, but never polling tokens, machine credentials, signatures,
or private keys. A credential-exchange receipt confirms the server step; the
final registration result confirms local storage and activation also succeeded.

Receipt display is best effort. If output stalls after successful registration,
the CLI keeps the stored credential and returns success without a final display
message. Do not repeat registration or rotation because receipt output is missing.

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
explicit resolution first. After checking the physical printer, use
[`print-jobs resolve`](#investigate-a-print-job) to record what happened.

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

## Save or resolve a label

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

## Render a complete JSON selection

`stuffstash labels render ASSET_ID --input selection.json --output label.png`
uses the complete `media`, `template` and `format` from your JSON. It does not
look up or replace your selection with saved defaults. Do not combine `--input`
with media, template or format flags. Use the label API's snake_case field names,
such as `width_micrometers` and `show_reference`; command help lists the fields.

Add `--json` to include the complete render response under `render`, alongside
the saved path, format and checksum. This includes fingerprints, expiry and
response metadata. PNG/PDF validation, scoped downloads and private file
publication still apply. Existing files are never overwritten.

## Update the loaded label size

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

## Investigate a print job

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

## Submit a reviewed print request

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

## Manage printer configuration

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

## Approve a connector from the CLI

On the administrator's computer, use `connectors print pairings review PAIRING_ID`
to inspect the connector name, public-key fingerprint and discovered candidates.
`connectors print pairings approve PAIRING_ID` asks for the short code privately,
then lets you choose compatible printers with the keyboard. Review the household,
inventory and bindings before confirmation. These commands use your normal sign-in.

Scripts use `--input FILE|-` containing `userCode`, `tenantId` and `inventoryId`;
approval also needs `bindings` with `candidateId` and `printerId` pairs, and
`--yes`. Input scope must match the selected scope. The code never belongs in a
command argument. Approval does not reserve scope or bypass server checks.

Use `connectors print rotations approve CONNECTOR_ID` for a rotation pairing.
It requires `generation`, `pairingId` and `userCode` in JSON, or asks for those
values privately where appropriate. Review the fingerprint and replacement
warning before confirmation. Scripts add `--yes`. The CLI never substitutes a
new generation after a conflict or retries approval automatically.
