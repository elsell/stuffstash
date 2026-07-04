import { describe, expect, it } from 'vitest';
import type { AssetAttachment, AssetViewModel } from '$lib/domain/inventory';
import { assetActionHref, assetActionIsAvailable, assetDetailHref, attachmentDeleteHref } from './workspaceAssetActions';

describe('workspace asset actions', () => {
  it('builds canonical asset detail and action hrefs', () => {
    expect(assetDetailHref(asset())).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one');
    expect(assetDetailHref({ ...asset(), kind: 'location' })).toBe('/tenants/tenant-one/inventories/inventory-one/locations/asset-one');
    expect(assetActionHref(asset(), 'move')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one/move');
    expect(assetActionHref(asset(), 'archive')).toBe('/tenants/tenant-one/inventories/inventory-one/assets/asset-one/archive');
    expect(assetActionHref({ ...asset(), kind: 'location' }, 'edit')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/locations/asset-one/edit'
    );
  });

  it('builds canonical attachment delete hrefs under the owning asset', () => {
    expect(attachmentDeleteHref(asset(), attachment())).toBe(
      '/tenants/tenant-one/inventories/inventory-one/assets/asset-one/attachments/file-one/delete'
    );
  });

  it('allows actions according to edit permission, saving state, and lifecycle state', () => {
    expect(assetActionIsAvailable(asset(), 'edit', { canEdit: true, saving: false })).toBe(true);
    expect(assetActionIsAvailable(asset(), 'move', { canEdit: false, saving: false })).toBe(false);
    expect(assetActionIsAvailable(asset(), 'delete', { canEdit: true, saving: true })).toBe(false);
    expect(assetActionIsAvailable(asset(), 'restore', { canEdit: true, saving: false })).toBe(false);
    expect(assetActionIsAvailable({ ...asset(), lifecycleState: 'archived' }, 'restore', { canEdit: true, saving: false })).toBe(true);
    expect(assetActionIsAvailable({ ...asset(), lifecycleState: 'archived' }, 'archive', { canEdit: true, saving: false })).toBe(false);
    expect(assetActionIsAvailable({ ...asset(), lifecycleState: 'archived' }, 'delete', { canEdit: true, saving: false })).toBe(true);
  });
});

function asset(): AssetViewModel {
  return {
    id: 'asset-one',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    kind: 'item',
    title: 'Tape measure',
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail: 'Inventory root'
  };
}

function attachment(): AssetAttachment {
  return {
    id: 'file-one',
    tenantId: 'tenant-one',
    inventoryId: 'inventory-one',
    assetId: 'asset-one',
    fileName: 'manual.pdf',
    contentType: 'application/pdf',
    sizeBytes: 1024,
    lifecycleState: 'active'
  };
}
