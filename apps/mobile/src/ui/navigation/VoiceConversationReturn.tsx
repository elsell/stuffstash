import { router, usePathname } from 'expo-router';
import { Pressable, Text, StyleSheet } from 'react-native';
import { useVoiceInteractionState } from './VoiceInteractionStateContext';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing } from '../theme/tokens';

export function VoiceConversationReturn() {
  const { state, history } = useVoiceInteractionState();
  const pathname = usePathname();
  const colors = useAppearancePalette();
  if (!pathname.startsWith('/assets/') && !pathname.startsWith('/locations/')) return null;
  if (state.status !== 'ready' || (!state.realtime && !history.length)) return null;
  const review = state.stage === 'review';
  return <Pressable accessibilityRole="button" accessibilityLabel={review ? 'Review ready. Return to conversation' : 'Return to conversation'}
    onPress={() => router.navigate('/voice')} style={[styles.accessory, { backgroundColor: colors.surface, borderTopColor: colors.border }]}>
    <Text style={{ color: colors.action, fontWeight: '600' }}>{review ? 'Review ready · ' : ''}Return to conversation</Text>
  </Pressable>;
}
const styles = StyleSheet.create({ accessory: { minHeight: 48, paddingHorizontal: spacing.md, paddingVertical: spacing.sm, borderTopWidth: StyleSheet.hairlineWidth, justifyContent: 'center', alignItems: 'center' } });
