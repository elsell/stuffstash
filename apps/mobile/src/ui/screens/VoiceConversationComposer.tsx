import { NativeConversationButton } from '../components/NativeConversationButton';
import { ActivityIndicator, StyleSheet, View } from 'react-native';
import { AppTextInput } from '../components/AppTextInput';
import { VoiceLevelMeter } from '../components/VoiceLevelMeter';
import { useVoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { canCancelConversation, canSubmitConversation } from '../navigation/VoiceConversationHistory';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing } from '../theme/tokens';

export function VoiceConversationComposer({ onMic }: { readonly onMic: () => void }) {
  const { composerText, setComposerText, sendText, pauseMedia, cancelRealtime, state } = useVoiceInteractionState();
  const colors = useAppearancePalette();
  const listening = state.stage === 'listening';
  const available = canSubmitConversation(state.stage);
  const busy = !available && !listening;
  const cardOwnsProgress = state.status === 'ready' && !!state.realtime?.actionPlan;
  return <View style={[styles.row, { backgroundColor: colors.surface }]}>
    <AppTextInput accessibilityLabel="Message Stuff Stash" placeholder="Ask or add something…" placeholderTextColor={colors.textMuted}
      value={composerText} onChangeText={setComposerText} multiline maxLength={8000} editable={!busy}
      onFocus={() => { if (listening) void pauseMedia(); }}
      style={[styles.input, { color: colors.text, backgroundColor: colors.surfaceMuted, borderColor: colors.border }]} />
    {listening ? <VoiceLevelMeter level={state.status === 'ready' ? state.realtime?.recordingLevel ?? 0 : 0} size="regular" /> : null}
    {busy ? canCancelConversation(state.stage, state.status === 'ready' ? state.realtime : null)
      ? <NativeConversationButton kind="cancel" label="Cancel request" onPress={() => { void cancelRealtime(); }} />
      : !cardOwnsProgress ? <View style={styles.progress}><ActivityIndicator accessibilityLabel="Working on your request" color={colors.action} /></View> : null
      : <NativeConversationButton kind={composerText.trim() || listening ? 'send' : 'record'}
        label={composerText.trim() && !listening ? 'Send message' : listening ? 'Finish recording and send' : 'Start recording'}
        onPress={() => { if (composerText.trim() && !listening) void sendText(); else onMic(); }} />}

  </View>;
}
const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, width: '100%' },
  input: { flex: 1, minHeight: 44, maxHeight: 120, borderRadius: radius.lg, borderWidth: StyleSheet.hairlineWidth, padding: spacing.sm, fontSize: 16 },
  progress: { width: 48, height: 48, alignItems: 'center', justifyContent: 'center' }
});
