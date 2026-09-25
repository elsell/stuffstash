import { typeInNativeSearch } from '../../test-support/NativeSearchDriver';
import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { resetNavigation } from '../../test-support/navigation';
import { AssetTagSelectionScreen } from './AssetTagSelectionScreen';
const tags = Array.from({ length: 30 }, (_, i) => ({ id: `tag-${i}`, label: `Tag ${i}` }));
const search = typeInNativeSearch;

it('selects beyond the first choices, preserves hidden selections, and commits only on Done', async () => {
  const h = new MobileRenderHarness(); const applied: string[][] = []; resetNavigation();
  try {
    await h.render(<AssetTagSelectionScreen tags={tags} initialSelectedIds={['tag-1']} onDone={ids => applied.push([...ids])} onCancel={() => {}} />);
    await h.press(h.byLabel('Select tag Tag 29'));
    await h.run(() => search('Tag 2'));
    expect(h.byLabel('Select tag Tag 1')).toBeUndefined();
    expect(applied).toEqual([]);
    await h.press(h.byLabel('Select tag Tag 2'));
    await h.press(h.byLabel('Done selecting tags'));
    expect(applied).toEqual([['tag-1', 'tag-29', 'tag-2']]);
  } finally { await h.unmount(); resetNavigation(); }
});

it('cancels selection edits without changing the parent draft', async () => {
  const h = new MobileRenderHarness(); let canceled = 0;
  try {
    await h.render(<AssetTagSelectionScreen tags={tags} initialSelectedIds={['tag-1']} onDone={() => { throw new Error('Must not commit'); }} onCancel={() => canceled++} />);
    await h.press(h.byLabel('Select tag Tag 1'));
    await h.press(h.byLabel('Cancel selecting tags'));
    expect(canceled).toBe(1);
  } finally { await h.unmount(); }
});

it('rejects a retained Done callback when its owner becomes unavailable', async () => {
  const h = new MobileRenderHarness(); const applied: unknown[] = [];
  const render = (available: boolean) => h.render(<AssetTagSelectionScreen tags={tags} initialSelectedIds={[]} available={available} onDone={ids => applied.push(ids)} onCancel={() => {}} />);
  try {
    await render(true);
    await h.press(h.byLabel('Select tag Tag 29'));
    const complete = h.byLabel('Done selecting tags')?.props.onPress;
    await render(false);
    await h.run(() => complete?.());
    expect(applied).toEqual([]);
    expect(h.byLabel('Cancel selecting tags')?.props.disabled).not.toBe(true);
  } finally { await h.unmount(); }
});


it('reviews selected tags without losing the search query', async () => {
  const h = new MobileRenderHarness(); resetNavigation();
  try {
    await h.render(<AssetTagSelectionScreen tags={tags} initialSelectedIds={['tag-29']} onDone={() => {}} onCancel={() => {}} />);
    await h.run(() => search('unmatched'));
    expect(h.byText('No matching tags')).toBeDefined();
    const control = () => h.all().find(node => node.props.values?.includes('All tags'));
    expect(control()).toBeDefined();
    await h.run(() => control()!.props.onValueChange('Selected'));
    expect(h.byLabel('Select tag Tag 29')).toBeDefined();
    await h.run(() => control()!.props.onValueChange('All tags'));
    expect(h.byLabel('Select tag Tag 29')).toBeUndefined();
    expect(h.byText('No matching tags')).toBeDefined();
  } finally { await h.unmount(); resetNavigation(); }
});


it('keeps unavailable assignments until explicitly removed and lets people review them', async () => {
  const h = new MobileRenderHarness(); const applied: string[][] = [];
  try {
    await h.render(<AssetTagSelectionScreen tags={[]} initialSelectedIds={['former-tag']} onDone={ids => applied.push([...ids])} onCancel={() => {}} />);
    await h.press(h.byLabel('Done selecting tags'));
    expect(applied).toEqual([['former-tag']]);
    const control = h.all().find(node => node.props.values?.includes('All tags'));
    await h.run(() => control!.props.onValueChange('Selected'));
    expect(h.byText('No tags selected')).toBeUndefined();
    await h.press(h.byLabel('Remove unavailable tag 1'));
    await h.press(h.byLabel('Done selecting tags'));
    expect(applied[1]).toEqual([]);
  } finally { await h.unmount(); }
});


it('creates a tag in a separate form and stages it only when selection is confirmed', async () => {
  const h = new MobileRenderHarness(); const applied: unknown[] = [];
  try {
    await h.render(<AssetTagSelectionScreen tags={tags} initialSelectedIds={['tag-1']} initialNewTags={[]}
      onDone={(ids, newTags) => applied.push({ ids, newTags })} onCancel={() => {}} />);
    await h.press(h.byLabel('New tag'));
    expect(h.byLabel('Select tag Tag 1')).toBeUndefined();
    await h.changeText(h.byLabel('New tag name'), 'Camping');
    await h.press(h.byLabel('Add tag'));
    expect(h.byLabel('New tag name')).toBeUndefined();
    expect(applied).toEqual([]);
    expect(h.byLabel('Remove new tag Camping')).toBeDefined();
    await h.press(h.byLabel('New tag'));
    await h.changeText(h.byLabel('New tag name'), 'Discarded');
    await h.press(h.byLabel('Cancel new tag'));
    await h.press(h.byLabel('Done selecting tags'));
    expect(applied).toEqual([{ ids: ['tag-1'], newTags: [{ displayName: 'Camping' }] }]);
  } finally { await h.unmount(); }
});
