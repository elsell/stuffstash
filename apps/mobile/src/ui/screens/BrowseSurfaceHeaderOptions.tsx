import type { HeaderOptions } from '../components/NativeHeaderActions.types';
import { BrowseSurfaceControl } from './BrowseSurfaceControl';
import type { InventoryMapSurface } from './InventoryMapPresentation';

export function browseSurfaceHeaderOptions(surface: InventoryMapSurface,
  onChange: (surface: InventoryMapSurface) => void): HeaderOptions {
  return {
    headerTitle: '',
    headerLeft: () => <BrowseSurfaceControl selectedSurface={surface} onChangeSurface={onChange} />
  };
}
