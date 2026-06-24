import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import InventorySettings from './InventorySettings.svelte';
import type { Inventory, Tenant } from '$lib/domain/inventory';
import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('InventorySettings', () => {
  it('uses edit_asset access for asset edit status and disables unsupported future entry points', () => {
    const tenant: Tenant = {
      id: 'tenant-one',
      name: 'Household',
      access: { relationship: 'owner', permissions: ['view', 'configure'] }
    };
    const inventory: Inventory = {
      id: 'inventory-one',
      tenantId: 'tenant-one',
      name: 'Garage',
      access: {
        relationship: 'editor',
        permissions: ['view', 'configure', 'edit_asset']
      }
    };

    component = mount(InventorySettings, {
      target: document.body,
      props: {
        tenant,
        inventory,
        inventoryCount: 2,
        accessRepository: fakeAccessRepository(),
        auditRepository: fakeAuditRepository(),
        customizationRepository: fakeCustomizationRepository(),
        customAssetTypes: [],
        customFieldDefinitions: [],
        onCustomizationChange: () => {}
      }
    });

    expect(document.body.textContent).toContain('Asset editsAllowed');
    expect(document.body.textContent).toContain('You can view this inventory, but you cannot manage sharing.');
    expect(
      Array.from(document.body.querySelectorAll('button')).map((button) => ({
        text: button.textContent,
        disabled: button.disabled
      })).filter((button) => button.text.includes('unavailable'))
    ).toEqual([
      { text: 'Tenant administration unavailable', disabled: true }
    ]);
    expect(document.body.textContent).toContain('Custom fields');
  });

  it('does not treat create-only access as asset edit access', () => {
    const inventory: Inventory = {
      id: 'inventory-one',
      tenantId: 'tenant-one',
      name: 'Garage',
      access: {
        relationship: 'editor',
        permissions: ['view', 'create_asset']
      }
    };

    component = mount(InventorySettings, {
      target: document.body,
      props: {
        tenant: null,
        inventory,
        inventoryCount: 1,
        accessRepository: fakeAccessRepository(),
        auditRepository: fakeAuditRepository(),
        customizationRepository: fakeCustomizationRepository(),
        customAssetTypes: [],
        customFieldDefinitions: [],
        onCustomizationChange: () => {}
      }
    });

    expect(document.body.textContent).toContain('Asset editsView only');
  });
});

function fakeAccessRepository(): InventoryAccessRepository {
  return {
    listInventoryAccessGrants: async () => ({ items: [], pagination: { limit: 50, nextCursor: null, hasMore: false } }),
    grantInventoryAccess: async () => failRepositoryCall(),
    revokeInventoryAccess: async () => failRepositoryCall(),
    listInventoryAccessInvitations: async () => ({ items: [], pagination: { limit: 50, nextCursor: null, hasMore: false } }),
    createInventoryAccessInvitation: async () => failRepositoryCall(),
    updateInventoryAccessInvitationExpiration: async () => failRepositoryCall(),
    cancelInventoryAccessInvitation: async () => failRepositoryCall(),
    deleteInventoryAccessInvitation: async () => failRepositoryCall()
  };
}

function fakeAuditRepository(): InventoryAuditRepository {
  return {
    listTenantAuditRecords: async () => ({ items: [], pagination: { limit: 50, nextCursor: null, hasMore: false } }),
    listInventoryAuditRecords: async () => ({ items: [], pagination: { limit: 50, nextCursor: null, hasMore: false } })
  };
}

function fakeCustomizationRepository(): InventoryCustomizationRepository {
  return {
    listInventoryCustomAssetTypes: async () => ({ items: [], pagination: { limit: 50, nextCursor: null, hasMore: false } }),
    createCustomAssetType: async () => failRepositoryCall(),
    archiveCustomAssetType: async () => failRepositoryCall(),
    listInventoryCustomFieldDefinitions: async () => ({ items: [], pagination: { limit: 50, nextCursor: null, hasMore: false } }),
    createCustomFieldDefinition: async () => failRepositoryCall(),
    archiveCustomFieldDefinition: async () => failRepositoryCall()
  };
}

function failRepositoryCall(): never {
  throw new Error('Unexpected repository call.');
}
