import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { BrowseFiltersScreen, type BrowseFilterDraft } from './BrowseFiltersScreen';
const initial: BrowseFilterDraft = { scope: 'all', lifecycleState: 'active', checkoutState: 'any', tagIds: [], sort: 'updated_desc' };
it('selects short choices in place, keeps draft until Show results and resets all selections', async () => {
  const h = new MobileRenderHarness(); const applied: BrowseFilterDraft[] = [];
  try {
    await h.render(<BrowseFiltersScreen initial={initial} query="" tags={[{ id: 'tag', key: 'tools', label: 'Tools' }]} onApply={draft => applied.push(draft)} onCancel={() => {}} onExpiration={() => {}} />);
    await h.press(h.byLabel('Choose availability'));
    expect(h.byLabel('Choose tags')).toBeDefined();
    await h.press(h.byLabel('Checked out'));
    expect(applied).toEqual([]);
    await h.press(h.byLabel('Choose tags'));
    await h.press(h.byLabel('Filter by tag Tools'));
    await h.press(h.byLabel('Back to filters'));
    await h.press(h.byLabel('Show results'));
    expect(applied[0]).toMatchObject({ checkoutState: 'checked_out', tagIds: ['tag'] });
    await h.press(h.byLabel('Reset all filters'));
    await h.press(h.byLabel('Show results'));
    expect(applied[1]).toEqual(initial);
  } finally { await h.unmount(); }
});
it('cancels without applying and carries draft into expiration review', async () => {
  const h = new MobileRenderHarness(); let cancelled = 0; const expiration: unknown[] = [];
  try {
    await h.render(<BrowseFiltersScreen initial={initial} query="medicine" tags={[]} onApply={() => { throw new Error('Must not apply'); }} onCancel={() => cancelled++} onExpiration={(mode, draft) => expiration.push({ mode, draft })} />);
    await h.press(h.byLabel('Choose availability')); await h.press(h.byLabel('Available'));
    await h.press(h.byLabel('Choose expiration review')); await h.press(h.byLabel('Review expired items'));
    expect(expiration).toEqual([{ mode: 'expired', draft: { ...initial, checkoutState: 'available' } }]);
    await h.press(h.byLabel('Back to filters'));
    await h.press(h.byLabel('Cancel filters'));
    expect(cancelled).toBe(1);
  } finally { await h.unmount(); }
});

it('cancels pending verification when Back returns to the filter overview', async () => {
  const h = new MobileRenderHarness(); let cancelled = 0;
  try {
    await h.render(<BrowseFiltersScreen initial={initial} query="" tags={[]} busy onApply={() => {}} onCancel={() => {}} onCancelPending={() => cancelled++} onExpiration={() => {}} />);
    await h.press(h.byLabel('Choose expiration review'));
    await h.press(h.byLabel('Back to filters'));
    expect(cancelled).toBe(1);
    expect(h.byLabel('Choose availability')).toBeDefined();
  } finally { await h.unmount(); }
});

it('shows the effective relevance order during search instead of an inactive saved sort', async () => {
  const h = new MobileRenderHarness();
  try {
    await h.render(<BrowseFiltersScreen initial={initial} query="medicine" tags={[]} onApply={() => {}} onCancel={() => {}} onExpiration={() => {}} />);
    expect(h.byLabel('Sort, Relevance while searching')).toBeDefined();
  } finally { await h.unmount(); }
});
