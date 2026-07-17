import { StyleSheet, View } from 'react-native';
import { NativeSegmentedControl } from '../components/NativeSegmentedControl';
import type { MobileColorPalette } from '../theme/tokens';
import { buildBrowseSurfaceOptions, type InventoryMapSurface } from './InventoryMapPresentation';

export function BrowseSurfaceControl({
  palette,
  selectedSurface,
  onChangeSurface
}: {
  readonly palette: MobileColorPalette;
  readonly selectedSurface: InventoryMapSurface;
  readonly onChangeSurface: (surface: InventoryMapSurface) => void;
}) {
  return (
    <View accessibilityLabel="Browse view" accessibilityRole="tablist" style={styles.container}>
      <NativeSegmentedControl
        colors={palette}
        onChange={(surface) => onChangeSurface(surface as InventoryMapSurface)}
        segments={buildBrowseSurfaceOptions()}
        style={styles.control}
        value={selectedSurface}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { minWidth: 142 },
  control: { width: 142 }
});
