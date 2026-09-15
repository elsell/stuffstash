import type { ReactNode } from 'react';
import { StyleSheet, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAppearancePalette } from '../theme/AppearanceContext';

/** The tab host owns iOS26's accessory; other platforms reserve a sibling area. */
export function VoiceTabContent({ children, accessory, platform, version }: {
  readonly children: ReactNode;
  readonly accessory: ReactNode;
  readonly platform: string;
  readonly version: string | number;
}) {
  const palette = useAppearancePalette();
  if (platform === 'ios' && Number.parseInt(String(version), 10) >= 26) return <>{children}</>;
  return <SafeAreaView edges={['bottom']} style={[styles.root, { backgroundColor: palette.background }]}>
    <View style={styles.body}>{children}</View>
    <View style={styles.accessory}>{accessory}</View>
  </SafeAreaView>;
}
const styles = StyleSheet.create({
  root: { flex: 1 },
  body: { flex: 1 },
  accessory: { minHeight: 72, flexShrink: 0 }
});
