import { Stack } from 'expo-router';
import { Platform } from 'react-native';
import { useAppearancePalette } from '../../../ui/theme/AppearanceContext';
import { nativeTabHeaderOptions } from '../../../ui/navigation/NativeTabHeader';

export default function BrowseLayout() {
  const palette = useAppearancePalette();
  return <Stack screenOptions={{ title: 'Browse', ...nativeTabHeaderOptions(palette, Platform.OS, Platform.Version) }} />;
}
