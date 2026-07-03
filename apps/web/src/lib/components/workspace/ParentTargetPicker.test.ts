import { mount, tick, unmount } from 'svelte';
import { afterEach, describe, expect, it } from 'vitest';
import type { AssetViewModel } from '$lib/domain/inventory';
import ParentTargetPicker from './ParentTargetPicker.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('ParentTargetPicker', () => {
  it('filters valid parent targets and exposes grouped selected state', async () => {
    const selectedIds: Array<string | null> = [];
    component = mount(ParentTargetPicker, {
      target: document.body,
      props: {
        legend: 'Parent',
        searchId: 'parent-target-search',
        groupLabel: 'Move target',
        search: '',
        selectedId: null,
        targets: [
          parentTarget('garage-shelf', 'Garage shelf', 'Garage'),
          parentTarget('hall-closet', 'Hall closet', 'Hall'),
          parentTarget('closet-bin', 'Closet bin', 'Hall / Hall closet')
        ],
        onSelect: (id) => {
          selectedIds.push(id);
        }
      }
    });

    expect(document.body.querySelector('[role="group"]')?.getAttribute('aria-label')).toBe('Move target');
    expect(button('Inventory root').getAttribute('aria-pressed')).toBe('true');

    setInputValue(requiredInput('#parent-target-search'), 'closet');
    await flush();

    expect(document.body.textContent).toContain('Hall closet');
    expect(document.body.textContent).toContain('Closet bin');
    expect(document.body.textContent).not.toContain('Garage shelf');

    button('Hall closet').click();
    await flush();

    expect(selectedIds).toEqual(['hall-closet']);
  });

  it('shows empty and overflow states consistently', async () => {
    component = mount(ParentTargetPicker, {
      target: document.body,
      props: {
        legend: 'Place in existing parent',
        searchId: 'parent-target-search',
        groupLabel: 'Parent target',
        search: '',
        selectedId: 'target-1',
        targets: [parentTarget('target-1', 'Target 1', 'Root'), parentTarget('target-2', 'Target 2', 'Root')],
        visibleLimit: 1,
        onSelect: () => {}
      }
    });

    expect(document.body.textContent).toContain('Selected Target 1');
    expect(document.body.textContent).toContain('Showing the first 1 of 2 matches.');

    setInputValue(requiredInput('#parent-target-search'), 'missing');
    await flush();

    expect(document.body.textContent).toContain('No matching locations or containers.');
  });
});

function parentTarget(id: string, title: string, containmentTrail: string): AssetViewModel {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind: 'container',
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail
  };
}

function requiredInput(selector: string): HTMLInputElement {
  const input = document.body.querySelector<HTMLInputElement>(selector);
  if (!input) {
    throw new Error(`Missing input ${selector}`);
  }
  return input;
}

function button(text: string): HTMLButtonElement {
  const target = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!target) {
    throw new Error(`Missing button ${text}`);
  }
  return target;
}

function setInputValue(input: HTMLInputElement, value: string): void {
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set;
  setter?.call(input, value);
  input.dispatchEvent(new Event('input', { bubbles: true }));
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
