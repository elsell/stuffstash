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
      accessibilityActions: [{ name: 'increment', label: 'Increase saturation' }, { name: 'decrement', label: 'Decrease saturation' }, { name: 'increaseBrightness', label: 'Increase brightness' }, { name: 'decreaseBrightness', label: 'Decrease brightness' }],
      accessibilityLabel: 'Saturation and brightness',
      accessibilityRole: 'adjustable' as const,
      accessibilityState: { disabled },
      accessibilityValue: { text: `${Math.round(value.saturation * 100)} percent saturation, ${Math.round(value.brightness * 100)} percent brightness` }
    },
    hue: {
      accessibilityActions: [{ name: 'increment', label: 'Increase hue' }, { name: 'decrement', label: 'Decrease hue' }],
      accessibilityLabel: 'Hue',
      accessibilityRole: 'adjustable' as const,
      accessibilityState: { disabled },
      accessibilityValue: { min: 0, max: 360, now: Math.round(value.hue), text: `${Math.round(value.hue)} degrees` }
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
