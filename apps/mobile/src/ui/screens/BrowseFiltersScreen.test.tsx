import { navigationOptions, resetNavigation } from '../../test-support/navigation';
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
    await h.press(h.byLabel('Choose expiration review'));
    expect(h.byLabel('Choose availability')).toBeDefined();
    expect(h.byLabel('Back to filters')).toBeUndefined();
    await h.press(h.byText('Expired')?.parent ?? undefined);
    expect(expiration).toEqual([{ mode: 'expired', draft: { ...initial, checkoutState: 'available' } }]);
    await h.press(h.byLabel('Cancel filters'));
    expect(cancelled).toBe(1);
  } finally { await h.unmount(); }
});

it('cancels pending verification when Back returns to the filter overview', async () => {
  const h = new MobileRenderHarness(); let cancelled = 0;
  try {
    await h.render(<BrowseFiltersScreen initial={initial} query="" tags={[]} busy onApply={() => {}} onCancel={() => {}} onCancelPending={() => cancelled++} onExpiration={() => {}} />);
    await h.press(h.byLabel('Choose tags'));
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

it('explains an empty tag inventory while retaining Back and Show results', async () => {
  const h = new MobileRenderHarness(); const applied: BrowseFilterDraft[] = [];
  try {
    await h.render(<BrowseFiltersScreen initial={initial} query="" tags={[]} onApply={value => applied.push(value)} onCancel={() => {}} onExpiration={() => {}} />);
    await h.press(h.byLabel('Choose tags'));
    expect(h.allText()).toContain('No tags available');
    await h.press(h.byLabel('Back to filters'));
    await h.press(h.byLabel('Show results'));
    expect(applied).toEqual([initial]);
  } finally { await h.unmount(); }
});

it('explains unmatched tag search and restores the selected tag when search is cleared', async () => {
  resetNavigation();
  const h = new MobileRenderHarness(); const applied: BrowseFilterDraft[] = [];
  try {
    await h.render(<BrowseFiltersScreen initial={initial} query="" tags={[{ id: 'tag', key: 'tools', label: 'Tools' }]} onApply={value => applied.push(value)} onCancel={() => {}} onExpiration={() => {}} />);
    await h.press(h.byLabel('Choose tags'));
    await h.press(h.byLabel('Filter by tag Tools'));
    const options = () => (Object.assign({}, ...navigationOptions()) as { headerSearchBarOptions?: { onFocus: () => void; onChangeText: (event: { nativeEvent: { text: string } }) => void } }).headerSearchBarOptions;
    expect(options()).toBeDefined();
    await h.run(() => {options()!.onFocus(); options()!.onChangeText({ nativeEvent: { text: 'unmatched' } });});
    expect(h.allText()).toContain('No matching tags');
    expect(options()).toBeDefined();
    await h.run(() => {options()!.onFocus(); options()!.onChangeText({ nativeEvent: { text: '' } });});
    expect(h.allText()).not.toContain('No matching tags');
    expect(h.byLabel('Filter by tag Tools')?.props.accessibilityState.checked).toBe(true);
    await h.press(h.byLabel('Show results'));
    expect(applied[0].tagIds).toEqual(['tag']);
    await h.press(h.byLabel('Back to filters'));
    expect(options()).toBeUndefined();
  } finally { await h.unmount(); resetNavigation(); }
});

it('keeps verification recovery inside scrollable content with the draft and actions intact', async () => {
  const h = new MobileRenderHarness(); const applied: BrowseFilterDraft[] = [];
  try {
    await h.render(<BrowseFiltersScreen initial={{ ...initial, tagIds: ['tag'] }} query="" tags={[{ id: 'tag', key: 'tools', label: 'Tools' }]}
      error="Inventory could not be verified" onApply={value => applied.push(value)} onCancel={() => {}} onExpiration={() => {}} />);
    const error = h.allByType('Text').find(node => node.children.includes('Inventory could not be verified'));
    expect(error?.props.accessibilityRole).toBe('alert');
    let parent = error?.parent;
    while (parent && parent.type !== 'ScrollView') parent = parent.parent;
    expect(parent?.type).toBe('ScrollView');
    await h.press(h.byLabel('Show results'));
    expect(applied[0].tagIds).toEqual(['tag']);
  } finally { await h.unmount(); }
});

for (const [label, mode] of [['Expiring soon', 'soon'], ['Expired', 'expired'], ['All dates', 'all']] as const) {
  it(`opens ${mode} directly with current draft and rejects retained commands while busy`, async () => {
    const h = new MobileRenderHarness(); const calls: unknown[] = [];
    const render = (busy = false) => h.render(<BrowseFiltersScreen initial={initial} query="medicine" tags={[]} busy={busy}
      onApply={() => { throw new Error('Review is not filter apply'); }} onCancel={() => {}}
      onExpiration={(mode, draft) => calls.push({ mode, draft })} />);
    try {
      await render();
      expect(h.allText()).toContain('Reviews active items only.');
      await h.press(h.byLabel('Choose expiration review'));
      expect(h.byLabel('Choose tags')).toBeDefined();
      const retained = h.byText(label)?.parent?.props.onPress;
      expect(retained).toBeTypeOf('function');
      await render(true); await h.run(retained);
      expect(calls).toEqual([]);
      await render();
      await h.press(h.byLabel('Choose availability')); await h.press(h.byLabel('Checked out'));
      await h.press(h.byLabel('Choose expiration review')); await h.press(h.byText(label)?.parent ?? undefined);
      expect(calls).toEqual([{ mode, draft: { ...initial, checkoutState: 'checked_out' } }]);
    } finally { await h.unmount(); }
  });
}

it('dismisses the review menu without navigation or changing draft choices', async () => {
  const h = new MobileRenderHarness(); const applied: BrowseFilterDraft[] = [];
  try {
    await h.render(<BrowseFiltersScreen initial={initial} query="" tags={[]}
      onApply={draft => applied.push(draft)} onCancel={() => { throw new Error('Menu dismissal is not filter cancellation'); }}
      onExpiration={() => { throw new Error('Menu dismissal is not navigation'); }} />);
    await h.press(h.byLabel('Choose availability')); await h.press(h.byLabel('Checked out'));
    await h.press(h.byLabel('Choose expiration review'));
    await h.press(h.byLabel('Choose expiration review'));
    expect(h.byText('Expired')).toBeUndefined();
    await h.press(h.byLabel('Show results'));
    expect(applied).toEqual([{ ...initial, checkoutState: 'checked_out' }]);
  } finally { await h.unmount(); }
});
