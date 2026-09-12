import type { Client } from 'openapi-fetch';
import type { components, paths } from './generated/schema';
import type { ClientTransport } from './clientTransport';

const expirationPath = '/tenants/{tenantId}/inventories/{inventoryId}/expiration-assets';
export type ExpirationWorkspaceOptions = NonNullable<paths[typeof expirationPath]['get']['parameters']['query']> & { signal?: AbortSignal };
export type ExpirationWorkspaceAsset = components['schemas']['ExpirationWorkspaceAsset'];
export type ExpirationWorkspacePage = Omit<components['schemas']['ExpirationWorkspaceData'], 'items'> & { items: ExpirationWorkspaceAsset[] } & { pagination: { limit: number; nextCursor: string | null; hasMore: boolean } };

export class ExpirationClient {
 constructor(private readonly client: Client<paths>, private readonly transport: ClientTransport) {}
 async list(tenantId: string, inventoryId: string, options: ExpirationWorkspaceOptions = {}): Promise<ExpirationWorkspacePage> {
  const { signal, ...query } = options;
  const response = await this.transport.unwrap(this.client.GET(expirationPath, { params: { path: { tenantId, inventoryId }, query }, headers: await this.transport.headers(), signal }));
  const pagination = response.meta.pagination;
  const counts = response.data.counts;
  if (!Array.isArray(response.data.items) || !pagination || !counts || ![counts.all,counts.soon,counts.expired].every(value => Number.isSafeInteger(value) && value >= 0) || counts.soon + counts.expired > counts.all || typeof pagination.hasMore !== 'boolean' || !Number.isSafeInteger(pagination.limit) || pagination.limit! < 1 || pagination.limit! > 100 || pagination.hasMore && (!pagination.nextCursor || pagination.nextCursor === options.cursor || response.data.items.length === 0)) throw new Error('Expiration results are incomplete. Try again.');
  return { ...response.data, items: response.data.items, pagination: { limit: pagination.limit!, nextCursor: pagination.nextCursor ?? null, hasMore: pagination.hasMore } };
 }
}
