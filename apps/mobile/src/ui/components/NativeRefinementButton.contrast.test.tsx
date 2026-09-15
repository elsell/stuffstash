import { contrastRatio } from '../../test-support/color-contrast';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { AppearancePreferenceController, type AppearancePreference } from '../../application/settings/AppearancePreference';
import { MobileRenderHarness } from '../../test-support/render';
import { resetNativeTestState, setDarkerSystemColorsEnabledForTest, setHighTextContrastEnabledForTest } from '../../test-support/react-native';
import { AppearanceProvider } from '../theme/AppearanceContext';
import { NativeRefinementButton as FallbackButton } from './NativeRefinementButton';
import { NativeRefinementButton as IOSButton } from './NativeRefinementButton.ios';

let harness: MobileRenderHarness;
beforeEach(() => { resetNativeTestState(); harness = new MobileRenderHarness(); });
afterEach(async () => { await harness.unmount(); resetNativeTestState(); });

for (const [platform, Button] of [['fallback', FallbackButton], ['iOS', IOSButton]] as const) {
  describe(`${platform} refinement badge contrast`, () => {
    it.each([
      ['light', false], ['dark', false], ['light', true], ['dark', true]
    ] as const)('is legible in %s with increased contrast %s', async (appearance, contrast) => {
      setDarkerSystemColorsEnabledForTest(contrast);
      setHighTextContrastEnabledForTest(contrast);
      let saved: AppearancePreference = appearance;
      const controller = new AppearancePreferenceController({
        async load() { return saved; }, async save(value) { saved = value; }
      });
      await harness.render(<AppearanceProvider controller={controller}>
        <Button badgeCount={2} accessibilityLabel="Filters, 2 applied" label="Filters" onPress={() => {}} />
      </AppearanceProvider>);
      const count = harness.byText('2');
      const badge = harness.allByType('View').find(node => node.props.pointerEvents === 'none');
      expect(count).toBeDefined();
      expect(badge).toBeDefined();
      const foreground = flatten(count!.props.style).color;
      const background = flatten(badge!.props.style).backgroundColor;
      expect(contrastRatio(foreground, background)).toBeGreaterThanOrEqual(4.5);
    });
  });
}

function flatten(style: unknown): Record<string, string> {
  return Object.assign({}, ...(Array.isArray(style) ? style.flat(Infinity) : [style]));
}
