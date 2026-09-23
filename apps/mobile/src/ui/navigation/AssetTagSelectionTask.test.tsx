import SelectionRoute from '../screens/AssetTagSelectionRouteScreen';
import { useState } from 'react';
import { Pressable, Text } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { resetNavigation, dispatchedActions, setScreenFocused } from '../../test-support/navigation';
import { AssetTagSelectionTaskProvider, useAssetTagSelectionTask, useAssetTagSelectionVisit } from './AssetTagSelectionTask';

function Owner({ scope = 'one', disabled = false }: { scope?: string; disabled?: boolean }) {
  const [ids, setIds] = useState<readonly string[]>(['a']);
  const open = useAssetTagSelectionVisit({ scope, disabled, tags: [{ id: 'a', label: 'Alpha' }, { id: 'b', label: 'Beta' }], selectedIds: ids, onChange: setIds });
  return <><Pressable accessibilityLabel="Choose tags" onPress={open} /><Text>{`Draft: ${ids.join(',')}`}</Text></>;
}
function PresentedTask() { const task = useAssetTagSelectionTask(); return <>{task?.content}<Pressable accessibilityLabel="Native back" onPress={() => task?.cancel()} /></>; }
it('retains the owning draft until Done and starts a clean visit after cancellation', async () => {
  const h = new MobileRenderHarness(); resetNavigation();
  try {
    await h.render(<AssetTagSelectionTaskProvider><Owner /><PresentedTask /></AssetTagSelectionTaskProvider>);
    await h.press(h.byLabel('Choose tags'));
    await h.press(h.byLabel('Select tag Beta'));
    expect(h.byText('Draft: a')).toBeDefined();
    await h.press(h.byLabel('Native back'));
    await h.press(h.byLabel('Choose tags'));
    expect(h.byLabel('Select tag Beta')?.props.accessibilityState.checked).toBe(false);
    await h.press(h.byLabel('Select tag Beta'));
    await h.press(h.byLabel('Done selecting tags'));
    expect(h.byText('Draft: a,b')).toBeDefined();
    expect(dispatchedActions().filter(action => action.type === 'push')).toHaveLength(2);
  } finally { await h.unmount(); resetNavigation(); }
});
it('invalidates an open visit when its inventory changes', async () => {
  const h = new MobileRenderHarness();
  const render = (scope: string) => h.render(<AssetTagSelectionTaskProvider><Owner scope={scope} /><PresentedTask /></AssetTagSelectionTaskProvider>);
  try {
    await render('one'); await h.press(h.byLabel('Choose tags'));
    await h.press(h.byLabel('Select tag Beta'));
    const done = h.byLabel('Done selecting tags')!.props.onPress;
    await render('two'); await h.run(done);
    expect(h.byText('Draft: a')).toBeDefined();
    expect(h.byLabel('Select tag Beta')).toBeUndefined();
  } finally { await h.unmount(); resetNavigation(); }
});

it('does not open from a retained callback after the form loses focus', async () => {
  const h = new MobileRenderHarness(); resetNavigation(); setScreenFocused(true);
  try {
    await h.render(<AssetTagSelectionTaskProvider><Owner /><PresentedTask /></AssetTagSelectionTaskProvider>);
    const open = h.byLabel('Choose tags')!.props.onPress;
    await h.run(() => setScreenFocused(false));
    await h.run(open);
    expect(dispatchedActions()).toEqual([]);
    expect(h.byLabel('Select tag Alpha')).toBeUndefined();
  } finally { await h.unmount(); resetNavigation(); setScreenFocused(true); }
});

it('cancels on native route removal without applying selection edits', async () => {
  const h = new MobileRenderHarness(); resetNavigation();
  const render = (route: boolean) => h.render(<AssetTagSelectionTaskProvider><Owner />{route ? <SelectionRoute /> : null}</AssetTagSelectionTaskProvider>);
  try {
    await render(false); await h.press(h.byLabel('Choose tags')); await render(true);
    await h.press(h.byLabel('Select tag Beta'));
    await render(false);
    expect(h.byText('Draft: a')).toBeDefined();
    await h.press(h.byLabel('Choose tags')); await render(true);
    expect(h.byLabel('Select tag Beta')?.props.accessibilityState.checked).toBe(false);
    await h.press(h.byLabel('Select tag Beta'));
    await h.press(h.byLabel('Done selecting tags'));
    expect(h.byText('Draft: a,b')).toBeDefined();
    expect(dispatchedActions().filter(action => action.type === 'back')).toHaveLength(1);
  } finally { await h.unmount(); resetNavigation(); }
});
