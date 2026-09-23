import SelectionRoute from '../screens/AddDestinationRouteScreen';
import { useState } from 'react';
import { Pressable, Text } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { resetNavigation, attemptNavigation, dispatchedActions } from '../../test-support/navigation';
import { AddDestinationTaskProvider, useAddDestinationTask, useAddDestinationPresentation } from './AddDestinationTask';

function Owner({ blocked = false }: { blocked?: boolean }) {
  const [open, setOpen] = useState(false);
  useAddDestinationPresentation(open ? { blocked, onClose: () => setOpen(false), content: <Text>Location selection</Text> } : undefined);
  return <Pressable accessibilityLabel="Open location" onPress={() => setOpen(true)} />;
}
function Task() { const task = useAddDestinationTask(); return <>{task?.content}<Pressable accessibilityLabel="Cancel location" onPress={task?.cancel} /></>; }
it('keeps one selection visit through updates and protects pending creation', async () => {
  const h = new MobileRenderHarness(); resetNavigation();
  const render = (blocked: boolean) => h.render(<AddDestinationTaskProvider><Owner blocked={blocked} /><Task /></AddDestinationTaskProvider>);
  try {
    await render(false); await h.press(h.byLabel('Open location'));
    const cancel = h.byLabel('Cancel location')!.props.onPress;
    await render(true); await h.run(cancel);
    expect(h.byText('Location selection')).toBeDefined();
    expect(dispatchedActions().filter(action => action.type === 'push')).toHaveLength(1);
    await render(false); await h.run(cancel);
    expect(h.byText('Location selection')).toBeUndefined();
    await h.press(h.byLabel('Open location')); await h.run(cancel);
    expect(h.byText('Location selection')).toBeDefined();
  } finally { await h.unmount(); resetNavigation(); }
});

it('retires the selection visit on native route removal and can reopen cleanly', async () => {
  const h = new MobileRenderHarness(); resetNavigation();
  const render = (route: boolean) => h.render(<AddDestinationTaskProvider><Owner /><Task />{route ? <SelectionRoute /> : null}</AddDestinationTaskProvider>);
  try {
    await render(false); await h.press(h.byLabel('Open location')); await render(true);
    await render(false); expect(h.byText('Location selection')).toBeUndefined();
    await h.press(h.byLabel('Open location')); await render(true);
    await h.press(h.byLabel('Cancel location'));
    expect(h.byText('Location selection')).toBeUndefined();
    expect(dispatchedActions().filter(action => action.type === 'back')).toHaveLength(1);
  } finally { await h.unmount(); resetNavigation(); }
});

it('blocks native removal during creation and allows it after the operation ends', async () => {
  const h = new MobileRenderHarness(); resetNavigation();
  const render = (blocked: boolean) => h.render(<AddDestinationTaskProvider><Owner blocked={blocked} /><SelectionRoute /></AddDestinationTaskProvider>);
  try {
    await render(false); await h.press(h.byLabel('Open location'));
    await render(true);
    await h.run(() => attemptNavigation({ type: 'GO_BACK' }));
    expect(dispatchedActions().filter(action => action.type === 'GO_BACK')).toEqual([]);
    await render(false);
    await h.run(() => attemptNavigation({ type: 'GO_BACK' }));
    expect(dispatchedActions().filter(action => action.type === 'GO_BACK')).toHaveLength(1);
  } finally { await h.unmount(); resetNavigation(); }
});
