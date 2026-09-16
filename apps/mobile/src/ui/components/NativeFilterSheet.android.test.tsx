import React from 'react';
import { Text } from 'react-native';
import { expect, it } from 'vitest';
import { NativeFilterSheet } from './NativeFilterSheet.android';
import { MobileRenderHarness } from '../../test-support/render';
import { navigationOptions, setCanGoBack } from '../../test-support/navigation';

it('keeps native footer presentation stable while delivering current draft actions and guarding teardown', async () => {
  const h = new MobileRenderHarness();
  const calls: string[] = [];
  const screen = (draft: string, disabled = false, secondaryDisabled = false) =>
    <NativeFilterSheet title="Filters" footerTestID="filter-footer" actions={{
      primaryLabel: 'Show results', secondaryLabel: 'Cancel', disabled, secondaryDisabled,
      onApply: () => calls.push(draft), onBack: () => calls.push(`cancel ${draft}`)
    }}><Text>Content</Text></NativeFilterSheet>;
  let apply: () => void = () => {}; let cancel: () => void = () => {};
  try {
    await h.render(screen('old'));
    apply = h.byLabel('Show results')!.props.onPress;
    cancel = h.byLabel('Cancel')!.props.onPress;
    const options = navigationOptions().at(-1);
    await h.render(screen('current'));
    expect(navigationOptions().at(-1)).toBe(options);
    await h.run(apply); await h.run(cancel);
    expect(calls).toEqual(['current', 'cancel current']);
    await h.render(screen('locked', true, true));
    await h.run(apply); await h.run(cancel);
    expect(calls).toEqual(['current', 'cancel current']);
    expect(h.byLabel('Show results')!.props.disabled).toBe(true);
  } finally { await h.unmount(); }
  await h.run(apply); await h.run(cancel);
  expect(calls).toEqual(['current', 'cancel current']);
});

it('retains visible actions when a cold link has no underlying sheet route', async () => {
  const h = new MobileRenderHarness();
  setCanGoBack(false);
  try {
    await h.render(<NativeFilterSheet title="Filters" footerTestID="root-footer" actions={{
      primaryLabel: 'Show results', secondaryLabel: 'Cancel', disabled: false,
      onApply: () => {}, onBack: () => {}
    }}><Text>Content</Text></NativeFilterSheet>);
    expect(navigationOptions().at(-1)).toMatchObject({ headerShown: false, unstable_sheetFooter: undefined });
    expect(h.byLabel('Show results')).toBeDefined();
    expect(h.byLabel('Cancel')).toBeDefined();
  } finally { await h.unmount(); setCanGoBack(true); }
});
