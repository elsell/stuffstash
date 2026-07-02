import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import AddAssetTray from './AddAssetTray.svelte';
import AddAssetTrayHarness from './AddAssetTray.test-harness.svelte';
import type { AddAssetSubmission } from '$lib/domain/inventory';

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
  it('submits a quick-created parent with the asset draft', async () => {
    let savedDraft: AddAssetSubmission | null = null;
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
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
    input('#quick-parent-title', 'Garage shelf');
    clickLast('Container');
    await flush();
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

  it('selects a created parent for retry after a partial save', async () => {
    const submissions: AddAssetSubmission[] = [];
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
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

  it('filters parent targets and exposes grouped picker controls', async () => {
    component = mount(AddAssetTray, {
      target: document.body,
      props: {
        open: true,
        parentTargets: [
          parentTarget('garage', 'Garage shelf', 'Garage'),
          parentTarget('closet', 'Hall closet', 'Hall'),
          parentTarget('closet-bin', 'Closet bin', 'Hall'),
          parentTarget('pantry', 'Pantry bin', 'Kitchen')
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
    uploadPhoto('first.jpg', 'image/jpeg', 1200);
    await flush();
    expect(document.body.textContent).toContain('first.jpg');

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
    const closeButton = document.body.querySelector<HTMLButtonElement>('button[aria-label="Close add tray"]');
    const saveButton = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find(
      (candidate) => candidate.textContent === 'Save'
    );
    if (!closeButton || !saveButton) throw new Error('Missing focus controls');

    saveButton.focus();
    dialog.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }));
    expect(document.activeElement).toBe(closeButton);

    closeButton.focus();
    dialog.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true }));
    expect(document.activeElement).toBe(saveButton);

    button('Cancel').click();
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

function input(selector: string, value: string): void {
  const element = document.querySelector<HTMLInputElement>(selector);
  if (!element) throw new Error(`Missing input ${selector}`);
  element.value = value;
  element.dispatchEvent(new Event('input', { bubbles: true }));
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
