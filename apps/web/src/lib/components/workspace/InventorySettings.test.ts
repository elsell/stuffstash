import { afterEach, describe, expect, it, vi } from 'vitest';
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
  it('uses edit_asset access for asset edit status and routes between focused settings sections', async () => {
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

    const onSectionChange = vi.fn();
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
        section: 'overview',
        invitationStatus: 'all',
        onSectionChange,
        onInvitationStatusChange: () => {},
        onCustomizationChange: () => {}
      }
    });

    expect(document.body.textContent).toContain('Asset editsAllowed');
    const settingsNav = document.body.querySelector<HTMLElement>('.settings-section-nav');
    expect(settingsNav?.tagName).toBe('NAV');
    expect(settingsNav?.getAttribute('aria-label')).toBe('Settings sections');
    expect(linkStartingWith('Overview').getAttribute('aria-current')).toBe('page');
    expect(linkStartingWith('Fields').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/fields');
    expect(document.body.textContent).toContain('Custom asset types and fields');
    expect(document.body.querySelector('.settings-section-context')?.textContent).toContain('Overview');
    expect(document.body.querySelector('.settings-section-context')?.textContent).toContain('Inventory context and access summary');

    linkStartingWith('Fields').click();
    expect(onSectionChange).toHaveBeenCalledWith('fields');
  });

  it('shows administration as a focused settings section', () => {
    component = mount(InventorySettings, {
      target: document.body,
      props: {
        tenant: { id: 'tenant-one', name: 'Household', access: { relationship: 'owner', permissions: ['view'] } },
        inventory: {
          id: 'inventory-one',
          tenantId: 'tenant-one',
          name: 'Garage',
          access: { relationship: 'viewer', permissions: ['view'] }
        },
        inventoryCount: 1,
        accessRepository: fakeAccessRepository(),
        auditRepository: fakeAuditRepository(),
        customizationRepository: fakeCustomizationRepository(),
        customAssetTypes: [],
        customFieldDefinitions: [],
        section: 'administration',
        onSectionChange: () => {},
        onInvitationStatusChange: () => {},
        onCustomizationChange: () => {}
      }
    });

    expect(document.body.textContent).toContain('Tenant administration unavailable');
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
        section: 'overview',
        onSectionChange: () => {},
        onInvitationStatusChange: () => {},
        onCustomizationChange: () => {}
      }
    });

    expect(document.body.textContent).toContain('Asset editsView only');
    expect(linkStartingWith('Activity').getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings/activity');
  });

  it('passes route-backed access invitation status links through settings', async () => {
    const tenant: Tenant = {
      id: 'tenant-one',
      name: 'Household',
      access: { relationship: 'owner', permissions: ['view'] }
    };
    const inventory: Inventory = {
      id: 'inventory-one',
      tenantId: 'tenant-one',
      name: 'Garage',
      access: { relationship: 'owner', permissions: ['view', 'share'] }
    };
    let selectedStatus = '';

    component = mount(InventorySettings, {
      target: document.body,
      props: {
        tenant,
        inventory,
        inventoryCount: 1,
        accessRepository: fakeAccessRepository(),
        auditRepository: fakeAuditRepository(),
        customizationRepository: fakeCustomizationRepository(),
        customAssetTypes: [],
        customFieldDefinitions: [],
        section: 'access',
        invitationStatus: 'revoked',
        onSectionChange: () => {},
        onInvitationStatusChange: (status) => {
          selectedStatus = status;
        },
        onCustomizationChange: () => {}
      }
    });
    await flush();

    expect(linkStartingWith('Revoked').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/settings/access?invitationStatus=revoked'
    );
    expect(linkStartingWith('Revoked').getAttribute('aria-current')).toBe('page');

    linkStartingWith('Pending').click();
    expect(selectedStatus).toBe('pending');
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

function linkStartingWith(text: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.trim().startsWith(text)
  );
  if (!link) {
    throw new Error(`Missing link starting with ${text}`);
  }
  return link;
}

async function flush(): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 0));
}
