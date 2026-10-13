import { expect, it } from 'vitest';
import { createRequire } from 'node:module';
import { pathToFileURL } from 'node:url';
import { openScannedAsset } from './openScannedAsset';
const localRequire = createRequire(import.meta.url);
const nativeRequire = createRequire(localRequire.resolve('@react-navigation/native/package.json'));
const coreRequire = createRequire(nativeRequire.resolve('@react-navigation/core/package.json'));
const { StackRouter } = await import(pathToFileURL(coreRequire.resolve('@react-navigation/routers')).href);

it.each([true, false])('requests anchored Home detail and reuses the root tab shell (existing tabs=%s)', warm => {
  const stack = StackRouter({ initialRouteName: '(tabs)' });
  const options = { routeNames: ['(tabs)', 'scan-label'], routeParamList: {}, routeGetIdList: {} };
  const state = stack.getRehydratedState({ stale: true, index: warm ? 1 : 0,
    routes: warm ? [{ name: '(tabs)', key: 'existing-tabs' }, { name: 'scan-label' }] : [{ name: 'scan-label' }] }, options);
  let result = state;
  openScannedAsset({ dismissTo(href, navigationOptions) {
    expect(href).toEqual({ pathname: '/(tabs)/(home)/assets/[assetId]', params: { assetId: 'drill' } });
    expect(navigationOptions).toEqual({ withAnchor: true });
    // Expo maps dismissTo to POP_TO at the divergent root stack. withAnchor
    // supplies initial:false to nested navigators, retaining their root screen.
    result = stack.getStateForAction(state, { type: 'POP_TO', payload: { name: '(tabs)' } }, options)!;
  } }, 'drill');
  expect(result.routes.map((route: { name: string }) => route.name)).toEqual(['(tabs)']);
  if (warm) expect(result.routes[0].key).toBe('existing-tabs');
});
