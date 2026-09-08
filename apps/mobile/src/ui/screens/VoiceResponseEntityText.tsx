import { voiceResponseMarkdown, type VoiceMarkdownSpan } from './VoiceResponseMarkdown';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import type { VoiceResponseArtifact } from '../../application/voice/RealtimeVoiceSession';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { spacing } from '../theme/tokens';
import { buildVoiceResponseEntityLinks, voiceResponseEntityOpenLabel } from './VoiceResponseEntityLinks';

export function VoiceResponseEntityText({
  markdown = false,
  showFallbackReferences = true,
  enabled,
  onOpen,
  references,
  text
}: {
  readonly markdown?: boolean;
  readonly showFallbackReferences?: boolean;
  readonly enabled: boolean;
  readonly onOpen: (artifact: VoiceResponseArtifact) => void;
  readonly references: readonly VoiceResponseArtifact[];
  readonly text: string;
}) {
  const styles = createStyles(useAppearancePalette());
  const blocks = markdown ? voiceResponseMarkdown(text) : [{ prefix: '', spans: [{ text }] }];
  const linkedBlocks = blocks.map(block => ({ ...block, links: buildVoiceResponseEntityLinks(block.spans.map(span => span.text).join(''), references) }));
  const placed = new Set(linkedBlocks.flatMap(block => block.links.segments.flatMap(segment => segment.reference ? [segment.reference.assetId] : [])));
  const fallbackReferences = references.filter(reference => !placed.has(reference.assetId));
  return (
    <View style={styles.responseTextGroup}>
      {linkedBlocks.map((block, blockIndex) => <Text key={blockIndex} accessibilityLiveRegion="polite" style={[styles.responseText, block.heading && styles.strong]}>
        {block.prefix}
        {block.links.segments.map((segment, index) => (
          <Text
            accessibilityHint={enabled && segment.reference ? 'Opens this asset' : undefined}
            accessibilityLabel={enabled && segment.reference ? `Open ${segment.reference.title}` : undefined}
            accessibilityRole={enabled && segment.reference ? 'link' : undefined}
            key={index}
            onPress={enabled && segment.reference ? () => onOpen(segment.reference!) : undefined}
            style={enabled && segment.reference ? styles.responseEntityLink : undefined}
          >{formattedSlice(block.spans, block.links.segments.slice(0, index).reduce((length, previous) => length + previous.text.length, 0), segment.text.length).map((span, spanIndex) => <Text key={spanIndex} style={[span.bold && styles.strong, span.italic && styles.emphasis, span.code && styles.code]}>{span.text}</Text>)}</Text>
        ))}
      </Text>)}
      {showFallbackReferences && fallbackReferences.length ? (
        <View style={styles.responseEntityActions}>
          {fallbackReferences.map((reference) => {
            const label = voiceResponseEntityOpenLabel(reference, fallbackReferences);
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

function formattedSlice(spans: readonly VoiceMarkdownSpan[], start: number, length: number): VoiceMarkdownSpan[] {
  let offset = 0;
  return spans.flatMap(span => {
    const from = Math.max(0, start - offset);
    const to = Math.min(span.text.length, start + length - offset);
    offset += span.text.length;
    return to > from ? [{ ...span, text: span.text.slice(from, to) }] : [];
  });
}

function createStyles(colors: ReturnType<typeof useAppearancePalette>) {
  return StyleSheet.create({
    strong: { fontWeight: '700' },
    emphasis: { fontStyle: 'italic' },
    code: { fontFamily: 'monospace' },
    responseText: {
      color: colors.text,
      fontSize: 17,
      fontWeight: '400',
      lineHeight: 24
    },
    responseTextGroup: {
      minWidth: 0,
      flexShrink: 1,
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
