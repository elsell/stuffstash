import React from 'react';
import { expect, it } from 'vitest';
import { Stack } from 'expo-router';
import { useNativeHeaderActionOptions } from './useNativeHeaderActionOptions';
import type { NativeHeaderAction } from './NativeHeaderActions.types';
import { MobileRenderHarness } from '../../test-support/render';

function Header({ actions }: { actions: readonly NativeHeaderAction[] }) {
  const options = useNativeHeaderActionOptions(actions);
  return <Stack.Screen options={options} />;
}

it('uses the committed callback and guards disabled, removed and unmounted native actions', async () => {
  const h = new MobileRenderHarness();
  const calls: string[] = [];
  const action = (value: string, disabled = false): NativeHeaderAction => ({
    kind: 'save', label: 'Save item', disabled, onPress: () => { calls.push(value); }
  });
  let stale: () => void;
  try {
    await h.render(<Header actions={[action('old')]} />);
    stale = h.byLabel('Save item')!.props.onPress;
    await h.render(<Header actions={[action('current')]} />);
    await h.run(stale);
    expect(calls).toEqual(['current']);
    await h.render(<Header actions={[action('disabled', true)]} />);
    await h.run(stale);
    expect(calls).toEqual(['current']);
    await h.render(<Header actions={[]} />);
    await h.run(stale);
    expect(calls).toEqual(['current']);
    await h.render(<Header actions={[action('new')]} />);
    await h.run(stale);
    expect(calls).toEqual(['current', 'new']);
  } finally { await h.unmount(); }
  await h.run(stale!);
  expect(calls).toEqual(['current', 'new']);
});
