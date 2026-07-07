import { describe, expect, it } from 'vitest';
import type {
  AddAssetDraft,
  Asset,
  AssetAttachment,
  AssetTag,
  Inventory,
  SelectedPhoto,
  WorkspaceData
} from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';
import { createAssetWorkflow, replaceWorkspaceAsset } from './workspaceAssetWorkflow';

const inventory: Inventory = {
  id: 'inventory-household',
  tenantId: 'tenant-home',
  name: 'Household',
  access: { relationship: 'owner', permissions: ['view', 'create_asset', 'edit_asset'] }
};

function workspaceData(
  assets: Asset[] = [],
  lifecycleState: WorkspaceData['context']['assetLifecycleState'] = 'active',
  assetTags: AssetTag[] = []
): WorkspaceData {
  return {
    context: {
      principal: { id: 'principal-one', email: 'owner@example.test' },
      tenants: [{ id: 'tenant-home', name: 'Home', access: { relationship: 'owner', permissions: ['view'] } }],
      inventories: [inventory],
      selectedTenantId: 'tenant-home',
      selectedInventoryId: inventory.id,
      assetLifecycleState: lifecycleState,
      mediaUploadPolicy: { supportedContentTypes: ['image/jpeg'], maxBytes: 1024 },
      customAssetTypes: [],
      customFieldDefinitions: [],
      assetTags,
      capability: 'editor'
    },
    assets,
    checkedOutAssets: []
  };
}

function asset(id: string, title = id, kind: Asset['kind'] = 'item'): Asset {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: inventory.id,
    kind,
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active'
  };
}

function assetTag(id: string, displayName: string, color?: string): AssetTag {
  return {
    id,
    key: displayName.toLowerCase().replaceAll(' ', '-'),
    displayName,
    color
  };
}

function photo(): SelectedPhoto {
  return selectedPhoto('front.jpg', 'blob:front');
}

function selectedPhoto(name: string, previewUrl: string): SelectedPhoto {
  return {
    id: name,
    name,
    sizeBytes: 12,
    contentType: 'image/jpeg',
    previewUrl,
    file: new File(['photo'], name, { type: 'image/jpeg' })
  };
}

describe('workspace asset workflow', () => {
  it('creates an asset with a quick parent and uploaded primary photo', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('parent-one', 'Garage bin', 'container'), asset('asset-one', 'Tape measure')],
      uploadedPhotos: [attachment('photo-one', 'asset-one')]
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      parentQuickCreate: { kind: 'container', title: 'Garage bin' },
      customFields: {},
      photos: [photo()]
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.closeAdd).toBe(true);
    expect(result.message).toBe('Saved Tape measure in Garage bin with 1 photo upload.');
    expect(result.route).toMatchObject({ mode: 'asset', assetId: 'asset-one' });
    expect(result.data.assets.map((item) => item.id)).toEqual(['asset-one', 'parent-one']);
    expect(result.data.assets[0]?.photo).toMatchObject({ id: 'photo-one', assetId: 'asset-one', url: 'blob:front' });
  });

  it('treats photo upload failure as a saved asset with a warning', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('asset-one', 'Tape measure')],
      uploadFailure: true
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      customFields: {},
      photos: [photo()]
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.closeAdd).toBe(true);
    expect(result.message).toBe('Saved Tape measure. 1 photo upload failed.');
    expect(result.route).toMatchObject({ mode: 'asset', assetId: 'asset-one' });
    expect(result.selectedAsset?.id).toBe('asset-one');
    expect(result.data.assets.map((item) => item.id)).toEqual(['asset-one']);
  });

  it('includes safe photo upload failure reasons when creating an asset', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('asset-one', 'Tape measure')],
      uploadFailure: true,
      uploadFailureError: safeUploadError('Attachment content does not match its file type.')
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      customFields: {},
      photos: [photo()]
    });

    expect(result.message).toBe(
      'Saved Tape measure. 1 photo upload failed. Attachment content does not match its file type.'
    );
  });

  it('deduplicates repeated safe photo upload failure reasons', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('asset-one', 'Tape measure')],
      uploadFailure: true,
      uploadFailureError: safeUploadError('Attachment content does not match its file type.')
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      customFields: {},
      photos: [selectedPhoto('front.jpg', 'blob:front'), selectedPhoto('back.jpg', 'blob:back')]
    });

    expect(result.message).toBe(
      'Saved Tape measure. 2 photo uploads failed. Attachment content does not match its file type.'
    );
  });

  it('routes created locations to the focused location surface', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('location-one', 'Garage shelf', 'location')]
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'location',
      title: 'Garage shelf',
      description: '',
      parentAssetId: null,
      customFields: {},
      photos: []
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.closeAdd).toBe(true);
    expect(result.mode).toBe('location');
    expect(result.selectedAsset?.id).toBe('location-one');
    expect(result.route).toMatchObject({ mode: 'location', locationId: 'location-one' });
  });

  it('keeps quick-created parent context when photo upload fails', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('parent-one', 'Garage bin', 'container'), asset('asset-one', 'Tape measure')],
      uploadFailure: true
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      parentQuickCreate: { kind: 'container', title: 'Garage bin' },
      customFields: {},
      photos: [photo()]
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.message).toBe('Saved Tape measure in Garage bin. 1 photo upload failed.');
  });

  it('uses count-aware photo saved feedback', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('asset-one', 'Tape measure')],
      uploadedPhotos: [attachment('photo-one', 'asset-one'), attachment('photo-two', 'asset-one')]
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      customFields: {},
      photos: [selectedPhoto('front.jpg', 'blob:front'), selectedPhoto('back.jpg', 'blob:back')]
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.message).toBe('Saved Tape measure with 2 photo uploads.');
  });

  it('reports mixed photo upload outcomes and pairs the primary photo with the upload that succeeded', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('asset-one', 'Tape measure')],
      uploadedPhotos: [attachment('photo-two', 'asset-one')],
      failedUploadIndexes: [0]
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      customFields: {},
      photos: [selectedPhoto('front.jpg', 'blob:front'), selectedPhoto('back.jpg', 'blob:back')]
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.message).toBe('Saved Tape measure with 1 photo upload. 1 photo upload failed.');
    expect(result.selectedAsset?.photo).toMatchObject({ id: 'photo-two', assetId: 'asset-one', url: 'blob:back' });
  });

  it('keeps quick-created parent context for mixed photo upload outcomes', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('parent-one', 'Garage bin', 'container'), asset('asset-one', 'Tape measure')],
      uploadedPhotos: [attachment('photo-two', 'asset-one')],
      failedUploadIndexes: [0]
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      parentQuickCreate: { kind: 'container', title: 'Garage bin' },
      customFields: {},
      photos: [selectedPhoto('front.jpg', 'blob:front'), selectedPhoto('back.jpg', 'blob:back')]
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.message).toBe('Saved Tape measure in Garage bin with 1 photo upload. 1 photo upload failed.');
  });

  it('uses plural failure feedback when multiple photo uploads fail', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('asset-one', 'Tape measure')],
      uploadFailure: true
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      customFields: {},
      photos: [selectedPhoto('front.jpg', 'blob:front'), selectedPhoto('back.jpg', 'blob:back')]
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.message).toBe('Saved Tape measure. 2 photo uploads failed.');
  });

  it('returns the created parent id when child creation fails', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('parent-one', 'Garage bin', 'container')],
      createFailureAfter: 1
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      parentQuickCreate: { kind: 'container', title: 'Garage bin' },
      customFields: {},
      photos: []
    });

    expect(result.saveResult).toEqual({ saved: false, createdParentId: 'parent-one' });
    expect(result.error).toBe('Created Garage bin, but could not save Tape measure. Create failed.');
    expect(result.data.assets.map((item) => item.id)).toEqual(['parent-one']);
  });

  it('keeps inline-created tags visible when asset creation fails', async () => {
    const repository = fakeRepository({
      createdAssets: [],
      createdTags: [assetTag('tag-workshop', 'Workshop', '#2F80ED')],
      createFailureAfter: 0
    });

    const result = await createAssetWorkflow(repository, workspaceData(), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      customFields: {},
      tagIds: [],
      newTags: [{ displayName: 'Workshop', color: '#2f80ed' }],
      photos: []
    });

    expect(result.saveResult).toEqual({ saved: false });
    expect(result.closeAdd).toBe(false);
    expect(result.data.context.assetTags).toMatchObject([
      { id: 'tag-workshop', key: 'workshop', displayName: 'Workshop', color: '#2F80ED' }
    ]);
  });

  it('reuses known inventory tags before creating asset drafts', async () => {
    const createdDrafts: AddAssetDraft[] = [];
    const repository = fakeRepository({
      createdAssets: [asset('asset-one', 'Tape measure')],
      onCreateAsset: (draft) => {
        createdDrafts.push(draft);
      }
    });

    const result = await createAssetWorkflow(repository, workspaceData([], 'active', [assetTag('tag-ted', 'Ted')]), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      customFields: {},
      tagIds: [],
      newTags: [{ displayName: 'ted' }],
      photos: []
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.data.assets[0]?.id).toBe('asset-one');
    expect(createdDrafts[0]?.tagIds).toEqual(['tag-ted']);
  });

  it('switches back to active lifecycle when creating from an archived view', async () => {
    const activeData = workspaceData([asset('asset-one', 'Tape measure')], 'active');
    const repository = fakeRepository({
      createdAssets: [asset('asset-one', 'Tape measure')],
      uploadedPhotos: [attachment('photo-one', 'asset-one')],
      selectedLifecycleData: activeData
    });

    const result = await createAssetWorkflow(repository, workspaceData([], 'archived'), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      customFields: {},
      photos: [photo()]
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.data).toBe(activeData);
    expect(result.mode).toBe('asset');
    expect(result.selectedAsset?.photo).toMatchObject({ id: 'photo-one', assetId: 'asset-one', url: 'blob:front' });
    expect(result.route).toMatchObject({ mode: 'asset', assetId: 'asset-one' });
  });

  it('does not reopen a duplicate-create path when active lifecycle refresh fails after create', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('asset-one', 'Tape measure')],
      selectLifecycleFailure: true
    });

    const result = await createAssetWorkflow(repository, workspaceData([], 'archived'), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      customFields: {},
      photos: []
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.closeAdd).toBe(true);
    expect(result.error).toBe('Saved Tape measure, but could not refresh the active view. Refresh failed.');
    expect(result.selectedAsset?.id).toBe('asset-one');
    expect(result.route).toMatchObject({ mode: 'asset', assetId: 'asset-one' });
  });

  it('keeps created locations focused when active lifecycle refresh fails after create', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('location-one', 'Garage shelf', 'location')],
      selectLifecycleFailure: true
    });

    const result = await createAssetWorkflow(repository, workspaceData([], 'archived'), inventory, {
      kind: 'location',
      title: 'Garage shelf',
      description: '',
      parentAssetId: null,
      customFields: {},
      photos: []
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.closeAdd).toBe(true);
    expect(result.error).toBe('Saved Garage shelf, but could not refresh the active view. Refresh failed.');
    expect(result.mode).toBe('location');
    expect(result.selectedAsset?.id).toBe('location-one');
    expect(result.route).toMatchObject({ mode: 'location', locationId: 'location-one' });
  });

  it('keeps quick parent and photo context when active lifecycle refresh fails after create', async () => {
    const repository = fakeRepository({
      createdAssets: [asset('parent-one', 'Garage bin', 'container'), asset('asset-one', 'Tape measure')],
      uploadedPhotos: [attachment('photo-one', 'asset-one')],
      selectLifecycleFailure: true
    });

    const result = await createAssetWorkflow(repository, workspaceData([], 'archived'), inventory, {
      kind: 'item',
      title: 'Tape measure',
      description: '',
      parentAssetId: null,
      parentQuickCreate: { kind: 'container', title: 'Garage bin' },
      customFields: {},
      photos: [photo()]
    });

    expect(result.saveResult).toEqual({ saved: true });
    expect(result.message).toBe('Saved Tape measure in Garage bin with 1 photo upload.');
    expect(result.error).toBe('Saved Tape measure, but could not refresh the active view. Refresh failed.');
    expect(result.selectedAsset?.photo).toMatchObject({ id: 'photo-one', assetId: 'asset-one', url: 'blob:front' });
    expect(result.route).toMatchObject({ mode: 'asset', assetId: 'asset-one' });
  });

  it('replaces only assets in the selected tenant, inventory, and lifecycle', () => {
    const original = asset('asset-one', 'Old title');
    const updated = { ...original, title: 'New title' };
    const otherInventory = { ...original, inventoryId: 'inventory-other', title: 'Wrong inventory' };
    const archived = { ...original, lifecycleState: 'archived' as const, title: 'Archived title' };

    expect(replaceWorkspaceAsset(workspaceData([original]), updated).assets[0]?.title).toBe('New title');
    expect(replaceWorkspaceAsset(workspaceData([original]), otherInventory).assets[0]?.title).toBe('Old title');
    expect(replaceWorkspaceAsset(workspaceData([original]), archived).assets[0]?.title).toBe('Old title');
  });
});

function attachment(id: string, assetId: string): AssetAttachment {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: inventory.id,
    assetId,
    fileName: 'front.jpg',
    contentType: 'image/jpeg',
    sizeBytes: 12,
    lifecycleState: 'active',
    thumbnailUrl: 'blob:thumbnail'
  };
}

function fakeRepository({
  createdAssets,
  uploadedPhotos = [],
  uploadFailure = false,
  uploadFailureError = new Error('Upload failed.'),
  failedUploadIndexes = [],
  createFailureAfter,
  createdTags = [],
  selectedLifecycleData,
  selectLifecycleFailure = false,
  onCreateAsset
}: {
  createdAssets: Asset[];
  uploadedPhotos?: AssetAttachment[];
  uploadFailure?: boolean;
  uploadFailureError?: Error;
  failedUploadIndexes?: number[];
  createFailureAfter?: number;
  createdTags?: AssetTag[];
  selectedLifecycleData?: WorkspaceData;
  selectLifecycleFailure?: boolean;
  onCreateAsset?: (draft: AddAssetDraft) => void;
}): Pick<InventoryRepository, 'createAsset' | 'selectAssetLifecycle' | 'uploadAssetPhoto' | 'createAssetTag'> {
  let createCount = 0;
  let uploadCount = 0;
  let tagCreateCount = 0;
  return {
    async createAssetTag(): Promise<AssetTag> {
      const created = createdTags[tagCreateCount];
      tagCreateCount += 1;
      if (!created) {
        throw new Error('Missing created tag fixture.');
      }
      return created;
    },
    async createAsset(_tenantId: string, _inventoryId: string, draft: AddAssetDraft): Promise<Asset> {
      onCreateAsset?.(draft);
      if (createFailureAfter !== undefined && createCount >= createFailureAfter) {
        throw new Error('Create failed.');
      }
      const created = createdAssets[createCount];
      createCount += 1;
      if (!created) {
        throw new Error('Missing created asset fixture.');
      }
      return { ...created, parentAssetId: draft.parentAssetId };
    },
    async selectAssetLifecycle(): Promise<WorkspaceData> {
      if (selectLifecycleFailure) {
        throw new Error('Refresh failed.');
      }
      return selectedLifecycleData ?? workspaceData(createdAssets);
    },
    async uploadAssetPhoto(): Promise<AssetAttachment> {
      if (uploadFailure || failedUploadIndexes.includes(uploadCount)) {
        uploadCount += 1;
        throw uploadFailureError;
      }
      const uploaded = uploadedPhotos.shift();
      uploadCount += 1;
      if (!uploaded) {
        throw new Error('Missing uploaded photo fixture.');
      }
      return uploaded;
    }
  };
}

function safeUploadError(message: string): Error & { safeForUser: true } {
  const error = new Error(message) as Error & { safeForUser: true };
  error.safeForUser = true;
  return error;
}
