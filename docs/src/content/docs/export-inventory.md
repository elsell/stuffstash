---
title: Export and restore an inventory
description: Move inventories between instances, or download data for other tools.
---

## Back up or move an inventory

On mobile, open **Inventory Settings → Import and export** and select
**Backup archive** under Export. On the web, choose **Export archive** in
inventory settings. Inventory data is always included. Leave **Photos** and **Other files** selected for a complete inventory
backup, then choose **Create archive**. The server prepares a ZIP in the background;
you can leave the screen and return to download it. Mobile opens the system share
sheet so you can save the ZIP to Files or another app.

The ZIP contains a JSON inventory document and the original attachments you
selected. Turning off either attachment option makes a partial backup. CSV is for
spreadsheets, not backup or restore.

To restore on another Stuff Stash instance, sign in there first. On mobile,
choose **Import** on the **Import and export** screen. You can also open
**Restore inventory** from the household/inventory switcher, including when the
household has no inventory. On the web, open the destination household's settings. You must
have permission to create inventories in that household; it can be empty.

1. Choose the ZIP and select **Upload and validate**.
2. Choose **Review restore** to check its contents and any omitted attachments.
3. Enter a name and choose **Restore inventory**.
4. When it finishes, choose **Open inventory**.

Restore creates a new inventory with new IDs and reconnects its internal
relationships. It never overwrites an existing inventory. Conflicting custom
definition keys are renamed and reported in the review. Selecting or validating a
file alone does not create an inventory.

On mobile, the activity list shows your archive imports and exports across the
current household. Work in progress and archives waiting for review appear first,
followed by past jobs. Return to this screen to review, download or open the
result; accepted jobs continue even if you leave or restart the app.

Jobs are private to the person who started them. Leaving the screen cancels a
local file transfer, but does not cancel work already accepted by the server.
Use a job's **Cancel** action to cancel server work, or **Retry** after a failure.
Cancellation is no longer available once the restored inventory has been created
and its access is being finalized. **Open inventory** appears when it is ready to use.
Download availability expires after 24 hours by default; the job shows its expiry.
Restored inventory data remains after the job expires.

An inventory archive is not a server backup. It does not transfer accounts,
credentials, sharing access, provider configuration, or audit history. Keep separate
database and blob-storage backups for disaster recovery.

Archive uploads default to a 1 GiB limit and a 30-minute transfer timeout.
Administrators can configure `STUFF_STASH_ARCHIVE_MAX_BYTES` and
`STUFF_STASH_ARCHIVE_TRANSFER_TIMEOUT`; reverse-proxy upload limits and timeouts
must also allow the configured transfers. The server needs writable temporary
disk space for archive preparation and validation.

## Download JSON or CSV

On mobile, open **Inventory Settings → Import and export**, select Export, then
choose **JSON data** or **CSV spreadsheet** and start the export. On the web,
choose **Export inventory** in inventory settings. Anyone who can view the
inventory can export it.

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

On mobile, stay on the screen while a JSON or CSV file is prepared, then use
the system share sheet to save or send it. These direct downloads do not appear
in archive activity; use a backup archive for a saved background job. On iPhone and iPad, **Save to Files** keeps a copy in your chosen folder.
Canceling preparation prevents the share sheet from opening. Once you send a
copy to another app, canceling cannot take it back.

## If a download fails

The app keeps you on the export screen with an error and a way to try again. If your session has
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
