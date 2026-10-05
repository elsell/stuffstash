import { useLayoutEffect, useMemo, useRef } from 'react';
import { Stack } from 'expo-router';
import { browseSurfaceHeaderOptions } from './BrowseSurfaceHeaderOptions';
import type { InventoryMapSurface } from './InventoryMapPresentation';

/** One navigation-owned switcher survives changes to the Browse content. */
export function BrowseSurfaceHeader({ surface, onChange }: {
  readonly surface: InventoryMapSurface;
  readonly onChange: (surface: InventoryMapSurface) => void;
}) {
  const current = useRef<typeof onChange | undefined>(onChange);
  useLayoutEffect(() => {
    current.current = onChange;
    return () => { current.current = undefined; };
  }, [onChange]);
  const options = useMemo(() => browseSurfaceHeaderOptions(surface,
    next => current.current?.(next)), [surface]);
  return <Stack.Screen options={options} />;
}
