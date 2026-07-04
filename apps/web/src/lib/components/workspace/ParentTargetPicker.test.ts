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
  it('offers bounded suggestions before search and grouped results after search', async () => {
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
          parentTarget('closet-bin', 'Closet bin', 'Hall / Hall closet'),
          parentTarget('attic', 'Attic', 'Upstairs', 'location'),
          parentTarget('pantry-bin', 'Pantry bin', 'Kitchen'),
          parentTarget('toolbox', 'Toolbox', 'Garage')
        ],
        visibleLimit: 4,
        onSelect: (id) => {
          selectedIds.push(id);
        }
      }
    });

    expect(document.body.querySelector('[role="group"]')?.getAttribute('aria-label')).toBe('Move target current destination');
    expect(button('Inventory root').getAttribute('aria-pressed')).toBe('true');
    expect(document.body.textContent).toContain('6 possible destinations');
    expect(group('Move target suggested destinations')).toBeTruthy();
    expect(document.body.textContent).toContain('Suggested destinations');
    expect(destinationButtons('Move target suggested destinations')).toHaveLength(4);
    expect(document.body.textContent).toContain('Attic');
    expect(document.body.textContent).toContain('Hall closet');
    expect(document.body.textContent).toContain('Garage shelf');
    expect(document.body.textContent).not.toContain('Pantry bin');

    setInputValue(requiredInput('#parent-target-search'), 'closet');
    await flush();

    expect(group('Move target search results')).toBeTruthy();
    expect(document.body.textContent).toContain('2 matches');
    expect(document.body.querySelector('.selection-summary[aria-live="polite"]')?.textContent).toBe('2 matches');
    expect(group('Locations')).toBeTruthy();
    expect(group('Containers')).toBeTruthy();
    expect(destinationButtons('Move target search results')).toHaveLength(2);
    expect(document.body.textContent).toContain('Hall closet');
    expect(resultButton('Hall closet').textContent).toContain('Location');
    expect(document.body.textContent).toContain('Closet bin');
    expect(resultButton('Closet bin').textContent).toContain('Container');
    expect(document.body.textContent).not.toContain('Garage shelf');

    button('Hall closet').click();
    await flush();

    expect(selectedIds).toEqual(['hall-closet']);
  });

  it('orders search results by title relevance before loose path matches', async () => {
    component = mount(ParentTargetPicker, {
      target: document.body,
      props: {
        legend: 'Parent',
        searchId: 'parent-target-search',
        groupLabel: 'Parent target',
        search: '',
        selectedId: null,
        targets: [
          parentTarget('garage-shelf', 'Garage shelf', 'Garage'),
          parentTarget('shelf-rack', 'Shelf rack', 'Storage'),
          parentTarget('storage-shelf', 'Shelf', 'Storage'),
          parentTarget('bin', 'Utility bin', 'Garage / Shelf')
        ],
        visibleLimit: 4,
        onSelect: () => {}
      }
    });

    setInputValue(requiredInput('#parent-target-search'), 'shelf');
    await flush();

    const labels = destinationButtons('Parent target search results').map((target) => target.textContent ?? '');
    expect(labels[0]).toContain('Shelf');
    expect(labels[1]).toContain('Shelf rack');
    expect(labels[2]).toContain('Garage shelf');
    expect(labels[3]).toContain('Utility bin');
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
        targets: [
          parentTarget('target-1', 'Target 1', 'Root'),
          parentTarget('target-2', 'Target 2', 'Root'),
          parentTarget('target-3', 'Target 3', 'Root')
        ],
        visibleLimit: 1,
        onSelect: () => {}
      }
    });

    expect(document.body.textContent).toContain('Current destination');
    expect(document.body.textContent).toContain('Container');
    expect(document.body.textContent).toContain('Root');
    expect(button('Clear parent').getAttribute('aria-label')).toBe('Clear parent selection');
    expect(document.body.textContent).toContain('3 possible destinations');
    expect(document.body.textContent).toContain('Showing 1 suggested destination.');
    expect(destinationButtons('Parent target suggested destinations')).toHaveLength(1);
    expect(group('Parent target suggested destinations').textContent).not.toContain('Target 1');
    expect(group('Parent target suggested destinations').textContent).toContain('Target 2');
    expect(document.body.textContent).not.toContain('Target 3');

    setInputValue(requiredInput('#parent-target-search'), 'target');
    await flush();

    expect(document.body.textContent).toContain('Showing the first 1 of 3 matches.');

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

function destinationButtons(groupLabel: string): HTMLButtonElement[] {
  return Array.from(group(groupLabel).querySelectorAll<HTMLButtonElement>('button.parent-target-button'));
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
