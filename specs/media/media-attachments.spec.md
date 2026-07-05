# Media Attachments Spec

## Purpose

Stuff Stash needs first-class media and file attachments for inventory records.

Photos, receipts, manuals, and arbitrary files help users identify and manage physical things.

## Scope

This spec covers media and attachment requirements, the asset attachment API slice, production-shaped blob storage, direct upload readiness, image derivative processing, thumbnails, and image preparation for future model adapters.

This spec does not define multipart upload, virus scanning, retention policy, model provider selection, model prompting, inference behavior, or media export packaging.

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
- Images may be prepared for model providers through media ports, but model provider calls, prompts, inference behavior, and provider-specific request formats are not part of the media domain.

## First Slice

The first implementation supports asset-scoped attachments.

The first upload protocol is JSON with base64 content:

- It is not the final high-performance upload protocol.
- It exists to prove the domain, authorization, metadata, storage port, and generated OpenAPI contract.
- Multipart upload is not part of the first implementation.
- Direct-to-object-storage upload uses an application direct upload flow and provider-specific adapter behind a port.

Initial endpoints:

- `POST /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads`
- `POST /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads/{uploadId}/complete`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/content`
- `GET /tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/thumbnail`

Initial upload request fields:

- `fileName`: user-facing file name.
- `contentType`: MIME type.
- `contentBase64`: base64-encoded file content.

Direct upload initiation request fields:

- `fileName`: user-facing file name.
- `contentType`: MIME type.
- `sizeBytes`: expected decoded content size.

Direct upload initiation response fields:

- `uploadId`: opaque upload identifier.
- `attachmentId`: attachment ID reserved for completion.
- `method`: HTTP method the client should use with the returned upload target.
- `url`: opaque adapter-provided upload target.
- `headers`: adapter-provided request headers the client must send.
- `formFields`: adapter-provided form fields the client must send when the upload method requires form upload.
- `expiresAt`: upload expiration timestamp.

Direct upload completion:

- The client calls the completion endpoint after successfully uploading content through the returned upload target.
- The application must authorize completion with `inventory.edit_asset`.
- The direct upload adapter must verify that the uploaded object exists, that its content type is expected, that the size is positive and within `STUFF_STASH_MAX_ATTACHMENT_BYTES`, and that a SHA-256 hash is available before metadata is committed.
- Attachment metadata and audit history are committed only after the direct upload adapter verifies completion.
- Provider URLs, object keys, bucket names, filesystem paths, credentials, and provider internals must not appear in completion responses.
- Expired, unknown, oversized, mismatched, or incomplete direct uploads must fail safely and must not create attachment metadata.

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

- 25 MiB decoded content per attachment.
- The limit must come from `STUFF_STASH_MAX_ATTACHMENT_BYTES`.
- The default is 25 MiB so original mobile camera and library photos can be attached without client-side compression in normal local and production configurations.
- The JSON upload route must set a larger request body cap derived from the attachment limit so base64 overhead does not reject valid uploads before application validation runs.
- Blob adapters that read from external storage must enforce the same maximum before buffering content in memory.
- Mobile clients should prefer direct upload targets for camera and library images so original selected file bytes are uploaded to blob storage instead of carried through the JSON attachment route. The JSON route remains a compatibility fallback for the explicit local-development sentinel `stuffstash-local://direct-uploads/` or clients that cannot use direct upload, but it must not become the preferred mobile media path.
- Mobile clients may request base64 fallback content from the native camera or photo picker at selection time so the local-development sentinel can still attach photos when platform photo-library URIs cannot be read later. This fallback content must not be compressed, resized, logged, sent to the language provider, or stored in the database; it exists only to call the authorized JSON attachment API when direct upload is not available.
- Mobile clients must surface safe upload-stage failure reasons when all selected photos fail, such as unsupported direct-upload targets, direct-upload transport failures, missing fallback content, or server attachment validation errors. The UI must not collapse every upload failure into an opaque generic message when the mobile application layer has a safe error message.
- Mobile clients must require HTTPS for production direct-upload targets. For local S3/Garage development where the API is intentionally configured with `STUFF_STASH_S3_SECURE=false`, mobile may upload to HTTP targets only when the target host is loopback, localhost, `.local`, or an RFC1918/private-network address. Public cleartext HTTP direct-upload targets must still be rejected.
- Attachment upload validation errors must be safe but specific enough for clients to guide the user. Oversized uploads must report that the file is too large, unsupported media types must report unsupported file type, content-signature mismatches must report that the file content does not match the declared type, empty fallback content must report that attachment content is empty, and unreadable base64 fallback content must report that attachment content could not be read.
- Mobile API clients must preserve safe server validation details when a transport adapter returns a generic invalid-request envelope with client-safe detail messages, so users and developers can see which upload field or constraint failed.

Attachment listing:

- Is scoped to one asset.
- Requires cursor pagination with `limit` and `cursor`.
- Must return the standard success envelope.

Attachment content download:

- Must authorize every request.
- Must return the bytes stored for the attachment.
- Must set the stored content type.
- Must not return presigned object storage URLs in the first slice.

Thumbnails and image processing:

- Thumbnails are generated through an image processing port.
- Thumbnail variants are `small`, `medium`, and `large`.
- `small` is optimized for dense cards and list rows.
- `medium` is optimized for asset detail hero presentation.
- `large` is optimized for focused full-screen mobile preview without requiring the original attachment bytes.
- Thumbnail variants must preserve aspect ratio and use a quality-oriented resize algorithm rather than nearest-neighbor scaling.
- Generated thumbnails should prefer compact, broadly supported image encodings suitable for photos, such as bounded-quality JPEG, so camera-photo derivatives remain small enough for mobile card and list browsing.
- Image processors must inspect image dimensions before full decode and reject excessive dimensions or pixel counts before resizing, so small-byte images with hostile dimensions cannot force unbounded memory or CPU work.
- Thumbnail generation must be authorized with `inventory.view` and scoped to the requested tenant, inventory, asset, and attachment.
- Thumbnail generation is supported only for image attachments.
- The application must not contain image codec, resize, crop, EXIF, or color-management implementation details.
- Image processing adapters must enforce bounded input and output sizes and must not leak implementation paths, command lines, or provider internals in user-facing errors.
- Thumbnail derivatives must be cached behind the blob storage port with deterministic keys per attachment and variant so dense mobile surfaces such as Home and Browse do not repeatedly read, decode, and resize original camera images.
- Cached thumbnail derivatives are rebuildable artifacts, not attachment metadata. They must not be returned as storage-provider URLs and must remain authorized through the thumbnail endpoint on every request.
- Hard-deleting an attachment must clean up the original blob and the known thumbnail derivative keys so derivative caching does not create unbounded orphaned storage.
- Thumbnail responses must return bytes and the derivative content type; they must not expose derivative storage keys.

Asset summary media:

- API asset summary responses used by dense browse surfaces, including asset list and asset search results, must include a bounded first active image attachment summary when one exists.
- The first image summary must expose only safe attachment metadata and authorized thumbnail endpoint references for `small`, `medium`, and `large`; it must not expose storage keys, bucket names, presigned provider URLs, filesystem paths, credentials, or original attachment bytes.
- The first image summary must be assembled server-side using repository-level batch lookup for the returned assets. Clients must not need to issue one attachment-list request per asset to render Home or Browse cards.
- After an authorized asset list or search builds primary image summaries, the API should best-effort warm a bounded number of `small` thumbnail derivatives for those primary images so first-visible Home and Browse cards do not each pay lazy thumbnail generation latency. Warm scheduling must not block the list or search response on image resizing, warm work must be concurrency bounded, and warm generation must share the same in-flight derivative coordination used by direct thumbnail downloads so immediate card image requests do not duplicate resize work. Warm limit, concurrency, and timeout must be configurable through environment-backed runtime config. Warm failures must be observed but must not fail the list or search response.
- Mobile Home and Browse must prefer the asset summary image references for cards and only fall back to attachment listing when a detail/photo-management surface needs the complete attachment set or when the API summary does not include media.

Model image readiness:

- Media may prepare image attachments for future model providers through an image preparation port.
- Model-image preparation must require `inventory.view`, must use the same tenant/inventory/asset/attachment scoping as downloads, and must support only image attachments.
- Prepared image data must include only the bounded bytes and media metadata needed by a future model adapter, such as content type, byte size, hash, and optional dimensions.
- The media application layer must not call model providers, construct provider prompts, or bypass authorization, validation, audit, tenant isolation, inventory isolation, or attachment lifecycle rules.
- Model provider adapters must be specified separately before any provider is implemented.

Deletion:

- Archive, restore, and hard-delete behavior is defined by `specs/platform/resource-lifecycle.spec.md`.
- Hard delete must preserve audit history and remove attachment metadata through an attachment command/unit-of-work port.
- Hard delete must enqueue durable blob-deletion intent in the same database transaction that removes attachment metadata.
- Blob deletion must be processed from an outbox after metadata deletion commits.
- Blob deletion must be retryable and idempotent. Missing blobs should be treated as successful cleanup.
- The API must not delete blob content inline before metadata removal commits, because that can create broken attachment records if metadata deletion fails after blob deletion succeeds.
- Until asynchronous workers are separated from the API process, the API may drain blob-deletion events opportunistically after successful hard-delete requests, but the durable outbox is the consistency mechanism.
- The API runtime must also drain blob-deletion events on an environment-configured interval so failed opportunistic cleanup does not wait for a future attachment delete.
- Blob-deletion outbox processing must emit domain-oriented observability when events are claimed, processed, failed, and dead-lettered.
- Blob-deletion events must move to a dead-letter terminal state once they reach the environment-configured maximum attempt count.

## Storage Direction

- Production blob storage must use Garage through an S3-compatible storage adapter.
- Local development may use a filesystem adapter.
- Storage credentials, bucket names, endpoints, and region-style settings must come from environment-backed configuration.
- Blob storage dependencies and images must be pinned to reviewed versions.
- Blob storage must be accessed through a port.
- Application and domain code must not depend on S3, Garage, local filesystem paths, or provider SDK types.
- Blob cleanup must be accessed through a blob-deletion outbox port, not by coupling attachment metadata persistence to a specific storage adapter.
- Blob-deletion outbox batch size, retry interval, claim lease duration, and maximum attempt count must come from environment-backed configuration.

The first storage adapters are:

- In-memory fake storage for tests.
- Local filesystem storage for local development.
- S3-compatible storage for Garage.

Blob storage mode:

- `STUFF_STASH_BLOB_STORAGE_MODE=filesystem` uses local filesystem storage.
- `STUFF_STASH_BLOB_STORAGE_MODE=s3` uses the S3-compatible adapter.
- Memory repository mode may still use the in-memory fake blob storage and must wire the local direct-upload sentinel adapter so browser and mobile clients can fall back to the authorized JSON attachment route in local tracer-bullet runs.

S3-compatible adapter configuration:

- `STUFF_STASH_S3_ENDPOINT`: endpoint host and port, without scheme.
- `STUFF_STASH_S3_PUBLIC_ENDPOINT`: browser-reachable endpoint host and port, without scheme, used for presigned direct-upload targets when it differs from the API's internal S3 endpoint. Defaults to `STUFF_STASH_S3_ENDPOINT`.
- `STUFF_STASH_S3_ACCESS_KEY`: access key ID.
- `STUFF_STASH_S3_SECRET_KEY`: secret key.
- `STUFF_STASH_S3_BUCKET`: bucket name.
- `STUFF_STASH_S3_REGION`: region value. Defaults to `garage`.
- `STUFF_STASH_S3_SECURE`: whether to use HTTPS. Defaults to `true`.

The adapter must not create buckets at API startup. Buckets and credentials are deployment concerns.
Local plain-HTTP Garage verification must set `STUFF_STASH_S3_SECURE=false`.
When a browser client cannot use an advertised direct-upload target because the target is not browser-fetchable or the local object-storage topology rejects the upload, the web client may fall back to the authorized JSON attachment upload route. The fallback must still use the API authorization boundary and must not bypass attachment size, type, tenancy, audit, or storage validation.

## Security

- Uploads must enforce size limits.
- Uploads must enforce allowed MIME types.
- Uploads must inspect content signatures for supported types and reject content that does not match the claimed MIME type.
- Image uploads must also decode as the claimed image format before attachment metadata is committed, so thumbnail and search-card image paths do not store broken image attachments that later fail at read time.
- Download URLs must not bypass authorization.
- Direct upload targets must be opaque, bounded, expiring, and usable only for the intended tenant, inventory, asset, attachment ID, content type, and size.
- Direct upload completion must verify uploaded object metadata before committing attachment metadata.
- Garage/S3 direct upload targets must use presigned object-storage POST policies behind the direct upload port.
- S3-compatible direct upload initiation must include policy form fields that constrain object key, content type, content length, and expiration.
- S3-compatible direct upload completion must not rely on process-local pending state. Completion state must be durable or encoded in a signed opaque upload token so restarts and multi-replica routing do not break completion.
- Local filesystem and memory runtime modes may advertise only the explicit non-uploadable local-development sentinel `stuffstash-local://direct-uploads/{uploadId}` until a real local HTTP upload target adapter is implemented. Browser and mobile clients must treat this sentinel as a signal to use the JSON attachment route; they must not treat arbitrary non-HTTP schemes as safe fallback targets.
- S3-compatible direct-upload adapters may advertise HTTP upload targets only when the deployment is explicitly configured for local/plain-HTTP development. Clients must treat these as development targets and should bound support to loopback, localhost, `.local`, or private-network hosts.
- Public buckets must not be used for private tenant inventory files.
- Error messages must not leak storage keys, credentials, bucket internals, or filesystem paths.
- Oversized or externally mutated blobs must not cause unbounded memory reads.
- Attachment upload requires `inventory.edit_asset`.
- Attachment direct upload initiation and completion require `inventory.edit_asset`.
- Attachment list and content download require `inventory.view`.
- Attachment thumbnail and model-image preparation require `inventory.view`.
- The target asset must belong to the tenant and inventory in the route.
- Hidden or cross-tenant assets must return safe not-found behavior after the caller passes inventory authorization.
- Viewers may list and download attachment content but must not upload.

## Audit And Observability

- Uploading an attachment must emit `attachment.created` audit history.
- Attachment detail, list, content download, archive, restore, and hard delete must emit safe audit history where specified by the lifecycle contract.
- Listing attachments must record domain observability through the injected observer.
- Downloading attachment content must record domain observability through the injected observer.
- Creating and completing direct uploads must record domain observability through the injected observer.
- Generating thumbnails and preparing model images must record domain observability through the injected observer.
- Blob storage failures must be recorded through domain-oriented observability without leaking provider internals.
- Blob deletion outbox events must be recorded through domain-oriented observability without leaking storage keys.

## Testing

- Tests must verify upload, direct upload initiation and completion, metadata persistence, authorization, tenant isolation, inventory isolation, download, thumbnail generation, model-image preparation, and failure behavior.
- Tests must verify unsupported MIME type rejection, oversize rejection, invalid base64 rejection, viewer upload denial, viewer download success, intruder denial, cross-tenant hiding, and safe storage errors.
- Tests must verify direct upload viewer denial, direct upload completion failure safety, thumbnail image-only behavior, and model-image preparation through a fake image processing adapter.
- HTTP boundary tests must include at least one real decodable image upload that is downloaded and processed through the real image processor for thumbnail generation.
- Real-image HTTP boundary tests must verify authorization for image download and thumbnail access, including authorized viewer access and unauthorized principal denial.
- Tests must use fake blob storage adapters where appropriate.
- Integration tests must verify the production-style S3-compatible adapter against Garage running in Docker.

## Open Questions

- Are attachments versioned?
- Are thumbnails persisted durably, cached opportunistically, or always generated lazily?
- What virus scanning or content safety workflow is required before arbitrary file uploads?
