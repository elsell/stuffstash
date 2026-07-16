import { afterEach, describe, expect, it, vi } from 'vitest';
import { SeededInventoryRepository } from './seededInventoryRepository';
import type { WorkspaceSeed } from '$lib/ports/inventoryRepository';

const seed: WorkspaceSeed = {
  principal: { id: 'person-one', email: 'person@example.test' },
  tenants: [
    { id: 'tenant-home', name: 'Home', access: { relationship: 'owner', permissions: ['view'] } },
    { id: 'tenant-cabin', name: 'Cabin', access: { relationship: 'editor', permissions: ['view'] } },
    { id: 'tenant-empty', name: 'Empty', access: { relationship: 'owner', permissions: ['view', 'create_inventory'] } },
    { id: 'tenant-viewer-empty', name: 'Viewer Empty', access: { relationship: 'viewer', permissions: ['view'] } }
  ],
  inventories: [
    {
      id: 'inventory-household',
      tenantId: 'tenant-home',
      name: 'Household',
      access: { relationship: 'owner', permissions: ['view', 'create_asset'] }
    },
    {
      id: 'inventory-cabin',
      tenantId: 'tenant-cabin',
      name: 'Cabin Gear',
      access: { relationship: 'viewer', permissions: ['view'] }
    }
  ],
  customAssetTypes: [],
  customFieldDefinitions: [],
  assets: [
    {
      id: 'asset-home',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Passport',
      description: 'Blue folder',
      parentAssetId: null,
      lifecycleState: 'active'
    },
    {
      id: 'asset-cabin',
      tenantId: 'tenant-cabin',
      inventoryId: 'inventory-cabin',
      kind: 'item',
      title: 'Lantern',
      description: 'Shelf',
      parentAssetId: null,
      lifecycleState: 'active'
    },
    {
      id: 'asset-archived',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Archived Passport',
      description: 'Old folder',
      parentAssetId: null,
      lifecycleState: 'archived'
    }
  ]
};

afterEach(() => {
  vi.useRealTimers();
});

describe('SeededInventoryRepository tenant selection', () => {
  it('loads the selected tenant inventories and scopes assets to its first inventory', async () => {
    const repository = new SeededInventoryRepository(seed);

    const data = await repository.selectTenant('tenant-cabin');

    expect(data.context.selectedTenantId).toBe('tenant-cabin');
    expect(data.context.inventories.map((inventory) => inventory.id)).toEqual(['inventory-cabin']);
    expect(data.context.selectedInventoryId).toBe('inventory-cabin');
    expect(data.context.capability).toBe('viewer');
    expect(data.assets.map((asset) => asset.id)).toEqual(['asset-cabin']);
  });

  it('keeps an empty tenant selected without leaking another tenant inventory', async () => {
    const repository = new SeededInventoryRepository(seed);

    const data = await repository.selectTenant('tenant-empty');

    expect(data.context.selectedTenantId).toBe('tenant-empty');
    expect(data.context.inventories).toEqual([]);
    expect(data.context.selectedInventoryId).toBe('');
    expect(data.assets).toEqual([]);
  });

  it('creates a starter inventory inside the selected tenant', async () => {
    const repository = new SeededInventoryRepository(seed);
    await repository.selectTenant('tenant-empty');

    const data = await repository.createInventory('tenant-empty', 'Household');

    expect(data.context.selectedTenantId).toBe('tenant-empty');
    expect(data.context.inventories).toMatchObject([{ tenantId: 'tenant-empty', name: 'Household' }]);
    expect(data.context.selectedInventoryId).toBe(data.context.inventories[0]?.id);
  });

  it('rejects starter inventory creation when the selected tenant lacks permission', async () => {
    const repository = new SeededInventoryRepository(seed);
    await repository.selectTenant('tenant-viewer-empty');

    await expect(repository.createInventory('tenant-viewer-empty', 'Household')).rejects.toThrow(
      'You do not have permission'
    );
  });

  it('does not leak assets when inventory selection is mismatched across tenants', async () => {
    const repository = new SeededInventoryRepository(seed);

    const data = await repository.selectInventory('tenant-home', 'inventory-cabin');

    expect(data.context.selectedTenantId).toBe('tenant-home');
    expect(data.context.selectedInventoryId).toBe('inventory-household');
    expect(data.assets.map((asset) => asset.id)).toEqual(['asset-home']);
    await expect(
      repository.searchAssets({ tenantId: 'tenant-home', inventoryId: 'inventory-household', query: 'Lantern', lifecycleState: 'active', mode: 'fuzzy' })
    ).resolves.toEqual([]);
  });

  it('loads and updates asset detail inside the selected inventory', async () => {
    const repository = new SeededInventoryRepository(seed);

    await expect(repository.getAsset('tenant-home', 'inventory-household', 'asset-home')).resolves.toMatchObject({
      id: 'asset-home',
      title: 'Passport'
    });

    const updated = await repository.updateAsset('tenant-home', 'inventory-household', 'asset-home', {
      title: 'Updated Passport',
      description: 'Fire safe',
      parentAssetId: null
    });

    expect(updated).toMatchObject({
      id: 'asset-home',
      title: 'Updated Passport',
      description: 'Fire safe',
      parentAssetId: null
    });
    await expect(repository.getAsset('tenant-home', 'inventory-household', 'asset-home')).resolves.toMatchObject({
      title: 'Updated Passport'
    });
  });

  it('moves an asset while preserving its editable fields', async () => {
    const repository = new SeededInventoryRepository(seed);
    const shelf = await repository.createAsset('tenant-home', 'inventory-household', {
      kind: 'container',
      title: 'Office shelf',
      description: '',
      parentAssetId: null,
      photos: []
    });

    const moved = await repository.moveAsset('tenant-home', 'inventory-household', 'asset-home', shelf.id);

    expect(moved).toMatchObject({
      id: 'asset-home',
      title: 'Passport',
      description: 'Blue folder',
      parentAssetId: shelf.id
    });
  });

  it('creates assets inside active locations and containers only', async () => {
    const repository = new SeededInventoryRepository(seed);

    const shelf = await repository.createAsset('tenant-home', 'inventory-household', {
      kind: 'container',
      title: 'Garage shelf',
      description: '',
      parentAssetId: null,
      photos: []
    });
    const tape = await repository.createAsset('tenant-home', 'inventory-household', {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: shelf.id,
      photos: []
    });

    expect(tape).toMatchObject({
      title: 'Tape measure',
      parentAssetId: shelf.id
    });
    expect(tape.id).not.toBe(shelf.id);
  });

  it('rejects invalid containment parents in the local adapter', async () => {
    const repository = new SeededInventoryRepository(seed);

    await expect(
      repository.createAsset('tenant-home', 'inventory-household', {
        kind: 'item',
        title: 'Folder tab',
        description: '',
        parentAssetId: 'asset-home',
        photos: []
      })
    ).rejects.toThrow('Parent asset must be a container or location');
    await expect(
      repository.createAsset('tenant-home', 'inventory-household', {
        kind: 'item',
        title: 'Old folder tab',
        description: '',
        parentAssetId: 'asset-archived',
        photos: []
      })
    ).rejects.toThrow('Parent asset must be active');
    await expect(
      repository.createAsset('tenant-home', 'inventory-household', {
        kind: 'item',
        title: 'Lost tab',
        description: '',
        parentAssetId: 'missing-parent',
        photos: []
      })
    ).rejects.toThrow('Parent asset not found');
  });

  it('rejects containment cycles on asset moves', async () => {
    const repository = new SeededInventoryRepository(seed);

    const shelf = await repository.createAsset('tenant-home', 'inventory-household', {
      kind: 'container',
      title: 'Garage shelf',
      description: '',
      parentAssetId: null,
      photos: []
    });
    const bin = await repository.createAsset('tenant-home', 'inventory-household', {
      kind: 'container',
      title: 'Garage bin',
      description: '',
      parentAssetId: shelf.id,
      photos: []
    });

    await expect(
      repository.updateAsset('tenant-home', 'inventory-household', shelf.id, {
        title: shelf.title,
        description: shelf.description,
        parentAssetId: shelf.id
      })
    ).rejects.toThrow('Asset cannot contain itself');
    await expect(
      repository.updateAsset('tenant-home', 'inventory-household', shelf.id, {
        title: shelf.title,
        description: shelf.description,
        parentAssetId: bin.id
      })
    ).rejects.toThrow('Asset cannot be moved inside its own contents');
  });

  it('switches between active and archived asset views without mixing lifecycle states', async () => {
    const repository = new SeededInventoryRepository(seed);

    const archived = await repository.selectAssetLifecycle('tenant-home', 'inventory-household', 'archived');
    const active = await repository.selectAssetLifecycle('tenant-home', 'inventory-household', 'active');

    expect(archived.context.assetLifecycleState).toBe('archived');
    expect(archived.assets.map((asset) => asset.id)).toEqual(['asset-archived']);
    expect(active.context.assetLifecycleState).toBe('active');
    expect(active.assets.map((asset) => asset.id)).toEqual(['asset-home']);
  });

  it('keeps search scoped to active assets like the API search adapter', async () => {
    const repository = new SeededInventoryRepository(seed);

    await repository.selectAssetLifecycle('tenant-home', 'inventory-household', 'archived');

    await expect(
      repository.searchAssets({ tenantId: 'tenant-home', inventoryId: 'inventory-household', query: 'Archived Passport', lifecycleState: 'active', mode: 'fuzzy' })
    ).resolves.toEqual([]);
    await expect(
      repository.searchAssets({ tenantId: 'tenant-home', inventoryId: 'inventory-household', query: 'Passport', lifecycleState: 'active', mode: 'fuzzy' })
    ).resolves.toMatchObject([
      { asset: { id: 'asset-home', lifecycleState: 'active' } }
    ]);
  });

  it('supports exact archived search in the selected inventory', async () => {
    const repository = new SeededInventoryRepository(seed);

    await expect(
      repository.searchAssets({ tenantId: 'tenant-home', inventoryId: 'inventory-household', query: 'Archived Passport', lifecycleState: 'archived', mode: 'exact' })
    ).resolves.toMatchObject([{ asset: { id: 'asset-archived', lifecycleState: 'archived' } }]);
    await expect(
      repository.searchAssets({ tenantId: 'tenant-home', inventoryId: 'inventory-household', query: 'Passport', lifecycleState: 'archived', mode: 'exact' })
    ).resolves.toEqual([]);
  });

  it('supports all-lifecycle fuzzy search in the selected inventory', async () => {
    const repository = new SeededInventoryRepository(seed);

    await expect(
      repository.searchAssets({ tenantId: 'tenant-home', inventoryId: 'inventory-household', query: 'Passport', lifecycleState: 'all', mode: 'fuzzy' })
    ).resolves.toMatchObject([
      { asset: { id: 'asset-home', lifecycleState: 'active' } },
      { asset: { id: 'asset-archived', lifecycleState: 'archived' } }
    ]);
  });

  it('filters local search by selected tag IDs without replacing query text', async () => {
    const taggedSeed: WorkspaceSeed = {
      ...seed,
      assets: [
        {
          ...seed.assets[0],
          tags: [
            { id: 'tag-travel', key: 'travel', displayName: 'Travel' },
            { id: 'tag-documents', key: 'documents', displayName: 'Documents' }
          ]
        },
        {
          ...seed.assets[2],
          lifecycleState: 'active',
          tags: [{ id: 'tag-travel', key: 'travel', displayName: 'Travel' }]
        }
      ]
    };
    const repository = new SeededInventoryRepository(taggedSeed);

    await expect(
      repository.searchAssets({ tenantId: 'tenant-home', inventoryId: 'inventory-household', query: '', tagIds: ['tag-travel'], lifecycleState: 'active', mode: 'fuzzy' })
    ).resolves.toHaveLength(2);
    await expect(
      repository.searchAssets({ tenantId: 'tenant-home', inventoryId: 'inventory-household', query: 'Passport', tagIds: ['tag-travel', 'tag-documents'], lifecycleState: 'active', mode: 'fuzzy' })
    ).resolves.toMatchObject([
      { asset: { id: 'asset-home' } }
    ]);
  });

  it('refreshes collapsed local import progress history with latest safe counts', async () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-07-06T12:00:00Z'));
    const repository = new SeededInventoryRepository(seed);
    const previewed = await repository.previewImportJob('tenant-home', 'inventory-household', {
      sourceType: 'legacy_homebox',
      baseUrl: 'https://homebox.example.test',
      username: 'owner@example.test',
      password: 'secret',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false
    });

    const firstStart = await repository.startImportJob('tenant-home', 'inventory-household', previewed.id, {
      sourceType: 'legacy_homebox',
      baseUrl: 'https://homebox.example.test',
      username: 'owner@example.test',
      password: 'secret',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false
    });
    const firstProgressUpdatedAt = firstStart.progress.updatedAt;
    vi.setSystemTime(new Date('2026-07-06T12:01:00Z'));
    const secondStart = await repository.startImportJob('tenant-home', 'inventory-household', previewed.id, {
      sourceType: 'legacy_homebox',
      baseUrl: 'https://homebox.example.test',
      username: 'owner@example.test',
      password: 'secret',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false
    });

    const last = secondStart.progressHistory.at(-1);
    expect(secondStart.progressHistory.map((progress) => progress.phase)).toEqual(['ready', 'reading_source']);
    expect(secondStart.progress.updatedAt).not.toBe(firstProgressUpdatedAt);
    expect(last).toMatchObject({
      phase: 'reading_source',
      message: 'Queued locally',
      done: secondStart.progress.done,
      total: secondStart.progress.total,
      updatedAt: secondStart.progress.updatedAt
    });
  });

  it('keeps distinct local import progress messages in the same phase', async () => {
    const repository = new SeededInventoryRepository(seed);
    const previewed = await repository.previewImportJob('tenant-home', 'inventory-household', {
      sourceType: 'legacy_homebox',
      baseUrl: 'https://homebox.example.test',
      username: 'owner@example.test',
      password: 'secret',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false
    });

    await repository.startImportJob('tenant-home', 'inventory-household', previewed.id, {
      sourceType: 'legacy_homebox',
      baseUrl: 'https://homebox.example.test',
      username: 'owner@example.test',
      password: 'secret',
      includeImages: true,
      allowPrivateNetwork: false,
      allowInsecureTLS: false
    });
    const cancelled = await repository.cancelImportJob('tenant-home', 'inventory-household', previewed.id, 'keep_partial_progress');

    expect(cancelled.progressHistory.map((progress) => [progress.phase, progress.message])).toEqual([
      ['ready', undefined],
      ['reading_source', 'Queued locally'],
      ['reading_source', 'Cancellation requested']
    ]);
  });

  it('archives, restores, and deletes assets inside the selected inventory', async () => {
    const repository = new SeededInventoryRepository(seed);

    await expect(repository.archiveAsset('tenant-home', 'inventory-household', 'asset-home')).resolves.toMatchObject({
      id: 'asset-home',
      lifecycleState: 'archived'
    });
    await expect(repository.selectAssetLifecycle('tenant-home', 'inventory-household', 'active')).resolves.toMatchObject({
      assets: []
    });
    await expect(repository.restoreAsset('tenant-home', 'inventory-household', 'asset-home')).resolves.toMatchObject({
      id: 'asset-home',
      lifecycleState: 'active'
    });
    await expect(repository.deleteAsset('tenant-home', 'inventory-household', 'asset-home')).resolves.toBeUndefined();
    await expect(repository.getAsset('tenant-home', 'inventory-household', 'asset-home')).rejects.toThrow('Asset not found');
  });

  it('appends local audit records for mutations and preserves them after deletion', async () => {
    const repository = new SeededInventoryRepository(seed);

    const created = await repository.createAsset('tenant-home', 'inventory-household', {
      kind: 'item',
      title: 'Fire extinguisher',
      description: 'Kitchen',
      parentAssetId: null,
      photos: []
    });
    await repository.updateAsset('tenant-home', 'inventory-household', created.id, {
      title: 'Updated fire extinguisher',
      description: 'Kitchen cabinet',
      parentAssetId: null
    });
    await repository.archiveAsset('tenant-home', 'inventory-household', created.id);
    await repository.restoreAsset('tenant-home', 'inventory-household', created.id);
    await repository.deleteAsset('tenant-home', 'inventory-household', created.id);

    const audit = await repository.listInventoryAuditRecords('tenant-home', 'inventory-household');

    expect(audit.items.map((record) => record.action)).toEqual([
      'asset.deleted',
      'asset.restored',
      'asset.archived',
      'asset.updated',
      'asset.created'
    ]);
    expect(audit.items.every((record) => record.targetId === created.id)).toBe(true);
    await expect(repository.getAsset('tenant-home', 'inventory-household', created.id)).rejects.toThrow('Asset not found');
  });

  it('appends local audit records for tenant and inventory creation', async () => {
    const repository = new SeededInventoryRepository(seed);

    const workspace = await repository.createTenantWithInventory({ tenantName: 'Studio', inventoryName: 'Equipment' });
    const createdTenantId = workspace.context.selectedTenantId;
    await repository.createInventory(createdTenantId, 'Attic');
    const tenantAudit = await repository.listTenantAuditRecords(createdTenantId);

    expect(tenantAudit.items.map((record) => record.action)).toEqual(['inventory.created', 'inventory.created', 'tenant.created']);
    expect(tenantAudit.items[0]).toMatchObject({
      tenantId: createdTenantId,
      action: 'inventory.created',
      targetType: 'inventory'
    });
  });

  it('manages effective custom asset types and field definitions in memory', async () => {
    const repository = new SeededInventoryRepository(seed);

    const assetType = await repository.createCustomAssetType('tenant-home', 'inventory-household', {
      scope: 'inventory',
      key: 'medicine',
      displayName: 'Medicine',
      description: 'Medication'
    });
    const field = await repository.createCustomFieldDefinition('tenant-home', 'inventory-household', {
      scope: 'inventory',
      key: 'expiration-date',
      displayName: 'Expiration date',
      type: 'date',
      enumOptions: [],
      applicability: 'custom_asset_types',
      customAssetTypeIds: [assetType.id]
    });

    await expect(repository.listInventoryCustomAssetTypes('tenant-home', 'inventory-household')).resolves.toMatchObject({
      items: [{ id: assetType.id, displayName: 'Medicine' }]
    });
    await expect(repository.listInventoryCustomFieldDefinitions('tenant-home', 'inventory-household')).resolves.toMatchObject({
      items: [{ id: field.id, customAssetTypeIds: [assetType.id] }]
    });

    await repository.archiveCustomFieldDefinition('tenant-home', 'inventory-household', field.id, 'inventory');
    await repository.archiveCustomAssetType('tenant-home', 'inventory-household', assetType.id, 'inventory');

    await expect(repository.listInventoryCustomAssetTypes('tenant-home', 'inventory-household')).resolves.toMatchObject({ items: [] });
    await expect(repository.listInventoryCustomFieldDefinitions('tenant-home', 'inventory-household')).resolves.toMatchObject({ items: [] });
    await expect(
      repository.createCustomFieldDefinition('tenant-home', 'inventory-household', {
        scope: 'tenant',
        key: 'tenant-expiration',
        displayName: 'Tenant expiration',
        type: 'date',
        enumOptions: [],
        applicability: 'custom_asset_types',
        customAssetTypeIds: [assetType.id]
      })
    ).rejects.toThrow('Custom field target is not available.');
  });

  it('appends local audit records for attachments and access mutations', async () => {
    const repository = new SeededInventoryRepository(seed);
    const uploaded = await repository.uploadAssetPhoto('tenant-home', 'inventory-household', 'asset-home', {
      id: 'selected-photo-one',
      file: new File(['fake'], 'photo.jpg', { type: 'image/jpeg' }),
      name: 'photo.jpg',
      contentType: 'image/jpeg',
      sizeBytes: 4,
      previewUrl: 'blob:photo'
    });

    await repository.archiveAssetAttachment('tenant-home', 'inventory-household', 'asset-home', uploaded.id);
    await repository.restoreAssetAttachment('tenant-home', 'inventory-household', 'asset-home', uploaded.id);
    await repository.deleteAssetAttachment('tenant-home', 'inventory-household', 'asset-home', uploaded.id);
    await repository.grantInventoryAccess('tenant-home', 'inventory-household', 'principal-two', 'viewer');
    await repository.revokeInventoryAccess('tenant-home', 'inventory-household', 'principal-two', 'viewer');
    const invitation = await repository.createInventoryAccessInvitation('tenant-home', 'inventory-household', 'new@example.test', 'editor');
    await repository.updateInventoryAccessInvitationExpiration(
      'tenant-home',
      'inventory-household',
      invitation.invitation.id,
      '2026-07-01T00:00:00Z'
    );
    await repository.cancelInventoryAccessInvitation('tenant-home', 'inventory-household', invitation.invitation.id);
    await repository.deleteInventoryAccessInvitation('tenant-home', 'inventory-household', invitation.invitation.id);

    const audit = await repository.listInventoryAuditRecords('tenant-home', 'inventory-household');

    expect(audit.items.map((record) => record.action)).toEqual([
      'inventory_access_invitation.deleted',
      'inventory_access_invitation.cancelled',
      'inventory_access_invitation.expiration_updated',
      'inventory_access_invitation.created',
      'inventory_access.revoked',
      'inventory_access.granted',
      'asset_photo.deleted',
      'asset_photo.restored',
      'asset_photo.archived',
      'attachment.created'
    ]);
    expect(audit.items.every((record) => record.source === 'local_demo')).toBe(true);
  });

  it('stores uploaded photos as asset attachments with lifecycle controls', async () => {
    const repository = new SeededInventoryRepository(seed);
    const file = new File(['fake'], 'photo.jpg', { type: 'image/jpeg' });

    const attachment = await repository.uploadAssetPhoto('tenant-home', 'inventory-household', 'asset-home', {
      id: 'photo-one',
      name: 'photo.jpg',
      sizeBytes: file.size,
      contentType: 'image/jpeg',
      previewUrl: 'blob:photo-one',
      file
    });

    await expect(repository.listAssetAttachments('tenant-home', 'inventory-household', 'asset-home')).resolves.toMatchObject([
      { id: attachment.id, fileName: 'photo.jpg', lifecycleState: 'active' }
    ]);
    await expect(
      repository.archiveAssetAttachment('tenant-home', 'inventory-household', 'asset-home', attachment.id)
    ).resolves.toMatchObject({ lifecycleState: 'archived' });
    await expect(repository.listAssetAttachments('tenant-home', 'inventory-household', 'asset-home')).resolves.toEqual([]);
    await expect(
      repository.restoreAssetAttachment('tenant-home', 'inventory-household', 'asset-home', attachment.id)
    ).resolves.toMatchObject({ lifecycleState: 'active' });
    await expect(
      repository.deleteAssetAttachment('tenant-home', 'inventory-household', 'asset-home', attachment.id)
    ).resolves.toBeUndefined();
    await expect(repository.listAssetAttachments('tenant-home', 'inventory-household', 'asset-home')).resolves.toEqual([]);
  });

  it('pages authoritative Browse collections across lifecycle state', async () => {
    const repository = new SeededInventoryRepository(seed);

    const first = await repository.browseAssets({
      tenantId: 'tenant-home', inventoryId: 'inventory-household', query: '', tagIds: [], lifecycleState: 'all',
      checkoutState: 'any', scope: 'all', sort: 'id_asc', mode: 'fuzzy', limit: 1
    });
    const second = await repository.browseAssets({
      tenantId: 'tenant-home', inventoryId: 'inventory-household', query: '', tagIds: [], lifecycleState: 'all',
      checkoutState: 'any', scope: 'all', sort: 'id_asc', mode: 'fuzzy', limit: 1, cursor: first.nextCursor ?? undefined
    });

    expect(first.assets).toHaveLength(1);
    expect(first.hasMore).toBe(true);
    expect(second.assets).toHaveLength(1);
    expect(new Set([...first.assets, ...second.assets].map((asset) => asset.lifecycleState))).toEqual(new Set(['active', 'archived']));
  });

  it('checks inventory existence across lifecycle state for honest Browse empty copy', async () => {
    const repository = new SeededInventoryRepository(seed);
    const emptyRepository = new SeededInventoryRepository({ ...seed, assets: [] });

    await expect(repository.hasAnyAssets('tenant-home', 'inventory-household')).resolves.toBe(true);
    await expect(emptyRepository.hasAnyAssets('tenant-home', 'inventory-household')).resolves.toBe(false);
  });
});
