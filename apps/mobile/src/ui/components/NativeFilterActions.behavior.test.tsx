import React from 'react';
import { Text } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { setScreenFocused } from '../../test-support/navigation';
import { NativeFilterSheet as IOSFilter } from './NativeFilterSheet';
import { NativeFilterSheet as AndroidFilter } from './NativeFilterSheet.android';

it.each([['iOS', IOSFilter], ['Android', AndroidFilter]] as const)(
  '%s footer owns current draft, page, enabled state and focused lifetime', async (_, Filter) => {
    const h = new MobileRenderHarness();
    const calls: string[] = [];
    const render = (draft: string, page: string, disabled = false, secondaryDisabled = false) =>
      h.render(<Filter title={page} footerTestID="footer" actions={{
        primaryLabel: 'Apply', secondaryLabel: page === 'Filters' ? 'Cancel' : 'Back',
        disabled, secondaryDisabled,
        onApply: () => calls.push(`apply ${draft}`), onBack: () => calls.push(`back ${page}`)
      }}><Text>{draft}</Text></Filter>);
    let apply!: () => void;
    let back!: () => void;
    try {
      await render('old valid range', 'Filters');
      apply = h.byLabel('Apply')!.props.onPress;
      back = h.byLabel('Cancel')!.props.onPress;
      await render('new valid range', 'Dates');
      await h.run(apply); await h.run(back);
      expect(calls).toEqual(['apply new valid range', 'back Dates']);
      calls.length = 0;
      await render('invalid range', 'Dates', true, true);
      await h.run(apply); await h.run(back);
      expect(calls).toEqual([]);
      await render('latest valid range', 'Dates');
      await h.run(() => setScreenFocused(false));
      await h.run(apply); await h.run(back);
      expect(calls).toEqual([]);
      await h.run(() => setScreenFocused(true));
      await h.run(apply); await h.run(back);
      expect(calls).toEqual(['apply latest valid range', 'back Dates']);
      calls.length = 0;
    } finally { await h.unmount(); setScreenFocused(true); }
    await h.run(apply); await h.run(back);
    expect(calls).toEqual([]);
  }
);
