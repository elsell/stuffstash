import { t } from '../../presentation/localization';
export type ExpoViewConfigLookup = (moduleName: string, viewName: string) => unknown;

export function expoUIColorPickerAvailable(platform: string, lookup: ExpoViewConfigLookup = nativeExpoViewConfig): boolean {
  if (platform !== 'ios') return false;
  try {
    return Boolean(lookup('ExpoUI', 'HostView') && lookup('ExpoUI', 'ColorPickerView'));
  } catch {
    return false;
  }
}

export function fullSpectrumPickerKind(platform: string, nativeExpoUIAvailable: boolean): 'native-ios' | 'project-spectrum' {
  return platform === 'ios' && nativeExpoUIAvailable ? 'native-ios' : 'project-spectrum';
}

function nativeExpoViewConfig(moduleName: string, viewName: string): unknown {
  const expoGlobal = (globalThis as typeof globalThis & {
    readonly expo?: { readonly getViewConfig?: ExpoViewConfigLookup };
  }).expo;
  return expoGlobal?.getViewConfig?.(moduleName, viewName);
}

export type SpectrumValue = { readonly hue: number; readonly saturation: number; readonly brightness: number };

export const spectrumGestureOwnership = {
  onPanResponderTerminationRequest: () => false,
  onShouldBlockNativeResponder: () => true
};

export function androidSpectrumAccessibility(value: SpectrumValue, disabled: boolean) {
  return {
    spectrum: {
      accessibilityActions: [{ name: 'increment', label: t('mobile.FullSpectrumTagColorPickerPresentation.increaseSaturation') }, { name: 'decrement', label: t('mobile.FullSpectrumTagColorPickerPresentation.decreaseSaturation') }, { name: 'increaseBrightness', label: t('mobile.FullSpectrumTagColorPickerPresentation.increaseBrightness') }, { name: 'decreaseBrightness', label: t('mobile.FullSpectrumTagColorPickerPresentation.decreaseBrightness') }],
      accessibilityLabel: t('mobile.FullSpectrumTagColorPickerPresentation.saturationAndBrightness'),
      accessibilityRole: 'adjustable' as const,
      accessibilityState: { disabled },
      accessibilityValue: { text: t('mobile.FullSpectrumTagColorPickerPresentation.percentSaturationPercentBrightness', { value: String(Math.round(value.saturation * 100)), value2: String(Math.round(value.brightness * 100)) }) }
    },
    hue: {
      accessibilityActions: [{ name: 'increment', label: t('mobile.FullSpectrumTagColorPickerPresentation.increaseHue') }, { name: 'decrement', label: t('mobile.FullSpectrumTagColorPickerPresentation.decreaseHue') }],
      accessibilityLabel: t('mobile.FullSpectrumTagColorPickerPresentation.hue'),
      accessibilityRole: 'adjustable' as const,
      accessibilityState: { disabled },
      accessibilityValue: { min: 0, max: 360, now: Math.round(value.hue), text: t('mobile.FullSpectrumTagColorPickerPresentation.degrees', { value: String(Math.round(value.hue)) }) }
    }
  };
}

export function adjustSpectrumValue(value: SpectrumValue, control: 'spectrum' | 'hue', action: string): SpectrumValue {
  if (control === 'spectrum' && action === 'increaseBrightness') return { ...value, brightness: Math.min(1, value.brightness + 0.05) };
  if (control === 'spectrum' && action === 'decreaseBrightness') return { ...value, brightness: Math.max(0, value.brightness - 0.05) };
  const direction = action === 'increment' ? 1 : action === 'decrement' ? -1 : 0;
  if (!direction) return value;
  if (control === 'hue') return { ...value, hue: (value.hue + direction * 5 + 360) % 360 };
  return { ...value, saturation: Math.max(0, Math.min(1, value.saturation + direction * 0.05)) };
}
