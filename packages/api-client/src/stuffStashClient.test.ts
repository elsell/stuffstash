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
              customFields: {},
              createdAt: '2026-06-23T10:00:00Z',
              updatedAt: '2026-06-24T10:00:00Z'
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
    expect(page.items[0]?.updatedAt).toBe('2026-06-24T10:00:00Z');
    expect(page.pagination).toEqual({ limit: 1, nextCursor: 'next-page', hasMore: true });
    expect(requests[0]?.url).toBe('http://api.local/tenants/tenant-one/inventories/inventory-one/assets?limit=1&lifecycleState=archived');

    await client.listAssets('tenant-one', 'inventory-one', 5, undefined, 'all', 'updated_desc');

    expect(requests[1]?.url).toBe('http://api.local/tenants/tenant-one/inventories/inventory-one/assets?limit=5&lifecycleState=all&sort=updated_desc');
  });

  it('maps compact current checkout state on asset reads', async () => {
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async () => Response.json({
        data: {
          id: 'asset-one',
          tenantId: 'tenant-one',
          inventoryId: 'inventory-one',
          kind: 'item',
          title: 'Cordless drill',
          description: '',
          lifecycleState: 'active',
          customFields: {},
          createdAt: '2026-06-23T10:00:00Z',
          updatedAt: '2026-06-24T10:00:00Z',
          currentCheckout: {
            id: 'checkout-one',
            state: 'open',
            checkedOutAt: '2026-06-24T11:00:00Z',
            checkedOutByPrincipalId: 'user-one',
            checkedOutByPrincipal: { id: 'user-one', email: 'user-one@example.test' }
          }
        },
        meta: {}
      })
    });

    await expect(client.getAsset('tenant-one', 'inventory-one', 'asset-one')).resolves.toMatchObject({
      id: 'asset-one',
      currentCheckout: {
        id: 'checkout-one',
        state: 'open',
        checkedOutByPrincipalId: 'user-one',
        checkedOutByPrincipal: { id: 'user-one', email: 'user-one@example.test' }
      }
    });
  });

  it('checks out and returns assets through inventory scoped routes', async () => {
    const requests: Request[] = [];
    const checkoutEnvelope = {
      data: {
        id: 'checkout-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        assetId: 'asset-one',
        state: 'open',
        checkedOutAt: '2026-06-24T11:00:00Z',
        checkedOutByPrincipalId: 'user-one',
        checkoutDetails: 'using at bench',
        createdAt: '2026-06-24T11:00:00Z',
        updatedAt: '2026-06-24T11:00:00Z'
      },
      meta: {}
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        return Response.json(checkoutEnvelope);
      }
    });

    await expect(
      client.checkoutAsset('tenant-one', 'inventory-one', 'asset-one', { details: 'using at bench' })
    ).resolves.toMatchObject({
      id: 'checkout-one',
      assetId: 'asset-one',
      checkoutDetails: 'using at bench'
    });
    await expect(client.returnAsset('tenant-one', 'inventory-one', 'asset-one')).resolves.toMatchObject({
      id: 'checkout-one'
    });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/checkout',
      'POST http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/return'
    ]);
    expect(await requests[0]?.json()).toEqual({ details: 'using at bench' });
    expect(await requests[1]?.json()).toEqual({});
  });

  it('lists checkout history and checked-out assets through inventory scoped routes', async () => {
    const requests: Request[] = [];
    const checkout = {
      id: 'checkout-one',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      assetId: 'asset-one',
      state: 'returned',
      checkedOutAt: '2026-06-24T11:00:00Z',
      checkedOutByPrincipalId: 'user-one',
      checkoutDetails: 'using at bench',
      returnedAt: '2026-06-24T12:00:00Z',
      returnedByPrincipalId: 'user-two',
      returnDetails: 'back in the bin',
      createdAt: '2026-06-24T11:00:00Z',
      updatedAt: '2026-06-24T12:00:00Z'
    };
    const checkedOutAsset = {
      asset: {
        id: 'asset-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        kind: 'item',
        title: 'Socket set',
        description: '',
        parentAssetId: null,
        lifecycleState: 'archived',
        customFields: {},
        createdAt: '2026-06-20T10:00:00Z',
        updatedAt: '2026-06-24T11:00:00Z',
        currentCheckout: {
          id: 'checkout-open',
          state: 'open',
          checkedOutAt: '2026-06-24T11:00:00Z',
          checkedOutByPrincipalId: 'user-one',
          checkedOutByPrincipal: { id: 'user-one', email: 'user-one@example.test' }
        }
      },
      checkout: {
        id: 'checkout-open',
        state: 'open',
        checkedOutAt: '2026-06-24T11:00:00Z',
        checkedOutByPrincipalId: 'user-one',
        checkedOutByPrincipal: { id: 'user-one', email: 'user-one@example.test' }
      }
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        if (request.url.includes('/checkouts')) {
          return Response.json({
            data: [checkout],
            meta: { pagination: { limit: 10, nextCursor: 'next-history', hasMore: true } }
          });
        }
        return Response.json({
          data: [checkedOutAsset],
          meta: { pagination: { limit: 5, nextCursor: null, hasMore: false } }
        });
      }
    });

    await expect(client.listAssetCheckoutHistory('tenant-one', 'inventory-one', 'asset-one', 10, 'after-one')).resolves.toMatchObject({
      items: [{ id: 'checkout-one', state: 'returned', checkoutDetails: 'using at bench', returnDetails: 'back in the bin' }],
      pagination: { limit: 10, nextCursor: 'next-history', hasMore: true }
    });
    await expect(client.listCheckedOutAssets('tenant-one', 'inventory-one', 5)).resolves.toMatchObject({
      items: [
        {
          asset: {
            id: 'asset-one',
            lifecycleState: 'archived',
            currentCheckout: {
              id: 'checkout-open',
              state: 'open',
              checkedOutByPrincipal: { id: 'user-one', email: 'user-one@example.test' }
            }
          },
          checkout: {
            id: 'checkout-open',
            state: 'open',
            checkedOutByPrincipal: { id: 'user-one', email: 'user-one@example.test' }
          }
        }
      ],
      pagination: { limit: 5, nextCursor: null, hasMore: false }
    });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/checkouts?limit=10&cursor=after-one',
      'GET http://api.local/tenants/tenant-one/inventories/inventory-one/checked-out-assets?limit=5'
    ]);
  });

  it('maps durable import jobs and sends job actions through inventory scoped routes', async () => {
    const requests: Request[] = [];
    const jobEnvelope = {
      data: {
        id: 'job-one',
        status: 'previewed',
        actorId: 'owner',
        source: {
          type: 'legacy_homebox_csv',
          name: 'Homebox CSV',
          imageImport: 'unavailable',
          allowPrivateNetwork: true,
          allowInsecureTLS: true,
          fingerprint: 'sha256:test'
        },
        counts: {
          fields: 1,
          tags: 1,
          locations: 0,
          assets: 1,
          attachments: 0,
          warnings: 1,
          errors: 0,
          fieldsCreated: 0,
          fieldsExisting: 0,
          tagsCreated: 0,
          tagsExisting: 1,
          locationsCreated: 0,
          assetsCreated: 0,
          assetsSkipped: 0,
          attachmentsCreated: 0,
          attachmentsSkipped: 0,
          recordsDiscarded: 0,
          sourceLinksDiscarded: 0
        },
        progress: {
          phase: 'ready',
          done: 2,
          total: 2,
          message: 'Preview ready',
          updatedAt: '2026-07-06T12:00:00Z'
        },
        progressHistory: [
          {
            phase: 'ready',
            done: 2,
            total: 2,
            message: 'Preview ready',
            updatedAt: '2026-07-06T12:00:00Z'
          }
        ],
        preview: {
          fields: [{ key: 'homebox-source-id', displayName: 'Homebox Source ID', type: 'text' }],
          tags: [{ key: 'workshop', displayName: 'Workshop' }],
          locations: [{ sourceId: 'location:garage', kind: 'location', title: 'Garage', archived: false }],
          assets: [{ sourceId: 'source:drill', kind: 'item', title: 'Drill', archived: false }],
          attachments: [],
          messages: [{ code: 'csv-images-unavailable', severity: 'warning', summary: 'Images are unavailable' }],
          fieldsTruncated: false,
          tagsTruncated: false,
          locationsTruncated: false,
          assetsTruncated: false,
          attachmentsTruncated: false,
          messagesTruncated: false
        },
        createdAt: '2026-07-06T12:00:00Z',
        updatedAt: '2026-07-06T12:00:00Z',
        resources: [{
          resourceType: 'asset',
          resourceId: 'asset-one',
          displayName: 'Drill',
          sourceEntityType: 'asset',
          sourceEntityId: 'source:drill',
          createdAt: '2026-07-06T12:00:01Z'
        }],
        messages: [{ code: 'csv-images-unavailable', severity: 'warning', summary: 'Images are unavailable' }]
      },
      meta: {}
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        if (request.method === 'GET' && request.url.endsWith('/imports/jobs')) {
          return Response.json({ data: { jobs: [jobEnvelope.data] }, meta: {} });
        }
        if (request.method === 'DELETE') {
          return new Response(null, { status: 204 });
        }
        return Response.json(jobEnvelope);
      }
    });
    const source = { sourceType: 'legacy_homebox_csv' as const, fileName: 'homebox.csv', contentBase64: 'SEIuCg==' };

    await expect(client.previewImportJob('tenant-one', 'inventory-one', source, { requestId: 'preview-request' })).resolves.toMatchObject({
      id: 'job-one',
      status: 'previewed',
      actorId: 'owner',
      source: { fingerprint: 'sha256:test', allowPrivateNetwork: true, allowInsecureTLS: true },
      counts: { assets: 1, recordsDiscarded: 0 },
      preview: { fields: [{ key: 'homebox-source-id' }], locations: [{ title: 'Garage' }], assets: [{ title: 'Drill' }] },
      resources: [{ resourceId: 'asset-one', displayName: 'Drill', sourceEntityId: 'source:drill' }],
      progress: { phase: 'ready' },
      progressHistory: [{ phase: 'ready' }],
      messages: [{ code: 'csv-images-unavailable' }]
    });
    await expect(client.listImportJobs('tenant-one', 'inventory-one')).resolves.toHaveLength(1);
    await client.getImportJob('tenant-one', 'inventory-one', 'job-one');
    await client.startImportJob('tenant-one', 'inventory-one', 'job-one', source, { requestId: 'start-request' });
    await client.cancelImportJob('tenant-one', 'inventory-one', 'job-one', 'discard_partial_progress', { requestId: 'cancel-request' });
    await expect(
      client.removeImportJobFromHistory('tenant-one', 'inventory-one', 'job-one', { requestId: 'remove-request' })
    ).resolves.toBeUndefined();

    expect(requests.map((request) => `${request.method} ${new URL(request.url).pathname}`)).toEqual([
      'POST /tenants/tenant-one/inventories/inventory-one/imports/jobs/preview',
      'GET /tenants/tenant-one/inventories/inventory-one/imports/jobs',
      'GET /tenants/tenant-one/inventories/inventory-one/imports/jobs/job-one',
      'POST /tenants/tenant-one/inventories/inventory-one/imports/jobs/job-one/start',
      'POST /tenants/tenant-one/inventories/inventory-one/imports/jobs/job-one/cancel',
      'DELETE /tenants/tenant-one/inventories/inventory-one/imports/jobs/job-one'
    ]);
    expect(await requests[4]?.json()).toEqual({ mode: 'discard_partial_progress' });
    expect(requests[0]?.headers.get('X-Request-ID')).toBe('preview-request');
    expect(requests[3]?.headers.get('X-Request-ID')).toBe('start-request');
    expect(requests[4]?.headers.get('X-Request-ID')).toBe('cancel-request');
    expect(requests[5]?.headers.get('X-Request-ID')).toBe('remove-request');
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
        customFields: { 'expiration-date': '2027-01-01', count: 2 },
        tags: [{ id: 'tag-one', key: 'medicine', displayName: 'Medicine', color: '#2F80ED' }]
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
        customFields: { 'expiration-date': '2027-01-01', count: 2 },
        tagIds: ['tag-one']
      })
    ).resolves.toMatchObject({
      customAssetTypeId: 'type-medicine',
      customFields: { 'expiration-date': '2027-01-01', count: 2 },
      tags: [{ id: 'tag-one', key: 'medicine', displayName: 'Medicine', color: '#2F80ED' }]
    });
    await expect(
      client.updateAsset('tenant-one', 'inventory-one', 'asset-one', {
        title: 'Ibuprofen',
        customFields: { count: 3 },
        tagIds: ['tag-one']
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
      customFields: { 'expiration-date': '2027-01-01', count: 2 },
      tagIds: ['tag-one']
    });
    expect(await requests[1]?.json()).toEqual({
      title: 'Ibuprofen',
      customFields: { count: 3 },
      tagIds: ['tag-one']
    });
  });

  it('manages inventory tags through tag routes', async () => {
    const requests: Request[] = [];
    const tagEnvelope = {
      data: {
        id: 'tag-one',
        tenantId: 'tenant-one',
        inventoryId: 'inventory-one',
        key: 'workshop',
        displayName: 'Workshop',
        color: '#2F80ED',
        lifecycleState: 'active',
        createdAt: '2026-07-06T12:00:00Z',
        updatedAt: '2026-07-06T12:00:00Z'
      },
      meta: { pagination: { limit: 10, nextCursor: null, hasMore: false } }
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        if (request.method === 'GET') {
          return Response.json({ data: [tagEnvelope.data], meta: tagEnvelope.meta });
        }
        return Response.json(tagEnvelope);
      }
    });

    await expect(client.listAssetTags('tenant-one', 'inventory-one', 10)).resolves.toMatchObject({
      items: [{ id: 'tag-one', key: 'workshop', color: '#2F80ED' }]
    });
    await expect(client.createAssetTag('tenant-one', 'inventory-one', { displayName: 'Workshop', color: '#2f80ed' })).resolves.toMatchObject({
      id: 'tag-one',
      displayName: 'Workshop'
    });
    await client.updateAssetTag('tenant-one', 'inventory-one', 'tag-one', { displayName: 'Shop' });
    await client.archiveAssetTag('tenant-one', 'inventory-one', 'tag-one');

    expect(requests.map((request) => `${request.method} ${new URL(request.url).pathname}`)).toEqual([
      'GET /tenants/tenant-one/inventories/inventory-one/tags',
      'POST /tenants/tenant-one/inventories/inventory-one/tags',
      'PATCH /tenants/tenant-one/inventories/inventory-one/tags/tag-one',
      'DELETE /tenants/tenant-one/inventories/inventory-one/tags/tag-one'
    ]);
    expect(requests[0]?.url).toBe('http://api.local/tenants/tenant-one/inventories/inventory-one/tags?limit=10');
    expect(await requests[1]?.json()).toEqual({ displayName: 'Workshop', color: '#2f80ed' });
    expect(await requests[2]?.json()).toEqual({ displayName: 'Shop' });
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

  it('lists and tests provider profiles with redacted metadata', async () => {
    const requests: Request[] = [];
    const providerProfile = {
      id: 'profile-language',
      tenantId: 'tenant-one',
      capability: 'language_inference',
      providerKind: 'gemini',
      displayName: 'Gemini cheap language',
      endpointUrl: 'https://generativelanguage.googleapis.com',
      modelName: 'gemini-2.5-flash-lite',
      runtimeOptions: { credentialType: 'api_key' },
      capabilityMetadata: { structuredOutput: true },
      promptTemplate: 'Prefer concise spoken answers.',
      credentialStatus: 'configured',
      lifecycleState: 'enabled',
      lastTestedAt: '2026-06-26T12:00:00Z',
      createdAt: '2026-06-26T11:00:00Z',
      updatedAt: '2026-06-26T12:00:00Z'
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        if (request.method === 'POST') {
          return Response.json({
            data: {
              providerProfileId: 'profile-language',
              capability: 'language_inference',
              providerKind: 'gemini',
              status: 'success',
              message: 'Provider profile test succeeded.',
              testedAt: '2026-06-26T12:01:00Z'
            },
            meta: {}
          });
        }
        return Response.json({ data: [providerProfile], meta: {} });
      }
    });

    await expect(client.listProviderProfiles('tenant-one')).resolves.toEqual([providerProfile]);
    await expect(client.testProviderProfile('tenant-one', 'profile-language')).resolves.toEqual({
      providerProfileId: 'profile-language',
      capability: 'language_inference',
      providerKind: 'gemini',
      status: 'success',
      message: 'Provider profile test succeeded.',
      testedAt: '2026-06-26T12:01:00Z'
    });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-one/provider-profiles',
      'POST http://api.local/tenants/tenant-one/provider-profiles/profile-language/test'
    ]);
    expect(requests[0]?.headers.get('Authorization')).toBe('Bearer id-token');
  });

  it('manages provider profiles through tenant-scoped endpoints', async () => {
    const requests: Request[] = [];
    const providerProfile = {
      id: 'profile-language',
      tenantId: 'tenant-one',
      capability: 'language_inference',
      providerKind: 'gemini',
      displayName: 'Gemini language',
      endpointUrl: '',
      modelName: 'gemini-2.5-flash-lite',
      runtimeOptions: {},
      capabilityMetadata: {},
      promptTemplate: 'Answer briefly.',
      credentialStatus: 'missing',
      lifecycleState: 'disabled',
      createdAt: '2026-06-26T11:00:00Z',
      updatedAt: '2026-06-26T11:00:00Z'
    };
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        return Response.json({ data: providerProfile, meta: {} });
      }
    });

    await client.createProviderProfile('tenant-one', {
      capability: 'language_inference',
      providerKind: 'gemini',
      displayName: 'Gemini language',
      modelName: 'gemini-2.5-flash-lite',
      promptTemplate: 'Answer briefly.'
    });
    await client.updateProviderProfile('tenant-one', 'profile-language', {
      promptTemplate: 'Answer in one sentence.'
    });
    await client.enableProviderProfile('tenant-one', 'profile-language');
    await client.disableProviderProfile('tenant-one', 'profile-language');
    await client.archiveProviderProfile('tenant-one', 'profile-language');

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'POST http://api.local/tenants/tenant-one/provider-profiles',
      'PATCH http://api.local/tenants/tenant-one/provider-profiles/profile-language',
      'POST http://api.local/tenants/tenant-one/provider-profiles/profile-language/enable',
      'POST http://api.local/tenants/tenant-one/provider-profiles/profile-language/disable',
      'POST http://api.local/tenants/tenant-one/provider-profiles/profile-language/archive'
    ]);
    expect(await requests[0]?.json()).toEqual({
      capability: 'language_inference',
      providerKind: 'gemini',
      displayName: 'Gemini language',
      modelName: 'gemini-2.5-flash-lite',
      promptTemplate: 'Answer briefly.'
    });
    expect(await requests[1]?.json()).toEqual({
      promptTemplate: 'Answer in one sentence.'
    });
  });

  it('replaces provider credentials only in the credential request body', async () => {
    const requests: Request[] = [];
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async (input: RequestInfo | URL, init?: RequestInit) => {
        const request = new Request(input, init);
        requests.push(request);
        return Response.json({
          data: {
            id: 'profile-language',
            tenantId: 'tenant-one',
            capability: 'language_inference',
            providerKind: 'gemini',
            displayName: 'Gemini language',
            endpointUrl: '',
            modelName: 'gemini-2.5-flash-lite',
            runtimeOptions: {},
            capabilityMetadata: {},
            credentialStatus: 'configured',
            lifecycleState: 'disabled',
            createdAt: '2026-06-26T11:00:00Z',
            updatedAt: '2026-06-26T12:00:00Z'
          },
          meta: {}
        });
      }
    });

    await expect(
      client.replaceProviderProfileCredential('tenant-one', 'profile-language', {
        purpose: 'api_key',
        credential: 'secret-api-key'
      })
    ).resolves.toMatchObject({
      id: 'profile-language',
      credentialStatus: 'configured'
    });

    expect(requests[0]?.url).toBe(
      'http://api.local/tenants/tenant-one/provider-profiles/profile-language/credential'
    );
    expect(requests[0]?.method).toBe('PUT');
    expect(await requests[0]?.json()).toEqual({
      purpose: 'api_key',
      credential: 'secret-api-key'
    });
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
        'attachment/one',
        'medium'
      )
    ).resolves.toEqual({
      uri: 'http://api.local/tenants/tenant%20one/inventories/inventory%2Fone/assets/asset%20one/attachments/attachment%2Fone/thumbnail?variant=medium',
      headers: { Authorization: 'Bearer id-token' }
    });
  });

  it('falls back to small thumbnail references for invalid runtime variants', async () => {
    const client = new StuffStashClient({
      baseUrl: 'http://api.local/',
      tokenProvider: () => 'id-token',
      fetch: async () => Response.json({ data: {}, meta: {} })
    });

    await expect(
      client.assetAttachmentThumbnailReference('tenant-one', 'inventory-one', 'asset-one', 'attachment-one', 'large&bad=true' as never)
    ).resolves.toMatchObject({
      uri: 'http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/attachments/attachment-one/thumbnail?variant=small'
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
                lifecycleState: 'archived',
                tags: [{ id: 'tag-travel', key: 'travel', displayName: 'Travel', color: '#2F80ED' }],
                currentCheckout: {
                  id: 'checkout-open',
                  state: 'open',
                  checkedOutAt: '2026-06-24T11:00:00Z',
                  checkedOutByPrincipalId: 'user-one',
                  checkedOutByPrincipal: { id: 'user-one', email: 'user-one@example.test' }
                }
              },
              matches: [{ field: 'title', value: 'Passport' }]
            }
          ],
          meta: { pagination: { limit: 5, nextCursor: null, hasMore: false } }
        });
      }
    });

    await expect(
      client.searchAssets('tenant-one', 'Passport', { limit: 5, inventoryId: 'inventory-one', lifecycleState: 'archived', mode: 'exact' })
    ).resolves.toMatchObject({
      items: [{
        asset: {
          id: 'asset-one',
          lifecycleState: 'archived',
          tags: [{ id: 'tag-travel', key: 'travel', displayName: 'Travel', color: '#2F80ED' }],
          currentCheckout: {
            id: 'checkout-open',
            checkedOutByPrincipal: { id: 'user-one', email: 'user-one@example.test' }
          }
        }
      }]
    });

    expect(requests[0]?.url).toBe(
      'http://api.local/tenants/tenant-one/search/assets?q=Passport&limit=5&inventoryId=inventory-one&lifecycleState=archived&mode=exact'
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

  it('lists tenant, inventory, and asset audit records through generated paths', async () => {
    const requests: Request[] = [];
    const record = {
      id: 'audit-one',
      tenantId: 'tenant-one',
      inventoryId: 'inventory-one',
      principalId: 'principal-one',
      principal: { id: 'principal-one', email: 'alex@example.test' },
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
    await expect(client.listAssetAuditRecords('tenant-one', 'inventory-one', 'asset-one', 2)).resolves.toEqual({
      items: [record],
      pagination: { limit: 1, nextCursor: 'next-page', hasMore: true }
    });

    expect(requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET http://api.local/tenants/tenant-one/audit-records?limit=1',
      'GET http://api.local/tenants/tenant-one/inventories/inventory-one/audit-records?limit=1&cursor=next-page',
      'GET http://api.local/tenants/tenant-one/inventories/inventory-one/assets/asset-one/audit-records?limit=2'
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

  it('awaits an async refreshed token before sending API requests', async () => {
    const requests: Request[] = [];
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: async () => 'refreshed-id-token',
      fetch: async (input, init) => {
        const request = new Request(input, init);
        requests.push(request);
        return Response.json({ data: { id: 'principal-1' } });
      }
    });

    await expect(client.me()).resolves.toMatchObject({ id: 'principal-1' });
    expect(requests[0]?.headers.get('Authorization')).toBe('Bearer refreshed-id-token');
  });

  it('does not send API requests when mobile auth cannot provide a token', async () => {
    let fetchCalled = false;
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: async () => {
        throw new Error('Sign in to Stuff Stash.');
      },
      fetch: async () => {
        fetchCalled = true;
        return Response.json({ data: { id: 'principal-1' } });
      }
    });

    await expect(client.me()).rejects.toThrow('Sign in to Stuff Stash.');
    expect(fetchCalled).toBe(false);
  });

  it('preserves safe validation details when API errors are generic', async () => {
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async () =>
        Response.json(
          {
            error: {
              code: 'invalid_request',
              message: 'validation failed',
              details: [{ message: 'contentType must be one of image/jpeg, image/png, image/webp, application/pdf' }]
            }
          },
          { status: 400 }
        )
    });

    await expect(
      client.createAssetAttachment('tenant-one', 'inventory-one', 'asset-one', {
        fileName: 'photo.jpg',
        contentType: 'image/jpeg',
        contentBase64: 'ZmFrZQ=='
      })
    ).rejects.toMatchObject({
      status: 400,
      code: 'invalid_request',
      message: 'contentType must be one of image/jpeg, image/png, image/webp, application/pdf'
    });
  });

  it('keeps non-validation API error details out of the thrown message', async () => {
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async () =>
        Response.json(
          {
            error: {
              code: 'internal_error',
              message: 'Internal server error.',
              details: [{ message: 'database detail that must not be promoted' }]
            }
          },
          { status: 500 }
        )
    });

    await expect(client.me()).rejects.toMatchObject({
      status: 500,
      code: 'internal_error',
      message: 'Internal server error.'
    });
  });

  it('includes status when API errors have no safe envelope message', async () => {
    const client = new StuffStashClient({
      baseUrl: 'http://api.local',
      tokenProvider: () => 'id-token',
      fetch: async () => Response.json({}, { status: 404 })
    });

    await expect(client.listAssetAuditRecords('tenant-one', 'inventory-one', 'asset-one')).rejects.toMatchObject({
      status: 404,
      code: 'request_failed',
      message: 'Request failed with status 404.'
    });
  });
});
