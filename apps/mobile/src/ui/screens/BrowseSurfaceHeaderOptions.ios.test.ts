import { expect, it } from 'vitest';
import { browseSurfaceHeaderOptions } from './BrowseSurfaceHeaderOptions.ios';

it('uses a leading native selection menu and leaves trailing actions untouched', () => {
  const selected: string[] = [];
  const options = browseSurfaceHeaderOptions('map', next => selected.push(next));
  expect(options.headerTitle).toBe('');
  expect(options.headerRight).toBeUndefined();
  expect(options.unstable_headerRightItems).toBeUndefined();
  const menu = options.unstable_headerLeftItems?.({ canGoBack: false })[0];
  expect(menu).toMatchObject({ type: 'menu', label: 'Map', accessibilityLabel: 'Browse view: Map' });
  if (menu?.type !== 'menu') throw new Error('Expected a native view menu');
  expect(menu.menu.items).toMatchObject([
    { type: 'action', label: 'List', state: 'off' },
    { type: 'action', label: 'Map', state: 'on' }
  ]);
  const list = menu.menu.items[0];
  if (list.type !== 'action') throw new Error('Expected a view selection');
  list.onPress();
  expect(selected).toEqual(['list']);
});
