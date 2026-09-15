import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { BrowseEmptyState, BrowseLoadError, BrowsePaginationRetry } from './BrowseResultStates';
import { lightPalette } from '../theme/tokens';

it('offers Add only to an editor of an empty inventory', async () => {
  const h = new MobileRenderHarness();
  let additions = 0;
  try {
    await h.render(<BrowseEmptyState kind="inventory" inventoryName="Home inventory"
      palette={lightPalette} onAdd={() => additions++} />);
    expect(h.allText()).toContain('No items in Home inventory');
    await h.press(h.byLabel('Add item'));
    expect(additions).toBe(1);
    await h.render(<BrowseEmptyState kind="inventory" inventoryName="Shared inventory" palette={lightPalette} />);
    expect(h.allText()).toContain('An inventory editor can add the first item, container, or place.');
    expect(h.byLabel('Add item')).toBeUndefined();
  } finally { await h.unmount(); }
});

it('keeps search and filter clearing distinct', async () => {
  const h = new MobileRenderHarness();
  const cleared: string[] = [];
  try {
    await h.render(<BrowseEmptyState kind="search" query=" bike pump " palette={lightPalette}
      onClearSearch={() => cleared.push('search')} />);
    expect(h.allText()).toContain('No results for “bike pump”');
    await h.press(h.byLabel('Clear search'));
    expect(cleared).toEqual(['search']);
    await h.render(<BrowseEmptyState kind="filters" palette={lightPalette}
      onClearFilters={() => cleared.push('filters')} />);
    expect(h.allText()).toContain('No items match these filters');
    await h.press(h.byLabel('Clear filters'));
    expect(cleared).toEqual(['search', 'filters']);
  } finally { await h.unmount(); }
});

it('retries the failed load or next page beside its specific explanation', async () => {
  const h = new MobileRenderHarness();
  const retries: string[] = [];
  try {
    await h.render(<BrowseLoadError palette={lightPalette} message="The server could not be reached."
      onRetry={() => retries.push('load')} />);
    expect(h.allText()).toContain('Could not load this inventory');
    expect(h.allText()).toContain('The server could not be reached.');
    await h.press(h.byLabel('Retry'));
    expect(retries).toEqual(['load']);
    await h.render(<BrowsePaginationRetry palette={lightPalette} message="Could not load more items."
      onRetry={() => retries.push('page')} />);
    expect(h.allText()).toContain('Could not load more items.');
    await h.press(h.byLabel('Try again'));
    expect(retries).toEqual(['load', 'page']);
  } finally { await h.unmount(); }
});
