---
title: Use the CLI
description: Sign in and work with your inventory from a terminal.
---

The Stuff Stash CLI uses the same permissions as the web and mobile apps. You can
list inventories, find assets, and create, move, archive, or restore them.

## Install from source

Published CLI downloads are not available yet. With Go 1.25.8 installed, build
from the repository root:

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

Printer registration and background printing are being implemented separately.
They are not available in this CLI slice yet.
