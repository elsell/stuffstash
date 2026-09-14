import { Stack } from 'expo-router';
import { useAppearancePalette } from '../../../ui/theme/AppearanceContext';
export default function HomeLayout() {
  const palette = useAppearancePalette();
  return <Stack screenOptions={{ title: 'Home', headerStyle: { backgroundColor: palette.background }, headerTintColor: palette.action, headerTitleStyle: { color: palette.text }, contentStyle: { backgroundColor: palette.background } }} />;
}
