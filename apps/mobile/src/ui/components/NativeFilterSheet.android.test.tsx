import React from 'react';
import { Text } from 'react-native';
import { expect, it } from 'vitest';
import { NativeFilterSheet } from './NativeFilterSheet.android';
import { MobileRenderHarness } from '../../test-support/render';
import { setCanGoBack, setScreenFocused } from '../../test-support/navigation';

it('delivers current draft actions and guarding teardown', async () => {
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
    await h.render(screen('current'));
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
    expect(h.byLabel('Show results')).toBeDefined();
    expect(h.byLabel('Cancel')).toBeDefined();
  } finally { await h.unmount(); setCanGoBack(true); }
});

it('provides editable search inside the Android sheet and removes it on page return', async () => {
  const h = new MobileRenderHarness();
  const changes: string[] = [];
  const actions = { disabled: false, primaryLabel: 'Show results', secondaryLabel: 'Back', onApply: () => {}, onBack: () => {} };
  try {
    await h.render(<NativeFilterSheet title="Tags" footerTestID="footer" actions={actions}
      search={{ query: '', placeholder: 'Search tags', onChange: value => changes.push(value), onSubmit: value => changes.push(`submit ${value}`), onClear: () => changes.push('clear') }}><Text>Tags</Text></NativeFilterSheet>);
    const input = h.byLabel('Search tags');
    expect(input).toBeDefined();
    await h.run(() => input!.props.onChangeText('tools'));
    await h.run(() => input!.props.onSubmitEditing({ nativeEvent: { text: 'tools' } }));
    expect(changes).toEqual(['tools', 'submit tools']);
    await h.run(() => setScreenFocused(false));
    await h.run(() => input!.props.onChangeText('blurred'));
    expect(changes).toEqual(['tools', 'submit tools']);
    await h.run(() => setScreenFocused(true));
    await h.run(() => input!.props.onChangeText('current'));
    expect(changes).toEqual(['tools', 'submit tools', 'current']);
    await h.render(<NativeFilterSheet title="Filters" footerTestID="footer" actions={actions}><Text>Filters</Text></NativeFilterSheet>);
    expect(h.byLabel('Search tags')).toBeUndefined();
    await h.run(() => input!.props.onChangeText('late'));
    await h.run(() => input!.props.onSubmitEditing({ nativeEvent: { text: 'late' } }));
    expect(changes).toEqual(['tools', 'submit tools', 'current']);
  } finally { await h.unmount(); }
});
