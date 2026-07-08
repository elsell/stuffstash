import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Asset, SearchResult } from '$lib/domain/inventory';
import SearchPanel from './SearchPanel.svelte';

let component: ReturnType<typeof mount> | null = null;

interface SearchPanelProps {
  tenantId: string;
  inventoryId: string;
  query: string;
	  lifecycleState: 'active' | 'archived' | 'all';
	  searchMode: 'fuzzy' | 'exact';
	  checkoutState: 'any' | 'checked_out' | 'available';
  results: SearchResult[];
  suggestions: Asset[];
  submitted: boolean;
  error: string;
  busy: boolean;
  onSearch: () => void;
  onOpenAsset: (asset: Asset) => void;
}

function asset(id: string, title: string, kind: Asset['kind'] = 'item', photoUrl?: string): Asset {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind,
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active',
    ...(photoUrl ? { photo: { id: `${id}-photo`, assetId: id, url: photoUrl, alt: title } } : {})
  };
}

function mountSearchPanel(props: Partial<SearchPanelProps> = {}) {
  const openedAssetIds: string[] = [];
  const searches: string[] = [];

  component = mount(SearchPanel, {
    target: document.body,
    props: {
      tenantId: 'tenant-home',
      inventoryId: 'inventory-household',
      query: 'ta',
	      lifecycleState: 'active',
	      searchMode: 'fuzzy',
	      checkoutState: 'any',
      results: [],
      suggestions: [asset('tape', 'Tape measure'), asset('tags', 'Gift tags'), asset('table', 'Hall table', 'container')],
      submitted: false,
      error: '',
      busy: false,
      onSearch: () => {
        searches.push('search');
      },
      onOpenAsset: (selected) => {
        openedAssetIds.push(selected.id);
      },
      ...props
    }
  });

  return { openedAssetIds, searches };
}

afterEach(() => {
  if (component) {
    if (component) {
      unmount(component);
      component = null;
    }
  }
  document.body.innerHTML = '';
  vi.useRealTimers();
});

describe('SearchPanel', () => {
  it('opens autocomplete suggestions directly from the search page field', async () => {
    const { openedAssetIds } = mountSearchPanel({
      suggestions: [asset('tape', 'Tape measure', 'item', 'blob:tape-photo')]
    });
    const input = searchInput();

    input.focus();
    await flush();

    expect(document.body.querySelector('#search-page-suggestions')).not.toBeNull();
    expect(document.body.querySelector<HTMLImageElement>('#search-page-suggestions img')?.src).toBe('blob:tape-photo');
    expect(document.body.querySelector<HTMLImageElement>('#search-page-suggestions img')?.alt).toBe('Tape measure');
    expect(controlWithLabel('Open Tape measure').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/assets/tape'
    );

    controlWithLabel('Open Tape measure').click();
    await flush();

    expect(openedAssetIds).toEqual(['tape']);
    expect(input.value).toBe('Tape measure');
    expect(document.body.querySelector('#search-page-suggestions')).toBeNull();
  });

  it('uses kind fallbacks for search suggestions without their own photo', async () => {
    const mismatchedPhotoAsset = {
      ...asset('box', 'Holiday box', 'container'),
      photo: { id: 'wrong-photo', assetId: 'different-asset', url: 'blob:wrong-photo', alt: 'Wrong photo' }
    };
    mountSearchPanel({
      suggestions: [mismatchedPhotoAsset, asset('garage', 'Garage shelf', 'location')],
      query: 'g'
    });
    const input = searchInput();

    input.focus();
    await flush();

    expect(document.body.querySelector('#search-page-suggestions img')).toBeNull();
    expect(document.body.querySelectorAll('#search-page-suggestions .asset-thumb svg')).toHaveLength(2);
  });

  it('marks suggestions and results when a primary photo cannot render', async () => {
    const unavailablePhotoAsset = { ...asset('tape', 'Tape measure'), photoUnavailable: true };
    const results: SearchResult[] = [
      {
        type: 'asset',
        asset: unavailablePhotoAsset,
        inventory: { id: 'inventory-household', name: 'Household' },
        matches: [{ field: 'title', value: 'Tape measure' }]
      }
    ];
    mountSearchPanel({ query: 'tape', suggestions: [unavailablePhotoAsset], results, submitted: true });

    searchInput().focus();
    await flush();

    const suggestion = controlWithLabel('Open Tape measure');
    const result = document.body.querySelector<HTMLAnchorElement>('.asset-list a');
    expect(result).not.toBeNull();
    expect(suggestion.getAttribute('aria-describedby')).toBe('search-page-suggestion-0-photo-unavailable');
    expect(document.getElementById('search-page-suggestion-0-photo-unavailable')?.textContent).toBe('Photo unavailable');
    expect(result?.getAttribute('aria-describedby')).toBe('search-result-tape-photo-unavailable');
    expect(document.getElementById('search-result-tape-photo-unavailable')?.textContent).toBe('Photo unavailable');
    expect(document.body.querySelectorAll('.photo-unavailable-mark')).toHaveLength(2);
    expect(document.body.querySelector<HTMLImageElement>('#search-page-suggestions img')).toBeNull();
    expect(document.body.querySelector<HTMLImageElement>('.asset-list img')).toBeNull();
  });

  it('shows assigned tag chips on search result rows without overwhelming the row', () => {
    const results: SearchResult[] = [
      {
        type: 'asset',
        asset: {
          ...asset('tape', 'Tape measure'),
          tags: [
            { id: 'tag-tools', key: 'tools', displayName: 'Tools', color: '#2F80ED' },
            { id: 'tag-camping', key: 'camping', displayName: 'Camping', color: '#2E7D32' },
            { id: 'tag-kids', key: 'kids', displayName: 'Kids', color: '#7C3AED' }
          ]
        },
        inventory: { id: 'inventory-household', name: 'Household' },
        matches: [{ field: 'tag_display_name', value: 'Tools' }]
      }
    ];

    mountSearchPanel({ query: 'tools', results, submitted: true });

    const tagList = document.body.querySelector<HTMLElement>('.asset-list [aria-label="Asset tags"]');
    expect(tagList).not.toBeNull();
    expect(tagList?.textContent).toContain('Tools');
    expect(tagList?.textContent).toContain('Camping');
    expect(tagList?.textContent).not.toContain('Kids');
    expect(tagList?.querySelector('.tag-chip-overflow')?.textContent).toBe('+1');
    expect(document.body.textContent).toContain('Tag');
  });

  it('supports keyboard traversal for autocomplete suggestions', async () => {
    const { openedAssetIds } = mountSearchPanel();
    const input = searchInput();

    input.focus();
    await flush();
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();

    expect(document.activeElement?.id).toBe('search-page-suggestion-0');

    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();
    expect(document.activeElement?.id).toBe('search-page-suggestion-1');

    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowUp', bubbles: true }));
    await flush();
    expect(document.activeElement?.id).toBe('search-page-suggestion-0');

    (document.activeElement as HTMLElement | null)?.click();
    await flush();

    expect(openedAssetIds).toEqual(['tape']);
  });

  it('closes suggestions with Escape from a focused suggestion', async () => {
    mountSearchPanel();
    const input = searchInput();

    input.focus();
    await flush();
    input.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
    await flush();

    expect(document.activeElement?.id).toBe('search-page-suggestion-0');

    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
    await flush();

    expect(document.activeElement).toBe(input);
    expect(document.body.querySelector('#search-page-suggestions')).toBeNull();
    expect(input.getAttribute('aria-expanded')).toBeNull();
  });

  it('closes suggestions on submit and runs the search', async () => {
    const { searches } = mountSearchPanel();
    const input = searchInput();

    input.focus();
    await flush();
    expect(document.body.querySelector('#search-page-suggestions')).not.toBeNull();

    document.body.querySelector('form')?.dispatchEvent(new SubmitEvent('submit', { bubbles: true, cancelable: true }));
    await flush();

    expect(searches).toEqual(['search']);
    expect(document.body.querySelector('#search-page-suggestions')).toBeNull();
  });

  it('shows clear feedback for empty submitted searches and empty autocomplete suggestions', async () => {
    mountSearchPanel({ query: 'box', suggestions: [], results: [], submitted: true });

    expect(document.body.textContent).toContain('No results for "box"');
    expect(document.body.textContent).not.toContain('Search this inventory');

    searchInput().focus();
    await flush();

    const noSuggestions = document.body.querySelector<HTMLElement>('.search-suggestions-empty');
    expect(noSuggestions?.getAttribute('role')).toBe('status');
    expect(noSuggestions?.textContent).toBe('No suggestions for "box". Press Search to run a full search.');
  });

  it('announces loading and error states to assistive technology', () => {
    mountSearchPanel({ busy: true, submitted: true, suggestions: [] });
    expect(document.body.querySelector('[role="status"]')?.textContent).toContain('Searching');

    const mountedComponent = component;
    if (mountedComponent) {
      unmount(mountedComponent);
      component = null;
    }
    document.body.innerHTML = '';

    mountSearchPanel({ error: 'Search service unavailable.', submitted: true, suggestions: [] });
    expect(document.body.querySelector('[role="alert"]')?.textContent).toContain('Search failed');
    expect(document.body.querySelector('[role="alert"]')?.textContent).toContain('Search service unavailable.');
  });

  it('exposes route-backed search filter hrefs and preserves modified clicks', async () => {
    const { searches } = mountSearchPanel({
      query: 'garage shelf',
      lifecycleState: 'active',
      searchMode: 'fuzzy',
      suggestions: []
    });

    expect(linkWithText('Archived').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/search?q=garage+shelf&lifecycle=archived'
    );
    expect(linkWithText('Exact').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/search?q=garage+shelf&mode=exact'
    );

    linkWithText('Exact').click();
    await flush();

    expect(searches).toEqual(['search']);

    searches.length = 0;
    let componentPreventedModifiedClick = true;
    const target = linkWithText('Archived');
    target.addEventListener('click', (event) => {
      componentPreventedModifiedClick = event.defaultPrevented;
      event.preventDefault();
    });
    target.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }));

    expect(searches).toEqual([]);
    expect(componentPreventedModifiedClick).toBe(false);
  });

  it('keeps the result list behavior independent from autocomplete suggestions', async () => {
    const resultAsset = asset('passport', 'Passport', 'item', 'blob:passport-photo');
    const results: SearchResult[] = [
      {
        type: 'asset',
        asset: resultAsset,
        inventory: { id: 'inventory-household', name: 'Household' },
        matches: [{ field: 'title', value: 'Passport' }]
      }
    ];
    const { openedAssetIds } = mountSearchPanel({ query: '', results, suggestions: [], submitted: true });

    expect(document.body.querySelector<HTMLImageElement>('.asset-list img')?.src).toBe('blob:passport-photo');
    expect(document.body.querySelector<HTMLImageElement>('.asset-list img')?.alt).toBe('Passport');
    expect(linkWithText('Passport').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/assets/passport'
    );

    linkWithText('Passport').click();
    await flush();

    expect(openedAssetIds).toEqual(['passport']);
  });

  it('renders tag search matches with user-facing labels', () => {
    const resultAsset = asset('tent', 'Family tent');
    const results: SearchResult[] = [
      {
        type: 'asset',
        asset: resultAsset,
        inventory: { id: 'inventory-household', name: 'Household' },
        matches: [{ field: 'tag_display_name', value: 'Camping' }]
      }
    ];
    mountSearchPanel({ query: 'camping', results, suggestions: [], submitted: true });

    expect(document.body.querySelector('.asset-row-meta')?.textContent).toContain('Tag');
    expect(document.body.querySelector('.asset-row-meta')?.textContent).toContain('Camping');
    expect(document.body.querySelector('.asset-row-meta')?.textContent).not.toContain('tag_display_name');
  });

  it('routes location suggestions and results to the focused location surface', async () => {
    const locationAsset = asset('garage', 'Garage', 'location');
    const results: SearchResult[] = [
      {
        type: 'asset',
        asset: locationAsset,
        inventory: { id: 'inventory-household', name: 'Household' },
        matches: [{ field: 'title', value: 'Garage' }]
      }
    ];
    const { openedAssetIds } = mountSearchPanel({
      query: 'ga',
      suggestions: [locationAsset],
      results,
      submitted: true
    });

    searchInput().focus();
    await flush();

    expect(controlWithLabel('Open Garage').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/garage'
    );
    expect(linkWithText('Garage').getAttribute('href')).toBe(
      '/tenants/tenant-home/inventories/inventory-household/locations/garage'
    );

    controlWithLabel('Open Garage').click();
    await flush();

    expect(openedAssetIds).toEqual(['garage']);
  });

  it('preserves modified clicks on search result links', () => {
    const resultAsset = asset('passport', 'Passport');
    const results: SearchResult[] = [
      {
        type: 'asset',
        asset: resultAsset,
        inventory: { id: 'inventory-household', name: 'Household' },
        matches: [{ field: 'title', value: 'Passport' }]
      }
    ];
    const { openedAssetIds } = mountSearchPanel({ query: '', results, suggestions: [], submitted: true });

    let componentPreventedModifiedClick = false;
    const target = linkWithText('Passport');
    target.addEventListener('click', (event) => {
      componentPreventedModifiedClick = event.defaultPrevented;
      event.preventDefault();
    });
    target.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, metaKey: true }));

    expect(openedAssetIds).toEqual([]);
    expect(componentPreventedModifiedClick).toBe(false);
  });

  it('uses the kind fallback for search results without their own photo', () => {
    const resultAsset = {
      ...asset('garage-bin', 'Garage bin', 'container'),
      photo: { id: 'wrong-photo', assetId: 'different-asset', url: 'blob:wrong-photo', alt: 'Wrong photo' }
    };
    const results: SearchResult[] = [
      {
        type: 'asset',
        asset: resultAsset,
        inventory: { id: 'inventory-household', name: 'Household' },
        matches: [{ field: 'title', value: 'Garage bin' }]
      }
    ];
    mountSearchPanel({ query: '', results, suggestions: [], submitted: true });

    expect(document.body.querySelector('.asset-list img')).toBeNull();
    expect(document.body.querySelectorAll('.asset-list .asset-thumb svg')).toHaveLength(1);
  });
});

function searchInput(): HTMLInputElement {
  const input = document.body.querySelector<HTMLInputElement>('#search-page-query');
  if (!input) {
    throw new Error('Missing search page input');
  }
  return input;
}

function controlWithLabel(label: string): HTMLElement {
  const control = document.body.querySelector<HTMLElement>(`button[aria-label="${label}"], a[aria-label="${label}"]`);
  if (!control) {
    throw new Error(`Missing control labelled ${label}`);
  }
  return control;
}

function linkWithText(text: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!link) {
    throw new Error(`Missing link containing ${text}`);
  }
  return link;
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
