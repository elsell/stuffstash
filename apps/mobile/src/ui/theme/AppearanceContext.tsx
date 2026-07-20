import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState
} from 'react';
import { AccessibilityInfo, Appearance, Platform, useColorScheme } from 'react-native';
import {
  type AppearancePreference,
  type AppearancePreferenceController
} from '../../application/settings/AppearancePreference';
import {
  appearanceAwarePalette,
  nativeAppearanceColorScheme,
  resolveAppearanceColorScheme,
  type ResolvedColorScheme
} from './appearancePalette';
import { lightPalette, type MobileColorPalette } from './tokens';

export type AppearanceContextValue = {
  readonly preference: AppearancePreference;
  readonly resolvedColorScheme: ResolvedColorScheme;
  readonly palette: MobileColorPalette;
  readonly isHydrated: boolean;
  setPreference(preference: AppearancePreference): Promise<void>;
};

const defaultAppearance: AppearanceContextValue = {
  preference: 'system',
  resolvedColorScheme: 'light',
  palette: lightPalette,
  isHydrated: false,
  async setPreference() {}
};

const AppearanceContext = createContext<AppearanceContextValue>(defaultAppearance);

export function AppearanceProvider({
  children,
  controller
}: {
  readonly children: ReactNode;
  readonly controller: AppearancePreferenceController;
}) {
  const systemColorScheme = useColorScheme();
  const [preference, setPreferenceState] = useState<AppearancePreference>('system');
  const persistedPreference = useRef<AppearancePreference>('system');
  const preferenceRequestSequence = useRef(0);
  const [isPreferenceHydrated, setIsPreferenceHydrated] = useState(false);
  const [isContrastHydrated, setIsContrastHydrated] = useState(false);
  const [increasedContrast, setIncreasedContrast] = useState(false);

  useEffect(() => {
    let isCurrent = true;
    controller.getPreference().then((savedPreference) => {
      if (isCurrent) {
        applyNativeAppearancePreference(savedPreference);
        persistedPreference.current = savedPreference;
        setPreferenceState(savedPreference);
        setIsPreferenceHydrated(true);
      }
    }).catch(() => {
      if (isCurrent) {
        applyNativeAppearancePreference('system');
        persistedPreference.current = 'system';
        setIsPreferenceHydrated(true);
      }
    });
    return () => {
      isCurrent = false;
    };
  }, [controller]);

  useEffect(() => {
    let isCurrent = true;
    let receivedContrastEvent = false;
    const contrastEvent = Platform.OS === 'ios'
      ? 'darkerSystemColorsChanged' as const
      : 'highTextContrastChanged' as const;
    const subscription = AccessibilityInfo.addEventListener(
      contrastEvent,
      (enabled) => {
        receivedContrastEvent = true;
        setIncreasedContrast(enabled);
      }
    );
    const contrastEnabled = Platform.OS === 'ios'
      ? AccessibilityInfo.isDarkerSystemColorsEnabled()
      : AccessibilityInfo.isHighTextContrastEnabled();
    contrastEnabled.then((enabled) => {
      if (isCurrent) {
        if (!receivedContrastEvent) {
          setIncreasedContrast(enabled);
        }
        setIsContrastHydrated(true);
      }
    }).catch(() => {
      if (isCurrent) {
        setIsContrastHydrated(true);
      }
    });
    return () => {
      isCurrent = false;
      subscription.remove();
    };
  }, []);

  const setPreference = useCallback(async (nextPreference: AppearancePreference) => {
    const requestId = preferenceRequestSequence.current + 1;
    preferenceRequestSequence.current = requestId;
    applyNativeAppearancePreference(nextPreference);
    setPreferenceState(nextPreference);
    try {
      await controller.setPreference(nextPreference);
      persistedPreference.current = nextPreference;
    } catch (error) {
      if (preferenceRequestSequence.current === requestId) {
        applyNativeAppearancePreference(persistedPreference.current);
        setPreferenceState(persistedPreference.current);
      }
      throw error;
    }
  }, [controller]);

  const resolvedColorScheme = resolveAppearanceColorScheme(preference, systemColorScheme);
  const isHydrated = isPreferenceHydrated && isContrastHydrated;
  const palette = useMemo(
    () => appearanceAwarePalette(resolvedColorScheme, increasedContrast),
    [increasedContrast, resolvedColorScheme]
  );
  const value = useMemo<AppearanceContextValue>(() => ({
    preference,
    resolvedColorScheme,
    palette,
    isHydrated,
    setPreference
  }), [isHydrated, palette, preference, resolvedColorScheme, setPreference]);

  return <AppearanceContext.Provider value={value}>{children}</AppearanceContext.Provider>;
}

export function useAppearance(): AppearanceContextValue {
  return useContext(AppearanceContext);
}

export function useAppearancePalette(): MobileColorPalette {
  return useAppearance().palette;
}

function applyNativeAppearancePreference(preference: AppearancePreference): void {
  // React Native 0.83 uses `unspecified` to remove an app override and resume
  // the operating system appearance.
  Appearance.setColorScheme(nativeAppearanceColorScheme(preference));
}
