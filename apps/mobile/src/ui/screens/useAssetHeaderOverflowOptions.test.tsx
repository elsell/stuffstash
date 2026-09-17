import React, { useLayoutEffect } from 'react';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { useAssetHeaderOverflowOptions } from './useAssetHeaderOverflowOptions';
import type { AssetHeaderOverflowProps } from './AssetHeaderOverflow.types';

type Options = ReturnType<typeof useAssetHeaderOverflowOptions>;
function Probe({ value, observe, scope = 'a' }: { value?: AssetHeaderOverflowProps; observe: (options: Options) => void; scope?: string }) {
  const options = useAssetHeaderOverflowOptions(value, scope);
  useLayoutEffect(() => observe(options), [options, observe]);
  return null;
}
it('retains presentation while invoking only current eligible committed actions', async () => {
  const h = new MobileRenderHarness(); const calls: string[] = [];
  let options!: Options;
  const observe = (value: Options) => { options = value; };
  const props = (name: string, disabled = false, canArchive = true): AssetHeaderOverflowProps => ({
    asset: { title: 'Drill', canArchive, canRestore: false, canDeletePermanently: false }, disabled,
    onHistory: () => calls.push(name), onCheckoutHistory: () => calls.push(`${name}-checkout`),
    onLifecycleAction: action => calls.push(`${name}-${action}`)
  });
  let history!: () => void; let archive!: () => void; let replacementHistory!: () => void;
  try {
    await h.render(<Probe value={props('old')} observe={observe} />);
    const original = options;
    const menu = (options.headerRight as () => React.ReactElement<AssetHeaderOverflowProps>)();
    history = menu.props.onHistory; archive = () => menu.props.onLifecycleAction('archive');
    await h.render(<Probe value={props('current')} observe={observe} />);
    expect(options).toBe(original);
    await h.run(history); await h.run(archive);
    expect(calls).toEqual(['current', 'current-archive']);
    await h.render(<Probe value={props('disabled', true)} observe={observe} />);
    expect(options).not.toBe(original); await h.run(history);
    await h.render(<Probe value={props('restricted', false, false)} observe={observe} />);
    await h.run(archive);
    expect(calls).toEqual(['current', 'current-archive']);
    await h.render(<Probe value={props('replacement')} scope="b" observe={observe} />);
    await h.run(history); await h.run(archive);
    expect(calls).toEqual(['current', 'current-archive']);
    const replacementMenu = (options.headerRight as () => React.ReactElement<AssetHeaderOverflowProps>)();
    replacementHistory = replacementMenu.props.onHistory;
    await h.run(replacementHistory);
    expect(calls).toEqual(['current', 'current-archive', 'replacement']);
    await h.render(<Probe scope="b" observe={observe} />);
    expect(options.headerRight).toBeUndefined(); await h.run(history); await h.run(replacementHistory);
    expect(calls).toEqual(['current', 'current-archive', 'replacement']);
    await h.render(<Probe value={props('restored')} scope="b" observe={observe} />);
    const restoredMenu = (options.headerRight as () => React.ReactElement<AssetHeaderOverflowProps>)();
    replacementHistory = restoredMenu.props.onHistory;
  } finally { await h.unmount(); }
  await h.run(history); await h.run(replacementHistory);
  expect(calls).toEqual(['current', 'current-archive', 'replacement']);
});
