import { useCallback, useState } from 'react';
import { useFocusEffect } from 'expo-router';
import { ScrollView, Text } from 'react-native';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing } from '../theme/tokens';

export function VoicePreviewRecovery({ message, identity, onRetry }: {
  readonly message: string; readonly identity: string; readonly onRetry: () => Promise<void>;
}) {
  const colors = useAppearancePalette();
  const [visit, setVisit] = useState<{ active: boolean; pending: boolean } | null>(null);
  const [retrying, setRetrying] = useState(false);
  useFocusEffect(useCallback(() => {
    const owner = { active: true, pending: false };
    setVisit(owner); setRetrying(false);
    return () => { owner.active = false; };
  }, [identity]));
  const retry = async () => {
    if (!visit?.active || visit.pending) return;
    visit.pending = true; setRetrying(true);
    try { await onRetry(); }
    catch { /* The query retains its recovery message on failure. */ }
    finally {
      visit.pending = false;
      if (visit.active) setRetrying(false);
    }
  };
  return <ScrollView style={{ flex: 1 }} contentContainerStyle={{ flexGrow: 1, justifyContent: 'center', padding: spacing.lg, gap: spacing.md }}>
    <Text accessibilityRole="header" style={{ color: colors.text, fontSize: 20, fontWeight: '700' }}>Voice unavailable</Text>
    <Text accessibilityLiveRegion="polite" style={{ color: colors.textMuted }}>{message}</Text>
    <NativeCommandButton label="Retry conversation" disabled={retrying} onPress={() => { void retry(); }} />
  </ScrollView>;
}
