import { describe, expect, it } from 'vitest';
import {
  assetTagDisplayNameByteLength,
  assetTagKeyFromDisplayName,
  canCreateAsset,
  canCreateImportJob,
  canEditAsset,
  canEditInventory,
  canViewImportJobs
} from './inventory';
import type { Inventory } from './inventory';

describe('inventory permissions', () => {
  it('distinguishes asset creation from broader edit capability', () => {
    const editOnlyInventory: Inventory = {
      id: 'inventory-one',
      tenantId: 'tenant-one',
      name: 'Household',
      access: { relationship: 'editor', permissions: ['view', 'edit_asset'] }
    };

    expect(canEditInventory(editOnlyInventory)).toBe(true);
    expect(canEditAsset(editOnlyInventory)).toBe(true);
    expect(canCreateAsset(editOnlyInventory)).toBe(false);
  });

  it('uses explicit import job permissions for import access', () => {
    const importInventory: Inventory = {
      id: 'inventory-one',
      tenantId: 'tenant-one',
      name: 'Household',
      access: { relationship: 'editor', permissions: ['view', 'view_import_job', 'create_import_job'] }
    };
    const viewerInventory: Inventory = {
      ...importInventory,
      access: { relationship: 'viewer', permissions: ['view'] }
    };

    expect(canViewImportJobs(importInventory)).toBe(true);
    expect(canCreateImportJob(importInventory)).toBe(true);
    expect(canViewImportJobs(viewerInventory)).toBe(false);
    expect(canCreateImportJob(viewerInventory)).toBe(false);
  });
});

describe('asset tag keys', () => {
  it('normalizes display names like the backend tag key rule', () => {
    expect(assetTagKeyFromDisplayName(' Camp / Kitchen ')).toBe('camp-kitchen');
    expect(assetTagKeyFromDisplayName('Kids & Toys')).toBe('kids-toys');
    expect(assetTagKeyFromDisplayName('###')).toBe('');
  });

  it('truncates long keys and trims trailing separators', () => {
    expect(assetTagKeyFromDisplayName(`${'a'.repeat(79)} / camping`)).toBe('a'.repeat(79));
    expect(assetTagKeyFromDisplayName(`${'a'.repeat(80)}b`)).toBe('a'.repeat(80));
  });

  it('measures display-name limits with UTF-8 bytes', () => {
    expect(assetTagDisplayNameByteLength('a'.repeat(80))).toBe(80);
    expect(assetTagDisplayNameByteLength('é'.repeat(40))).toBe(80);
    expect(assetTagDisplayNameByteLength(`${'é'.repeat(40)}a`)).toBe(81);
  });
});
