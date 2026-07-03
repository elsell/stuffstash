import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it } from 'vitest';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import type {
  Asset,
  AssetAttachment,
  AssetLifecycleFilter,
  AuditScope,
  InventoryAccessInvitation,
  InvitationStatusFilter,
  SelectedPhoto,
  WorkspaceData
} from '$lib/domain/inventory';
import type { InventoryAccessPage } from '$lib/ports/inventoryAccessRepository';
import type { AuditRecordPage } from '$lib/ports/inventoryAuditRepository';
import type { WorkspaceSeed } from '$lib/ports/inventoryRepository';
import InventoryWorkspaceApp from './InventoryWorkspaceApp.svelte';

let component: ReturnType<typeof mount> | null = null;

const seed: WorkspaceSeed = {
  principal: { id: 'principal-one', email: 'owner@example.test' },
  tenants: [{ id: 'tenant-home', name: 'Home', access: { relationship: 'owner', permissions: ['view'] } }],
  inventories: [
    {
      id: 'inventory-household',
      tenantId: 'tenant-home',
      name: 'Household',
      access: { relationship: 'owner', permissions: ['view', 'create_asset', 'edit_asset'] }
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
      id: 'location-garage',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'location',
      title: 'Garage',
      description: 'Main storage area',
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

class PhotoUploadFailingRepository extends SeededInventoryRepository {
  async uploadAssetPhoto(
    _tenantId: string,
    _inventoryId: string,
    _assetId: string,
    _photo: SelectedPhoto
  ): Promise<AssetAttachment> {
    throw new Error('Upload failed.');
  }
}

class LifecycleSelectionFailingRepository extends SeededInventoryRepository {
  async selectAssetLifecycle(
    _tenantId: string,
    _inventoryId: string,
    _lifecycleState: AssetLifecycleFilter
  ): Promise<WorkspaceData> {
    throw new Error('Search routes must not mutate the home lifecycle.');
  }
}

class InvitationStatusRecordingRepository extends SeededInventoryRepository {
  invitationStatuses: InvitationStatusFilter[] = [];

  async listInventoryAccessInvitations(
    tenantId: string,
    inventoryId: string,
    status: InvitationStatusFilter,
    cursor?: string
  ): Promise<InventoryAccessPage<InventoryAccessInvitation>> {
    this.invitationStatuses.push(status);
    return super.listInventoryAccessInvitations(tenantId, inventoryId, status, cursor);
  }
}

class AuditScopeRecordingRepository extends SeededInventoryRepository {
  auditScopes: AuditScope[] = [];

  async listTenantAuditRecords(tenantId: string, cursor?: string, signal?: AbortSignal): Promise<AuditRecordPage> {
    this.auditScopes.push('tenant');
    return super.listTenantAuditRecords(tenantId, cursor, signal);
  }

  async listInventoryAuditRecords(
    tenantId: string,
    inventoryId: string,
    cursor?: string,
    signal?: AbortSignal
  ): Promise<AuditRecordPage> {
    this.auditScopes.push('inventory');
    return super.listInventoryAuditRecords(tenantId, inventoryId, cursor, signal);
  }
}

async function mountWorkspace(path: string, repository = new SeededInventoryRepository(structuredClone(seed))): Promise<SeededInventoryRepository> {
  window.history.replaceState({}, '', path);
  component = mount(InventoryWorkspaceApp, {
    target: document.body,
    props: {
      repository,
      initialData: await repository.loadWorkspace(),
      onSignOut: () => {}
    }
  });
  return repository;
}

async function waitFor(assertion: () => void): Promise<void> {
  let lastError: unknown;
  for (let attempt = 0; attempt < 30; attempt += 1) {
    await tick();
    await new Promise((resolve) => window.setTimeout(resolve, 0));
    try {
      assertion();
      return;
    } catch (caught) {
      lastError = caught;
    }
  }
  throw lastError;
}

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
  window.history.replaceState({}, '', '/');
});

describe('InventoryWorkspaceApp route application', () => {
  it('canonicalizes inventory-only asset aliases after loading the asset detail', async () => {
    await mountWorkspace('/inventories/inventory-household/assets/asset-home');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-home');
      expect(document.body.textContent).toContain('Passport');
    });
  });

  it('shows a calm unavailable state for an inventory outside the visible workspace', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/not-visible');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Workspace unavailable');
      expect(document.body.textContent).toContain('That inventory is not available in the current workspace.');
    });

    expect(controlContaining('Go home').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household');
    controlContaining('Go home').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household');
      expect(document.body.textContent).toContain('Recently added');
      expect(document.body.textContent).not.toContain('Workspace unavailable');
    });
  });

  it('disables home add-location controls for inventories without create access', async () => {
    const viewerSeed = structuredClone(seed);
    viewerSeed.inventories[0].access = { relationship: 'viewer', permissions: ['view'] };
    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household',
      new SeededInventoryRepository(viewerSeed)
    );

    await waitFor(() => {
      expect(document.body.textContent).toContain('Garage');
      expect(document.body.textContent).toContain('Creating locations is unavailable for this inventory.');
    });

    const addLocation = controlContaining('Add location');
    expect(addLocation.hasAttribute('href')).toBe(false);
    expect(addLocation.getAttribute('aria-disabled')).toBe('true');
    expect(addLocation.getAttribute('aria-describedby')).toBe('home-add-location-denied');
  });

  it('keeps unavailable recovery clicks aligned with filtered home hrefs', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household?lifecycle=archived');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household');
      expect(window.location.search).toBe('?lifecycle=archived');
      expect(document.body.textContent).toContain('Archived Passport');
    });

    window.history.pushState({}, '', '/tenants/tenant-home/inventories/not-visible?lifecycle=archived');
    window.dispatchEvent(new PopStateEvent('popstate'));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Workspace unavailable');
    });

    expect(controlContaining('Go home').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household?lifecycle=archived'
    );
    controlContaining('Go home').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household');
      expect(window.location.search).toBe('?lifecycle=archived');
      expect(document.body.textContent).toContain('Archived Passport');
      expect(document.body.textContent).not.toContain('PassportBlue folder');
    });
  });

  it('normalizes unavailable asset action routes back to asset detail', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/assets/asset-archived/move?lifecycle=archived');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Archived Passport');
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-archived');
    });
  });

  it('keeps search lifecycle route state independent from the home lifecycle', async () => {
    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/search?q=Passport&lifecycle=archived',
      new LifecycleSelectionFailingRepository(structuredClone(seed))
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/search');
      expect(window.location.search).toBe('?q=Passport&lifecycle=archived');
      expect(document.body.textContent).toContain('Search');
      expect(document.body.textContent).toContain('Archived Passport');
    });
  });

  it('updates the search filter URL when no query has been submitted', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/search');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/search');
      expect(document.body.textContent).toContain('Search this inventory');
    });

    controlContaining('Exact').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/search');
      expect(window.location.search).toBe('?mode=exact');
      expect(document.body.textContent).toContain('Search this inventory');
    });
  });

  it('deep-links and updates the import source route', async () => {
    const importSeed = structuredClone(seed);
    importSeed.inventories[0].access.permissions.push('configure');

    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox-csv',
      new SeededInventoryRepository(importSeed)
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe(
        '/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox-csv'
      );
      expect(document.body.textContent).toContain('CSV file');
      expect(importSourceControl('CSV').getAttribute('href')).toBe(
        '/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox-csv'
      );
      expect(importSourceControl('CSV').getAttribute('aria-current')).toBe('page');
      expect(importSourceControl('Connect').getAttribute('href')).toBe(
        '/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox'
      );
    });

    importSourceControl('Connect').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/import/legacy-homebox');
      expect(document.body.textContent).toContain('Homebox URL');
      expect(importSourceControl('Connect').getAttribute('aria-current')).toBe('page');
    });
  });

  it('deep-links and updates the access invitation status filter', async () => {
    const accessSeed = structuredClone(seed);
    accessSeed.inventories[0].access.permissions.push('share');
    const repository = new InvitationStatusRecordingRepository(accessSeed);

    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/settings/access?invitationStatus=revoked', repository);

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/access');
      expect(window.location.search).toBe('?invitationStatus=revoked');
      expect(document.body.textContent).toContain('Sharing');
      expect(controlContaining('Revoked').getAttribute('href')).toBe(
        '/tenants/tenant-home/inventories/inventory-household/settings/access?invitationStatus=revoked'
      );
      expect(repository.invitationStatuses).toContain('revoked');
    });

    controlContaining('Pending').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/access');
      expect(window.location.search).toBe('?invitationStatus=pending');
      expect(repository.invitationStatuses).toContain('pending');
    });
  });

  it('deep-links access invitation action confirmations', async () => {
    const accessSeed = structuredClone(seed);
    accessSeed.inventories[0].access.permissions.push('share');
    const repository = new SeededInventoryRepository(accessSeed);
    const cancelTarget = await repository.createInventoryAccessInvitation(
      'tenant-home',
      'inventory-household',
      'friend@example.test',
      'viewer'
    );
    await new Promise((resolve) => window.setTimeout(resolve, 2));
    const deleteTarget = await repository.createInventoryAccessInvitation(
      'tenant-home',
      'inventory-household',
      'delete-me@example.test',
      'viewer'
    );

    await mountWorkspace(
      `/tenants/tenant-home/inventories/inventory-household/settings/access/invitations/${cancelTarget.invitation.id}/cancel?invitationStatus=pending`,
      repository
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe(
        `/tenants/tenant-home/inventories/inventory-household/settings/access/invitations/${cancelTarget.invitation.id}/cancel`
      );
      expect(window.location.search).toBe('?invitationStatus=pending');
      expect(document.body.textContent).toContain('Cancel invitation');
      expect(document.body.textContent).toContain('friend@example.test');
      expect(controlContaining('Cancel').getAttribute('href')).toBe(
        '/tenants/tenant-home/inventories/inventory-household/settings/access?invitationStatus=pending'
      );
    });

    controlContaining('Cancel invitation').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/access');
      expect(window.location.search).toBe('?invitationStatus=pending');
      expect(document.body.textContent).not.toContain('Cancel invitation');
      expect(document.body.textContent).not.toContain('friend@example.test');
      expect(document.body.textContent).toContain('delete-me@example.test');
    });

    window.history.pushState(
      {},
      '',
      `/tenants/tenant-home/inventories/inventory-household/settings/access/invitations/${deleteTarget.invitation.id}/delete?invitationStatus=pending`
    );
    window.dispatchEvent(new PopStateEvent('popstate'));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Delete invitation');
      expect(document.body.textContent).toContain('delete-me@example.test');
    });

    buttonContaining('Delete').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/access');
      expect(window.location.search).toBe('?invitationStatus=pending');
      expect(document.body.textContent).not.toContain('Delete invitation');
      expect(document.body.textContent).not.toContain('delete-me@example.test');
    });
  });

  it('does not resurrect invitation status query state from non-access settings routes', async () => {
    const accessSeed = structuredClone(seed);
    accessSeed.inventories[0].access.permissions.push('share');

    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/settings/fields?invitationStatus=revoked', new SeededInventoryRepository(accessSeed));

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/fields');
      expect(settingsLink('Access').getAttribute('href')).toBe(
        '/tenants/tenant-home/inventories/inventory-household/settings/access'
      );
    });

    settingsLink('Access').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/access');
      expect(window.location.search).toBe('');
      expect(document.body.textContent).toContain('Sharing');
    });
  });

  it('deep-links and updates the activity audit scope filter', async () => {
    const auditSeed = structuredClone(seed);
    auditSeed.tenants[0].access.permissions.push('configure');
    const repository = new AuditScopeRecordingRepository(auditSeed);

    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/settings/activity?auditScope=tenant', repository);

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/activity');
      expect(window.location.search).toBe('?auditScope=tenant');
      expect(document.body.textContent).toContain('Activity');
      expect(auditScopeControl('Tenant').getAttribute('href')).toBe(
        '/tenants/tenant-home/inventories/inventory-household/settings/activity?auditScope=tenant'
      );
      expect(repository.auditScopes).toContain('tenant');
    });

    auditScopeControl('Inventory').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/activity');
      expect(window.location.search).toBe('');
      expect(repository.auditScopes).toContain('inventory');
    });
  });

  it('deep-links custom schema archive confirmations from settings fields', async () => {
    const schemaSeed = structuredClone(seed);
    schemaSeed.inventories[0].access.permissions.push('configure');
    schemaSeed.customAssetTypes.push({
      id: 'type-medicine',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      scope: 'inventory',
      key: 'medicine',
      displayName: 'Medicine',
      description: 'Medication',
      lifecycleState: 'active'
    });
    const repository = new SeededInventoryRepository(schemaSeed);

    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/settings/fields/asset-types/type-medicine/archive',
      repository
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe(
        '/tenants/tenant-home/inventories/inventory-household/settings/fields/asset-types/type-medicine/archive'
      );
      expect(document.body.textContent).toContain('Archive asset type');
      expect(document.body.textContent).toContain('Medicine');
      expect(controlContaining('Cancel').getAttribute('href')).toBe(
        '/tenants/tenant-home/inventories/inventory-household/settings/fields'
      );
    });

    controlContaining('Cancel').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/fields');
      expect(document.body.textContent).not.toContain('Archive asset type');
    });

    window.history.pushState(
      {},
      '',
      '/tenants/tenant-home/inventories/inventory-household/settings/fields/asset-types/type-medicine/archive'
    );
    window.dispatchEvent(new PopStateEvent('popstate'));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Archive asset type');
    });

    buttonContaining('Archive').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/fields');
      expect(document.body.textContent).not.toContain('Archive asset type');
      expect(document.body.textContent).not.toContain('Medicine');
    });
  });

  it('keeps add tray cancel clicks aligned with the exposed home href', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/add/item');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');
      expect(document.body.querySelector('[role="dialog"]')?.textContent).toContain('Add item');
    });

    expect(controlContaining('Cancel').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household');
    controlContaining('Cancel').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household');
      expect(document.body.querySelector('[role="dialog"]')).toBeNull();
      expect(document.body.textContent).toContain('Recently added');
    });
  });

  it('passes add kind routes into contextual add tray copy', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/add/location');

    await waitFor(() => {
      const dialog = document.body.querySelector('[role="dialog"]');
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/add/location');
      expect(dialog?.textContent).toContain('Add location');
      expect(document.body.querySelector<HTMLLabelElement>('label[for="asset-title"]')?.textContent).toBe('Location name');
      expect(document.body.querySelector<HTMLInputElement>('#asset-title')?.getAttribute('placeholder')).toBe('Garage shelf');
      expect(buttonContaining('Save location')).toBeTruthy();
    });
  });

  it('applies browser popstate route changes', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Home');
    });

    window.history.pushState({}, '', '/tenants/tenant-home/inventories/inventory-household/assets/asset-home');
    window.dispatchEvent(new PopStateEvent('popstate'));

    await waitFor(() => {
      expect(document.body.textContent).toContain('Passport');
      expect(document.body.textContent).toContain('Blue folder');
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-home');
    });
  });

  it('opens top-level locations through a durable locations route', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/locations');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations');
      expect(document.body.textContent).toContain('Locations');
      expect(document.body.textContent).toContain('The places where your things live.');
      expect(document.body.textContent).toContain('Garage');
      expect(document.body.textContent).not.toContain('Recently added');
      expect(buttonMaybeContaining('Archived')).toBeUndefined();
    });
  });

  it('deep-links location edit from the focused location view', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household');

    await waitFor(() => {
      expect(controlWithLabel('Open location Garage')).toBeTruthy();
    });
    controlWithLabel('Open location Garage').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
      expect(document.body.textContent).toContain('Main storage area');
    });

    expect(controlContaining('Edit location').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/location-garage/edit'
    );
    controlContaining('Edit location').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage/edit');
      expect(document.body.textContent).toContain('Edit asset');
      expect(document.body.querySelector<HTMLInputElement>('#edit-asset-title')?.value).toBe('Garage');
    });

    expect(controlContaining('Cancel').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/location-garage'
    );
    controlContaining('Cancel').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
      expect(document.body.textContent).toContain('Main storage area');
      expect(document.body.textContent).not.toContain('Edit asset');
    });
  });

  it('keeps ordinary location back clicks aligned with the exposed locations href', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
      expect(document.body.textContent).toContain('Main storage area');
    });

    expect(controlContaining('Back').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/locations');
    controlContaining('Back').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations');
      expect(document.body.textContent).toContain('The places where your things live.');
    });
  });

  it('keeps ordinary asset detail back clicks aligned with the exposed previous-location href', async () => {
    const locationSeed = structuredClone(seed);
    locationSeed.assets.push({
      id: 'asset-wrench',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Garage wrench',
      description: 'Hanging by the bench',
      parentAssetId: 'location-garage',
      lifecycleState: 'active'
    });
    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/locations/location-garage',
      new SeededInventoryRepository(locationSeed)
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
      expect(document.body.textContent).toContain('Garage wrench');
    });

    controlContaining('Garage wrench').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-wrench');
      expect(document.body.textContent).toContain('Hanging by the bench');
    });

    expect(controlContaining('Back').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/location-garage'
    );
    controlContaining('Back').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
      expect(document.body.textContent).toContain('Main storage area');
    });
  });

  it('keeps normal location asset edit clicks on the canonical location edit route', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/assets/location-garage');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Garage');
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/location-garage');
    });

    controlContaining('Edit').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage/edit');
      expect(document.body.textContent).toContain('Edit asset');
    });
  });

  it('rejects location edit routes for non-location assets', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/locations/asset-home/edit');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Workspace unavailable');
      expect(document.body.textContent).toContain('That location is not available in this inventory.');
    });
  });

  it('deep-links asset archive and restore confirmations', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/assets/asset-home');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Passport');
    });

    controlContaining('Archive').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-home/archive');
      expect(document.body.textContent).toContain('Archive asset');
    });

    if (component) {
      unmount(component);
      component = null;
    }
    document.body.innerHTML = '';

    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/assets/asset-archived/restore');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-archived/restore');
      expect(document.body.textContent).toContain('Restore asset');
      expect(document.body.textContent).toContain('Archived Passport');
    });
  });

  it('deep-links attachment delete confirmations under the selected asset', async () => {
    const repository = new SeededInventoryRepository(structuredClone(seed));
    const attachment = await repository.uploadAssetAttachment('tenant-home', 'inventory-household', 'asset-home', {
      id: 'selected-manual',
      name: 'manual.pdf',
      sizeBytes: 6,
      contentType: 'application/pdf',
      file: new File(['manual'], 'manual.pdf', { type: 'application/pdf' })
    });

    await mountWorkspace(
      `/tenants/tenant-home/inventories/inventory-household/assets/asset-home/attachments/${attachment.id}/delete`,
      repository
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe(
        `/tenants/tenant-home/inventories/inventory-household/assets/asset-home/attachments/${attachment.id}/delete`
      );
      expect(document.body.textContent).toContain('Passport');
      expect(document.body.textContent).toContain('Delete attachment');
      expect(document.body.textContent).toContain('Delete manual.pdf permanently?');
    });

    expect(controlContaining('Cancel').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/assets/asset-home'
    );
    controlContaining('Cancel').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-home');
      expect(document.body.textContent).not.toContain('Delete attachment');
    });
  });

  it('keeps location attachment delete cancel aligned with the exposed location href', async () => {
    const repository = new SeededInventoryRepository(structuredClone(seed));
    const attachment = await repository.uploadAssetAttachment('tenant-home', 'inventory-household', 'location-garage', {
      id: 'garage-manual',
      name: 'garage-photo.pdf',
      sizeBytes: 6,
      contentType: 'application/pdf',
      file: new File(['manual'], 'garage-photo.pdf', { type: 'application/pdf' })
    });

    await mountWorkspace(
      `/tenants/tenant-home/inventories/inventory-household/assets/location-garage/attachments/${attachment.id}/delete`,
      repository
    );

    await waitFor(() => {
      expect(document.body.textContent).toContain('Delete attachment');
      expect(document.body.textContent).toContain('Delete garage-photo.pdf permanently?');
    });

    expect(controlContaining('Cancel').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/location-garage'
    );
    controlContaining('Cancel').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
      expect(document.body.textContent).toContain('Main storage area');
      expect(document.body.textContent).not.toContain('Delete attachment');
    });
  });

  it('closes the add tray after a saved asset with a photo upload warning', async () => {
    const repository = await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/add/item',
      new PhotoUploadFailingRepository(structuredClone(seed))
    );

    await waitFor(() => {
      expect(document.body.querySelector('[role="dialog"]')?.textContent).toContain('Add item');
    });

    const titleInput = document.body.querySelector<HTMLInputElement>('#asset-title');
    if (!titleInput) throw new Error('Missing add title input');
    setInputValue(titleInput, 'Camera bag');
    await tick();

    const photoInput = document.body.querySelector<HTMLInputElement>('#asset-photos');
    if (!photoInput) throw new Error('Missing photo input');
    Object.defineProperty(photoInput, 'files', {
      value: [new File(['photo'], 'front.jpg', { type: 'image/jpeg' })],
      configurable: true
    });
    photoInput.dispatchEvent(new Event('change', { bubbles: true }));

    const saveButton = await waitForSaveButton();
    saveButton.click();

    await waitFor(() => {
      expect(document.body.querySelector('[role="dialog"]')).toBeNull();
      expect(document.body.textContent).toContain('Camera bag');
      expect(document.body.textContent).toContain('1 photo upload failed');
      expect(window.location.pathname).toMatch(/\/assets\/asset-local-\d+$/);
    });

    const savedAssets = await repository.searchAssets({
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      query: 'Camera bag',
      lifecycleState: 'active',
      mode: 'exact'
    });
    expect(savedAssets).toHaveLength(1);
  });
});

function setInputValue(input: HTMLInputElement, value: string): void {
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set;
  setter?.call(input, value);
  input.dispatchEvent(new Event('input', { bubbles: true }));
}

async function waitForSaveButton(): Promise<HTMLButtonElement> {
  let button: HTMLButtonElement | undefined;
  await waitFor(() => {
    button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find(
      (candidate) => candidate.textContent?.trim().startsWith('Save')
    );
    expect(button).toBeTruthy();
    expect(button?.disabled).toBe(false);
  });
  if (!button) throw new Error('Missing Save button');
  return button;
}

function buttonContaining(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) {
    throw new Error(`Missing button containing ${text}`);
  }
  return button;
}

function controlContaining(text: string): HTMLElement {
  const control = Array.from(document.body.querySelectorAll<HTMLElement>('button, a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!control) {
    throw new Error(`Missing control containing ${text}`);
  }
  return control;
}

function settingsLink(label: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('.settings-section-link')).find((candidate) =>
    candidate.textContent?.trim().startsWith(label)
  );
  if (!link) {
    throw new Error(`Missing settings link ${label}`);
  }
  return link;
}

function auditScopeControl(label: string): HTMLElement {
  const group = document.body.querySelector<HTMLElement>('[role="group"][aria-label="Audit scope"]');
  const control = Array.from(group?.querySelectorAll<HTMLElement>('button, a') ?? []).find((candidate) => candidate.textContent === label);
  if (!control) {
    throw new Error(`Missing audit scope control ${label}`);
  }
  return control;
}

function importSourceControl(label: string): HTMLElement {
  const group = document.body.querySelector<HTMLElement>('[role="group"][aria-label="Import source"]');
  const control = Array.from(group?.querySelectorAll<HTMLElement>('button, a') ?? []).find((candidate) => candidate.textContent === label);
  if (!control) {
    throw new Error(`Missing import source control ${label}`);
  }
  return control;
}

function buttonMaybeContaining(text: string): HTMLButtonElement | undefined {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
}

function buttonWithLabel(label: string): HTMLButtonElement {
  const button = document.body.querySelector<HTMLButtonElement>(`button[aria-label="${label}"]`);
  if (!button) {
    throw new Error(`Missing button labelled ${label}`);
  }
  return button;
}

function controlWithLabel(label: string): HTMLElement {
  const control = document.body.querySelector<HTMLElement>(`button[aria-label="${label}"], a[aria-label="${label}"]`);
  if (!control) {
    throw new Error(`Missing control labelled ${label}`);
  }
  return control;
}
