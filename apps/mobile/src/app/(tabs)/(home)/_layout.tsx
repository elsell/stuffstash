import { AppNoticeScreenLayout } from '../../../ui/feedback/AppNoticeScreenLayout';
import { Stack } from 'expo-router';
import { Platform } from 'react-native';
import { useAppearancePalette } from '../../../ui/theme/AppearanceContext';
import { nativeTabHeaderOptions } from '../../../ui/navigation/NativeTabHeader';

export default function HomeLayout() {
  const palette = useAppearancePalette();
  return <Stack screenLayout={AppNoticeScreenLayout} screenOptions={{ title: 'Home', ...nativeTabHeaderOptions(palette, Platform.OS, Platform.Version, palette.background) }} />;
}
