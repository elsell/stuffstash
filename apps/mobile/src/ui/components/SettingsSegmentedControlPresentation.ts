export type ExpoViewConfigLookup = (moduleName: string, viewName: string) => unknown;

const requiredViews = {
  ios: ['HostView', 'PickerView', 'TextView'],
  android: ['HostView', 'SingleChoiceSegmentedButtonRowView', 'SegmentedButtonView', 'TextView']
} as const;

export function expoSegmentedControlAvailable(
  platform: string,
  lookup: ExpoViewConfigLookup = nativeExpoViewConfig
): boolean {
  if (platform !== 'ios' && platform !== 'android') return false;
  try {
    return requiredViews[platform].every((view) => Boolean(lookup('ExpoUI', view)));
  } catch {
    return false;
  }
}

function nativeExpoViewConfig(moduleName: string, viewName: string): unknown {
  const expoGlobal = (globalThis as typeof globalThis & {
    readonly expo?: { readonly getViewConfig?: ExpoViewConfigLookup };
  }).expo;
  return expoGlobal?.getViewConfig?.(moduleName, viewName);
}
