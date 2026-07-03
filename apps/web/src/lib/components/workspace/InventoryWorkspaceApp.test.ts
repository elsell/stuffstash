import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it } from 'vitest';
import { SeededInventoryRepository } from '$lib/adapters/memory/seededInventoryRepository';
import type { Asset, AssetAttachment, SelectedPhoto } from '$lib/domain/inventory';
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
  });

  it('normalizes unavailable asset action routes back to asset detail', async () => {
    await mountWorkspace('/tenants/tenant-home/inventories/inventory-household/assets/asset-archived/move?lifecycle=archived');

    await waitFor(() => {
      expect(document.body.textContent).toContain('Archived Passport');
      expect(window.location.pathname).toBe('/tenants/tenant-home/inventories/inventory-household/assets/asset-archived');
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

  it('closes the add tray after a saved asset with a photo upload warning', async () => {
    const repository = await mountWorkspace(
      '/tenants/tenant-home/inventories/inventory-household/add/item',
      new PhotoUploadFailingRepository(structuredClone(seed))
    );

    await waitFor(() => {
      expect(document.body.querySelector('[role="dialog"]')?.textContent).toContain('Add stuff');
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
      (candidate) => candidate.textContent?.trim() === 'Save'
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
