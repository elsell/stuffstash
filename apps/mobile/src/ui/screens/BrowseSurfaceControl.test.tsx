import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { lightPalette } from '../theme/tokens';
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

  it('maps the durable Browse surface state through the shared native segmented control', async () => {
    const changes: string[] = [];
    await harness.render(
      <BrowseSurfaceControl
        palette={lightPalette}
        selectedSurface="map"
        onChangeSurface={(surface) => changes.push(surface)}
      />
    );

    const control = harness.byType('NativeSegmentedControl');
    expect(control?.props).toMatchObject({
      enabled: true,
      selectedIndex: 1,
      values: ['List', 'Map']
    });

    await harness.change(control, 'List');
    expect(changes).toEqual(['list']);
  });
});
