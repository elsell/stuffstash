import { afterEach, describe, expect, it } from 'vitest';
import { mount, tick, unmount, type ComponentProps } from 'svelte';
import type { Asset } from '$lib/domain/inventory';
import BrowsePanel from './BrowsePanel.svelte';

let component: ReturnType<typeof mount> | null = null;

const assets: Asset[] = [
  { id: 'garage', tenantId: 't', inventoryId: 'i', kind: 'location', title: 'Garage', description: '', parentAssetId: null, lifecycleState: 'active' },
  { id: 'drill', tenantId: 't', inventoryId: 'i', kind: 'item', title: 'Drill', description: '', parentAssetId: 'garage', lifecycleState: 'active' },
  { id: 'bin', tenantId: 't', inventoryId: 'i', kind: 'container', title: 'Tool bin', description: '', parentAssetId: 'garage', lifecycleState: 'active' }
];

afterEach(() => {
  if (component) unmount(component);
  component = null;
  document.body.innerHTML = '';
});

function render(surface: 'list' | 'map' = 'list', scope: 'all' | 'places' | 'containers' | 'items' = 'all', overrides: Partial<ComponentProps<typeof BrowsePanel>> = {}) {
  component = mount(BrowsePanel, {
    target: document.body,
    props: {
      tenantId: 't', inventoryId: 'i', inventoryName: 'Household', assets, placementAssets: assets, results: [], suggestions: [], assetTags: [], query: '', submitted: false,
      error: '', busy: false, surface, scope, lifecycleState: 'active', searchMode: 'fuzzy', checkoutState: 'any',
      sort: 'updated_desc', selectedTagIds: [], onStateChange: () => {}, onSearch: () => {}, onOpenAsset: () => {},
      canCreateAsset: true, onOpenAdd: () => {}, hasMore: false, loadingMore: false, errorPhase: null,
      inventoryEmpty: false,
      onLoadMore: () => {}, onRetry: () => {}, ...overrides
    }
  });
}

function linkWithText(text: string): HTMLAnchorElement {
  const link = Array.from(document.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) => candidate.textContent?.includes(text));
  if (!link) throw new Error(`Missing link ${text}`);
  return link;
}

describe('BrowsePanel', () => {
  it('presents the shared Browse header, List/Map surfaces, and all mobile scopes', () => {
    render();
    expect(document.querySelector('h1')?.textContent).toBe('Browse');
    expect(document.body.textContent).toContain('Household');
    expect(Array.from(document.querySelectorAll('[role="tab"]')).map((node) => node.textContent?.trim())).toEqual([
      'List', 'Map', 'All', 'Places', 'Containers', 'Items'
    ]);
    expect(document.body.textContent).toContain('3 shown');
    expect(document.querySelector('#browse-list-panel')?.getAttribute('aria-labelledby')).toBe('browse-surface-list-tab');
    expect(document.querySelector('#browse-results')?.getAttribute('aria-labelledby')).toBe('browse-scope-all-tab');
  });

  it('exposes canonical hrefs for durable surface, scope, and sort state while preserving modified clicks', () => {
    const changes: Array<Record<string, unknown>> = [];
    render('list', 'all', { query: 'paint', selectedTagIds: ['tag-blue'], onStateChange: (state) => changes.push(state) });

    expect(linkWithText('Map').getAttribute('href')).toBe('/tenants/t/inventories/i/browse?surface=map&q=paint&tag=tag-blue');
    expect(linkWithText('Places').getAttribute('href')).toBe('/tenants/t/inventories/i/browse?scope=places&q=paint&tag=tag-blue');

    linkWithText('Places').dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }));
    expect(changes).toEqual([]);

    unmount(component!);
    component = null;
    document.body.innerHTML = '';
    render('list', 'all', { selectedTagIds: ['tag-blue'] });
    expect(linkWithText('Default order').getAttribute('href')).toBe('/tenants/t/inventories/i/browse?tag=tag-blue&sort=id_asc');
  });

  it('keeps Browse combobox focus on the input while selecting a suggestion', async () => {
    const opened: string[] = [];
    render('list', 'all', {
      query: 'dr',
      suggestions: [assets[1]!],
      onOpenAsset: (asset) => opened.push(asset.id)
    });
    const input = document.querySelector<HTMLInputElement>('input[aria-label="Search Browse"]')!;

    input.focus();
    await tick();
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await tick();

    expect(document.activeElement).toBe(input);
    expect(input.getAttribute('aria-activedescendant')).toBe('browse-suggestion-0');
    expect(document.querySelector('#browse-suggestion-0')?.getAttribute('tabindex')).toBe('-1');

    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }));
    await tick();
    expect(opened).toEqual(['drill']);
  });

  it('filters the photo-first list by scope and exposes a containment map', () => {
    render('list', 'items');
    expect(document.body.textContent).toContain('Drill');
    expect(document.body.textContent).not.toContain('Tool bin');
    unmount(component!);
    component = null;
    document.body.innerHTML = '';

    render('map');
    expect(document.querySelector('#browse-map-panel[role="tabpanel"]')).not.toBeNull();
    expect(document.querySelector('#browse-map-panel')?.getAttribute('aria-labelledby')).toBe('browse-surface-map-tab');
    expect(document.body.textContent).toContain('Inventory root');
    expect(document.body.textContent).toContain('Garage');
    expect(Array.from(document.querySelectorAll('.containment-node small')).map((node) => node.textContent)).toContain('Location');
    expect(Array.from(document.querySelectorAll('.containment-node small')).map((node) => node.textContent)).not.toContain('location');
    expect(document.querySelector('.containment-columns')?.classList.contains('root-only')).toBe(true);
  });

  it('switches from the constrained root Map to equal drilldown columns', async () => {
    render('map');
    const garage = Array.from(document.querySelectorAll<HTMLButtonElement>('.containment-node')).find((node) => node.textContent?.includes('Garage'));
    garage?.click();
    await tick();

    expect(document.querySelectorAll('.containment-columns section')).toHaveLength(2);
    expect(document.querySelector('.containment-columns')?.classList.contains('root-only')).toBe(false);
  });

  it('uses title-cased asset kinds in Map jump results', async () => {
    render('map');
    const input = document.querySelector<HTMLInputElement>('input[aria-label="Jump to a place or container"]')!;
    input.value = 'tool';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    await tick();

    expect(document.querySelector('.containment-jump-results small')?.textContent).toBe('Container');
  });

  it('offers a direct recovery action when a submitted search has no matches', async () => {
    let searches = 0;
    render('list', 'all', {
      query: 'missing drill',
      submitted: true,
      results: [],
      onSearch: () => { searches += 1; }
    });

    expect(document.body.textContent).toContain('No results for “missing drill”');
    expect(document.body.textContent).toContain('Try another search term or clear a filter.');
    Array.from(document.querySelectorAll<HTMLButtonElement>('button')).find((button) => button.textContent?.includes('Clear search'))?.click();
    await tick();
    expect(searches).toBe(1);
  });

  it('keeps filter drafts local until Apply and exposes removable applied filters', () => {
    const changes: Array<Record<string, unknown>> = [];
    render('list', 'all', { lifecycleState: 'archived', checkoutState: 'available', onStateChange: (state) => changes.push(state) });

    expect(document.querySelector('[aria-label="Remove Status: Archived"]')).not.toBeNull();
    expect(document.querySelector('[aria-label="Remove Availability: Available"]')).not.toBeNull();
    Array.from(document.querySelectorAll<HTMLButtonElement>('button')).find((button) => button.textContent?.includes('Filters'))?.click();
    document.querySelector<HTMLButtonElement>('.browse-filter-popover button[aria-pressed="false"]')?.click();
    expect(changes).toEqual([]);
    Array.from(document.querySelectorAll<HTMLButtonElement>('.browse-filter-popover button')).find((button) => button.textContent?.includes('Cancel'))?.click();
    expect(changes).toEqual([]);
  });

  it('sorts filter tags and selected tag tokens with locale-aware natural order', async () => {
    const assetTags = [
      { id: 'zebra', key: 'zebra', displayName: 'zebra' },
      { id: 'bin-10', key: 'bin-10', displayName: 'Bin 10' },
      { id: 'apple', key: 'apple', displayName: 'apple' },
      { id: 'bin-8', key: 'bin-8', displayName: 'Bin 8' }
    ];
    render('list', 'all', { assetTags, selectedTagIds: assetTags.map((tag) => tag.id) });

    expect(Array.from(document.querySelectorAll('.browse-applied-filters [aria-label^="Remove Tag:"]')).map((node) => node.textContent?.replace('×', '').trim())).toEqual([
      'Tag: apple', 'Tag: Bin 8', 'Tag: Bin 10', 'Tag: zebra'
    ]);
    document.querySelector<HTMLButtonElement>('.browse-filter-trigger')?.click();
    await tick();
    expect(Array.from(document.querySelectorAll('.browse-filter-tags button')).map((node) => node.textContent?.trim())).toEqual([
      'apple', 'Bin 8', 'Bin 10', 'zebra'
    ]);
  });

  it('distinguishes an empty inventory from filters that match nothing', () => {
    render('list', 'all', { assets: [], placementAssets: [], inventoryEmpty: true });
    expect(document.body.textContent).toContain('No stuff here yet');
    expect(document.body.textContent).not.toContain('filters');
    expect(linkWithText('Add item').getAttribute('href')).toBe('/tenants/t/inventories/i/add/item');
    expect(linkWithText('Add location').getAttribute('href')).toBe('/tenants/t/inventories/i/add/location');

    unmount(component!);
    component = null;
    document.body.innerHTML = '';
    render('list', 'all', { assetTags: [{ id: 'missing', key: 'missing', displayName: 'Missing' }], selectedTagIds: ['missing'] });
    expect(document.body.textContent).toContain('Nothing matches these filters');
  });

  it('keeps viewers in an honest non-mutating empty state', () => {
    render('list', 'all', { assets: [], placementAssets: [], inventoryEmpty: true, canCreateAsset: false });
    expect(document.body.textContent).toContain('This inventory is empty.');
    expect(document.querySelector('a[href*="/add/"]')).toBeNull();
  });

  it('does not call an archived-only inventory empty when active Browse has no rows', () => {
    render('list', 'all', { assets: [], placementAssets: [], inventoryEmpty: false });
    expect(document.body.textContent).toContain('Nothing matches these filters');
    expect(document.body.textContent).not.toContain('No stuff here yet');
  });

  it('distinguishes initial loading from a background refresh without hiding results', () => {
    render('list', 'all', { assets: [], placementAssets: [], busy: true });
    expect(document.querySelector('[role="status"]')?.textContent).toContain('Loading inventory…');
    expect(document.querySelector('.browse-result-grid')).toBeNull();

    unmount(component!);
    component = null;
    document.body.innerHTML = '';
    render('list', 'all', { busy: true });
    expect(document.querySelector('.browse-updating')?.textContent).toContain('Updating…');
    expect(document.querySelectorAll('.browse-card')).toHaveLength(3);
  });

  it('offers phase-appropriate retry actions for list and Map failures', () => {
    let retries = 0;
    render('list', 'all', { assets: [], placementAssets: [], error: 'Browse could not be loaded. Try again.', errorPhase: 'initial', onRetry: () => { retries += 1; } });
    expect(document.querySelector('[role="alert"]')?.textContent).toContain('Browse could not be loaded. Try again.');
    Array.from(document.querySelectorAll<HTMLButtonElement>('button')).find((button) => button.textContent?.includes('Try again'))?.click();
    expect(retries).toBe(1);

    unmount(component!);
    component = null;
    document.body.innerHTML = '';
    render('map', 'all', { error: 'Map could not be loaded. Try again.', errorPhase: 'map', onRetry: () => { retries += 1; } });
    expect(document.querySelector('[role="alert"]')?.textContent).toContain('Map could not be loaded. Try again.');
    Array.from(document.querySelectorAll<HTMLButtonElement>('button')).find((button) => button.textContent?.includes('Try map again'))?.click();
    expect(retries).toBe(2);
  });

  it('shows an honest initial Map loading state instead of an empty tree', () => {
    render('map', 'all', { assets: [], placementAssets: [], busy: true });
    expect(document.querySelector('[role="status"]')?.textContent).toContain('Loading map…');
    expect(document.body.textContent).not.toContain('Nothing is contained here.');
  });

  it('uses nonhydrating kind icons for Map nodes and restores jump focus after selection', async () => {
    render('map');
    expect(document.querySelectorAll('.containment-node .asset-thumb')).toHaveLength(0);
    expect(document.querySelectorAll('.containment-node .containment-node-kind')).not.toHaveLength(0);

    const input = document.querySelector<HTMLInputElement>('input[aria-label="Jump to a place or container"]')!;
    input.value = 'tool';
    input.dispatchEvent(new Event('input', { bubbles: true }));
    await tick();
    document.querySelector<HTMLButtonElement>('.containment-jump-results [role="option"]')?.click();
    await tick();
    expect(document.activeElement).toBe(input);
  });

  it('renders an explicitly safe replacement or pagination reason without discarding it', () => {
    render('list', 'all', {
      error: 'The selected tag is no longer available.',
      errorPhase: 'replacement'
    });

    expect(document.querySelector('.browse-inline-error')?.textContent).toContain('The selected tag is no longer available.');
    expect(document.querySelector('.browse-inline-error')?.textContent).not.toContain('Browse could not be refreshed.');
  });

  it('requests the next page and supports arrow-key tab selection', () => {
    let loads = 0;
    const changes: Array<Record<string, unknown>> = [];
    render('list', 'all', { hasMore: true, onLoadMore: () => { loads += 1; }, onStateChange: (state) => changes.push(state) });
    Array.from(document.querySelectorAll<HTMLButtonElement>('button')).find((button) => button.textContent?.includes('Load more'))?.click();
    document.querySelector<HTMLButtonElement>('[role="tab"][aria-selected="true"]')?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }));
    expect(loads).toBe(1);
    expect(changes).toContainEqual({ surface: 'map' });
  });

  it('keeps large Map columns bounded while making every sibling reachable', async () => {
    const manyAssets = Array.from({ length: 105 }, (_, index): Asset => ({
      id: `item-${index}`, tenantId: 't', inventoryId: 'i', kind: 'item', title: `Item ${index}`,
      description: '', parentAssetId: null, lifecycleState: 'active'
    }));
    render('map', 'all', { assets: manyAssets, placementAssets: manyAssets });

    expect(document.querySelectorAll('.containment-node')).toHaveLength(100);
    Array.from(document.querySelectorAll<HTMLButtonElement>('button')).find((button) => button.textContent?.includes('Show next 5'))?.click();
    await tick();
    expect(document.querySelectorAll('.containment-node')).toHaveLength(105);
    expect(document.body.textContent).toContain('Item 104');
  });
});
