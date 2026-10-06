import { t } from '../../presentation/localization';
import type { HeaderOptions } from '../components/NativeHeaderActions.types';
import { buildBrowseSurfaceOptions, type InventoryMapSurface } from './InventoryMapPresentation';

/** UIKit owns the menu, selection checkmark, touch target and glass treatment. */
export function browseSurfaceHeaderOptions(surface: InventoryMapSurface,
  onChange: (surface: InventoryMapSurface) => void): HeaderOptions {
  const choices = buildBrowseSurfaceOptions();
  const label = choices.find(choice => choice.value === surface)!.label;
  return {
    headerTitle: () => null,
    unstable_headerLeftItems: () => [{
      type: 'menu', label,
      accessibilityLabel: `${t('mobile.BrowseSurfaceControl.browseView')}: ${label}`,
      menu: { items: choices.map(choice => ({
        type: 'action', label: choice.label,
        state: choice.value === surface ? 'on' : 'off',
        onPress: () => onChange(choice.value)
      })) }
    }]
  };
}
