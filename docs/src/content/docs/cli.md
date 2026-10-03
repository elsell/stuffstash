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
commit. USB printer support remains limited to Linux and the supported Brother
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

The foreground print worker is delivered separately from registration.
