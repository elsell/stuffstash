import { AccessibilityInfo, Platform, View } from 'react-native';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  AppearancePreferenceController,
  type AppearancePreference
} from '../../application/settings/AppearancePreference';
import { MobileRenderHarness } from '../../test-support/render';
import {
  resetNativeTestState,
  setDarkerSystemColorsEnabledForTest,
  setHighTextContrastEnabledForTest,
  setSystemColorSchemeForTest
} from '../../test-support/react-native';
import { AppearanceProvider, useAppearance } from './AppearanceContext';
import {
  darkPalette,
  lightHighContrastPalette,
  lightPalette
} from './tokens';

let harness: MobileRenderHarness | undefined;
const originalPlatform = Platform.OS;
const mutablePlatform = Platform as unknown as { OS: string };

beforeEach(() => {
  mutablePlatform.OS = 'ios';
  resetNativeTestState();
});

afterEach(async () => {
  vi.restoreAllMocks();
  mutablePlatform.OS = originalPlatform;
  await harness?.unmount();
  harness = undefined;
});

describe('AppearanceProvider', () => {
  it('waits for the initial contrast setting before publishing its first hydrated frame', async () => {
    let resolveContrast!: (enabled: boolean) => void;
    vi.spyOn(AccessibilityInfo, 'isDarkerSystemColorsEnabled').mockReturnValue(
      new Promise<boolean>((resolve) => { resolveContrast = resolve; })
    );
    const observed: Array<{ readonly isHydrated: boolean; readonly border: string; readonly scheme: string }> = [];
    harness = new MobileRenderHarness();

    await harness.render(
      <AppearanceProvider controller={controller('light')}>
        <AppearanceObserver observed={observed} />
      </AppearanceProvider>
    );
    await harness.settle();

    expect(observed.some((entry) => entry.isHydrated)).toBe(false);
    await harness.run(() => resolveContrast(true));

    expect(observed.filter((entry) => entry.isHydrated)[0]).toEqual({
      border: lightHighContrastPalette.border,
      isHydrated: true,
      scheme: 'light'
    });
  });

  it('keeps a newer contrast event when the initial snapshot resolves later', async () => {
    let resolveContrast!: (enabled: boolean) => void;
    vi.spyOn(AccessibilityInfo, 'isDarkerSystemColorsEnabled').mockReturnValue(
      new Promise<boolean>((resolve) => { resolveContrast = resolve; })
    );
    const observed: Array<{ readonly isHydrated: boolean; readonly border: string; readonly scheme: string }> = [];
    harness = new MobileRenderHarness();

    await harness.render(
      <AppearanceProvider controller={controller('light')}>
        <AppearanceObserver observed={observed} />
      </AppearanceProvider>
    );
    await harness.settle();
    await harness.run(() => setDarkerSystemColorsEnabledForTest(true));
    await harness.run(() => resolveContrast(false));

    expect(observed.filter((entry) => entry.isHydrated)[0]).toEqual({
      border: lightHighContrastPalette.border,
      isHydrated: true,
      scheme: 'light'
    });
  });

  it('publishes the persisted concrete light palette on its first hydrated frame', async () => {
    setSystemColorSchemeForTest('dark');
    const observed: Array<{ readonly isHydrated: boolean; readonly border: string; readonly scheme: string }> = [];
    harness = new MobileRenderHarness();

    await harness.render(
      <AppearanceProvider controller={controller('light')}>
        <AppearanceObserver observed={observed} />
      </AppearanceProvider>
    );
    await harness.settle();

    expect(observed.filter((entry) => entry.isHydrated)[0]).toEqual({
      border: lightPalette.border,
      isHydrated: true,
      scheme: 'light'
    });
  });

  it('uses the current device scheme for a hydrated System preference', async () => {
    setSystemColorSchemeForTest('dark');
    const observed: Array<{ readonly isHydrated: boolean; readonly border: string; readonly scheme: string }> = [];
    harness = new MobileRenderHarness();
    await harness.render(
      <AppearanceProvider controller={controller('system')}>
        <AppearanceObserver observed={observed} />
      </AppearanceProvider>
    );
    await harness.settle();

    expect(observed.filter((entry) => entry.isHydrated)[0]).toEqual({
      border: darkPalette.border,
      isHydrated: true,
      scheme: 'dark'
    });
  });

  it('tracks the iOS Increase Contrast setting with the concrete light palette', async () => {
    harness = new MobileRenderHarness();
    await harness.render(
      <AppearanceProvider controller={controller('light')}>
        <AppearanceObserver observed={[]} />
      </AppearanceProvider>
    );
    await harness.settle();

    await harness.run(() => setDarkerSystemColorsEnabledForTest(true));

    const observer = harness.byLabel('Appearance observer');
    expect(observer?.props.style).toMatchObject({ borderColor: lightHighContrastPalette.border });
  });

  it('retains Android High Contrast Text event support', async () => {
    mutablePlatform.OS = 'android';
    harness = new MobileRenderHarness();
    await harness.render(
      <AppearanceProvider controller={controller('light')}>
        <AppearanceObserver observed={[]} />
      </AppearanceProvider>
    );
    await harness.settle();

    await harness.run(() => setHighTextContrastEnabledForTest(true));

    expect(harness.byLabel('Appearance observer')?.props.style).toMatchObject({
      borderColor: lightHighContrastPalette.border
    });
  });
});

function AppearanceObserver({
  observed
}: {
  readonly observed: Array<{ readonly isHydrated: boolean; readonly border: string; readonly scheme: string }>;
}) {
  const appearance = useAppearance();
  observed.push({
    border: appearance.palette.border,
    isHydrated: appearance.isHydrated,
    scheme: appearance.resolvedColorScheme
  });
  return (
    <View
      accessibilityLabel="Appearance observer"
      style={{ borderColor: appearance.palette.border }}
    />
  );
}

function controller(preference: AppearancePreference) {
  return new AppearancePreferenceController({
    async load() { return preference; },
    async save() {}
  });
}
