export {
  appearanceAwarePalette,
  nativeAppearanceColorScheme,
  resolveAppearanceColorScheme,
  type NativeAppearanceColorScheme,
  type ResolvedColorScheme
} from './appearancePalette';

// Kept as the shared component-facing hook name while appearance ownership
// moves into the app-level provider.
export { useAppearancePalette as useAppearanceAwarePalette } from './AppearanceContext';
