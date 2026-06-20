# Media Attachments Spec

## Purpose

Stuff Stash needs first-class media and file attachments for inventory records.

Photos, receipts, manuals, and arbitrary files help users identify and manage physical things.

## Scope

This spec covers initial media and attachment requirements, the first asset attachment API slice, and the first production-shaped blob storage adapter.

This spec does not define multipart upload, direct-to-object-storage upload, image processing, thumbnails, virus scanning, retention policy, model vision features, or media export packaging.

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

## First Slice

The first implementation supports asset-scoped attachments.

The first upload protocol is JSON with base64 content:

- It is not the final high-performance upload protocol.
- It exists to prove the domain, authorization, metadata, storage port, and generated OpenAPI contract.
- Multipart upload or direct-to-storage upload must be specified before larger production media workflows.

Initial endpoints:

- `POST /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/content`

Initial upload request fields:

- `fileName`: user-facing file name.
- `contentType`: MIME type.
- `contentBase64`: base64-encoded file content.

Initial attachment response fields:

- ID.
- Tenant ID.
- Inventory ID.
- Asset ID.
- File name.
- Content type.
- Size in bytes.
- SHA-256 hash.
- Created timestamp.

The API must not expose storage keys, bucket names, filesystem paths, credentials, or provider internals.

Initial supported content types:

- `image/jpeg`
- `image/png`
- `image/webp`
- `application/pdf`

Initial upload size limit:

- 5 MiB decoded content per attachment.
- The limit must come from `STUFF_STASH_MAX_ATTACHMENT_BYTES`.
- The default is 5 MiB.
- The JSON upload route must set a larger request body cap derived from the attachment limit so base64 overhead does not reject valid uploads before application validation runs.
- Blob adapters that read from external storage must enforce the same maximum before buffering content in memory.

Attachment listing:

- Is scoped to one asset.
- Requires cursor pagination with `limit` and `cursor`.
- Must return the standard success envelope.

Attachment content download:

- Must authorize every request.
- Must return the bytes stored for the attachment.
- Must set the stored content type.
- Must not return presigned object storage URLs in the first slice.

Deletion:

- Archive, restore, and hard-delete behavior is defined by `specs/platform/resource-lifecycle.spec.md`.
- Hard delete must use the blob storage port and must not leave orphaned blob content when blob deletion fails.

## Storage Direction

- Production blob storage must use Garage through an S3-compatible storage adapter.
- Local development may use a filesystem adapter.
- Storage credentials, bucket names, endpoints, and region-style settings must come from environment-backed configuration.
- Blob storage dependencies and images must be pinned to reviewed versions.
- Blob storage must be accessed through a port.
- Application and domain code must not depend on S3, Garage, local filesystem paths, or provider SDK types.

The first storage adapters are:

- In-memory fake storage for tests.
- Local filesystem storage for local development.
- S3-compatible storage for Garage.

Blob storage mode:

- `STUFF_STASH_BLOB_STORAGE_MODE=filesystem` uses local filesystem storage.
- `STUFF_STASH_BLOB_STORAGE_MODE=s3` uses the S3-compatible adapter.
- Memory repository mode may still use the in-memory fake blob storage.

S3-compatible adapter configuration:

- `STUFF_STASH_S3_ENDPOINT`: endpoint host and port, without scheme.
- `STUFF_STASH_S3_ACCESS_KEY`: access key ID.
- `STUFF_STASH_S3_SECRET_KEY`: secret key.
- `STUFF_STASH_S3_BUCKET`: bucket name.
- `STUFF_STASH_S3_REGION`: region value. Defaults to `garage`.
- `STUFF_STASH_S3_SECURE`: whether to use HTTPS. Defaults to `true`.

The adapter must not create buckets at API startup. Buckets and credentials are deployment concerns.
Local plain-HTTP Garage verification must set `STUFF_STASH_S3_SECURE=false`.

## Security

- Uploads must enforce size limits.
- Uploads must enforce allowed MIME types.
- Uploads must inspect content signatures for supported types and reject content that does not match the claimed MIME type.
- Download URLs must not bypass authorization.
- Public buckets must not be used for private tenant inventory files.
- Error messages must not leak storage keys, credentials, bucket internals, or filesystem paths.
- Oversized or externally mutated blobs must not cause unbounded memory reads.
- Attachment upload requires `inventory.edit_asset`.
- Attachment list and content download require `inventory.view`.
- The target asset must belong to the tenant and inventory in the route.
- Hidden or cross-tenant assets must return safe not-found behavior after the caller passes inventory authorization.
- Viewers may list and download attachment content but must not upload.

## Audit And Observability

- Uploading an attachment must emit `attachment.created` audit history.
- Attachment detail, list, content download, archive, restore, and hard delete must emit safe audit history where specified by the lifecycle contract.
- Listing attachments must record domain observability through the injected observer.
- Downloading attachment content must record domain observability through the injected observer.
- Blob storage failures must be recorded through domain-oriented observability without leaking provider internals.

## Testing

- Tests must verify upload, metadata persistence, authorization, tenant isolation, inventory isolation, download, and failure behavior.
- Tests must verify unsupported MIME type rejection, oversize rejection, invalid base64 rejection, viewer upload denial, viewer download success, intruder denial, cross-tenant hiding, and safe storage errors.
- Tests must use fake blob storage adapters where appropriate.
- Integration tests must verify the production-style S3-compatible adapter against Garage running in Docker.

## Open Questions

- Are attachments versioned?
- Are thumbnails generated immediately, lazily, or not at first?
- What virus scanning or content safety workflow is required before arbitrary file uploads?
- What direct upload protocol should replace JSON base64 for larger production media?
