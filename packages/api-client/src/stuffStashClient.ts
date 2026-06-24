import createClient, { type Client } from 'openapi-fetch';
import type { components, paths } from './generated/schema';

export type OpenAPIPaths = paths;

export type TokenProvider = () => string | null | Promise<string | null>;

export interface StuffStashClientOptions {
  baseUrl: string;
  tokenProvider: TokenProvider;
  fetch?: typeof fetch;
}

export interface Principal {
  id: string;
  email?: string;
}

export interface AccessSummary {
  relationship: string;
  permissions: string[];
}

export interface Tenant {
  id: string;
  name: string;
  access: AccessSummary;
}

export interface Inventory {
  id: string;
  tenantId: string;
  name: string;
  access: AccessSummary;
}

export type AssetKind = 'item' | 'container' | 'location';
export type AssetLifecycleState = 'active' | 'archived';
export type AssetLifecycleFilter = AssetLifecycleState | 'all';

export interface Asset {
  id: string;
  tenantId: string;
  inventoryId: string;
  kind: AssetKind;
  title: string;
  description: string;
  parentAssetId: string | null;
  lifecycleState: AssetLifecycleState;
}

export interface Attachment {
  id: string;
  tenantId: string;
  inventoryId: string;
  assetId: string;
  fileName: string;
  contentType: string;
  sizeBytes: number;
  lifecycleState: AssetLifecycleState;
}

export interface CreateAttachmentInput {
  fileName: string;
  contentType: 'image/jpeg' | 'image/png' | 'image/webp' | 'application/pdf';
  contentBase64: string;
}

export interface InitiateDirectUploadInput {
  fileName: string;
  contentType: 'image/jpeg' | 'image/png' | 'image/webp' | 'application/pdf';
  sizeBytes: number;
}

export interface DirectUpload {
  uploadId: string;
  attachmentId: string;
  method: string;
  url: string;
  headers: Record<string, string>;
  formFields: Record<string, string>;
  expiresAt: string;
}

export interface AssetPhotoReference {
  uri: string;
  headers: Record<string, string>;
}

export interface CreateAssetInput {
  kind: AssetKind;
  title: string;
  description?: string;
  parentAssetId?: string | null;
}

export interface UpdateAssetInput {
  title: string;
  description?: string;
  parentAssetId?: string | null;
}

export interface AssetSearchResult {
  type: 'asset';
  tenantId: string;
  inventory: {
    id: string;
    name: string;
  };
  asset: Asset;
  matches: Array<{
    field: string;
    value: string;
  }>;
}

export interface Pagination {
  limit: number;
  nextCursor: string | null;
  hasMore: boolean;
}

export interface Page<T> {
  items: T[];
  pagination: Pagination;
}

type ErrorEnvelope = components['schemas']['ErrorEnvelope'];
type Meta = components['schemas']['Meta'];
type PrincipalResponse = components['schemas']['PrincipalResponse'];
type TenantResponse = components['schemas']['TenantResponse'];
type InventoryResponse = components['schemas']['InventoryResponse'];
type AssetResponse = components['schemas']['AssetResponse'];
type AttachmentResponse = components['schemas']['AttachmentResponse'];
type DirectUploadResponse = components['schemas']['DirectUploadResponse'];

interface SuccessEnvelope<T> {
  data: T;
  meta: Meta;
}

export class StuffStashAPIError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'StuffStashAPIError';
    this.status = status;
    this.code = code;
  }
}

export class StuffStashClient {
  private readonly client: Client<paths>;
  private readonly baseUrl: string;
  private readonly tokenProvider: TokenProvider;

  constructor(options: StuffStashClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/+$/, '');
    this.tokenProvider = options.tokenProvider;
    this.client = createClient<paths>({
      baseUrl: this.baseUrl,
      fetch: options.fetch
    });
  }

  async me(): Promise<Principal> {
    const envelope = await this.unwrap(
      this.client.GET('/me', {
        headers: await this.authHeaders()
      })
    );
    return mapPrincipal(envelope.data);
  }

  async createTenant(name: string): Promise<Tenant> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants', {
        headers: await this.authHeaders(),
        body: { name }
      })
    );
    return mapTenant(envelope.data);
  }

  async listMyTenants(limit = 50, cursor?: string): Promise<Page<Tenant>> {
    const envelope = await this.unwrap(
      this.client.GET('/me/tenants', {
        headers: await this.authHeaders(),
        params: {
          query: { limit, cursor }
        }
      })
    );
    return mapPage(envelope, mapTenant);
  }

  async getTenant(tenantId: string): Promise<Tenant> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId } }
      })
    );
    return mapTenant(envelope.data);
  }

  async listInventories(tenantId: string, limit = 50, cursor?: string): Promise<Page<Inventory>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId },
          query: { limit, cursor }
        }
      })
    );
    return mapPage(envelope, mapInventory);
  }

  async createInventory(tenantId: string, name: string): Promise<Inventory> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories', {
        headers: await this.authHeaders(),
        params: { path: { tenantId } },
        body: { name }
      })
    );
    return mapInventory(envelope.data);
  }

  async listAssets(
    tenantId: string,
    inventoryId: string,
    limit = 50,
    cursor?: string,
    lifecycleState: AssetLifecycleFilter = 'active'
  ): Promise<Page<Asset>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/assets', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId, inventoryId },
          query: { limit, cursor, lifecycleState }
        }
      })
    );
    return mapPage(envelope, mapAsset);
  }

  async createAsset(tenantId: string, inventoryId: string, input: CreateAssetInput): Promise<Asset> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/assets', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId } },
        body: { ...input, parentAssetId: input.parentAssetId ?? undefined }
      })
    );
    return mapAsset(envelope.data);
  }

  async getAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } }
      })
    );
    return mapAsset(envelope.data);
  }

  async updateAsset(tenantId: string, inventoryId: string, assetId: string, input: UpdateAssetInput): Promise<Asset> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } },
        body: input
      })
    );
    return mapAsset(envelope.data);
  }

  async searchAssets(tenantId: string, query: string, limit = 20, cursor?: string): Promise<Page<AssetSearchResult>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/search/assets', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId },
          query: { q: query, limit, cursor, lifecycleState: 'active', mode: 'fuzzy' }
        }
      })
    );
    return mapPage(envelope, mapAssetSearchResult);
  }

  async archiveAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/archive', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } }
      })
    );
    return mapAsset(envelope.data);
  }

  async restoreAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/restore', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } }
      })
    );
    return mapAsset(envelope.data);
  }

  async deleteAsset(tenantId: string, inventoryId: string, assetId: string): Promise<void> {
    await this.unwrapNoContent(
      this.client.DELETE('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } }
      })
    );
  }

  async listAssetAttachments(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    limit = 10,
    cursor?: string
  ): Promise<Page<Attachment>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId, inventoryId, assetId },
          query: { limit, cursor }
        }
      })
    );
    return mapPage(envelope, mapAttachment);
  }

  async createAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    input: CreateAttachmentInput
  ): Promise<Attachment> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } },
        body: input
      })
    );
    return mapAttachment(envelope.data);
  }

  async initiateAssetAttachmentDirectUpload(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    input: InitiateDirectUploadInput
  ): Promise<DirectUpload> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId } },
        body: input
      })
    );
    return mapDirectUpload(envelope.data);
  }

  async completeAssetAttachmentDirectUpload(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    uploadId: string
  ): Promise<Attachment> {
    const envelope = await this.unwrap(
      this.client.POST('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/direct-uploads/{uploadId}/complete', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId, uploadId } }
      })
    );
    return mapAttachment(envelope.data);
  }

  async archiveAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<Attachment> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/archive', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId, attachmentId } }
      })
    );
    return mapAttachment(envelope.data);
  }

  async restoreAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<Attachment> {
    const envelope = await this.unwrap(
      this.client.PATCH('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}/restore', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId, attachmentId } }
      })
    );
    return mapAttachment(envelope.data);
  }

  async deleteAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<void> {
    await this.unwrapNoContent(
      this.client.DELETE('/tenants/{tenantId}/inventories/{inventoryId}/assets/{assetId}/attachments/{attachmentId}', {
        headers: await this.authHeaders(),
        params: { path: { tenantId, inventoryId, assetId, attachmentId } }
      })
    );
  }

  async assetAttachmentThumbnailReference(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<AssetPhotoReference> {
    return {
      uri: [
        this.baseUrl,
        'tenants',
        encodeURIComponent(tenantId),
        'inventories',
        encodeURIComponent(inventoryId),
        'assets',
        encodeURIComponent(assetId),
        'attachments',
        encodeURIComponent(attachmentId),
        'thumbnail?variant=small'
      ].join('/'),
      headers: await this.authHeaders()
    };
  }

  private async authHeaders(): Promise<Record<string, string>> {
    const headers: Record<string, string> = {};
    const token = await this.tokenProvider();
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    return headers;
  }

  private async unwrap<T>(request: Promise<{ data?: T; error?: ErrorEnvelope; response: Response }>): Promise<T> {
    const { data, error, response } = await request;
    if (!response.ok) {
      throw new StuffStashAPIError(
        response.status,
        error?.error?.code ?? 'request_failed',
        error?.error?.message ?? 'Request failed.'
      );
    }
    if (data === undefined) {
      throw new StuffStashAPIError(response.status, 'invalid_response', 'Invalid API response.');
    }
    return data;
  }

  private async unwrapNoContent(request: Promise<{ error?: ErrorEnvelope; response: Response }>): Promise<void> {
    const { error, response } = await request;
    if (!response.ok) {
      throw new StuffStashAPIError(
        response.status,
        error?.error?.code ?? 'request_failed',
        error?.error?.message ?? 'Request failed.'
      );
    }
  }
}

function mapAttachment(response: AttachmentResponse): Attachment {
  return {
    id: response.id,
    tenantId: response.tenantId,
    inventoryId: response.inventoryId,
    assetId: response.assetId,
    fileName: response.fileName,
    contentType: response.contentType,
    sizeBytes: response.sizeBytes,
    lifecycleState: mapAssetLifecycleState(response.lifecycleState)
  };
}

function mapDirectUpload(response: DirectUploadResponse): DirectUpload {
  return {
    uploadId: response.uploadId,
    attachmentId: response.attachmentId,
    method: response.method,
    url: response.url,
    headers: response.headers ?? {},
    formFields: response.formFields ?? {},
    expiresAt: response.expiresAt
  };
}

function mapPrincipal(response: PrincipalResponse): Principal {
  return { id: response.id, email: response.email };
}

function mapTenant(response: TenantResponse): Tenant {
  return { id: response.id, name: response.name, access: mapAccess(response.access) };
}

function mapInventory(response: InventoryResponse): Inventory {
  return {
    id: response.id,
    tenantId: response.tenantId,
    name: response.name,
    access: mapAccess(response.access)
  };
}

function mapAccess(response: components['schemas']['AccessResponse']): AccessSummary {
  return {
    relationship: response.relationship,
    permissions: response.permissions ?? []
  };
}

function mapAsset(response: AssetResponse): Asset {
  return {
    id: response.id,
    tenantId: response.tenantId,
    inventoryId: response.inventoryId,
    kind: mapAssetKind(response.kind),
    title: response.title,
    description: response.description,
    parentAssetId: response.parentAssetId ?? null,
    lifecycleState: mapAssetLifecycleState(response.lifecycleState)
  };
}

function mapAssetSearchResult(response: components['schemas']['AssetSearchResultResponse']): AssetSearchResult {
  return {
    type: 'asset',
    tenantId: response.tenantId,
    inventory: response.inventory,
    asset: {
      id: response.asset.id,
      tenantId: response.tenantId,
      inventoryId: response.asset.inventoryId,
      kind: mapAssetKind(response.asset.kind),
      title: response.asset.title,
      description: response.asset.description,
      parentAssetId: response.asset.parentAssetId ?? null,
      lifecycleState: mapAssetLifecycleState(response.asset.lifecycleState)
    },
    matches: response.matches ?? []
  };
}

function mapAssetKind(kind: string): AssetKind {
  switch (kind) {
    case 'item':
    case 'container':
    case 'location':
      return kind;
    default:
      throw new StuffStashAPIError(200, 'invalid_asset_kind', 'Invalid asset kind.');
  }
}

function mapAssetLifecycleState(lifecycleState: string): AssetLifecycleState {
  switch (lifecycleState) {
    case 'active':
    case 'archived':
      return lifecycleState;
    default:
      throw new StuffStashAPIError(200, 'invalid_asset_lifecycle_state', 'Invalid asset lifecycle state.');
  }
}

function mapPage<TResponse, TItem>(
  envelope: SuccessEnvelope<TResponse[] | null>,
  mapper: (response: TResponse) => TItem
): Page<TItem> {
  return {
    items: (envelope.data ?? []).map(mapper),
    pagination: {
      limit: envelope.meta.pagination?.limit ?? 0,
      nextCursor: envelope.meta.pagination?.nextCursor ?? null,
      hasMore: envelope.meta.pagination?.hasMore ?? false
    }
  };
}
