import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import type { Asset } from '$lib/domain/inventory';
import SearchSuggestions from './SearchSuggestions.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('SearchSuggestions', () => {
  it('renders route-backed thumbnail suggestions with ordinary list semantics', () => {
    const focused: number[] = [];
    const opened: string[] = [];
    component = mount(SearchSuggestions, {
      target: document.body,
      props: {
        id: 'suggestions',
        idPrefix: 'suggestion',
        suggestions: [
          asset('tape', 'Tape measure', 'item', 'blob:tape-photo'),
          asset('garage', 'Garage', 'location')
        ],
        activeIndex: 1,
        assetHref: (candidate) => `/assets/${candidate.id}`,
        onFocusIndex: (index) => {
          focused.push(index);
        },
        onSuggestionKeydown: () => {},
        onOpen: (_event, candidate) => {
          opened.push(candidate.id);
        }
      }
    });

    expect(document.body.querySelector('ul[aria-label="Search suggestions"]')).not.toBeNull();
    expect(document.body.querySelector('[role="listbox"]')).toBeNull();
    expect(link('Open Tape measure').getAttribute('href')).toBe('/assets/tape');
    expect(link('Open Garage').dataset.active).toBe('true');
    expect(document.body.querySelector<HTMLImageElement>('img')?.src).toBe('blob:tape-photo');
    expect(document.body.textContent).toContain('Location');

    link('Open Garage').focus();
    link('Open Tape measure').click();

    expect(focused).toEqual([1]);
    expect(opened).toEqual(['tape']);
  });

  it('uses kind fallbacks when a suggestion has no usable own photo', () => {
    component = mount(SearchSuggestions, {
      target: document.body,
      props: {
        id: 'suggestions',
        idPrefix: 'suggestion',
        suggestions: [
          {
            ...asset('box', 'Holiday box', 'container'),
            photo: { id: 'wrong-photo', assetId: 'other-asset', url: 'blob:wrong-photo', alt: 'Wrong photo' }
          }
        ],
        activeIndex: -1,
        assetHref: (candidate) => `/assets/${candidate.id}`,
        onFocusIndex: () => {},
        onSuggestionKeydown: () => {},
        onOpen: () => {}
      }
    });

    expect(document.body.querySelector('img')).toBeNull();
    expect(document.body.querySelector('.asset-thumb svg')).not.toBeNull();
  });
});

function asset(id: string, title: string, kind: Asset['kind'], photoUrl?: string): Asset {
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

function link(label: string): HTMLAnchorElement {
  const target = document.body.querySelector<HTMLAnchorElement>(`a[aria-label="${label}"]`);
  if (!target) {
    throw new Error(`Missing link ${label}`);
  }
  return target;
}
