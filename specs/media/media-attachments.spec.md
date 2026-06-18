# Media Attachments Spec

## Purpose

Stuff Stash needs first-class media and file attachments for inventory records.

Photos, receipts, manuals, and arbitrary files help users identify and manage physical things.

## Scope

This spec covers initial media and attachment requirements.

This spec does not define the final upload protocol, image processing pipeline, virus scanning, retention policy, or model vision features.

## Requirements

- Images must be first-class attachments.
- Arbitrary file uploads must be supported eventually.
- Attachments must be scoped to an inventory and protected by the tenant boundary.
- Attachments must be associated with assets, and may later be associated with locations, inventories, audit events, or custom fields if specified.
- Attachment metadata must be stored separately from blob content.
- Blob storage must be behind ports and adapters.
- File access must be authorized through the same tenant, inventory, and resource relationships as the attached resource.
- File names, MIME types, sizes, hashes, and storage keys must be handled safely.
- Attachments must not expose storage-provider internals to domain logic.
- Images may be used by model providers in the future, but model vision behavior is not required initially.

## Storage Direction

- Production blob storage should use Garage as an S3-compatible storage system.
- Local development may use a simpler adapter through Docker Compose.
- Storage credentials, bucket names, endpoints, and region-style settings must come from environment-backed configuration.
- Blob storage dependencies and images must be pinned to reviewed versions.

## Security

- Uploads must enforce size limits.
- Uploads must enforce allowed MIME types once the first upload feature is implemented.
- Download URLs must not bypass authorization.
- Public buckets must not be used for private tenant inventory files.
- Error messages must not leak storage keys, credentials, bucket internals, or filesystem paths.

## Testing

- Tests must verify upload, metadata persistence, authorization, tenant isolation, inventory isolation, download, deletion, and failure behavior.
- Tests must use fake blob storage adapters where appropriate.
- Integration tests must cover the production-style S3-compatible adapter before release.

## Open Questions

- Which attachment types are supported in the first release?
- Are attachments versioned?
- Are thumbnails generated immediately, lazily, or not at first?
- What virus scanning or content safety workflow is required before arbitrary file uploads?
