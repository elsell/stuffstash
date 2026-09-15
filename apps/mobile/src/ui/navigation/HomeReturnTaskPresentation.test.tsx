import { useRef } from 'react';
import { Text } from 'react-native';
import { afterEach, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { dispatchedActions, resetNavigation, setCanGoBack, setScreenFocused } from '../../test-support/navigation';
import HomeReturnDetailsRoute from '../../app/home-return-details';
import { HomeReturnTaskProvider, useHomeReturnTask, useHomeReturnTaskPresentation } from './HomeReturnTaskPresentation';

function Owner({ open }: { readonly open: boolean }) {
  useHomeReturnTaskPresentation(open ? { content: <Text>Retained task</Text>, requestClose: () => {}, focusChanged: () => {} } : undefined);
  return null;
}
function RouteFrame() {
  const task = useHomeReturnTask();
  const presented = useRef(false);
  if (task) presented.current = true;
  return presented.current ? <HomeReturnDetailsRoute /> : null;
}
let h = new MobileRenderHarness();
afterEach(async () => { await h.unmount(); resetNavigation(); setCanGoBack(true); setScreenFocused(true); });
it('defers self-dismissal until focused when its owner disappears behind another screen', async () => {
  h = new MobileRenderHarness(); resetNavigation(); setScreenFocused(true);
  const render = (open: boolean) => h.render(<HomeReturnTaskProvider><Owner open={open} /><RouteFrame /></HomeReturnTaskProvider>);
  await render(true);
  expect(h.byText('Retained task')).toBeDefined();
  await h.run(() => setScreenFocused(false));
  await render(false);
  expect(h.byText('Retained task')).toBeUndefined();
  expect(dispatchedActions().filter(action => action.type === 'back')).toHaveLength(0);
  await h.run(() => setScreenFocused(true));
  expect(dispatchedActions().filter(action => action.type === 'back')).toHaveLength(1);
});

it('dismisses its focused route when Save clears the task', async () => {
  h = new MobileRenderHarness(); resetNavigation(); setScreenFocused(true);
  const render = (open: boolean) => h.render(<HomeReturnTaskProvider><Owner open={open} /><RouteFrame /></HomeReturnTaskProvider>);
  await render(true);
  await render(false);
  expect(dispatchedActions().filter(action => action.type === 'back')).toHaveLength(1);
});

it('recovers an entry without an owned task to Home when there is no back destination', async () => {
  h = new MobileRenderHarness(); resetNavigation(); setScreenFocused(true); setCanGoBack(false);
  await h.render(<HomeReturnTaskProvider><HomeReturnDetailsRoute /></HomeReturnTaskProvider>);
  expect(dispatchedActions()).toEqual([{ type: 'replace', href: '/' }]);
});
