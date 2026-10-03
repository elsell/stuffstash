---
title: Printing labels
description: Choose a label layout and prepare your Brother printer.
---

Labels let you return to an item, container, or location by scanning its QR code.
They require sign-in and current inventory access; a label does not make your
inventory public.

## What is available

The source build includes label rendering, connector registration, and a foreground
CLI worker for registered printer queues. Client setup and print controls are still
being integrated. See [connector setup](../../cli/#register-a-printer-connector). This catalog describes
the candidate source, and must not be read as a promise that an older downloadable
CLI has the same features. See [CLI installation](../../cli/) for current downloads.

Initial hardware support is **Brother QL-800, USB, Linux, with 29 × 90 mm labels**.
Physical printing and scanning still need device verification. Other printer models,
roll sizes, and remote wake are not part of the initial supported setup.

## Choose the loaded size

When registering a printer, select the label size physically loaded in it. The app
trusts that setting; automatic roll detection is not required. After changing a
roll, update its setting before requesting more labels. Existing jobs retain their
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
