import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import AddAssetTray from './AddAssetTray.svelte';
import type { AddAssetSubmission } from '$lib/domain/inventory';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
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
});

function input(selector: string, value: string): void {
  const element = document.querySelector<HTMLInputElement>(selector);
  if (!element) throw new Error(`Missing input ${selector}`);
  element.value = value;
  element.dispatchEvent(new Event('input', { bubbles: true }));
}

function click(text: string): void {
  button(text).click();
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

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
