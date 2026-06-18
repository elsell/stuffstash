# Import And Export Spec

## Purpose

Stuff Stash should make user data portable.

Users should be able to export inventory data and later import data through supported formats without coupling the domain to one file format.

## Scope

This spec covers initial import and export requirements.

This spec does not define exact CSV columns, JSON schema, backup packaging, media export packaging, or import conflict resolution.

## Requirements

- Import and export must be behind ports and adapters.
- Export must support JSON.
- Export must support CSV.
- Import should support JSON and CSV once the corresponding import workflows are specified.
- Export must preserve tenant and inventory authorization boundaries.
- Users must only export inventories they are authorized to export.
- Imports must validate tenant, inventory, asset, location, and custom field behavior before applying changes.
- Imports must produce audit records for state-changing operations.
- Import and export adapters must not leak file-format details into domain logic.

## Media And Backups

- Photo and file export should be supported eventually.
- Initial export may exclude binary attachment content if the export clearly states that limitation.
- Tenant-level backups should be modeled as exports unless a future backup spec defines a separate mechanism.
- Export packaging for photos and files must be specified before binary media export is implemented.

## Testing

- Tests must verify JSON export, CSV export, authorization, tenant isolation, inventory isolation, custom field handling, and audit behavior.
- Tests must use fakes for storage and repositories where appropriate.
- Import tests must verify validation failures and partial-failure behavior before import is exposed to users.

## Open Questions

- What exact JSON export schema should be used first?
- What exact CSV columns should be used first?
- Should exports include audit history?
- Should exports include attachment metadata before binary file export is supported?
