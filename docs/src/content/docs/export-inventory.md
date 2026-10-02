---
title: Export and restore an inventory
description: Move inventories between instances, or download data for other tools.
---

## Back up or move an inventory

In inventory settings, choose **Export archive**. Inventory data is always
included. Leave **Photos** and **Other files** selected for a complete inventory
backup, then choose **Create archive**. The server prepares a ZIP in the background;
you can leave the screen and return to download it. Mobile opens the system share
sheet so you can save the ZIP to Files or another app.

The ZIP contains a JSON inventory document and the original attachments you
selected. Turning off either attachment option makes a partial backup. CSV is for
spreadsheets, not backup or restore.

To restore on another Stuff Stash instance, sign in there first. On mobile, open
the household/inventory switcher and choose **Restore inventory** beside
**New inventory**. On the web, open the destination household's settings. You must
have permission to create inventories in that household; it can be empty.

1. Choose the ZIP and select **Upload and validate**.
2. Choose **Review restore** to check its contents and any omitted attachments.
3. Enter a name and choose **Restore inventory**.
4. When it finishes, choose **Open inventory**.

Restore creates a new inventory with new IDs and reconnects its internal
relationships. It never overwrites an existing inventory. Conflicting custom
definition keys are renamed and reported in the review. Selecting or validating a
file alone does not create an inventory.

Jobs are private to the person who started them. Leaving the screen cancels a
local file transfer, but does not cancel work already accepted by the server.
Use a job's **Cancel** action to cancel server work, or **Retry** after a failure.
Download availability expires after 24 hours by default; the job shows its expiry.
Restored inventory data remains after the job expires.

An inventory archive is not a server backup. It does not transfer accounts,
credentials, sharing access, provider configuration, or audit history. Keep separate
database and blob-storage backups for disaster recovery.

## Download JSON or CSV

In the web or mobile app, open **Settings**, choose your inventory, then choose
**Export inventory**. Anyone who can view the inventory can export it.

- **JSON** keeps the complete inventory document, including field definitions and
  asset types. Choose this when you need structured data for another tool.
- **CSV** gives you one row per asset for a spreadsheet. Tags, custom fields,
  checkout details, and attachment metadata appear as JSON inside their cells.

Both formats include active and archived assets, their parent IDs, tags, expiration
values, and current checkout details. Archived tags and attachment metadata are
included too. Deleted records cannot be recovered through export.

Photo and file contents are **not** included. Neither are account credentials,
sharing tokens, provider settings, or audit history. Keep database and blob-storage
backups for a complete server recovery. An export may reflect edits made while it
is being prepared; it is not a point-in-time snapshot.

CSV protects against spreadsheet formulas by prefixing formula-like text with a
single quote. JSON preserves the original text. Both formats preserve Unicode,
commas, quotes, and multiline descriptions.

On mobile, choose a format and use the system share sheet to save or send the
file. On iPhone and iPad, **Save to Files** keeps a copy in your chosen folder.
Canceling preparation prevents the share sheet from opening. Once you send a
copy to another app, canceling cannot take it back.

## If a download fails

The app keeps you in Settings and offers **Retry export**. If your session has
expired, sign in again. If access was removed, ask the inventory owner to restore
it. A failed request does not produce a partial inventory file.

The server defaults to 10,000 records per collection and a 64 MiB encoded file.
Administrators can change these positive-integer environment settings:

| Setting | Default |
| --- | --- |
| `STUFF_STASH_EXPORT_MAX_RECORDS` | `10000` |
| `STUFF_STASH_EXPORT_MAX_BYTES` | `67108864` |

Raise limits only when the server has enough memory and request time to assemble
and encode the larger inventory. Attachment metadata shares one collection limit
across the whole inventory.

## API access

Use your normal authenticated API client to request:

```text
GET /tenants/{tenantId}/inventories/{inventoryId}/export?format=json
GET /tenants/{tenantId}/inventories/{inventoryId}/export?format=csv
```

The response is the file itself, with an attachment filename and private,
no-store cache policy. JSON is the default when `format` is omitted. The JSON
format has `schemaVersion: 1`; it is not an import or restore protocol.
