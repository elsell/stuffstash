import { t } from '../../presentation/localization';
import { NativeActionMenu } from '../components/NativeActionMenu';
import { buildBrowseSurfaceOptions, type InventoryMapSurface } from './InventoryMapPresentation';

/** Android's native menu and the non-native test renderer share these choices. */
export function BrowseSurfaceControl({ selectedSurface, onChangeSurface }: {
  readonly selectedSurface: InventoryMapSurface;
  readonly onChangeSurface: (surface: InventoryMapSurface) => void;
}) {
  const choices = buildBrowseSurfaceOptions();
  const label = choices.find(choice => choice.value === selectedSurface)!.label;
  return <NativeActionMenu
    accessibilityLabel={`${t('mobile.BrowseSurfaceControl.browseView')}: ${label}`}
    trigger={{ kind: 'label', label }}
    groups={[{ id: 'browse-view', items: choices.map(choice => ({
      id: choice.value, label: choice.label, isSelected: choice.value === selectedSurface,
      onPress: () => onChangeSurface(choice.value)
    })) }]}
  />;
}
