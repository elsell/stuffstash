import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it } from 'vitest';
import type { AssetViewModel, ParentTargetViewModel } from '$lib/domain/inventory';
import ParentTargetButton from './ParentTargetButton.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('ParentTargetButton', () => {
  it('renders a photo-backed destination row and selects the target', () => {
    const selectedIds: string[] = [];
    component = mount(ParentTargetButton, {
      target: document.body,
      props: {
        target: parentTarget('garage-shelf', 'Garage shelf', 'Garage', {
          id: 'photo-one',
          assetId: 'garage-shelf',
          url: 'blob:garage-shelf',
          alt: 'Garage shelf photo'
        }),
        selected: true,
        onSelect: (id) => selectedIds.push(id)
      }
    });

    const button = requiredButton();
    expect(button.getAttribute('aria-pressed')).toBe('true');
    expect(button.textContent).toContain('Garage shelf');
    expect(button.textContent).toContain('Container / Garage');
    expect(button.getAttribute('aria-label')).toBeNull();
    expect(document.body.querySelector<HTMLImageElement>('img')?.src).toBe('blob:garage-shelf');
    expect(document.body.querySelector('.parent-target-thumb')?.getAttribute('aria-hidden')).toBe('true');

    button.click();

    expect(selectedIds).toEqual(['garage-shelf']);
  });

  it('uses the kind fallback when the destination has no own photo', () => {
    component = mount(ParentTargetButton, {
      target: document.body,
      props: {
        target: parentTarget('hall-closet', 'Hall closet', 'Hall'),
        selected: false,
        onSelect: () => {}
      }
    });

    expect(requiredButton().getAttribute('aria-pressed')).toBe('false');
    expect(document.body.querySelector('img')).toBeNull();
    expect(document.body.querySelector('.parent-target-thumb')?.getAttribute('aria-hidden')).toBe('true');
  });
});

function parentTarget(
  id: string,
  title: string,
  containmentTrail: string,
  photo?: AssetViewModel['photo']
): ParentTargetViewModel {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind: 'container',
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail,
    photo
  };
}

function requiredButton(): HTMLButtonElement {
  const button = document.body.querySelector<HTMLButtonElement>('button');
  if (!button) {
    throw new Error('Missing destination button');
  }
  return button;
}
