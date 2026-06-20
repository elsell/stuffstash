import { describe, expect, it } from 'vitest';
import { StuffStashAPIError, StuffStashClient } from './stuffStashClient';

describe('StuffStashClient', () => {
  it('sends bearer tokens and unwraps response envelopes', async () => {
    const requests: Request[] = [];
    const client = new StuffStashClient({
      baseUrl: 'http://api.local/',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        return Response.json({
          data: {
            id: 'inventory-one',
            tenantId: 'tenant-one',
            name: 'Garage'
          },
          meta: {}
        });
      }
    });

    const inventory = await client.createInventory('tenant-one', 'Garage');

    expect(inventory.name).toBe('Garage');
    expect(requests[0]?.url).toBe('http://api.local/tenants/tenant-one/inventories');
    expect(requests[0]?.headers.get('Authorization')).toBe('Bearer id-token');
    expect(await requests[0]?.json()).toEqual({ name: 'Garage' });
  });

  it('maps paginated list envelopes', async () => {
    const requests: Request[] = [];
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        return Response.json({
          data: [
            {
              id: 'asset-one',
              tenantId: 'tenant-one',
              inventoryId: 'inventory-one',
              kind: 'item',
              title: 'Fertilizer',
              description: '',
              lifecycleState: 'active',
              customFields: {}
            }
          ],
          meta: {
            pagination: {
              limit: 1,
              nextCursor: 'next-page',
              hasMore: true
            }
          }
        });
      }
    });

    const page = await client.listAssets('tenant-one', 'inventory-one', 1, undefined, 'archived');

    expect(page.items).toHaveLength(1);
    expect(page.items[0]?.title).toBe('Fertilizer');
    expect(page.pagination).toEqual({ limit: 1, nextCursor: 'next-page', hasMore: true });
    expect(requests[0]?.url).toBe('http://api.local/tenants/tenant-one/inventories/inventory-one/assets?limit=1&lifecycleState=archived');
  });

  it('calls asset lifecycle endpoints', async () => {
    const requests: Request[] = [];
    const assetResponse = {
      data: {
        id: 'asset-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        kind: 'item',
        title: 'Fertilizer',
        description: '',
        lifecycleState: 'archived',
        customFields: {}
      },
      meta: {}
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        if (request.method === 'DELETE') {
          return new Response(null, { status: 204 });
        }
        return Response.json(assetResponse);
      }
    });

    await expect(client.archiveAsset('tenant-one', 'inventory-one', 'asset-one')).resolves.toMatchObject({
      id: 'asset-one',
      lifecycleState: 'archived'
    });
    await expect(client.restoreAsset('tenant-one', 'inventory-one', 'asset-one')).resolves.toMatchObject({
      id: 'asset-one'
    });
    await expect(client.deleteAsset('tenant-one', 'inventory-one', 'asset-one')).resolves.toBeUndefined();

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'PATCH http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/archive',
      'PATCH http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/restore',
      'DELETE http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one'
    ]);
  });

  it('maps API errors into typed client errors', async () => {
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => null,
      fetch: async () =>
        Response.json(
          {
            error: {
              code: 'authentication_required',
              message: 'Authentication required.'
            }
          },
          { status: 401 }
        )
    });

    await expect(client.me()).rejects.toMatchObject({
      status: 401,
      code: 'authentication_required',
      message: 'Authentication required.'
    });
  });
});
