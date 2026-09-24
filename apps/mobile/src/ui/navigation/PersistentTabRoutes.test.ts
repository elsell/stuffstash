import { expect, it } from 'vitest';
import { getRoutes } from 'expo-router/build/getRoutesCore';
import type { RequireContext } from 'expo-router/build/types';

// The real router expands the actual route filenames; screen modules are not
// loaded because this checks navigation ownership, not their rendering.
// @ts-expect-error Vite supplies the route manifest.
const files = import.meta.glob('../../app/**/*.tsx') as Record<string, unknown>;
const context = Object.assign(() => ({ default: () => null }), {
  keys: () => Object.keys(files).map(path => path.replace('../../app/', './')),
  resolve: (path: string) => path,
  id: 'production-mobile-routes'
}) as RequireContext;

it('keeps ordinary destinations inside both native tabs and scoped tasks above them', () => {
  const root = getRoutes(context, { platform: 'ios', ignoreEntryPoints: true,
    getSystemRoute: route => ({ ...route, contextKey: route.route, children: [], dynamic: null, loadRoute: () => ({}) }) });
  const tabs = root!.children.find(route => route.route === '(tabs)')!;
  expect(tabs).toBeDefined();
  for (const group of ['(home)', '(search)']) {
    const stack = tabs.children.find(route => route.route === group);
    expect(stack, group).toBeDefined();
    for (const destination of ['assets/[assetId]/index', 'expiration', 'notifications',
      'settings/index', 'settings/inventory/tags/[resourceId]', 'assets/[assetId]/history/index']) {
      expect(stack!.children.map(route => route.route), `${group}: ${destination}`).toContain(destination);
      expect(root!.children.map(route => route.route)).not.toContain(destination);
    }
  }
  for (const modal of ['add', 'assets/[assetId]/move', 'browse-filters', 'tenant-switcher']) {
    expect(root!.children.map(route => route.route)).toContain(modal);
  }
});
