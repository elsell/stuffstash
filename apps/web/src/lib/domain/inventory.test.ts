import { describe, expect, it } from 'vitest';
import { canCreateAsset, canCreateImportJob, canEditAsset, canEditInventory, canViewImportJobs } from './inventory';
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
