import type { VoicePlanCommandDrafts } from './VoicePlanEdits';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';
import { Check } from 'lucide-react-native';
import type { VoiceRealtimeState } from '../../application/voice/RealtimeVoiceSession';
import { useAppearanceAwarePalette } from '../theme/appearance';
import { voicePlanProgress } from './VoicePlanProgressPresentation';
export function VoicePlanProgress({ state, drafts }: { readonly state: VoiceRealtimeState | null; readonly drafts?: VoicePlanCommandDrafts }) {
  const palette = useAppearanceAwarePalette();
  const progress = voicePlanProgress(state, drafts);
  if (!progress) return null;
  return <View style={styles.group} accessible accessibilityLiveRegion="polite" accessibilityRole={progress.busy || progress.percent !== undefined ? 'progressbar' : 'text'}
    accessibilityLabel={`${progress.title}. ${progress.detail}`} accessibilityValue={progress.percent === undefined ? { text: progress.detail } : { min: 0, max: 100, now: progress.percent, text: progress.detail }}>
    <View style={styles.row}>
      {progress.busy ? <ActivityIndicator color={palette.action} /> : <Check color={palette.action} size={20} />}
      <Text style={[styles.title, { color: palette.text }]}>{progress.title}</Text>
      {progress.percent !== undefined ? <Text style={{ color: palette.textMuted }}>{`${progress.percent}%`}</Text> : null}
    </View>
    <Text style={{ color: palette.textMuted }}>{progress.detail}</Text>
    {progress.percent !== undefined ? <View style={[styles.track, { backgroundColor: palette.surface }]}><View style={[styles.fill, { width: `${`${progress.percent}%`}`, backgroundColor: palette.action }]} /></View> : null}
  </View>;
}
const styles = StyleSheet.create({ group: { gap: 8, paddingTop: 12 }, row: { flexDirection: 'row', alignItems: 'center', gap: 10 }, title: { flex: 1, fontWeight: '700', fontSize: 16 }, track: { height: 5, borderRadius: 3, overflow: 'hidden' }, fill: { height: '100%', borderRadius: 3 } });
