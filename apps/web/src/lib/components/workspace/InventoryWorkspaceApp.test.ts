import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import { AuthenticationRequiredError } from '$lib/application/authenticationRequired';
import { Toaster } from '$lib/components/ui/sonner/index.js';
import { toast } from 'svelte-sonner';
import {
  ResourcefulImportJobRepository,
  TerminalImportJobRepository,
  seed as importWorkspaceSeed
} from './InventoryImportWorkspace.test-helpers';
import type {
  Asset,
  AssetAttachment,
  AssetLifecycleFilter,
  AuditScope,
  InventoryAccessInvitation,
  InvitationStatusFilter,
  SelectedPhoto,
  UpdateAssetDraft,
  UndoableOperationDirection,
  WorkspaceData
} from '$lib/domain/inventory';
import type { InventoryAccessPage } from '$lib/ports/inventoryAccessRepository';
import type { AuditRecordPage } from '$lib/ports/inventoryAuditRepository';
import type { WorkspaceSeed } from '$lib/ports/inventoryRepository';
import InventoryWorkspaceApp from './InventoryWorkspaceApp.svelte';

const afterNavigateCallbacks = vi.hoisted(() => [] as Array<() => void>);

vi.mock('$app/navigation', () => ({
  afterNavigate: (callback: () => void) => {
    afterNavigateCallbacks.push(callback);
    queueMicrotask(callback);
    return () => {
      const index = afterNavigateCallbacks.indexOf(callback);
      if (index >= 0) {
        afterNavigateCallbacks.splice(index, 1);
      }
    };
  }
}));

let component: ReturnType<typeof mount> | null = null;
let toaster: ReturnType<typeof mount> | null = null;

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

class LifecycleSelectionExpiredRepository extends SeededInventoryRepository {
  async selectAssetLifecycle(
    _tenantId: string,
    _inventoryId: string,
    _lifecycleState: AssetLifecycleFilter
  ): Promise<WorkspaceData> {
    throw new AuthenticationRequiredError();
  }
}

class RefreshFailingAfterReturnRepository extends SeededInventoryRepository {
  private returned = false;

  async returnAsset(...args: Parameters<SeededInventoryRepository['returnAsset']>) {
    const result = await super.returnAsset(...args);
    this.returned = true;
    return { ...result, undoableOperationId: 'operation-home-return' };
  }

  async getAsset(...args: Parameters<SeededInventoryRepository['getAsset']>) {
    if (this.returned) throw new Error('Refresh failed after confirmed return.');
    return super.getAsset(...args);
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

class DelayedAssetRepository extends SeededInventoryRepository {
  releaseAssetLoad: (() => void) | null = null;

  async getAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    if (assetId === 'asset-home' && !this.releaseAssetLoad) {
      await new Promise<void>((resolve) => {
        this.releaseAssetLoad = resolve;
      });
    }
    return super.getAsset(tenantId, inventoryId, assetId);
  }
}

class DelayedBrowseRepository extends SeededInventoryRepository {
  releaseBrowseLoad: (() => void) | null = null;

  async browseAssets(...args: Parameters<SeededInventoryRepository['browseAssets']>) {
    if (!this.releaseBrowseLoad) {
      await new Promise<void>((resolve) => {
        this.releaseBrowseLoad = resolve;
      });
    }
    return super.browseAssets(...args);
  }
}

class UnsafeBrowseRepository extends SeededInventoryRepository {
  private unsafeFailure(): Error & { safeForUser: false } {
    return Object.assign(new Error('database host 10.0.0.8 rejected the query'), { safeForUser: false as const });
  }

  async browseAssets(): Promise<never> {
    throw this.unsafeFailure();
  }

  async loadActiveContainmentMap(): Promise<never> {
    throw this.unsafeFailure();
  }
}

class AssetUpdateFailingRepository extends SeededInventoryRepository {
  createdTagCount = 0;

  async createAssetTag(...args: Parameters<SeededInventoryRepository['createAssetTag']>) {
    this.createdTagCount += 1;
    return super.createAssetTag(...args);
  }

  async updateAsset(
    _tenantId: string,
    _inventoryId: string,
    _assetId: string,
    _draft: UpdateAssetDraft
  ): Promise<Asset> {
    throw new Error('Update failed.');
  }
}

let undoableOperationSequence = 0;

class UndoableCreateRepository extends SeededInventoryRepository {
  directions: string[] = [];
  readonly operationId = `operation-create-${++undoableOperationSequence}`;
  private createdAsset: Asset | null = null;

  async createAsset(...args: Parameters<SeededInventoryRepository['createAsset']>): Promise<Asset> {
    const created = await super.createAsset(...args);
    this.createdAsset = created;
    return { ...created, undoableOperationId: this.operationId };
  }

  async applyAssetOperation(
    tenantId: string,
    inventoryId: string,
    operationId: string,
    direction: UndoableOperationDirection
  ): Promise<Asset> {
    this.directions.push(`${operationId}:${direction}`);
    if (!this.createdAsset) throw new Error('The saved operation is no longer available.');
    const asset = direction === 'undo'
      ? await super.archiveAsset(tenantId, inventoryId, this.createdAsset.id)
      : await super.restoreAsset(tenantId, inventoryId, this.createdAsset.id);
    this.createdAsset = asset;
    return asset;
  }
}

class StaleUndoableCreateRepository extends UndoableCreateRepository {
  async applyAssetOperation(
    _tenantId: string,
    _inventoryId: string,
    operationId: string,
    direction: UndoableOperationDirection
  ): Promise<Asset> {
    this.directions.push(`${operationId}:${direction}`);
    throw Object.assign(new Error('This change is stale because the asset changed later.'), { safeForUser: true as const });
  }
}

class RefreshFailingUndoableCreateRepository extends UndoableCreateRepository {
  private applied = false;

  async applyAssetOperation(...args: Parameters<UndoableCreateRepository['applyAssetOperation']>): Promise<Asset> {
    const result = await super.applyAssetOperation(...args);
    this.applied = true;
    return result;
  }

  async selectAssetLifecycle(...args: Parameters<SeededInventoryRepository['selectAssetLifecycle']>): Promise<WorkspaceData> {
    if (this.applied) throw Object.assign(new Error('database host 10.0.0.8 failed'), { safeForUser: false as const });
    return super.selectAssetLifecycle(...args);
  }
}

class UnsafeUndoableCreateRepository extends UndoableCreateRepository {
  async applyAssetOperation(
    _tenantId: string,
    _inventoryId: string,
    operationId: string,
    direction: UndoableOperationDirection
  ): Promise<Asset> {
    this.directions.push(`${operationId}:${direction}`);
    throw Object.assign(new Error('database host 10.0.0.8 rejected operation'), { safeForUser: false as const });
  }
}

class AuthExpiredRefreshUndoableCreateRepository extends UndoableCreateRepository {
  private applied = false;

  async applyAssetOperation(...args: Parameters<UndoableCreateRepository['applyAssetOperation']>): Promise<Asset> {
    const result = await super.applyAssetOperation(...args);
    this.applied = true;
    return result;
  }

  async selectAssetLifecycle(...args: Parameters<SeededInventoryRepository['selectAssetLifecycle']>): Promise<WorkspaceData> {
    if (this.applied) throw new AuthenticationRequiredError();
    return super.selectAssetLifecycle(...args);
  }
}

async function mountWorkspace(
  path: string,
  repository = new SeededInventoryRepository(structuredClone(seed)),
  options: { onSessionExpired?: () => void } = {}
): Promise<SeededInventoryRepository> {
  window.history.replaceState({}, '', path);
  installMatchMedia();
  toaster = mount(Toaster, { target: document.body });
  component = mount(InventoryWorkspaceApp, {
    target: document.body,
    props: {
      repository,
      initialData: await repository.loadWorkspace(),
      onSignOut: () => {},
      onSessionExpired: options.onSessionExpired
    }
  });
  return repository;
}

function installMatchMedia(): void {
  Object.defineProperty(window, 'matchMedia', {
    configurable: true,
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn()
    }))
  });
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
  toast.dismiss();
  if (component) {
    unmount(component);
    component = null;
  }
  if (toaster) {
    unmount(toaster);
    toaster = null;
  }
  document.body.innerHTML = '';
  window.history.replaceState({}, '', '/');
  afterNavigateCallbacks.splice(0, afterNavigateCallbacks.length);
});

describe('InventoryWorkspaceApp route application', () => {
  it('normalizes the authenticated root workspace to the selected inventory URL', async () => {
    await mountWorkspace('/');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household');
      expect(document.body.textContent).toContain('Recently changed');
    });
  });

  it('removes a returned Home row without depending on a follow-up asset refresh', async () => {
    const repository = new RefreshFailingAfterReturnRepository(structuredClone(seed));
    await repository.checkoutAsset('tenant-home', 'inventory-household', 'asset-home', {});
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household', repository);

    await waitFor(() => expect(controlWithLabel('Return Passport')).toBeTruthy());
    controlWithLabel('Return Passport').click();

    await waitFor(() => {
      expect(document.body.querySelector('[aria-label="Return Passport"]')).toBeNull();
      expect(document.body.textContent).toContain('Returned Passport.');
      expect(controlContaining('Undo')).toBeTruthy();
      expect(document.body.textContent).not.toContain('Refresh failed after confirmed return.');
    });
  });

  it('preserves root lifecycle filters while canonicalizing to the selected inventory URL', async () => {
    await mountWorkspace('/?lifecycle=archived');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household');
      expect(window.location.search).toBe('?lifecycle=archived');
      expect(document.body.textContent).toContain('Archived Passport');
    });
  });

  it('notifies the shell when a workspace API request reports an expired session', async () => {
    let expired = false;
    await mountWorkspace('/?lifecycle=archived', new LifecycleSelectionExpiredRepository(structuredClone(seed)), {
      onSessionExpired: () => {
        expired = true;
      }
    });

    await waitFor(() => {
      expect(expired).toBe(true);
      expect(document.body.textContent).not.toContain('Authentication required.');
    });
  });

  it('does not canonicalize arbitrary catch-all paths as the inventory home', async () => {
    await mountWorkspace('/not-a-workspace-route');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/not-a-workspace-route');
      expect(document.body.textContent).toContain('Recently changed');
    });
  });

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
      expect(document.body.textContent).toContain('Recently changed');
      expect(document.body.textContent).not.toContain('Workspace unavailable');
    });
  });

  it('normalizes unsupported inventory descendant paths to the inventory home', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/assets/asset-home/share?lifecycle=archived');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household');
      expect(window.location.search).toBe('?lifecycle=archived');
      expect(document.body.textContent).toContain('Archived Passport');
      expect(document.body.textContent).not.toContain('PassportBlue folder');
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
      '/tenants/tenant-home/inventories/inventory-household/browse?q=Passport&lifecycle=archived',
      new LifecycleSelectionFailingRepository(structuredClone(seed))
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/browse');
      expect(window.location.search).toBe('?q=Passport&lifecycle=archived');
      expect(document.body.textContent).toContain('Browse');
      expect(document.body.textContent).toContain('Archived Passport');
    });
  });

  it('opens location search results as focused location routes', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/search?q=Garage');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Garage');
      expect(controlContaining('Garage').getAttribute('href')).toBe(
        '/tenants/tenant-home/inventories/inventory-household/locations/location-garage'
      );
    });

    controlContaining('Garage').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
      expect(document.body.textContent).toContain('Main storage area');
    });
  });

  it('updates the Browse sort URL when no query has been submitted', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/search');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/browse');
      expect(document.body.textContent).toContain('Browse');
    });

    controlContaining('Default order').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/browse');
      expect(window.location.search).toBe('?sort=id_asc');
      expect(document.body.textContent).toContain('Browse');
    });
  });

  it('does not expose unsafe Browse list or Map failure details', async () => {
    const repository = new UnsafeBrowseRepository(structuredClone(seed));
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/browse', repository);

    await waitFor(() => {
      expect(document.body.textContent).toContain('Browse could not be loaded. Try again.');
      expect(document.body.textContent).not.toContain('database host');
    });

    window.history.pushState({}, '', '/tenants/tenant-home/inventories/inventory-household/browse?surface=map');
    afterNavigateCallbacks.forEach((callback) => callback());

    await waitFor(() => {
      expect(document.body.textContent).toContain('Map could not be loaded. Try again.');
      expect(document.body.textContent).not.toContain('database host');
    });
  });

  it('does not present an archived-only inventory as a new empty inventory', async () => {
    const archivedOnlySeed = structuredClone(seed);
    archivedOnlySeed.assets = archivedOnlySeed.assets.filter((asset) =>
      asset.tenantId !== 'tenant-home' || asset.inventoryId !== 'inventory-household' || asset.lifecycleState === 'archived'
    );
    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/browse',
      new SeededInventoryRepository(archivedOnlySeed)
    );

    await waitFor(() => {
      expect(document.body.textContent).toContain('Nothing matches these filters');
      expect(document.body.textContent).not.toContain('No stuff here yet');
    });
  });

  it('deep-links to the import workspace', async () => {
    const importSeed = structuredClone(seed);
    importSeed.inventories[0].access.permissions.push('configure', 'view_import_job', 'create_import_job');

    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/import',
      new SeededInventoryRepository(importSeed)
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/import');
      expect(document.body.textContent).toContain('Imports');
      expect(document.body.textContent).toContain('No import runs yet');
    });
  });

  it('opens import detail audit history inside the workspace shell without a reload', async () => {
    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/import',
      new TerminalImportJobRepository(structuredClone(importWorkspaceSeed))
    );

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
    });

    buttonContaining('Details').click();

    await waitFor(() => {
      expect(document.body.querySelector('h1')?.textContent).toBe('Homebox import');
    });

    buttonContaining('More').click();

    await waitFor(() => {
      expect(controlContaining('Open inventory activity').getAttribute('href')).toBe(
        '/tenants/tenant-home/inventories/inventory-household/settings/activity'
      );
      expect(document.body.textContent).toContain('Shows the full inventory activity log.');
    });

    controlContaining('Open inventory activity').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/activity');
      expect(document.body.textContent).toContain('Activity');
      expect(document.body.querySelector('h1')?.textContent).not.toBe('Homebox import');
    });
  });

  it('deep-links to import detail tabs inside the workspace shell', async () => {
    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/import/jobs/job-terminal?tab=records',
      new ResourcefulImportJobRepository(structuredClone(importWorkspaceSeed))
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/import/jobs/job-terminal');
      expect(window.location.search).toBe('?tab=records');
      expect(document.body.querySelector('h1')?.textContent).toBe('Homebox import');
      expect(document.body.textContent).toContain('Imported records');
    });

    buttonContaining('Issues').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/import/jobs/job-terminal');
      expect(window.location.search).toBe('?tab=issues');
      expect(document.body.textContent).toContain('Grouped by cause');
    });
  });

  it('applies workspace route state after client-side navigation without a popstate event', async () => {
    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/import',
      new TerminalImportJobRepository(structuredClone(importWorkspaceSeed))
    );

    await waitFor(() => {
      expect(document.body.textContent).toContain('Imports');
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/import');
    });

    window.history.pushState({}, '', '/tenants/tenant-home/inventories/inventory-household/settings/activity');
    afterNavigateCallbacks.forEach((callback) => callback());

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/activity');
      expect(document.body.textContent).toContain('Activity');
      expect(document.body.textContent).not.toContain('Imports');
    });
  });

  it('opens imported records from import detail inside the workspace shell without a reload', async () => {
    const shellSeed = structuredClone(importWorkspaceSeed);
    shellSeed.assets.push({
      id: 'asset-imported-passport',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      kind: 'item',
      title: 'Imported Passport',
      description: 'Created by import',
      parentAssetId: null,
      lifecycleState: 'active'
    });
    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/import',
      new ResourcefulImportJobRepository(shellSeed)
    );

    await waitFor(() => {
      expect(document.body.textContent).toContain('Runs');
    });

    buttonContaining('Review Details').click();

    await waitFor(() => {
      expect(document.body.querySelector('h1')?.textContent).toBe('Homebox import');
      expect(buttonContaining('Records')).toBeTruthy();
    });

    buttonContaining('Records').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Imported records');
      expect(controlContaining('Open').getAttribute('href')).toBe(
        '/tenants/tenant-home/inventories/inventory-household/assets/asset-imported-passport'
      );
    });

    controlContaining('Open').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-imported-passport');
      expect(document.body.textContent).toContain('Imported Passport');
      expect(document.body.querySelector('h1')?.textContent).not.toBe('Homebox import');
    });
  });

  it('applies the newest client-side route after an older route finishes loading', async () => {
    const repository = new DelayedAssetRepository(structuredClone(seed));
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/assets/asset-home', repository);

    await waitFor(() => {
      expect(repository.releaseAssetLoad).toBeTruthy();
    });

    window.history.pushState({}, '', '/tenants/tenant-home/inventories/inventory-household/settings/activity');
    afterNavigateCallbacks.forEach((callback) => callback());
    repository.releaseAssetLoad?.();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/activity');
      expect(document.body.textContent).toContain('Activity');
      expect(document.body.textContent).not.toContain('Blue folder');
    });
  });

  it('does not leave global actions busy after leaving an in-flight Browse request', async () => {
    const repository = new DelayedBrowseRepository(structuredClone(seed));
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/browse', repository);
    await waitFor(() => expect(repository.releaseBrowseLoad).toBeTruthy());

    document.body.querySelector<HTMLAnchorElement>('.nav-button[href="/tenants/tenant-home/inventories/inventory-household"]')?.click();
    repository.releaseBrowseLoad?.();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household');
      expect(document.body.querySelector<HTMLButtonElement>('.header-add')?.disabled).toBe(false);
    });
  });

  it('keeps the queued route URL when an older alias route canonicalizes after navigation', async () => {
    const repository = new DelayedAssetRepository(structuredClone(seed));
    await mountWorkspace('/inventories/inventory-household/assets/asset-home', repository);

    await waitFor(() => {
      expect(repository.releaseAssetLoad).toBeTruthy();
    });

    window.history.pushState({}, '', '/tenants/tenant-home/inventories/inventory-household/settings/activity');
    afterNavigateCallbacks.forEach((callback) => callback());
    repository.releaseAssetLoad?.();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings/activity');
      expect(document.body.textContent).toContain('Activity');
      expect(document.body.textContent).not.toContain('Blue folder');
    });
  });

  it('preserves a queued non-workspace URL after an older route finishes loading', async () => {
    const repository = new DelayedAssetRepository(structuredClone(seed));
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/assets/asset-home', repository);

    await waitFor(() => {
      expect(repository.releaseAssetLoad).toBeTruthy();
    });

    window.history.pushState({}, '', '/not-a-workspace-route');
    afterNavigateCallbacks.forEach((callback) => callback());
    repository.releaseAssetLoad?.();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/not-a-workspace-route');
      expect(document.body.textContent).not.toContain('Blue folder');
    });
  });

  it('keeps the active route when a queued route is superseded before loading finishes', async () => {
    const repository = new DelayedAssetRepository(structuredClone(seed));
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/assets/asset-home', repository);

    await waitFor(() => {
      expect(repository.releaseAssetLoad).toBeTruthy();
    });

    window.history.pushState({}, '', '/tenants/tenant-home/inventories/inventory-household/settings/activity');
    afterNavigateCallbacks.forEach((callback) => callback());
    window.history.pushState({}, '', '/tenants/tenant-home/inventories/inventory-household/assets/asset-home');
    afterNavigateCallbacks.forEach((callback) => callback());
    repository.releaseAssetLoad?.();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-home');
      expect(document.body.textContent).toContain('Passport');
      expect(document.body.textContent).not.toContain('Activity');
    });
  });

  it('resets the import route to history when the route returns to bare import', async () => {
    const importSeed = structuredClone(seed);
    importSeed.inventories[0].access.permissions.push('configure', 'view_import_job', 'create_import_job');

    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/import/homebox', new SeededInventoryRepository(importSeed));

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/import/homebox');
      expect(document.body.textContent).toContain('New import');
    });

    window.history.pushState({}, '', '/tenants/tenant-home/inventories/inventory-household/import');
    window.dispatchEvent(new PopStateEvent('popstate'));

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/import');
      expect(document.body.textContent).toContain('Imports');
      expect(document.body.textContent).toContain('No import runs yet');
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

  it('normalizes unknown settings section slugs to the overview route', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/settings/nope?invitationStatus=revoked');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/settings');
      expect(window.location.search).toBe('');
      expect(document.body.querySelector('#settings-title')?.textContent).toBe('Settings');
      expect(document.body.textContent).toContain('Overview');
      expect(settingsLink('Overview').getAttribute('aria-current')).toBe('page');
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
      expect(document.body.textContent).toContain('Recently changed');
    });
  });

  it('keeps add tray cancel clicks aligned with the focused location href', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
      expect(document.body.textContent).toContain('Garage');
    });

    controlContaining('Add item here').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');
      expect(window.location.search).toBe('?parent=location-garage');
      expect(document.body.querySelector('[role="dialog"]')?.textContent).toContain('Add item');
    });

    expect(controlContaining('Cancel').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/location-garage'
    );
    controlContaining('Cancel').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
      expect(document.body.querySelector('[role="dialog"]')).toBeNull();
      expect(document.body.textContent).toContain('Main storage area');
    });
  });

  it('keeps global add and feedback overlays outside the product shell', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/add/item');

    await waitFor(() => {
      expect(document.body.querySelector('[role="dialog"]')?.textContent).toContain('Add item');
    });

    const productShell = document.body.querySelector<HTMLElement>('.product-shell');
    const dialog = document.body.querySelector<HTMLElement>('[role="dialog"]');
    expect(productShell).toBeTruthy();
    expect(dialog).toBeTruthy();
    expect(productShell?.contains(dialog)).toBe(false);
    expect(productShell ? isInert(productShell) : false).toBe(true);
    expect(productShell?.getAttribute('aria-hidden')).toBe('true');

    const titleInput = document.body.querySelector<HTMLInputElement>('#asset-title');
    if (!titleInput) throw new Error('Missing add title input');
    setInputValue(titleInput, 'Camera bag');
    await tick();

    const saveButton = await waitForSaveButton();
    saveButton.click();

    await waitFor(() => {
      expect(document.body.querySelector('[role="dialog"]')).toBeNull();
      expect(document.body.textContent).toContain('Saved Camera bag.');
      expect(document.body.textContent).toContain('View asset');
      expect(productShell ? isInert(productShell) : true).toBe(false);
      expect(productShell?.getAttribute('aria-hidden')).toBeNull();
    });

    const toast = document.body.querySelector<HTMLElement>('.stuffstash-toast');
    expect(toast).toBeTruthy();
    expect(productShell?.contains(toast)).toBe(false);
  });

  it('offers target-scoped Undo after a supported mutation and Redo after compensation', async () => {
    const repository = new UndoableCreateRepository(structuredClone(seed));
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/add/item', repository);

    await waitFor(() => expect(document.body.querySelector('[role="dialog"]')).toBeTruthy());
    const titleInput = document.body.querySelector<HTMLInputElement>('#asset-title');
    if (!titleInput) throw new Error('Missing add title input');
    setInputValue(titleInput, 'Camera bag');
    await tick();
    (await waitForSaveButton()).click();

    await waitFor(() => expect(controlContaining('Undo')).toBeTruthy());
    controlContaining('Undo').click();

    await waitFor(() => {
      expect(repository.directions).toEqual([`${repository.operationId}:undo`]);
      expect(document.body.textContent).toContain('Undid change to Camera bag.');
      expect(controlContaining('Redo')).toBeTruthy();
    });

    controlContaining('Redo').click();
    await waitFor(() => {
      expect(repository.directions).toEqual([`${repository.operationId}:undo`, `${repository.operationId}:redo`]);
      expect(document.body.textContent).toContain('Redid change to Camera bag.');
    });
  });

  it('keeps stale Undo failure visible without announcing false success', async () => {
    const repository = new StaleUndoableCreateRepository(structuredClone(seed));
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/add/item', repository);

    await waitFor(() => expect(document.body.querySelector('[role="dialog"]')).toBeTruthy());
    const titleInput = document.body.querySelector<HTMLInputElement>('#asset-title');
    if (!titleInput) throw new Error('Missing add title input');
    setInputValue(titleInput, 'Camera bag');
    await tick();
    (await waitForSaveButton()).click();
    await waitFor(() => expect(controlContaining('Undo')).toBeTruthy());
    controlContaining('Undo').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Couldn’t undo change.');
      expect(document.body.textContent).toContain('This change is stale because the asset changed later.');
      expect(document.body.textContent).not.toContain('Undid change to Camera bag.');
    });
  });

  it('keeps successful Undo and Redo available when the follow-up refresh fails', async () => {
    const repository = new RefreshFailingUndoableCreateRepository(structuredClone(seed));
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/add/item', repository);
    await saveNamedItem('Camera bag');
    await waitFor(() => expect(controlContaining('Undo')).toBeTruthy());

    controlContaining('Undo').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Undid change to Camera bag.');
      expect(document.body.textContent).toContain('Change applied, but this view could not be refreshed.');
      expect(controlContaining('Redo')).toBeTruthy();
    });
  });

  it('does not expose unsafe Undo failure details', async () => {
    const repository = new UnsafeUndoableCreateRepository(structuredClone(seed));
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/add/item', repository);
    await saveNamedItem('Camera bag');
    await waitFor(() => expect(controlContaining('Undo')).toBeTruthy());

    controlContaining('Undo').click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Couldn’t undo change.');
      expect(document.body.textContent).toContain('can’t be applied safely');
      expect(document.body.textContent).not.toContain('database host');
    });
  });

  it('expires the session when Undo succeeds but the reconciliation refresh is unauthenticated', async () => {
    const onSessionExpired = vi.fn();
    const repository = new AuthExpiredRefreshUndoableCreateRepository(structuredClone(seed));
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/add/item', repository, { onSessionExpired });
    await saveNamedItem('Camera bag');
    await waitFor(() => expect(controlContaining('Undo')).toBeTruthy());

    controlContaining('Undo').click();

    await waitFor(() => {
      expect(onSessionExpired).toHaveBeenCalledOnce();
      expect(repository.directions).toEqual([`${repository.operationId}:undo`]);
      expect(document.body.textContent).not.toContain('Couldn’t undo change.');
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

  it('shows an explicit denied state for add deep links without create access', async () => {
    const viewerSeed = structuredClone(seed);
    viewerSeed.inventories[0].access = { relationship: 'viewer', permissions: ['view'] };
    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/add/item',
      new SeededInventoryRepository(viewerSeed)
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');
      expect(document.body.textContent).toContain('Workspace unavailable');
      expect(document.body.textContent).toContain('You do not have permission to add assets in this inventory.');
      expect(document.body.querySelector('[role="dialog"]')).toBeNull();
    });
  });

  it('passes add parent routes into the tray destination picker', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/add/item?parent=location-garage');

    await waitFor(() => {
      const dialog = document.body.querySelector('[role="dialog"]');
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');
      expect(window.location.search).toBe('?parent=location-garage');
      expect(dialog?.textContent).toContain('Add item');
      expect(dialog?.textContent).toContain('Garage');
      expect(document.body.querySelector('.parent-current-card')?.textContent).toContain('Garage');
      expect(document.body.querySelector('.parent-current-card')?.getAttribute('data-selected')).toBe('target');
    });
  });

  it('validates add parent routes after switching to the routed inventory', async () => {
    const multiInventorySeed = structuredClone(seed);
    multiInventorySeed.inventories.push({
      id: 'inventory-yard',
      tenantId: 'tenant-home',
      name: 'Yard',
      access: { relationship: 'owner', permissions: ['view', 'create_asset', 'edit_asset'] }
    });
    multiInventorySeed.assets.push({
      id: 'location-shed',
      tenantId: 'tenant-home',
      inventoryId: 'inventory-yard',
      kind: 'location',
      title: 'Shed',
      description: '',
      parentAssetId: null,
      lifecycleState: 'active'
    });

    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-yard/add/item?parent=location-shed',
      new SeededInventoryRepository(multiInventorySeed)
    );

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-yard/add/item');
      expect(window.location.search).toBe('?parent=location-shed');
      expect(document.body.querySelector('[role="dialog"]')?.textContent).toContain('Shed');
      expect(document.body.querySelector('.parent-current-card')?.getAttribute('data-selected')).toBe('target');
    });
  });

  it('normalizes invalid add parent routes before showing the tray', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/add/item?parent=missing-location');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');
      expect(window.location.search).toBe('');
      expect(document.body.querySelector('[role="dialog"]')?.textContent).toContain('Inventory root');
      expect(document.body.querySelector('.parent-current-card')?.getAttribute('data-selected')).toBe('root');
    });
  });

  it('normalizes invalid add parent routes when canonicalizing inventory aliases', async () => {
    await mountWorkspace('/inventories/inventory-household/add/item?parent=missing-location');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/add/item');
      expect(window.location.search).toBe('');
      expect(document.body.querySelector('[role="dialog"]')?.textContent).toContain('Inventory root');
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

  it('normalizes top-level locations to the Places scope of Browse', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/locations');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/browse');
      expect(window.location.search).toBe('?scope=places');
      expect(document.body.textContent).toContain('Browse');
      expect(document.body.textContent).toContain('Garage');
      expect(document.body.querySelectorAll('[data-recent-card]')).toHaveLength(0);
      expect(document.body.querySelector('[role="tab"][aria-selected="true"]')?.textContent).toContain('List');
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

  it('keeps inline-created edit tags visible when asset update fails', async () => {
    const repository = new AssetUpdateFailingRepository(structuredClone(seed));
    await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/assets/asset-home/edit',
      repository
    );

    await waitFor(() => {
      expect(document.body.textContent).toContain('Edit asset');
      expect(document.body.querySelector<HTMLInputElement>('#new-tag-name')).toBeTruthy();
    });

    setInputValue(document.body.querySelector<HTMLInputElement>('#new-tag-name')!, 'Workshop');
    setInputValue(document.body.querySelector<HTMLInputElement>('#new-tag-color')!, '2f80ed');
    await tick();
    await waitFor(() => {
      expect(tagAddButton().disabled).toBe(false);
    });
    tagAddButton().click();
    await waitFor(() => {
      expect(document.body.querySelector('.pending-tag')?.textContent).toContain('Workshop');
    });
    const saveButton = await waitForSaveButton();
    saveButton.click();

    await waitFor(() => {
      expect(document.body.textContent).toContain('Update failed.');
    });
    expect(repository.createdTagCount).toBe(1);
    await waitFor(() => {
      expect(document.body.querySelector('.pending-tag')).toBeNull();
      expect(selectedTagOption('Workshop')).toBeTruthy();
    });
    saveButton.click();
    await waitFor(() => {
      expect(document.body.textContent).toContain('Update failed.');
    });
    expect(repository.createdTagCount).toBe(1);
    controlContaining('Cancel').click();
    await waitFor(() => {
      expect(document.body.textContent).not.toContain('Edit asset');
    });
    controlContaining('Edit').click();
    await waitFor(() => {
      expect(document.body.querySelector('.tag-options')?.textContent).toContain('Workshop');
    });
  });

  it('keeps ordinary location back clicks aligned with the exposed locations href', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/locations/location-garage');
      expect(document.body.textContent).toContain('Main storage area');
    });

    expect(controlContaining('Back').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household/browse?scope=places');
    controlContaining('Back').click();

    await waitFor(() => {
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/browse');
      expect(window.location.search).toBe('?scope=places');
      expect(document.body.textContent).toContain('Browse');
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

async function saveNamedItem(title: string): Promise<void> {
  await waitFor(() => expect(document.body.querySelector('[role="dialog"]')).toBeTruthy());
  const titleInput = document.body.querySelector<HTMLInputElement>('#asset-title');
  if (!titleInput) throw new Error('Missing add title input');
  setInputValue(titleInput, title);
  await tick();
  (await waitForSaveButton()).click();
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

function tagAddButton(): HTMLButtonElement {
  const fieldset = Array.from(document.body.querySelectorAll('fieldset')).find((candidate) =>
    candidate.textContent?.includes('Tags')
  );
  const button = Array.from(fieldset?.querySelectorAll<HTMLButtonElement>('button') ?? []).find(
    (candidate) => candidate.textContent?.trim() === 'Add'
  );
  if (!button) {
    throw new Error('Missing tag Add button');
  }
  return button;
}

function selectedTagOption(text: string): HTMLButtonElement | undefined {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('.tag-options button')).find(
    (candidate) => candidate.textContent?.includes(text) && candidate.getAttribute('aria-pressed') === 'true'
  );
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

function isInert(element: HTMLElement): boolean {
  const candidate = element as HTMLElement & { inert?: boolean };
  return typeof candidate.inert === 'boolean' ? candidate.inert : element.hasAttribute('inert');
}
