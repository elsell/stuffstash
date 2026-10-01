import { t } from '../../presentation/localization';
import { StyleSheet, Text, View } from 'react-native';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { useAppearancePalette } from '../theme/AppearanceContext';

export function VoiceResultRailNavigation({ position, count, onMove }: {
  readonly position: number; readonly count: number; readonly onMove: (position: number) => void;
}) {
  const palette = useAppearancePalette();
  return <View style={styles.row}>
    <View style={styles.command}><NativeCommandButton label={t('mobile.VoiceResultRailNavigation.previous')} disabled={position <= 0} onPress={() => onMove(position - 1)} /></View>
    <Text style={{ color: palette.textMuted }}>{t('pagination.position', { position: position + 1, total: count })}</Text>
    <View style={styles.command}><NativeCommandButton label={t('mobile.VoiceResultRailNavigation.next')} disabled={position >= count - 1} onPress={() => onMove(position + 1)} /></View>
  </View>;
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center' },
  command: { flex: 1, minWidth: 0 }
});
