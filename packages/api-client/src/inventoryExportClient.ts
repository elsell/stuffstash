import { createAuthenticatedTransport } from './authenticatedTransport';
import { StuffStashAPIError, type StuffStashClientOptions } from './stuffStashClient';

export type InventoryExportFormat = 'json' | 'csv';

/** File bytes remain text and are never stored in the client's query caches. */
export class InventoryExportClient {
  private readonly transport;
  constructor(options: StuffStashClientOptions) {
    this.transport = createAuthenticatedTransport(options);
  }

  async download(tenantId: string, inventoryId: string, format: InventoryExportFormat, signal?: AbortSignal): Promise<string> {
    const { data, error, response } = await this.transport.GET('/tenants/{tenantId}/inventories/{inventoryId}/export', {
      params: { path: { tenantId, inventoryId }, query: { format } }, parseAs: 'text', signal
    });
    if (!response.ok || error) {
      throw new StuffStashAPIError(response.status, error?.error.code ?? 'export_failed', error?.error.message ?? 'Could not export this inventory.');
    }
    if (signal?.aborted) throw signal.reason ?? Object.assign(new Error('Export cancelled'), { name: 'AbortError' });
    if (typeof data !== 'string') throw new Error('The server did not return an inventory file.');
    return data;
  }
}
