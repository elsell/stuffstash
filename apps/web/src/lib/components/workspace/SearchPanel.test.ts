import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Asset, SearchResult } from '$lib/domain/inventory';
import SearchPanel from './SearchPanel.svelte';

let component: ReturnType<typeof mount> | null = null;

interface SearchPanelProps {
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

function asset(id: string, title: string, kind: Asset['kind'] = 'item'): Asset {
  return {
    id,
    tenantId: 'tenant-home',
    inventoryId: 'inventory-household',
    kind,
    title,
    description: '',
    parentAssetId: null,
    lifecycleState: 'active'
  };
}

function mountSearchPanel(props: Partial<SearchPanelProps> = {}) {
  const openedAssetIds: string[] = [];
  const searches: string[] = [];

  component = mount(SearchPanel, {
    target: document.body,
    props: {
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
    const { openedAssetIds } = mountSearchPanel();
    const input = searchInput();

    input.focus();
    await flush();

    expect(document.body.querySelector('#search-page-suggestions')).not.toBeNull();

    buttonWithLabel('Open Tape measure').click();
    await flush();

    expect(openedAssetIds).toEqual(['tape']);
    expect(input.value).toBe('Tape measure');
    expect(document.body.querySelector('#search-page-suggestions')).toBeNull();
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

    (document.activeElement as HTMLButtonElement | null)?.click();
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

  it('keeps the result list behavior independent from autocomplete suggestions', async () => {
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

    buttonWithText('Passport').click();
    await flush();

    expect(openedAssetIds).toEqual(['passport']);
  });
});

function searchInput(): HTMLInputElement {
  const input = document.body.querySelector<HTMLInputElement>('#search-page-query');
  if (!input) {
    throw new Error('Missing search page input');
  }
  return input;
}

function buttonWithLabel(label: string): HTMLButtonElement {
  const button = document.body.querySelector<HTMLButtonElement>(`button[aria-label="${label}"]`);
  if (!button) {
    throw new Error(`Missing button labelled ${label}`);
  }
  return button;
}

function buttonWithText(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) {
    throw new Error(`Missing button containing ${text}`);
  }
  return button;
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
