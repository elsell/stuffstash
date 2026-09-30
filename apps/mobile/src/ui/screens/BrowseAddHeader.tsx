import { useAppearancePalette } from '../theme/AppearanceContext';
import { useMemo } from 'react';
import type { NativeHeaderAction } from '../components/NativeHeaderActions.types';
import { Stack } from 'expo-router';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';

export function BrowseAddHeader({ canAdd, onAdd, onFilters, filterCount = 0 }: {
  readonly canAdd: boolean;
  readonly onAdd: () => void;
  readonly onFilters?: () => void;
  readonly filterCount?: number;
}) {
  const palette = useAppearancePalette();
  const commands: NativeHeaderAction[] = canAdd ? [{ kind: 'add', label: 'Add an asset', onPress: onAdd }] : [];
  if (onFilters) commands.push({ kind: 'filter',
    label: filterCount > 0 ? `Filters, ${filterCount} applied` : 'Filters',
    tintColor: filterCount > 0 ? palette.action : palette.text,
    badgeCount: filterCount, onPress: onFilters });
  const actions = useNativeHeaderActionOptions(commands);
  const options = useMemo(() => ({ title: 'Browse', headerLeft: undefined, ...actions }), [actions]);
  return <Stack.Screen options={options} />;
}
