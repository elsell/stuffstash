import { expect, it } from 'vitest';
import { createRequire } from 'node:module';
import { pathToFileURL } from 'node:url';
// Resolve the installed transitive router through its owning packages rather
// than adding a second navigation dependency to the application.
const localRequire = createRequire(import.meta.url);
const nativeRequire = createRequire(localRequire.resolve('@react-navigation/native/package.json'));
const coreRequire = createRequire(nativeRequire.resolve('@react-navigation/core/package.json'));
const { StackRouter } = await import(pathToFileURL(coreRequire.resolve('@react-navigation/routers')).href);
// @ts-expect-error Vite provides the production layout source.
import layout from '../../app/_layout.tsx?raw';

// Exercise the real native-stack state machine with the application's declared
// screens and initial option. The onboarding gate mounts it without route state.
const routeNames = [...layout.matchAll(/<Stack\.Screen\s+name="([^"]+)"/g)].map(match => match[1]);
const initialRouteName = layout.match(/<Stack\s+initialRouteName="([^"]+)"/)?.[1];
const options = { routeNames, routeParamList: {}, routeGetIdList: {} };

it('opens the tab shell when onboarding mounts a stack without route state', () => {
  const router = StackRouter({ initialRouteName });
  const state = router.getInitialState(options);
  expect(state.routes[state.index].name).toBe('(tabs)');
});

it('uses the tab shell when restoring an empty navigation state', () => {
  const router = StackRouter({ initialRouteName });
  const state = router.getRehydratedState({ stale: true, routes: [] }, options);
  expect(state.routes[state.index].name).toBe('(tabs)');
});

it('preserves an explicit invitation destination instead of forcing Home', () => {
  const router = StackRouter({ initialRouteName });
  const state = router.getRehydratedState({ stale: true, index: 0,
    routes: [{ name: 'invitations/accept', params: { reference: 'synthetic-invitation' } }] }, options);
  expect(state.routes[state.index]).toMatchObject({ name: 'invitations/accept',
    params: { reference: 'synthetic-invitation' } });
});
