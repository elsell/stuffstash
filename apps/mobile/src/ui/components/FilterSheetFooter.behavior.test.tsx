import React from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { BrowseFiltersScreen } from '../screens/BrowseFiltersScreen';
import { ExpirationFiltersScreen } from '../expiration/ExpirationFiltersScreen';

const flatten = (style: unknown): Record<string, unknown> => Array.isArray(style) ? Object.assign({}, ...style.map(flatten)) : style && typeof style === 'object' ? style as Record<string, unknown> : {};

it.each(['browse', 'expiration'])('%s reserves its measured footer and preserves the last tag through resizing and Back', async kind => {
  const h = new MobileRenderHarness();
  const tags = Array.from({ length: 30 }, (_, index) => ({ id: `tag-${index + 1}`, key: `tag-${index + 1}`, label: `Tag ${index + 1}` }));
  const applied: unknown[] = [];
  try {
    await h.render(kind === 'browse'
      ? <BrowseFiltersScreen initial={{ scope: 'all', lifecycleState: 'active', checkoutState: 'any', tagIds: [], sort: 'updated_desc' }} query="" tags={tags} onApply={draft => applied.push(draft)} onCancel={() => {}} onExpiration={() => {}} />
      : <ExpirationFiltersScreen initial={{ mode: 'all' }} choices={{ types: [], locations: [], tags }} onApply={draft => applied.push(draft)} onCancel={() => {}} />);
    await h.press(h.byLabel('Choose tags'));
    await h.press(h.byLabel(kind === 'browse' ? 'Filter by tag Tag 30' : 'Tag 30'));
    const footer = h.byType('SafeAreaView');
    expect(footer?.props.onLayout).toBeTypeOf('function');
    for (const height of [204, 148]) {
      await h.run(() => footer!.props.onLayout({ nativeEvent: { layout: { height, width: 390, x: 0, y: 600 } } }));
      const scroll = h.byType('ScrollView');
      expect(flatten(scroll?.props.contentContainerStyle).paddingBottom).toBeGreaterThanOrEqual(height);
      expect(scroll?.props.scrollIndicatorInsets.bottom).toBe(height);
      const background = flatten(footer?.props.style).backgroundColor;
      expect(background).toBe(flatten(scroll?.props.style).backgroundColor);
      expect(background).toMatch(/^#[a-fA-F0-9]{6}$/);
    }
    await h.press(h.byLabel(kind === 'browse' ? 'Back to filters' : 'Cancel or return to filters'));
    expect(applied).toEqual([]);
    await h.press(h.byLabel(kind === 'browse' ? 'Show results' : 'Apply expiration filters'));
    expect(applied).toEqual([expect.objectContaining({ tagIds: ['tag-30'] })]);
  } finally { await h.unmount(); }
});
