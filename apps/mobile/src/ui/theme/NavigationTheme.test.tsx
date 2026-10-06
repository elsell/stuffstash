import { afterEach, expect, it } from 'vitest';
import { DarkTheme, DefaultTheme, useTheme, type Theme } from '@react-navigation/native';
import { View } from 'react-native';
import { AppearancePreferenceController } from '../../application/settings/AppearancePreference';
import { MobileRenderHarness } from '../../test-support/render';
import { resetNativeTestState, setSystemColorSchemeForTest, setDarkerSystemColorsEnabledForTest } from '../../test-support/react-native';
import { AppearanceProvider, useAppearance } from './AppearanceContext';
import { NavigationTheme } from './NavigationTheme';
import { darkPalette, lightPalette, lightHighContrastPalette } from './tokens';

let harness: MobileRenderHarness | undefined;
afterEach(async () => { await harness?.unmount(); resetNativeTestState(); });

it('propagates the resolved app appearance to native navigation and updates descendants', async () => {
  setSystemColorSchemeForTest('light');
  const themes: Theme[] = [];
  let setPreference!: ReturnType<typeof useAppearance>['setPreference'];
  function Observer() {
    setPreference = useAppearance().setPreference;
    themes.push(useTheme());
    return <View />;
  }
  harness = new MobileRenderHarness();
  await harness.render(<AppearanceProvider controller={new AppearancePreferenceController({
    async load() { return 'dark'; }, async save() {}
  })}><NavigationTheme><Observer /></NavigationTheme></AppearanceProvider>);
  await harness.settle();
  expect(themes.at(-1)).toMatchObject({ dark: true, colors: {
    primary: darkPalette.action, background: darkPalette.background,
    card: darkPalette.surface, text: darkPalette.text, border: darkPalette.border,
    notification: darkPalette.danger
  }, fonts: DarkTheme.fonts });
  await harness.run(() => setPreference('light'));
  expect(themes.at(-1)).toMatchObject({ dark: false, colors: {
    primary: lightPalette.action, background: lightPalette.background,
    card: lightPalette.surface, text: lightPalette.text, border: lightPalette.border,
    notification: lightPalette.danger
  }, fonts: DefaultTheme.fonts });
  await harness.run(() => setDarkerSystemColorsEnabledForTest(true));
  expect(themes.at(-1)?.colors.border).toBe(lightHighContrastPalette.border);
});
