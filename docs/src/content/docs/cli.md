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

Credentials use your operating system's keyring. On a headless machine without a
keyring, explicitly choose a private credential file before signing in:

```sh
mkdir -m 700 -p "$HOME/.config/stuffstash"
export STUFF_STASH_CLI_CREDENTIAL_FILE="$HOME/.config/stuffstash/session.json"
```

The file must remain accessible only to your account. `./stuffstash logout`
removes the locally stored session for the selected server.

## Work with assets

Set the tenant and inventory IDs from your instance. Listing inventories requires
a tenant; asset commands also require an inventory.

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

Use `--parent root` to move an asset out of a container. Add `--json` for structured
output. Lists include a continuation cursor when more results are available;
pass it with `--cursor`. Flags can also set `--server`, `--tenant`, and
`--inventory` for one command.

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

The worker and recovery flows are verified with stateful API and USB protocol
fakes. Physical printing, the example udev rule, and service operation still need
verification on your host; no attached QL-800 was available during development.

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
explicit resolution first.

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
rejected instead of being resized. See [supported label sizes](./printing/label-sizes/)
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
