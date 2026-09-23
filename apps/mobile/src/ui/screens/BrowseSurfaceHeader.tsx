import { useLayoutEffect, useMemo, useRef } from 'react';
import { Stack } from 'expo-router';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { BrowseSurfaceControl } from './BrowseSurfaceControl';
import type { InventoryMapSurface } from './InventoryMapPresentation';

/** One navigation-owned switcher survives changes to the Browse content. */
export function BrowseSurfaceHeader({ surface, onChange }: {
  readonly surface: InventoryMapSurface;
  readonly onChange: (surface: InventoryMapSurface) => void;
}) {
  const palette = useAppearancePalette();
  const current = useRef<typeof onChange | undefined>(onChange);
  useLayoutEffect(() => {
    current.current = onChange;
    return () => { current.current = undefined; };
  }, [onChange]);
  const options = useMemo(() => ({
    headerTitleAlign: 'center' as const,
    headerTitle: () => <BrowseSurfaceControl palette={palette} selectedSurface={surface}
      onChangeSurface={next => current.current?.(next)} />
  }), [palette, surface]);
  return <Stack.Screen options={options} />;
}
