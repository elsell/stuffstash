import { Pressable, StyleSheet, Text, View } from 'react-native';
import type { VoiceResponseArtifact } from '../../application/voice/RealtimeVoiceSession';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing } from '../theme/tokens';
import { buildVoiceResponseEntityLinks, voiceResponseEntityOpenLabel } from './VoiceResponseEntityLinks';

export function VoiceResponseEntityText({
  enabled,
  onOpen,
  references,
  text
}: {
  readonly enabled: boolean;
  readonly onOpen: (artifact: VoiceResponseArtifact) => void;
  readonly references: readonly VoiceResponseArtifact[];
  readonly text: string;
}) {
  const styles = createStyles(useAppearancePalette());
  const presentation = buildVoiceResponseEntityLinks(text, references);
  return (
    <View style={styles.responseTextGroup}>
      <Text accessibilityLiveRegion="polite" style={styles.responseText}>
        {presentation.segments.map((segment, index) => segment.reference ? (
          <Text
            accessibilityHint={enabled ? 'Opens this asset' : undefined}
            accessibilityLabel={enabled ? `Open ${segment.reference.title}` : undefined}
            accessibilityRole={enabled ? 'link' : undefined}
            key={`${segment.reference.assetId}-${index.toString()}`}
            onPress={enabled ? () => onOpen(segment.reference!) : undefined}
            style={enabled ? styles.responseEntityLink : undefined}
          >
            {segment.text}
          </Text>
        ) : segment.text)}
      </Text>
      {presentation.fallbackReferences.length ? (
        <View style={styles.responseEntityActions}>
          {presentation.fallbackReferences.map((reference) => {
            const label = voiceResponseEntityOpenLabel(reference, presentation.fallbackReferences);
            const unavailableLabel = `${reference.title}${reference.context ? ` in ${reference.context}` : ''}, available after the response finishes`;
            return (
              <Pressable
                accessibilityLabel={enabled ? label : unavailableLabel}
                accessibilityRole="button"
                accessibilityState={{ disabled: !enabled }}
                disabled={!enabled}
                key={reference.assetId}
                onPress={() => onOpen(reference)}
                style={[styles.responseEntityButton, !enabled && styles.responseEntityButtonDisabled]}
              >
                <Text style={styles.responseEntityButtonText}>{label}</Text>
              </Pressable>
            );
          })}
        </View>
      ) : null}
    </View>
  );
}

function createStyles(colors: ReturnType<typeof useAppearancePalette>) {
  return StyleSheet.create({
    responseText: {
      color: colors.text,
      fontSize: 17,
      fontWeight: '700',
      lineHeight: 24
    },
    responseTextGroup: {
      flex: 1,
      gap: spacing.sm
    },
    responseEntityLink: {
      color: colors.accentStrong,
      fontWeight: '700',
      textDecorationLine: 'underline'
    },
    responseEntityActions: {
      alignItems: 'flex-start',
      flexDirection: 'row',
      flexWrap: 'wrap',
      gap: spacing.xs
    },
    responseEntityButton: {
      backgroundColor: colors.surface,
      borderColor: colors.border,
      borderRadius: 999,
      borderWidth: StyleSheet.hairlineWidth,
      paddingHorizontal: spacing.sm,
      paddingVertical: spacing.xs
    },
    responseEntityButtonDisabled: {
      opacity: 0.5
    },
    responseEntityButtonText: {
      color: colors.accentStrong,
      fontSize: 13,
      fontWeight: '700'
    }
  });
}
