import { Text, View } from 'react-native';
import type { VoiceActionPlanProposal } from '../../application/voice/RealtimeVoiceSession';
import { formatExpirationChange } from '../presentation/ExpirationPresentation';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing } from '../theme/tokens';

export function VoicePlanHistorySummary({ plan }: { readonly plan: VoiceActionPlanProposal }) {
 const colors = useAppearancePalette();
 return <View style={{ gap: spacing.xs }}>
  {plan.commands.map((command, index) => { const expirationLabel = formatExpirationChange(command.expiration, command.expirationCleared); return <Text key={command.id ?? index} selectable style={{ color: colors.text }}>{`${command.title ?? command.summary}${expirationLabel ? ` · ${expirationLabel}` : ''}`}</Text>; })}
  <Text selectable style={{ color: colors.textMuted }}>{plan.status === 'executed' ? 'Saved' : plan.status}</Text>
 </View>;
}
