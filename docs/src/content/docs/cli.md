---
title: Use the CLI
description: Sign in, choose an inventory, and work from a terminal or script.
---

[Download the CLI](../cli-downloads/) for your computer. The commands below use
`stuffstash` on your `PATH`. If you keep it in the current directory, use
`./stuffstash` on Linux/macOS or `.\stuffstash.exe` in PowerShell instead.

The CLI uses the same permissions as the web and mobile apps. Start by signing
in and selecting an inventory, then create, find, or update your items.

- [Sign in](#sign-in) and [choose an inventory](#work-with-assets).
- [Run scripts and read every page](#read-every-page).
- [Set up printers and labels](../cli-printing/).
- [Manage access, backups, imports, and providers](../cli-administration/).

Shell examples use a POSIX shell unless marked PowerShell. Replace uppercase
placeholders such as `ASSET_ID` with IDs from command results.

## Help and shell completion

Use `stuffstash --help` for command groups, or `stuffstash assets create --help`
for the options, scope, confirmation behavior and examples for one command.
Help does not require sign-in.

With `stuffstash` on your `PATH`, enable completion for the current Bash session:

```sh
stuffstash completion bash > stuffstash.bash
source stuffstash.bash
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
stuffstash login
```

In PowerShell, set the server with:

```powershell
$env:STUFF_STASH_CLI_SERVER = 'https://api.example.com'
stuffstash login
```

Browser sign-in opens your provider's login page. On a computer without a browser,
use device-code sign-in if your provider and administrator enable it:

```sh
stuffstash login --device-code
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

The file must remain accessible only to your account. `stuffstash logout`
removes the locally stored session for the selected server.

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
stuffstash inventories list
export STUFF_STASH_CLI_INVENTORY=YOUR_INVENTORY_ID
stuffstash assets list --limit 20
stuffstash assets create --kind item --title 'Cordless drill'
stuffstash assets show ASSET_ID
stuffstash assets move ASSET_ID --parent CONTAINER_ID
stuffstash assets archive ASSET_ID
stuffstash assets restore ASSET_ID
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

### Search assets

```sh
stuffstash assets search --query "Cordless drill"
stuffstash assets search --query "Tools" --all-inventories
stuffstash assets search --query "Drill" --mode exact --tag-id TAG_ID
```

Search uses your selected inventory. Add `--all-inventories` to search the
selected household. A temporary `--inventory` override does not change your
saved selection. Use `--limit` and `--cursor` to page through results; `--json`
keeps the full match details and pagination for scripts.

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

## Read every page

Supported list commands accept `--all` to fetch all remaining pages:

```sh
stuffstash assets list --all --json --no-input
```

`--limit` sets each request's page size. Add `--cursor` to start at a saved
position. Filters and scope stay the same across requests. The CLI prints one
combined result after every page succeeds; cancellation or a failed request
produces an error without partial output. Large lists use more memory.

Data can change between requests, so the result is not a snapshot. JSON keeps
its existing shape and the final page's metadata. Use the command's `--help`
to check whether it supports `--all`.

## Create and rename households or inventories

```sh
stuffstash tenants create --name 'Home'
stuffstash inventories create --tenant HOUSEHOLD_ID --name 'Garage'
stuffstash inventories update --name 'Workshop'
```

Omit `--name` in an interactive terminal to enter the name in a prompt. Use
arrow keys to edit and Ctrl-C to cancel. Scripts must supply a name or JSON.

Use `--input request.json` instead of `--name` to send a JSON object, or pipe it
with `--input -`. The limit is 1 MiB. The CLI prints the target scope to stderr
before a write; `--json` keeps the result on stdout. New resources do not change
your current context.

These four create/update operations do not support idempotency keys. If a create
response is lost, check `tenants list` or `inventories list` before retrying.

## Custom asset types and fields

Use `asset-types` and `field-definitions` to list, show, create, update, archive,
restore, or delete definitions. Choose `--scope household` to share a definition
across the household, or `--scope inventory` for one inventory. Terminals offer a
scope picker; scripts must supply the scope. Inventory lists include inherited
household definitions.

```sh
stuffstash asset-types create --scope household --tenant HOME --key appliance --name Appliance
stuffstash field-definitions create --scope inventory --tenant HOME --inventory GARAGE --key warranty --name Warranty --field-type date
stuffstash asset-types list --scope inventory --tenant HOME --inventory GARAGE --lifecycle all
```

Omit common required fields in a terminal to enter them through prompts. Enum
fields prompt for options. For full control, use `--input definition.json` or
`--input -` instead of field flags. Each command's `--help` lists its JSON fields.
An asset type update can explicitly disable expiration with
`--expiration-enabled=false`. Complex field updates, such as appending enum
options or applicable asset types, use JSON; the server checks compatibility.

Changes require confirmation; scripts add `--yes`. Delete permanently removes
an eligible definition. There are no automatic retries: inspect `list` or `show`
if a change's result is uncertain. Lists accept `--limit` and `--cursor`; JSON
includes complete definition fields and response metadata.

## Archive, restore, or delete a household or inventory

Use `tenants archive`, `tenants restore`, or `tenants delete` for the selected
household. Use the equivalent `inventories` commands for the selected inventory.
Each command shows its target before it runs. Archive and delete ask for
confirmation with Cancel selected first. Scripts must supply `--yes`; restore
does not need confirmation. Delete permanently removes the selected resource.
After deletion, matching saved scope is cleared for your account.

```sh
stuffstash inventories archive --tenant HOUSEHOLD_ID --inventory INVENTORY_ID
stuffstash inventories restore --tenant HOUSEHOLD_ID --inventory INVENTORY_ID
```

## Manage tags

```sh
stuffstash tags list
stuffstash tags create --name 'Tools' --key tools --tag-color '#247BA0'
stuffstash tags update TAG_ID --name 'Hand tools'
stuffstash tags update TAG_ID --tag-color ''
stuffstash tags delete TAG_ID
```

Tags use the current household and inventory. Omit a required name to enter it
interactively, or use `--input FILE` for JSON. A color-only update leaves the name
unchanged; an empty color value clears the color. The stable key is set only when
you create the tag. Delete asks for confirmation; scripts must add `--yes`.
`--tag-color` sets the tag color, while `--color` controls terminal presentation.

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

## Upload and download files

Upload a JPEG, PNG, WebP, or PDF attachment with
`stuffstash attachments upload ASSET_ID --file receipt.pdf`. The CLI sends the
file directly to storage and completes the attachment after the transfer. Use
`--transfer api` when direct storage is unavailable. Uploads do not retry
automatically. After an uncertain result, inspect `attachments list ASSET_ID`
before starting another upload.

Download an attachment with `attachments download ASSET_ID ATTACHMENT_ID
--output receipt.pdf`, or use `attachments thumbnail` with optional
`--variant small|medium|large`. `labels download RENDER_ID --output label.pdf`
downloads an existing label render. Existing files are never replaced; a failed
download does not publish a partial file. `--output -` writes only file bytes to
stdout; do not combine it with `--json`.

`inventories export --format json --output inventory.json` preserves inventory
data and referenced names. `--format csv` creates a report, not a full backup.
Neither export includes photo or file contents; use an archive for those.

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

## Install from source

With Go 1.25.8 installed, build from the repository root and check the local binary:

```sh
go build -buildvcs=false -o ./stuffstash ./apps/cli/cmd/stuffstash
./stuffstash version
```

## Register a printer connector

Printer setup, worker operation, and label commands now have a dedicated
[CLI printing guide](../cli-printing/#register-a-printer-connector).
