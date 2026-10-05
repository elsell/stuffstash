import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { BrowseSurfaceControl } from './BrowseSurfaceControl';

describe('BrowseSurfaceControl', () => {
  let harness: MobileRenderHarness;

  beforeEach(() => {
    Reflect.set(globalThis, 'expo', {
      getViewConfig: () => ({ directEventTypes: {}, validAttributes: {} })
    });
    harness = new MobileRenderHarness();
  });

  afterEach(async () => {
    Reflect.deleteProperty(globalThis, 'expo');
    await harness.unmount();
  });

  it('shows the current Browse view and selects a view in place with a native menu', async () => {
    const changes: string[] = [];
    await harness.render(
      <BrowseSurfaceControl
        selectedSurface="map"
        onChangeSurface={(surface) => changes.push(surface)}
      />
    );

    expect(harness.byLabel('Browse view: Map')).toBeDefined();
    await harness.press(harness.byLabel('Browse view: Map'));
    const items = harness.all().filter(node => node.props.accessibilityRole === 'menuitem');
    expect(items.map(item => item.props.accessibilityState.selected)).toEqual([false, true]);
    await harness.press(items[0]);
    expect(changes).toEqual(['list']);
  });
});
