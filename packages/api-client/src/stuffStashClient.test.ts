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

  it('maps asset custom type and custom field values through create and update', async () => {
    const requests: Request[] = [];
    const assetEnvelope = {
      data: {
        id: 'asset-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        kind: 'item',
        title: 'Ibuprofen',
        description: '',
        lifecycleState: 'active',
        customAssetTypeId: 'type-medicine',
        customFields: { 'expiration-date': '2027-01-01', count: 2 }
      },
      meta: {}
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        return Response.json(assetEnvelope);
      }
    });

    await expect(
      client.createAsset('tenant-one', 'inventory-one', {
        kind: 'item',
        title: 'Ibuprofen',
        customAssetTypeId: 'type-medicine',
        customFields: { 'expiration-date': '2027-01-01', count: 2 }
      })
    ).resolves.toMatchObject({
      customAssetTypeId: 'type-medicine',
      customFields: { 'expiration-date': '2027-01-01', count: 2 }
    });
    await expect(
      client.updateAsset('tenant-one', 'inventory-one', 'asset-one', {
        title: 'Ibuprofen',
        customFields: { count: 3 }
      })
    ).resolves.toMatchObject({
      customAssetTypeId: 'type-medicine',
      customFields: { 'expiration-date': '2027-01-01', count: 2 }
    });

    expect(await requests[0]?.json()).toEqual({
      kind: 'item',
      title: 'Ibuprofen',
      parentAssetId: undefined,
      customAssetTypeId: 'type-medicine',
      customFields: { 'expiration-date': '2027-01-01', count: 2 }
    });
    expect(await requests[1]?.json()).toEqual({
      title: 'Ibuprofen',
      customFields: { count: 3 }
    });
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

  it('passes search mode and lifecycle filters through generated paths', async () => {
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
              type: 'asset',
              tenantId: 'tenant-one',
              inventory: { id: 'inventory-one', name: 'Household' },
              asset: {
                id: 'asset-one',
                kind: 'item',
                title: 'Passport',
                description: '',
                parentAssetId: null,
                lifecycleState: 'archived'
              },
              matches: [{ field: 'title', value: 'Passport' }]
            }
          ],
          meta: { pagination: { limit: 5, nextCursor: null, hasMore: false } }
        });
      }
    });

    await expect(
      client.searchAssets('tenant-one', 'Passport', { limit: 5, lifecycleState: 'archived', mode: 'exact' })
    ).resolves.toMatchObject({
      items: [{ asset: { id: 'asset-one', lifecycleState: 'archived' } }]
    });

    expect(requests[0]?.url).toBe(
      'http://api.local/tenants/tenant-one/search/assets?q=Passport&limit=5&lifecycleState=archived&mode=exact'
    );
  });

  it('manages direct inventory access grants through generated paths', async () => {
    const requests: Request[] = [];
    const grant = {
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      principalId: 'principal-two',
      relationship: 'viewer'
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        if (request.method === 'GET') {
          return Response.json({
            data: [grant],
            meta: { pagination: { limit: 50, nextCursor: null, hasMore: false } }
          });
        }
        if (request.method === 'DELETE') {
          return new Response(null, { status: 204 });
        }
        return Response.json({ data: grant, meta: {} }, { status: 201 });
      }
    });

    await expect(client.listInventoryAccessGrants('tenant-one', 'inventory-one')).resolves.toMatchObject({
      items: [grant]
    });
    await expect(
      client.grantInventoryAccess('tenant-one', 'inventory-one', {
        principalId: 'principal-two',
        relationship: 'viewer'
      })
    ).resolves.toEqual(grant);
    await expect(
      client.revokeInventoryAccess('tenant-one', 'inventory-one', 'principal-two', 'viewer')
    ).resolves.toBeUndefined();

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-one/inventories/inventory-one/access-grants?limit=50',
      'POST http://api.local/tenants/tenant-one/inventories/inventory-one/access-grants',
      'DELETE http://api.local/tenants/tenant-one/inventories/inventory-one/access-grants/principal-two/viewer'
    ]);
    expect(await requests[1]?.json()).toEqual({ principalId: 'principal-two', relationship: 'viewer' });
  });

  it('manages inventory access invitations through generated paths', async () => {
    const requests: Request[] = [];
    const invitation = {
      id: 'invite-one',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      email: 'person@example.test',
      relationship: 'editor',
      status: 'pending',
      isExpired: false,
      expiresAt: '2026-06-30T00:00:00Z',
      inviterPrincipalId: 'principal-one',
      acceptanceToken: 'raw-token'
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        if (request.method === 'GET') {
          return Response.json({
            data: [invitation],
            meta: { pagination: { limit: 20, nextCursor: null, hasMore: false } }
          });
        }
        if (request.method === 'PATCH' && request.url.endsWith('/cancel')) {
          return new Response(null, { status: 204 });
        }
        if (request.method === 'DELETE') {
          return new Response(null, { status: 204 });
        }
        if (request.url.endsWith('/accept')) {
          return Response.json({
            data: {
              grant: {
                tenantId: 'tenant-one',
                inventoryId: 'inventory-one',
                principalId: 'principal-two',
                relationship: 'editor'
              },
              invitation: { ...invitation, status: 'accepted', acceptedPrincipalId: 'principal-two' }
            },
            meta: {}
          });
        }
        return Response.json({ data: invitation, meta: {} }, { status: request.method === 'POST' ? 201 : 200 });
      }
    });

    await expect(
      client.listInventoryAccessInvitations('tenant-one', 'inventory-one', { limit: 20, status: 'pending' })
    ).resolves.toMatchObject({ items: [invitation] });
    await expect(
      client.createInventoryAccessInvitation('tenant-one', 'inventory-one', {
        email: 'person@example.test',
        relationship: 'editor'
      })
    ).resolves.toEqual(invitation);
    await expect(
      client.updateInventoryAccessInvitationExpiration(
        'tenant-one',
        'inventory-one',
        'invite-one',
        '2026-07-01T00:00:00Z'
      )
    ).resolves.toEqual(invitation);
    await expect(client.cancelInventoryAccessInvitation('tenant-one', 'inventory-one', 'invite-one')).resolves.toBeUndefined();
    await expect(client.deleteInventoryAccessInvitation('tenant-one', 'inventory-one', 'invite-one')).resolves.toBeUndefined();
    await expect(
      client.acceptInventoryAccessInvitation('tenant-one', 'inventory-one', 'invite-one', 'raw-token')
    ).resolves.toMatchObject({
      grant: { principalId: 'principal-two', relationship: 'editor' },
      invitation: { id: 'invite-one', status: 'accepted' }
    });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-one/inventories/inventory-one/access-invitations?limit=20&status=pending',
      'POST http://api.local/tenants/tenant-one/inventories/inventory-one/access-invitations',
      'PATCH http://api.local/tenants/tenant-one/inventories/inventory-one/access-invitations/invite-one/expiration',
      'PATCH http://api.local/tenants/tenant-one/inventories/inventory-one/access-invitations/invite-one/cancel',
      'DELETE http://api.local/tenants/tenant-one/inventories/inventory-one/access-invitations/invite-one',
      'POST http://api.local/tenants/tenant-one/inventories/inventory-one/access-invitations/invite-one/accept'
    ]);
    expect(await requests[1]?.json()).toEqual({ email: 'person@example.test', relationship: 'editor' });
    expect(await requests[2]?.json()).toEqual({ expiresAt: '2026-07-01T00:00:00Z' });
    expect(await requests[5]?.json()).toEqual({ acceptanceToken: 'raw-token' });
  });

  it('lists tenant and inventory audit records through generated paths', async () => {
    const requests: Request[] = [];
    const record = {
      id: 'audit-one',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      principalId: 'principal-one',
      action: 'asset.created',
      source: 'api',
      targetType: 'asset',
      targetId: 'asset-one',
      occurredAt: '2026-06-24T12:00:00Z',
      requestId: 'request-one',
      metadata: { operation_id: 'operation-one' }
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        return Response.json({
          data: [record],
          meta: { pagination: { limit: 1, nextCursor: 'next-page', hasMore: true } }
        });
      }
    });

    await expect(client.listTenantAuditRecords('tenant-one', 1)).resolves.toEqual({
      items: [record],
      pagination: { limit: 1, nextCursor: 'next-page', hasMore: true }
    });
    await expect(client.listInventoryAuditRecords('tenant-one', 'inventory-one', 1, 'next-page')).resolves.toEqual({
      items: [record],
      pagination: { limit: 1, nextCursor: 'next-page', hasMore: true }
    });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-one/audit-records?limit=1',
      'GET http://api.local/tenants/tenant-one/inventories/inventory-one/audit-records?limit=1&cursor=next-page'
    ]);
  });

  it('manages custom asset types and field definitions through generated paths', async () => {
    const requests: Request[] = [];
    const assetType = {
      id: 'type-medicine',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      scope: 'inventory',
      key: 'medicine',
      displayName: 'Medicine',
      description: 'Medication and vitamins',
      lifecycleState: 'active'
    };
    const definition = {
      id: 'field-expiration',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      scope: 'inventory',
      key: 'expiration-date',
      displayName: 'Expiration date',
      type: 'date',
      applicability: 'custom_asset_types',
      customAssetTypeIds: ['type-medicine'],
      enumOptions: null,
      lifecycleState: 'active'
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        const path = new URL(request.url).pathname;
        if (path.includes('custom-asset-types')) {
          return Response.json({
            data: request.method === 'GET' ? [assetType] : assetType,
            meta: request.method === 'GET' ? { pagination: { limit: 25, nextCursor: null, hasMore: false } } : {}
          });
        }
        return Response.json({
          data: request.method === 'GET' ? [definition] : definition,
          meta: request.method === 'GET' ? { pagination: { limit: 25, nextCursor: null, hasMore: false } } : {}
        });
      }
    });

    await expect(client.listInventoryCustomAssetTypes('tenant-one', 'inventory-one', 25)).resolves.toMatchObject({
      items: [{ id: 'type-medicine', displayName: 'Medicine', scope: 'inventory' }]
    });
    await expect(
      client.createInventoryCustomAssetType('tenant-one', 'inventory-one', {
        key: 'medicine',
        displayName: 'Medicine',
        description: 'Medication and vitamins'
      })
    ).resolves.toMatchObject({ id: 'type-medicine' });
    await expect(
      client.updateInventoryCustomAssetType('tenant-one', 'inventory-one', 'type-medicine', {
        displayName: 'Medicine and Vitamins'
      })
    ).resolves.toMatchObject({ id: 'type-medicine' });
    await expect(client.archiveInventoryCustomAssetType('tenant-one', 'inventory-one', 'type-medicine')).resolves.toMatchObject({
      id: 'type-medicine'
    });
    await expect(client.listInventoryCustomFieldDefinitions('tenant-one', 'inventory-one', 25)).resolves.toMatchObject({
      items: [{ id: 'field-expiration', type: 'date', customAssetTypeIds: ['type-medicine'] }]
    });
    await expect(
      client.createInventoryCustomFieldDefinition('tenant-one', 'inventory-one', {
        key: 'expiration-date',
        displayName: 'Expiration date',
        type: 'date',
        applicability: 'custom_asset_types',
        customAssetTypeIds: ['type-medicine']
      })
    ).resolves.toMatchObject({ id: 'field-expiration' });
    await expect(
      client.updateInventoryCustomFieldDefinition('tenant-one', 'inventory-one', 'field-expiration', {
        displayName: 'Use by'
      })
    ).resolves.toMatchObject({ id: 'field-expiration' });
    await expect(client.archiveInventoryCustomFieldDefinition('tenant-one', 'inventory-one', 'field-expiration')).resolves.toMatchObject({
      id: 'field-expiration'
    });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-one/inventories/inventory-one/custom-asset-types?limit=25',
      'POST http://api.local/tenants/tenant-one/inventories/inventory-one/custom-asset-types',
      'PATCH http://api.local/tenants/tenant-one/inventories/inventory-one/custom-asset-types/type-medicine',
      'PATCH http://api.local/tenants/tenant-one/inventories/inventory-one/custom-asset-types/type-medicine/archive',
      'GET http://api.local/tenants/tenant-one/inventories/inventory-one/custom-field-definitions?limit=25',
      'POST http://api.local/tenants/tenant-one/inventories/inventory-one/custom-field-definitions',
      'PATCH http://api.local/tenants/tenant-one/inventories/inventory-one/custom-field-definitions/field-expiration',
      'PATCH http://api.local/tenants/tenant-one/inventories/inventory-one/custom-field-definitions/field-expiration/archive'
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
