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
            name: 'Garage',
            access: {
              relationship: 'editor',
              permissions: ['view', 'create_asset', 'edit_asset']
            }
          },
          meta: {}
        });
      }
    });

    const inventory = await client.createInventory('tenant-one', 'Garage');

    expect(inventory.name).toBe('Garage');
    expect(inventory.access).toEqual({
      relationship: 'editor',
      permissions: ['view', 'create_asset', 'edit_asset']
    });
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

  it('fetches tenants by ID', async () => {
    const requests: Request[] = [];
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        return Response.json({
          data: {
            id: 'tenant-one',
            name: 'Home',
            access: {
              relationship: 'owner',
              permissions: ['view', 'create_inventory', 'configure']
            }
          },
          meta: {}
        });
      }
    });

    await expect(client.getTenant('tenant-one')).resolves.toEqual({
      id: 'tenant-one',
      name: 'Home',
      access: {
        relationship: 'owner',
        permissions: ['view', 'create_inventory', 'configure']
      }
    });
    expect(requests[0]?.url).toBe('http://api.local/tenants/tenant-one');
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

  it('creates and lists asset attachments through generated paths', async () => {
    const requests: Request[] = [];
    const attachmentEnvelope = {
      data: {
        id: 'attachment-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        assetId: 'asset-one',
        fileName: 'photo.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 12,
        sha256: 'hash',
        createdAt: '2026-06-23T00:00:00Z',
        lifecycleState: 'active'
      },
      meta: {}
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        if (request.method === 'GET') {
          return Response.json({
            data: [attachmentEnvelope.data],
            meta: {
              pagination: {
                limit: 1,
                nextCursor: null,
                hasMore: false
              }
            }
          });
        }
        return Response.json(attachmentEnvelope);
      }
    });

    await expect(
      client.createAssetAttachment('tenant-one', 'inventory-one', 'asset-one', {
        fileName: 'photo.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'ZmFrZQ=='
      })
    ).resolves.toMatchObject({
      id: 'attachment-one',
      fileName: 'photo.jpg',
      contentType: 'image/jpeg'
    });
    await expect(
      client.listAssetAttachments('tenant-one', 'inventory-one', 'asset-one', 1)
    ).resolves.toMatchObject({
      items: [{ id: 'attachment-one', lifecycleState: 'active' }],
      pagination: { limit: 1, nextCursor: null, hasMore: false }
    });

    expect(requests[0]?.url).toBe('http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/attachments');
    expect(requests[0]?.headers.get('Authorization')).toBe('Bearer id-token');
    expect(await requests[0]?.json()).toEqual({
      fileName: 'photo.jpg',
      contentType: 'image/jpeg',
      contentBase64: 'ZmFrZQ=='
    });
    expect(requests[1]?.url).toBe('http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/attachments?limit=1');
  });

  it('builds authenticated thumbnail references', async () => {
    const client = new StuffStashClient({
      baseUrl: 'http://api.local/',
      tokenProvider: () => 'id-token',
      fetch: async () => Response.json({ data: {}, meta: {} })
    });

    await expect(
      client.assetAttachmentThumbnailReference(
        'tenant one',
        'inventory/one',
        'asset one',
        'attachment/one'
      )
    ).resolves.toEqual({
      uri: 'http://api.local/tenants/tenant%20one/inventories/inventory%2Fone/assets/asset%20one/attachments/attachment%2Fone/thumbnail?variant=small',
      headers: { Authorization: 'Bearer id-token' }
    });
  });

  it('initiates and completes direct attachment uploads through generated paths', async () => {
    const requests: Request[] = [];
    const attachmentEnvelope = {
      data: {
        id: 'attachment-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        assetId: 'asset-one',
        fileName: 'photo.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 12,
        sha256: 'hash',
        createdAt: '2026-06-23T00:00:00Z',
        lifecycleState: 'active'
      },
      meta: {}
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        if (request.url.endsWith('/direct-uploads')) {
          return Response.json({
            data: {
              uploadId: 'upload-one',
              attachmentId: 'attachment-one',
              method: 'PUT',
              url: 'https://uploads.local/object-one',
              headers: { 'Content-Type': 'image/jpeg' },
              formFields: {},
              expiresAt: '2026-06-23T00:15:00Z'
            },
            meta: {}
          });
        }
        return Response.json(attachmentEnvelope, { status: 201 });
      }
    });

    await expect(
      client.initiateAssetAttachmentDirectUpload('tenant-one', 'inventory-one', 'asset-one', {
        fileName: 'photo.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 12
      })
    ).resolves.toMatchObject({
      uploadId: 'upload-one',
      attachmentId: 'attachment-one',
      method: 'PUT'
    });
    await expect(
      client.completeAssetAttachmentDirectUpload('tenant-one', 'inventory-one', 'asset-one', 'upload-one')
    ).resolves.toMatchObject({ id: 'attachment-one', fileName: 'photo.jpg' });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/attachments/direct-uploads',
      'POST http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/attachments/direct-uploads/upload-one/complete'
    ]);
    expect(await requests[0]?.json()).toEqual({
      fileName: 'photo.jpg',
      contentType: 'image/jpeg',
      sizeBytes: 12
    });
  });

  it('calls attachment lifecycle endpoints through generated paths', async () => {
    const requests: Request[] = [];
    const attachmentEnvelope = {
      data: {
        id: 'attachment-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        assetId: 'asset-one',
        fileName: 'photo.jpg',
        contentType: 'image/jpeg',
        sizeBytes: 12,
        sha256: 'hash',
        createdAt: '2026-06-23T00:00:00Z',
        lifecycleState: 'archived'
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
        return Response.json(attachmentEnvelope);
      }
    });

    await expect(
      client.archiveAssetAttachment('tenant-one', 'inventory-one', 'asset-one', 'attachment-one')
    ).resolves.toMatchObject({ id: 'attachment-one', lifecycleState: 'archived' });
    await expect(
      client.restoreAssetAttachment('tenant-one', 'inventory-one', 'asset-one', 'attachment-one')
    ).resolves.toMatchObject({ id: 'attachment-one' });
    await expect(
      client.deleteAssetAttachment('tenant-one', 'inventory-one', 'asset-one', 'attachment-one')
    ).resolves.toBeUndefined();

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'PATCH http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/attachments/attachment-one/archive',
      'PATCH http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/attachments/attachment-one/restore',
      'DELETE http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/attachments/attachment-one'
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
