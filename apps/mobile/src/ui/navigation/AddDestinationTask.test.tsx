import { useState } from 'react';
import { Pressable, Text } from 'react-native';
import { expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { resetNavigation, dispatchedActions } from '../../test-support/navigation';
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
