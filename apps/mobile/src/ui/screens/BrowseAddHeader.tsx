import { t } from '../../presentation/localization';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { useMemo } from 'react';
import type { NativeHeaderAction } from '../components/NativeHeaderActions.types';
import { Stack } from 'expo-router';
import { useNativeHeaderActionOptions } from '../components/useNativeHeaderActionOptions';

export function BrowseAddHeader({ canAdd, onAdd, onFilters, filterCount = 0, onScan }: {
  readonly onScan?: () => void;
  readonly canAdd: boolean;
  readonly onAdd: () => void;
  readonly onFilters?: () => void;
  readonly filterCount?: number;
}) {
  const palette = useAppearancePalette();
  const commands: NativeHeaderAction[] = canAdd ? [{ kind: 'add', label: t('browse.addAsset'), onPress: onAdd }] : [];
  if (onScan) commands.push({ kind: 'scan', label: t('labels.mobile.scan'), onPress: onScan });
  if (onFilters) commands.push({ kind: 'filter',
    label: filterCount > 0 ? t('browse.filtersApplied', { count: filterCount }) : t('browse.filters'),
    tintColor: filterCount > 0 ? palette.action : palette.text,
    badgeCount: filterCount, onPress: onFilters });
  const actions = useNativeHeaderActionOptions(commands);
  const options = useMemo(() => ({ title: t('browse.title'), ...actions }), [actions]);
  return <Stack.Screen options={options} />;
}
