import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import HomeboxImportPanel from './HomeboxImportPanel.svelte';
import HomeboxImportPanelHarness from './HomeboxImportPanel.test-harness.svelte';
import type { ImportPreview, Inventory, LegacyHomeboxImportRequest } from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('HomeboxImportPanel', () => {
  it('exposes import source choices as durable workspace links', async () => {
    let selectedSource: string | null = null;
    component = mount(HomeboxImportPanel, {
      target: document.body,
      props: {
        tenantId: 'tenant-one',
        inventory: inventory(),
        repository: fakeRepository(),
        sourceType: 'legacy_homebox_csv',
        onSourceChange: (sourceType) => {
          selectedSource = sourceType;
        },
        onImported: async () => {}
      }
    });
    await flush();

    expect(sourceControl('CSV').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/import/legacy-homebox-csv'
    );
    expect(sourceControl('CSV').getAttribute('aria-current')).toBe('page');
    expect(sourceControl('Connect').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/import/legacy-homebox'
    );
    expect(document.body.textContent).toContain('Homebox CSV export');

    sourceControl('Connect').click();
    await flush();

    expect(selectedSource).toBe('legacy_homebox');
    expect(document.body.textContent).toContain('Live Homebox API');
  });

  it('clears preview state when a parent-driven route changes the import source', async () => {
    component = mount(HomeboxImportPanelHarness, {
      target: document.body,
      props: {
        tenantId: 'tenant-one',
        inventory: inventory(),
        repository: fakeRepository(),
        onImported: async () => {}
      }
    });
    await flush();

    input('#homebox-url', 'https://homebox.local');
    input('#homebox-username', 'owner');
    input('#homebox-password', 'secret');
    await flush();
    click('Preview');
    await flush();

    expect(document.body.textContent).toContain('Field definitions');

    click('Switch source externally');
    await flush();

    expect(document.body.textContent).toContain('CSV file');
    expect(document.body.textContent).toContain('Preview an import');
    expect(document.body.textContent).not.toContain('Field definitions');
  });

  it('uses explicit switch controls for live import options', async () => {
    component = mount(HomeboxImportPanel, {
      target: document.body,
      props: {
        tenantId: 'tenant-one',
        inventory: inventory(),
        repository: fakeRepository(),
        onImported: async () => {}
      }
    });
    await flush();

    expect(switchControl('Images')?.getAttribute('aria-checked')).toBe('true');
    expect(switchControl('Images')?.textContent).toContain('Import Homebox image attachments when available.');
    expect(switchControl('Self-signed certificate')?.getAttribute('aria-checked')).toBe('false');
    expect(switchControl('Self-signed certificate')?.textContent).toContain('untrusted TLS certificate');
    expect(switchControl('Private network address')?.getAttribute('aria-checked')).toBe('false');
    expect(switchControl('Private network address')?.textContent).toContain('private LAN addresses');

    switchControl('Images')?.click();
    switchControl('Self-signed certificate')?.click();
    await flush();

    expect(switchControl('Images')?.getAttribute('aria-checked')).toBe('false');
    expect(switchControl('Images')?.textContent).toContain('Off');
    expect(switchControl('Self-signed certificate')?.getAttribute('aria-checked')).toBe('true');
    expect(switchControl('Self-signed certificate')?.textContent).toContain('On');
  });

  it('submits switch state through the legacy Homebox preview request', async () => {
    let previewRequest: LegacyHomeboxImportRequest | null = null;
    component = mount(HomeboxImportPanel, {
      target: document.body,
      props: {
        tenantId: 'tenant-one',
        inventory: inventory(),
        repository: fakeRepository((request) => {
          previewRequest = request;
        }),
        onImported: async () => {}
      }
    });
    await flush();

    input('#homebox-url', 'https://homebox.local');
    input('#homebox-username', 'owner');
    input('#homebox-password', 'secret');
    switchControl('Images')?.click();
    switchControl('Self-signed certificate')?.click();
    switchControl('Private network address')?.click();
    await flush();
    click('Preview');
    await flush();

    expect(previewRequest).toMatchObject({
      sourceType: 'legacy_homebox',
      baseUrl: 'https://homebox.local',
      username: 'owner',
      password: 'secret',
      includeImages: false,
      allowInsecureTLS: true,
      allowPrivateNetwork: true
    });
  });
});

function inventory(): Inventory {
  return {
    id: 'inventory-one',
    tenantId: 'tenant-one',
    name: 'Household',
    access: { relationship: 'owner', permissions: ['view', 'configure'] }
  };
}

function fakeRepository(onPreview?: (request: LegacyHomeboxImportRequest) => void): InventoryRepository {
  return {
    loadWorkspace: async () => failRepositoryCall(),
    createTenantWithInventory: async () => failRepositoryCall(),
    createInventory: async () => failRepositoryCall(),
    selectTenant: async () => failRepositoryCall(),
    selectInventory: async () => failRepositoryCall(),
    selectAssetLifecycle: async () => failRepositoryCall(),
    getAsset: async () => failRepositoryCall(),
    updateAsset: async () => failRepositoryCall(),
    createAsset: async () => failRepositoryCall(),
    archiveAsset: async () => failRepositoryCall(),
    restoreAsset: async () => failRepositoryCall(),
    deleteAsset: async () => failRepositoryCall(),
    listAssetAttachments: async () => failRepositoryCall(),
    uploadAssetAttachment: async () => failRepositoryCall(),
    uploadAssetPhoto: async () => failRepositoryCall(),
    archiveAssetAttachment: async () => failRepositoryCall(),
    restoreAssetAttachment: async () => failRepositoryCall(),
    deleteAssetAttachment: async () => failRepositoryCall(),
    searchAssets: async () => failRepositoryCall(),
    previewLegacyHomeboxImport: async (_tenantId, _inventoryId, request) => {
      onPreview?.(request);
      return importPreview();
    },
    applyLegacyHomeboxImport: async () => failRepositoryCall()
  };
}

function importPreview(): ImportPreview {
  return {
    source: { type: 'legacy_homebox', name: 'Homebox', imageImport: 'disabled' },
    counts: { fields: 0, locations: 0, assets: 0, attachments: 0, warnings: 0, errors: 0 },
    fields: [],
    assetSamples: [],
    imageSamples: [],
    messages: []
  };
}

function failRepositoryCall(): never {
  throw new Error('Unexpected repository call.');
}

function switchControl(label: string): HTMLButtonElement | null {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('button[role="switch"]')).find((button) =>
    button.textContent?.includes(label)
  ) ?? null;
}

function sourceControl(label: string): HTMLElement {
  const group = document.body.querySelector<HTMLElement>('[role="group"][aria-label="Import source"]');
  const control = Array.from(group?.querySelectorAll<HTMLElement>('button, a') ?? []).find((candidate) =>
    candidate.textContent?.trim().includes(label)
  );
  if (!control) {
    throw new Error(`Missing import source control ${label}`);
  }
  return control;
}

function input(selector: string, value: string): void {
  const element = document.body.querySelector<HTMLInputElement>(selector);
  if (!element) {
    throw new Error(`Missing input ${selector}`);
  }
  element.value = value;
  element.dispatchEvent(new Event('input', { bubbles: true }));
}

function click(text: string): void {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) {
    throw new Error(`Missing button ${text}`);
  }
  button.click();
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
