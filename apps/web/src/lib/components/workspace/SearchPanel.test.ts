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
  results: SearchResult[];
  suggestions: Asset[];
  submitted: boolean;
  error: string;
  busy: boolean;
  onSearch: () => void;
  onOpenAsset: (assetId: string) => void;
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
      results: [],
      suggestions: [asset('tape', 'Tape measure'), asset('tags', 'Gift tags'), asset('table', 'Hall table', 'container')],
      submitted: false,
      error: '',
      busy: false,
      onSearch: () => {
        searches.push('search');
      },
      onOpenAsset: (assetId) => {
        openedAssetIds.push(assetId);
      },
      ...props
    }
  });

  return { openedAssetIds, searches };
}

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
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
