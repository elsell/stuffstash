import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import AddAssetTray from './AddAssetTray.svelte';
import AddAssetTrayHarness from './AddAssetTray.test-harness.svelte';
import type { AddAssetSubmission, CustomAssetType, CustomFieldDefinition, CustomFieldType } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;
let objectUrlIndex = 0;
const revokedObjectUrls: string[] = [];

beforeEach(() => {
  objectUrlIndex = 0;
  revokedObjectUrls.length = 0;
  vi.stubGlobal('URL', {
    ...URL,
    createObjectURL: vi.fn(() => `blob:test-${++objectUrlIndex}`),
    revokeObjectURL: vi.fn((url: string) => {
      revokedObjectUrls.push(url);
    })
  });
});

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
  vi.unstubAllGlobals();
});

describe('AddAssetTray', () => {
  it('reflects the selected kind in the heading, name field, and save action', async () => {
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        initialKind: 'location',
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        parentTargets: [],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async () => ({ saved: true })
      }
    });

    await flush();

    expect(document.body.textContent).toContain('Add location');
    expect(labelFor('asset-title')?.textContent).toBe('Location name');
    expect(inputElement('#asset-title').getAttribute('placeholder')).toBe('Garage shelf');
    expect(button('Save location')).toBeTruthy();

    button('Container').click();
    await flush();

    expect(document.body.textContent).toContain('Add container');
    expect(labelFor('asset-title')?.textContent).toBe('Container name');
    expect(inputElement('#asset-title').getAttribute('placeholder')).toBe('Clear storage bin');
    expect(button('Save container')).toBeTruthy();
  });

  it('keeps quick parent creation hidden until the user opts in', async () => {
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        parentTargets: [],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async () => ({ saved: true })
      }
    });

    await flush();

    expect(document.body.querySelector('.add-summary')?.textContent).toContain('Item');
    expect(document.body.querySelector('.add-summary')?.textContent).toContain('Inventory root');
    expect(document.body.querySelector('.add-summary')?.textContent).toContain('No photos');
    expect(document.body.textContent).toContain('Create missing parent');
    expect(switchControl('Create a parent first')?.getAttribute('aria-checked')).toBe('false');
    expect(document.querySelector('#quick-parent-title')).toBeNull();

    switchControl('Create a parent first')?.click();
    await flush();

    expect(switchControl('Create a parent first')?.getAttribute('aria-checked')).toBe('true');
    expect(document.querySelector('#quick-parent-title')).not.toBeNull();
    expect(quickParentContextText()).toContain('Created under');
    expect(quickParentContextText()).toContain('Inventory root');
    expect(document.body.querySelector('.add-summary')?.textContent).toContain('New Location in Inventory root');
    expect(document.body.querySelector('.quick-parent-context')?.getAttribute('aria-live')).toBeNull();
  });

  it('seeds the parent destination from a route-backed parent id', async () => {
    let savedDraft: AddAssetSubmission | null = null;
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household/locations/garage',
        initialKind: 'item',
        initialParentAssetId: 'garage',
        parentTargets: [parentTarget('garage', 'Garage', 'Home')],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async (draft) => {
          savedDraft = draft;
          return { saved: true };
        }
      }
    });

    await flush();

    expect(document.body.querySelector('.add-summary')?.textContent).toContain('Garage');
    expect(document.body.querySelector('.parent-current-card')?.textContent).toContain('Garage');
    expect(document.body.querySelector('.parent-current-card')?.getAttribute('data-selected')).toBe('target');

    input('#asset-title', 'Tape measure');
    await flush();
    click('Save');
    await flush();

    expect(savedDraft).toMatchObject({
      title: 'Tape measure',
      parentAssetId: 'garage'
    });
  });

  it('exposes close hrefs for visible tray dismissal controls and preserves modified clicks', async () => {
    let closeCount = 0;
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        parentTargets: [],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {
          closeCount += 1;
        },
        onSave: async () => ({ saved: true })
      }
    });

    await flush();

    expect(linkWithLabel('Close add tray').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household');
    expect(link('Cancel').getAttribute('href')).toBe('/tenants/tenant-home/inventories/inventory-household');

    link('Cancel').click();
    expect(closeCount).toBe(1);

    closeCount = 0;
    let componentPreventedModifiedClick = false;
    const target = linkWithLabel('Close add tray');
    target.addEventListener('click', (event) => {
      componentPreventedModifiedClick = event.defaultPrevented;
      event.preventDefault();
    });
    target.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }));

    expect(closeCount).toBe(0);
    expect(componentPreventedModifiedClick).toBe(false);
  });

  it('submits an opted-in quick-created parent with the asset draft', async () => {
    let savedDraft: AddAssetSubmission | null = null;
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        parentTargets: [],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async (draft) => {
          savedDraft = draft;
          return { saved: true };
        }
      }
    });

    await flush();
    input('#asset-title', 'Tape measure');
    switchControl('Create a parent first')?.click();
    await flush();
    input('#quick-parent-title', 'Garage shelf');
    await flush();
    expect(quickParentContextText()).toContain('Created under');
    expect(quickParentContextText()).toContain('Inventory root');
    expect(document.body.querySelector('.add-summary')?.textContent).toContain('New Location: Garage shelf in Inventory root');
    clickLast('Container');
    await flush();
    expect(document.body.querySelector('.add-summary')?.textContent).toContain('New Container: Garage shelf in Inventory root');
    click('Save');
    await flush();

    expect(savedDraft).toMatchObject({
      title: 'Tape measure',
      parentQuickCreate: {
        kind: 'container',
        title: 'Garage shelf'
      }
    });
  });

  it('shows where a quick-created parent will be nested when an existing parent is selected', async () => {
    let savedDraft: AddAssetSubmission | null = null;
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        initialParentAssetId: 'garage',
        parentTargets: [parentTarget('garage', 'Garage', 'Home')],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async (draft) => {
          savedDraft = draft;
          return { saved: true };
        }
      }
    });

    await flush();
    input('#asset-title', 'Tape measure');
    switchControl('Create a parent first')?.click();
    await flush();
    input('#quick-parent-title', 'Garage shelf');
    await flush();

    expect(quickParentContextText()).toContain('Created under');
    expect(quickParentContextText()).toContain('Garage');
    expect(document.body.querySelector('.add-summary')?.textContent).toContain('New Location: Garage shelf in Garage / Home');

    click('Save');
    await flush();

    expect(savedDraft).toMatchObject({
      parentAssetId: 'garage',
      parentQuickCreate: {
        kind: 'location',
        title: 'Garage shelf'
      }
    });
  });

  it('disambiguates duplicate selected parent names with the containment trail', async () => {
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        initialParentAssetId: 'garage-basement',
        parentTargets: [
          parentTarget('garage-main', 'Garage shelf', 'Main garage'),
          parentTarget('garage-basement', 'Garage shelf', 'Basement / Utility room')
        ],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async () => ({ saved: true })
      }
    });

    await flush();
    switchControl('Create a parent first')?.click();
    await flush();

    expect(quickParentContextText()).toContain('Garage shelf');
    expect(quickParentContextText()).toContain('Basement / Utility room');
    expect(document.body.querySelector('.add-summary')?.textContent).toContain(
      'New Location in Garage shelf / Basement / Utility room'
    );
  });

  it('submits selected custom type and typed custom field values with the asset draft', async () => {
    let savedDraft: AddAssetSubmission | null = null;
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        parentTargets: [],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [customAssetType('tool', 'Tool')],
        customFieldDefinitions: [
          customFieldDefinition('quantity', 'Quantity', 'number'),
          customFieldDefinition('fragile', 'Fragile', 'boolean'),
          customFieldDefinition('condition', 'Condition', 'enum', ['new', 'open'], 'tool')
        ],
        saving: false,
        onClose: () => {},
        onSave: async (draft) => {
          savedDraft = draft;
          return { saved: true };
        }
      }
    });

    await flush();
    input('#asset-title', 'Socket wrench');
    click('Tool');
    await flush();
    input('#custom-field-quantity', '3');
    click('Yes');
    click('open');
    click('Save');
    await flush();

    expect(savedDraft).toMatchObject({
      title: 'Socket wrench',
      customAssetTypeId: 'tool',
      customFields: {
        quantity: 3,
        fragile: true,
        condition: 'open'
      }
    });
  });

  it('requires a parent name when quick parent creation is enabled', async () => {
    let saved = false;
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        parentTargets: [],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async () => {
          saved = true;
          return { saved: true };
        }
      }
    });

    await flush();
    input('#asset-title', 'Tape measure');
    switchControl('Create a parent first')?.click();
    await flush();

    const parentInput = document.querySelector<HTMLInputElement>('#quick-parent-title');
    expect(parentInput?.getAttribute('aria-invalid')).toBe('true');
    expect(parentInput?.getAttribute('aria-describedby')).toBe('quick-parent-error');
    expect(document.body.textContent).toContain('Enter a parent name or turn this option off.');
    expect(button('Save').disabled).toBe(true);

    button('Save').click();
    await flush();

    expect(saved).toBe(false);
  });

  it('clears quick parent draft data when the option is turned off', async () => {
    let savedDraft: AddAssetSubmission | null = null;
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        parentTargets: [],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async (draft) => {
          savedDraft = draft;
          return { saved: true };
        }
      }
    });

    await flush();
    input('#asset-title', 'Tape measure');
    switchControl('Create a parent first')?.click();
    await flush();
    input('#quick-parent-title', 'Garage shelf');
    switchControl('Create a parent first')?.click();
    await flush();
    click('Save');
    await flush();

    expect(document.querySelector('#quick-parent-title')).toBeNull();
    expect(savedDraft).toMatchObject({
      title: 'Tape measure',
      parentQuickCreate: undefined
    });
  });

  it('selects a created parent for retry after a partial save', async () => {
    const submissions: AddAssetSubmission[] = [];
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        parentTargets: [
          {
            id: 'parent-created',
            tenantId: 'tenant-home',
            inventoryId: 'inventory-household',
            title: 'Garage shelf',
            description: '',
            kind: 'container',
            parentAssetId: null,
            lifecycleState: 'active',
            containmentTrail: 'Garage shelf'
          }
        ],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async (draft) => {
          submissions.push(draft);
          return submissions.length === 1 ? { saved: false, createdParentId: 'parent-created' } : { saved: true };
        }
      }
    });

    await flush();
    input('#asset-title', 'Tape measure');
    switchControl('Create a parent first')?.click();
    await flush();
    input('#quick-parent-title', 'Garage shelf');
    clickLast('Container');
    await flush();
    click('Save');
    await flush();
    click('Save');
    await flush();

    expect(submissions[0]).toMatchObject({
      parentQuickCreate: {
        kind: 'container',
        title: 'Garage shelf'
      }
    });
    expect(submissions[1]).toMatchObject({
      parentAssetId: 'parent-created',
      parentQuickCreate: undefined
    });
  });

  it('focuses the title field, seeds the requested kind, and closes on Escape', async () => {
    let closeCount = 0;
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        initialKind: 'location',
        parentTargets: [],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {
          closeCount += 1;
        },
        onSave: async () => ({ saved: true })
      }
    });

    await flush();

    expect(document.activeElement?.id).toBe('asset-title');
    expect(button('Location').getAttribute('aria-pressed')).toBe('true');

    const dialog = document.body.querySelector<HTMLElement>('[role="dialog"]');
    if (!dialog) throw new Error('Missing dialog');
    dialog.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));

    expect(closeCount).toBe(1);
  });

  it('filters parent targets and exposes grouped picker controls without dumping every parent before search', async () => {
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        parentTargets: [
          parentTarget('garage', 'Garage shelf', 'Garage'),
          parentTarget('closet', 'Hall closet', 'Hall'),
          parentTarget('closet-bin', 'Closet bin', 'Hall'),
          parentTarget('pantry', 'Pantry bin', 'Kitchen'),
          parentTarget('attic', 'Attic', 'Upstairs'),
          parentTarget('laundry', 'Laundry shelf', 'Laundry')
        ],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async () => ({ saved: true })
      }
    });

    await flush();

    const fieldsets = Array.from(document.body.querySelectorAll('fieldset')).map((field) => field.textContent ?? '');
    expect(fieldsets.some((text) => text.includes('Place in existing parent'))).toBe(true);
    expect(fieldsets.some((text) => text.includes('Asset kind'))).toBe(true);
    expect(document.body.textContent).toContain('6 possible destinations');
    expect(document.body.textContent).toContain('Suggested destinations');
    expect(document.body.textContent).toContain('Showing 4 suggested destinations.');
    expect(document.body.textContent).toContain('Garage shelf');
    expect(document.body.textContent).toContain('Hall closet');
    expect(document.body.textContent).not.toContain('Laundry shelf');
    expect(parentTargetButtons('Parent target suggested destinations')).toHaveLength(4);

    input('#parent-search', 'closet');
    await flush();

    expect(document.body.textContent).toContain('Hall closet');
    expect(document.body.textContent).toContain('Closet bin');
    expect(document.body.textContent).not.toContain('Garage shelf');
    expect(document.body.textContent).not.toContain('Showing the first 2 matches');

    click('Hall closet');
    await flush();

    expect(button('Hall closet').getAttribute('aria-pressed')).toBe('true');
  });

  it('revokes preview URLs when photos are replaced or removed', async () => {
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        closeHref: '/tenants/tenant-home/inventories/inventory-household',
        parentTargets: [],
        mediaPolicy: { supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 2048 },
        customAssetTypes: [],
        customFieldDefinitions: [],
        saving: false,
        onClose: () => {},
        onSave: async () => ({ saved: true })
      }
    });

    await flush();
    expect(document.querySelector<HTMLInputElement>('#asset-camera')).toBeNull();
    expect(document.body.textContent).toContain('Choose photos');
    uploadPhoto('first.jpg', 'image/jpeg', 1200);
    await flush();
    expect(document.body.textContent).toContain('first.jpg');
    expect(document.body.querySelector('.add-summary')?.textContent).toContain('1 photo');
    expect(document.body.querySelector('[aria-label="Photo actions"]')?.textContent).toContain('1 photo');

    uploadPhoto('second.jpg', 'image/jpeg', 1200);
    await flush();
    expect(revokedObjectUrls).toContain('blob:test-1');

    clickLabel('Remove second.jpg');
    await flush();
    expect(revokedObjectUrls).toContain('blob:test-2');
  });

  it('keeps keyboard focus inside the tray and restores focus to the opener', async () => {
    component = mount(AddAssetTrayHarness, {
      target: document.body
    });

    await flush();
    const opener = document.body.querySelector<HTMLButtonElement>('#open-add');
    if (!opener) throw new Error('Missing opener');
    opener.focus();
    opener.click();
    await flush();

    expect(document.activeElement?.id).toBe('asset-title');
    input('#asset-title', 'Focus test');
    await flush();

    const dialog = document.body.querySelector<HTMLElement>('[role="dialog"]');
    if (!dialog) throw new Error('Missing dialog');
    const closeButton = document.body.querySelector<HTMLElement>('button[aria-label="Close add tray"], a[aria-label="Close add tray"]');
    const saveButton = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
      candidate.textContent?.startsWith('Save')
    );
    if (!closeButton || !saveButton) throw new Error('Missing focus controls');

    saveButton.focus();
    dialog.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }));
    expect(document.activeElement).toBe(closeButton);

    closeButton.focus();
    dialog.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true }));
    expect(document.activeElement).toBe(saveButton);

    link('Cancel').click();
    await flush();
    expect(document.activeElement).toBe(opener);
  });
});

function parentTarget(id: string, title: string, containmentTrail: string) {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    title,
    description: '',
    kind: 'container' as const,
    parentAssetId: null,
    lifecycleState: 'active' as const,
    containmentTrail
  };
}

function customAssetType(id: string, displayName: string): CustomAssetType {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    scope: 'inventory',
    key: id,
    displayName,
    description: '',
    lifecycleState: 'active'
  };
}

function customFieldDefinition(
  key: string,
  displayName: string,
  type: CustomFieldType,
  enumOptions: string[] = [],
  customAssetTypeId: string | null = null
): CustomFieldDefinition {
  return {
    id: `field-${key}`,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    scope: 'inventory',
    key,
    displayName,
    type,
    enumOptions,
    applicability: customAssetTypeId ? 'custom_asset_types' : 'all_assets',
    customAssetTypeIds: customAssetTypeId ? [customAssetTypeId] : [],
    lifecycleState: 'active'
  };
}

function input(selector: string, value: string): void {
  const element = document.querySelector<HTMLInputElement>(selector);
  if (!element) throw new Error(`Missing input ${selector}`);
  element.value = value;
  element.dispatchEvent(new Event('input', { bubbles: true }));
}

function inputElement(selector: string): HTMLInputElement {
  const element = document.querySelector<HTMLInputElement>(selector);
  if (!element) throw new Error(`Missing input ${selector}`);
  return element;
}

function labelFor(id: string): HTMLLabelElement | null {
  return document.body.querySelector<HTMLLabelElement>(`label[for="${id}"]`);
}

function click(text: string): void {
  button(text).click();
}

function clickLabel(label: string): void {
  const element = document.body.querySelector<HTMLButtonElement>(`button[aria-label="${label}"]`);
  if (!element) throw new Error(`Missing button label ${label}`);
  element.click();
}

function button(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll('button')).find((candidate) => candidate.textContent?.includes(text));
  if (!button) throw new Error(`Missing button ${text}`);
  return button;
}

function link(text: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!link) throw new Error(`Missing link ${text}`);
  return link;
}

function linkWithLabel(label: string): HTMLAnchorElement {
  const link = document.body.querySelector<HTMLAnchorElement>(`a[aria-label="${label}"]`);
  if (!link) throw new Error(`Missing link label ${label}`);
  return link;
}

function switchControl(label: string): HTMLButtonElement | null {
  return Array.from(document.body.querySelectorAll<HTMLButtonElement>('button[role="switch"]')).find((button) =>
    button.textContent?.includes(label)
  ) ?? null;
}

function parentTargetButtons(groupLabel: string): HTMLButtonElement[] {
  const group = Array.from(document.body.querySelectorAll<HTMLElement>('[role="group"]')).find(
    (candidate) => candidate.getAttribute('aria-label') === groupLabel
  );
  if (!group) throw new Error(`Missing parent target group ${groupLabel}`);
  return Array.from(group.querySelectorAll<HTMLButtonElement>('button.parent-target-button'));
}

function quickParentContextText(): string {
  const context = document.body.querySelector('.quick-parent-context');
  if (!context) throw new Error('Missing quick parent context');
  return context.textContent ?? '';
}

function clickLast(text: string): void {
  const matching = Array.from(document.body.querySelectorAll('button')).filter((candidate) => candidate.textContent?.includes(text));
  const button = matching[matching.length - 1];
  if (!button) throw new Error(`Missing button ${text}`);
  button.click();
}

function uploadPhoto(name: string, type: string, sizeBytes: number): void {
  const input = document.querySelector<HTMLInputElement>('#asset-photos');
  if (!input) throw new Error('Missing photo input');
  const file = new File([new Uint8Array(sizeBytes)], name, { type });
  Object.defineProperty(input, 'files', {
    configurable: true,
    value: [file]
  });
  input.dispatchEvent(new Event('change', { bubbles: true }));
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
