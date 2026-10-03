---
title: Printing labels
description: Choose a label layout and prepare your Brother printer.
---

Labels let you return to an item, container, or location by scanning its QR code.
They require sign-in and current inventory access; a label does not make your
inventory public.

## What is available

The source build includes label previews and scanning, web connector approval,
printer settings, and print controls in web and mobile. A foreground CLI worker
runs registered printer queues. See [connector setup](../../cli/#register-a-printer-connector).
Initialize label identities once using the [self-host instructions](../../self-host-operations/#set-up-labels-and-printing).
These instructions describe the current source build. An older downloadable CLI
may not include every command; check [versioned downloads](../../cli-downloads/)
or [build from source](../../cli/#install-from-source).

Initial hardware support is **Brother QL-800, USB, Linux, with 29 × 90 mm labels**.
Physical printing and scanning still need device verification. Other printer models,
roll sizes, and remote wake are not part of the initial supported setup.

## Request a label

In inventory settings, choose the default printer and layout, then choose whether
new items should print a label by default. You can change that choice on the create
form. On an existing item, container, or location, use its menu to print a label.
Preview the layout before sending it to the printer. Inventory printing settings
show registered printers and recent jobs. Use **Print test label** to check a
printer without creating an inventory item; it prints only after confirmation.

A queued job can wait while its printer is unavailable. Connector availability and
printer readiness are separate: a healthy computer cannot print through a powered-off
printer. If a job's outcome is uncertain, check the physical label and follow the
recovery controls after the connector confirms the printer is idle. Resolving that
status never prints another label automatically. When you need another copy,
choose **Print another label** on a finished job and review the new request.
If a response is lost, use the offered retry to recover the same request.

## Choose the loaded size

When registering a printer, select the label size physically loaded in it. The app
trusts that setting; automatic roll detection is not required. After changing a
roll, edit the printer in inventory printing settings before requesting more
labels. Initially, only the 29 × 90 mm preset is available. Existing jobs retain their
original size instead of silently stretching to another size.

A printer is independent of its layout. Use **QR and title** when a readable name
helps, or **QR** when the code is enough. The reference line is optional. See the
[rendered examples](../templates/) and [supported sizes](../label-sizes/).

## Connect the Brother

Connect it by USB to the Linux computer running the CLI. The user running the CLI
needs access to its `/dev/usb/lp…` device through the operating system's USB printer
permissions; avoid running the whole CLI as root. Check local discovery:

```sh
stuffstash printers discover --json
```

Discovery reports local device details and does not print a label. Keep that output
private when asking for support. The connector computer being online does not mean
the printer is ready: check its power, loaded roll, cover, and reported errors.
The built-in adapter does not remotely wake a powered-off printer.

## Scan after printing

Try the in-app scanner and your phone's ordinary camera. In-app scanning can resolve
an older label hostname against your currently configured instance; a normal camera
still needs the printed web address to remain reachable. Keep label URLs stable when
possible. Printed QR codes identify the inventory instance, and do not contain login
credentials.

## Regenerate this catalog

Contributors can refresh these pages without a server, credentials, or printer.
With the pinned Go toolchain installed, download the locked build dependencies once,
then generate offline:

```sh
(cd apps/api && GOWORK=off go mod download)
(cd apps/cli && GOWORK=off go mod download)
make printing-docs-generate
make printing-docs-check-generated
```

Commit the generated pages, images, and manifest with the adapter or layout change.
The check reports stale, missing, or extra generated files and preserves this guide.
