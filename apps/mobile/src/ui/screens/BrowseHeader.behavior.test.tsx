import React from 'react';
import { afterEach, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { SearchHeader } from './SearchScreen';
import { lightPalette } from '../theme/tokens';

let h: MobileRenderHarness;
afterEach(async () => { await h?.unmount(); });
async function mount(overrides: Partial<Parameters<typeof SearchHeader>[0]> = {}) {
  h = new MobileRenderHarness();
  await h.render(<SearchHeader isLoading={false} lifecycleState="active" checkoutState="any"
    palette={lightPalette} resultCount={0} scope="all" selectedTagIds={[]}
    sort="updated_desc" submittedQuery="" onClearFilters={() => {}}
    onRemoveFilter={() => {}} {...overrides} />);
}

it('keeps search and creation out of content headers and leaves the view switcher to navigation', async () => {
  await mount();
  expect(h.byLabel('Add an asset')).toBeUndefined();
  expect(h.all().some(node => node.props.placeholder === 'Search names, places, or tags')).toBe(false);
  expect(h.byLabel('Browse view')).toBeUndefined();
  expect(h.allText()).not.toContain('Home inventory');

});

it('describes submitted results without claiming a total', async () => {
  await mount({ resultCount: 20, submittedQuery: 'mug' });
  expect(h.allText()).toContain('20 shown for “mug” · relevance');
});

it('keeps the filter command out of scrolling content', async () => {
  await mount({ scope: 'containers', lifecycleState: 'archived' });
  expect(h.byLabel('Filters, 2 applied')).toBeUndefined();
});

it('removes a named filter and clears multiple applied refinements', async () => {
  const removed: unknown[] = []; let cleared = 0;
  await mount({ lifecycleState: 'archived', selectedTagIds: ['tag-tools'],
    tagFilters: [{ id: 'tag-tools', key: 'tools', label: 'Tools' }],
    onRemoveFilter: token => removed.push(token), onClearFilters: () => { cleared++; } });
  expect(h.allText()).toEqual(expect.arrayContaining(['Archived', 'Tools', 'Clear all']));
  await h.press(h.byLabel('Remove filter Tools'));
  expect(removed).toEqual([{ key: 'tag:tag-tools', label: 'Tools', type: 'tag', tagId: 'tag-tools' }]);
  await h.press(h.byText('Clear all')?.parent ?? undefined);
  expect(cleared).toBe(1);
});
