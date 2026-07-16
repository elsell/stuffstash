import { mount, tick, unmount } from 'svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Asset } from '$lib/domain/inventory';
import { withTrail } from '$lib/application/workspace';
import ContainedAssetWorkspace from './ContainedAssetWorkspace.svelte';

const asset = (id: string, title: string, kind: Asset['kind'], parentAssetId: string | null = null): Asset => ({
  id,
  tenantId: 'tenant-one',
  inventoryId: 'inventory-one',
  title,
  description: '',
  kind,
  parentAssetId,
  lifecycleState: 'active'
});

const container = asset('cabinet', 'Garage cabinet', 'container', 'garage');
const assets = [
  asset('garage', 'Garage', 'location'),
  container,
  asset('bin-10', 'Bin 10', 'container', 'cabinet'),
  asset('bin-2', 'Bin 2', 'container', 'cabinet'),
  asset('drill', 'Drill', 'item', 'cabinet'),
  asset('hammer', 'Hammer', 'item', 'garage'),
  asset('wrench', 'Wrench', 'item')
];

let component: ReturnType<typeof mount> | null = null;

afterEach(async () => {
  if (component) await unmount(component);
  component = null;
  document.body.innerHTML = '';
});

function render(target: Asset, overrides: Record<string, unknown> = {}) {
  const props = {
    target: withTrail(target, assets),
    assets,
    canCreate: true,
    canEdit: true,
    saving: false,
    moveHereOpen: false,
    onOpenAsset: vi.fn(),
    onOpenAdd: vi.fn(),
    onOpenMoveHere: vi.fn(),
    onCloseMoveHere: vi.fn(),
    onMoveHere: vi.fn(async () => {}),
    ...overrides
  };
  component = mount(ContainedAssetWorkspace, { target: document.body, props });
  return props;
}

describe('ContainedAssetWorkspace', () => {
  it('renders an ordered container workspace with durable spatial actions', () => {
    render(container);

    expect(document.body.textContent).toContain('Inside Garage cabinet');
    const rows = [...document.body.querySelectorAll('.contained-asset-row')];
    expect(rows.map((row) => row.textContent)).toEqual(expect.arrayContaining([
      expect.stringContaining('Bin 2'),
      expect.stringContaining('Bin 10'),
      expect.stringContaining('Drill')
    ]));
    expect(rows.map((row) => row.textContent?.match(/Bin 2|Bin 10|Drill/)?.[0])).toEqual(['Bin 2', 'Bin 10', 'Drill']);
    expect(link('Add item here').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/add/item?parent=cabinet'
    );
    expect(link('Move items here').getAttribute('href')).toBe(
      '/tenants/tenant-one/inventories/inventory-one/assets/cabinet/move-here'
    );
  });

  it('renders location spaces and descendant items with relative placement', () => {
    const place = asset('place', 'Garage', 'location');
    const shelf = asset('shelf', 'Shelf / East', 'container', 'place');
    const nested = asset('nested', 'Tape', 'item', 'shelf');
    render(place, { assets: [...assets, place, shelf, nested] });

    expect(document.body.textContent).toContain('Spaces in Garage');
    expect(document.body.textContent).toContain('Items in Garage');
    expect(document.body.textContent).toContain('Shelf / East');
    expect(document.body.textContent).toContain('Tape');
    expect(document.body.querySelector('.location-contained-workspace')).not.toBeNull();
    expect(document.body.querySelectorAll('.contained-section')).toHaveLength(2);
    expect(document.body.querySelectorAll('.contained-asset-list')).toHaveLength(2);
    expect(document.body.querySelectorAll('.contained-asset-item')).toHaveLength(2);
    expect(document.body.querySelectorAll('.contained-asset-row-indicator')).toHaveLength(2);
  });

  it('searches, selects, and confirms one move-here candidate', async () => {
    const onMoveHere = vi.fn(async () => {});
    render(container, { moveHereOpen: true, onMoveHere });
    await tick();

    const search = document.body.querySelector<HTMLInputElement>('#move-here-search');
    if (!search) throw new Error('Missing move-here search');
    expect(search.classList.contains('pl-9')).toBe(true);
    expect(document.body.querySelector('[data-move-here-candidates]')?.classList.contains('grid')).toBe(true);
    expect(document.body.querySelector('[data-move-here-candidate-copy]')?.classList.contains('min-w-0')).toBe(true);
    search.value = 'hammer';
    search.dispatchEvent(new Event('input', { bubbles: true }));
    await Promise.resolve();

    button('Select Hammer').click();
    await Promise.resolve();
    button('Move Hammer here').click();
    await Promise.resolve();

    expect(onMoveHere).toHaveBeenCalledWith(expect.objectContaining({ id: 'hammer' }));
  });

  it('hides spatial mutation actions without permission', () => {
    render(container, { canCreate: false, canEdit: false });
    expect(document.body.textContent).not.toContain('Add item here');
    expect(document.body.textContent).not.toContain('Move items here');
    expect(document.body.textContent).toContain('Inside Garage cabinet');
  });
});

function link(name: string): HTMLAnchorElement {
  const candidate = [...document.body.querySelectorAll<HTMLAnchorElement>('a')].find((element) => element.textContent?.includes(name));
  if (!candidate) throw new Error(`Missing link ${name}`);
  return candidate;
}

function button(name: string): HTMLButtonElement {
  const candidate = [...document.body.querySelectorAll<HTMLButtonElement>('button')].find(
    (element) => element.getAttribute('aria-label') === name || element.textContent?.trim() === name
  );
  if (!candidate) throw new Error(`Missing button ${name}`);
  return candidate;
}
