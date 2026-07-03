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
          parentTarget('garage-shelf', 'Garage shelf', 'Garage', 'container'),
          parentTarget('hall-closet', 'Hall closet', 'Hall', 'location'),
          parentTarget('closet-bin', 'Closet bin', 'Hall / Hall closet')
        ],
        onSelect: (id) => {
          selectedIds.push(id);
        }
      }
    });

    expect(document.body.querySelector('[role="group"]')?.getAttribute('aria-label')).toBe('Move target current destination');
    expect(button('Inventory root').getAttribute('aria-pressed')).toBe('true');
    expect(document.body.textContent).toContain('3 possible destinations');
    expect(document.body.textContent).toContain('Search to choose a location or container.');
    expect(document.body.textContent).not.toContain('Hall closet');

    setInputValue(requiredInput('#parent-target-search'), 'closet');
    await flush();

    expect(group('Move target search results')).toBeTruthy();
    expect(document.body.textContent).toContain('2 matches');
    expect(document.body.textContent).toContain('Hall closet');
    expect(resultButton('Hall closet').textContent).toContain('Location');
    expect(document.body.textContent).toContain('Closet bin');
    expect(resultButton('Closet bin').textContent).toContain('Container');
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

    expect(document.body.textContent).toContain('Current destination');
    expect(document.body.textContent).toContain('Container');
    expect(document.body.textContent).toContain('Root');
    expect(button('Clear parent').getAttribute('aria-label')).toBe('Clear parent selection');
    expect(document.body.textContent).toContain('2 possible destinations');
    expect(document.body.textContent).not.toContain('Target 2');

    setInputValue(requiredInput('#parent-target-search'), 'target');
    await flush();

    expect(document.body.textContent).toContain('Showing the first 1 of 2 matches.');

    setInputValue(requiredInput('#parent-target-search'), 'missing');
    await flush();

    expect(document.body.textContent).toContain('No matching locations or containers.');
  });

  it('clears selected parent back to the root destination', async () => {
    const selectedIds: Array<string | null> = [];
    component = mount(ParentTargetPicker, {
      target: document.body,
      props: {
        legend: 'Parent',
        searchId: 'parent-target-search',
        groupLabel: 'Parent target',
        search: '',
        selectedId: 'target-1',
        targets: [parentTarget('target-1', 'Target 1', 'Root')],
        onSelect: (id) => {
          selectedIds.push(id);
        }
      }
    });

    button('Clear parent').click();
    await flush();

    expect(selectedIds).toEqual([null]);
  });
});

function parentTarget(
  id: string,
  title: string,
  containmentTrail: string,
  kind: AssetViewModel['kind'] = 'container'
): AssetViewModel {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind,
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    containmentTrail
  };
}

function resultButton(text: string): HTMLButtonElement {
  const groupElement = group('Move target search results');
  const target = Array.from(groupElement.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!target) {
    throw new Error(`Missing result button ${text}`);
  }
  return target;
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

function group(label: string): HTMLElement {
  const target = Array.from(document.body.querySelectorAll<HTMLElement>('[role="group"]')).find(
    (candidate) => candidate.getAttribute('aria-label') === label
  );
  if (!target) {
    throw new Error(`Missing group ${label}`);
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
