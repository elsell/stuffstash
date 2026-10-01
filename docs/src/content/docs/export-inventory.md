---
title: Export an inventory
description: Download your inventory as JSON or CSV.
---

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
