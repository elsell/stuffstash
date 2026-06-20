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

export interface Tenant {
  id: string;
  name: string;
}

export interface Inventory {
  id: string;
  tenantId: string;
  name: string;
}

export type AssetKind = 'item' | 'container' | 'location';

export interface Asset {
  id: string;
  tenantId: string;
  inventoryId: string;
  kind: AssetKind;
  title: string;
  description: string;
  lifecycleState: string;
}

export interface CreateAssetInput {
  kind: AssetKind;
  title: string;
  description?: string;
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
  private readonly tokenProvider: TokenProvider;

  constructor(options: StuffStashClientOptions) {
    this.tokenProvider = options.tokenProvider;
    this.client = createClient<paths>({
      baseUrl: options.baseUrl.replace(/\/+$/, ''),
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

  async listAssets(tenantId: string, inventoryId: string, limit = 50, cursor?: string): Promise<Page<Asset>> {
    const envelope = await this.unwrap(
      this.client.GET('/tenants/{tenantId}/inventories/{inventoryId}/assets', {
        headers: await this.authHeaders(),
        params: {
          path: { tenantId, inventoryId },
          query: { limit, cursor }
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
        body: input
      })
    );
    return mapAsset(envelope.data);
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
}

function mapPrincipal(response: PrincipalResponse): Principal {
  return { id: response.id, email: response.email };
}

function mapTenant(response: TenantResponse): Tenant {
  return { id: response.id, name: response.name };
}

function mapInventory(response: InventoryResponse): Inventory {
  return { id: response.id, tenantId: response.tenantId, name: response.name };
}

function mapAsset(response: AssetResponse): Asset {
  return {
    id: response.id,
    tenantId: response.tenantId,
    inventoryId: response.inventoryId,
    kind: mapAssetKind(response.kind),
    title: response.title,
    description: response.description,
    lifecycleState: response.lifecycleState
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
