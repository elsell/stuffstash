import React from 'react';
import { describe, expect, it, vi } from 'vitest';
import { MobileRenderHarness } from '../../test-support/render';
import { NativeRefinementButton } from './NativeRefinementButton';

describe('NativeRefinementButton contract', () => {
  it('keeps a 44-point target and forwards expanded accessibility state', async () => {
    const harness = new MobileRenderHarness();
    const onPress = vi.fn();
    await harness.render(<NativeRefinementButton
      accessibilityLabel="Filters, 2 applied"
      accessibilityState={{ expanded: false }}
      badgeCount={2}
      iconOnly
      label="Filters 2"
      onPress={onPress}
      systemImage="line.3.horizontal.decrease"
    />);

    const control = harness.byLabel('Filters, 2 applied');
    expect(control?.props.accessibilityRole).toBe('button');
    expect(control?.props.accessibilityState).toEqual({ disabled: false, expanded: false });
    const styles = Array.isArray(control?.props.style) ? control.props.style : [control?.props.style];
    expect(styles).toContainEqual(expect.objectContaining({ height: 44, width: 44 }));
    expect(harness.byText('2')).toBeDefined();
    await harness.press(control);
    expect(onPress).toHaveBeenCalledOnce();
  });

  it('does not invoke a disabled control', async () => {
    const harness = new MobileRenderHarness();
    const onPress = vi.fn();
    await harness.render(<NativeRefinementButton
      accessibilityLabel="Filters unavailable"
      disabled
      label="Filters"
      onPress={onPress}
    />);

    const control = harness.byLabel('Filters unavailable');
    expect(control?.props.accessibilityState).toEqual({ disabled: true });
    await harness.press(control);
    expect(onPress).not.toHaveBeenCalled();
  });
});
