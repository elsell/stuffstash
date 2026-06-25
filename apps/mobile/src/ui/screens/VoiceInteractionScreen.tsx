import { MessageCircle, Mic, Pause } from 'lucide-react-native';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { VoiceInteractionPreviewViewModel } from '../../application/voice/VoiceInteractionPreviewQuery';
import { IdentityLabel } from '../components/IdentityIcon';
import { useVoiceInteractionState, VoiceInteractionStage } from '../navigation/VoiceInteractionStateContext';
import { colors, radius, spacing } from '../theme/tokens';

export function VoiceInteractionScreen() {
  const { reset, startRealtime, state, stopRealtime } = useVoiceInteractionState();

  return (
    <SafeAreaView style={styles.shell} edges={['top', 'left', 'right']}>
      {state.status === 'loading' ? <LoadingState /> : null}
      {state.status === 'error' ? <ErrorState message={state.message} /> : null}
      {state.status === 'ready' ? (
        <VoicePreview
          preview={state.preview}
          realtime={state.realtime}
          onReset={reset}
          onStart={startRealtime}
          onStop={stopRealtime}
          stage={state.stage}
        />
      ) : null}
    </SafeAreaView>
  );
}

function LoadingState() {
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.stateText}>Loading voice</Text>
    </View>
  );
}

function ErrorState({ message }: { readonly message: string }) {
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Could not load</Text>
      <Text style={styles.stateText}>{message}</Text>
    </View>
  );
}

function VoicePreview({
  onReset,
  onStart,
  onStop,
  preview,
  realtime,
  stage
}: {
  readonly onReset: () => void;
  readonly onStart: () => Promise<void>;
  readonly onStop: () => Promise<void>;
  readonly preview: VoiceInteractionPreviewViewModel;
  readonly realtime: {
    readonly status: string;
    readonly transcript?: string;
    readonly spokenResponse?: string;
    readonly progressLabel?: string;
    readonly debugEvents: readonly string[];
    readonly errorMessage?: string;
  } | null;
  readonly stage: VoiceInteractionStage;
}) {
  const isListening = stage === 'listening';
  const isReviewing = stage === 'review';
  const isProcessing = stage === 'processing' || stage === 'speaking';
  const hasRealtimeResult = stage === 'completed' || stage === 'failed' || realtime?.transcript || realtime?.spokenResponse;

  return (
    <ScrollView contentContainerStyle={styles.content}>
      <Text style={styles.title}>Voice</Text>
      <View style={styles.contextLine}>
        <IdentityLabel
          iconSize="xs"
          kind="inventory"
          label={preview.inventoryName}
          textStyle={styles.contextText}
        />
        <IdentityLabel
          iconSize="xs"
          kind="tenant"
          label={preview.tenantName}
          textStyle={styles.contextText}
        />
      </View>

      <View style={styles.voicePanel}>
        <Pressable
          accessibilityRole="button"
          accessibilityState={{ selected: isListening }}
          onPress={() => {
            if (isListening) {
              void onStop();
            } else {
              void onStart();
            }
          }}
          style={[styles.micButton, isListening ? styles.micButtonActive : null]}
        >
          {isListening ? (
            <Pause color={colors.onAction} size={36} strokeWidth={2.4} />
          ) : (
            <Mic color={colors.onAction} size={38} strokeWidth={2.4} />
          )}
        </Pressable>
        <Text style={styles.voiceStatus}>
          {isListening ? 'Listening' : isProcessing ? 'Working' : stage === 'completed' ? 'Done' : stage === 'failed' ? 'Could not finish' : isReviewing ? 'Ready to review' : 'Ready'}
        </Text>
        <Text style={styles.voiceSubstatus}>
          {realtime?.progressLabel ?? (isListening ? 'Tap again to stop.' : 'Tap to ask about this inventory.')}
        </Text>
      </View>

      {realtime?.transcript ? (
        <View style={styles.transcriptDetailCard}>
          <Text style={styles.cardLabel}>Full transcript</Text>
          <Text selectable style={styles.fullTranscriptText}>{realtime.transcript}</Text>
        </View>
      ) : null}

      {realtime?.debugEvents.length ? (
        <View style={styles.debugPanel}>
          <Text style={styles.cardLabel}>Tool progress</Text>
          {realtime.debugEvents.map((event, index) => (
            <View key={`${event}-${index.toString()}`} style={styles.debugEventRow}>
              <Text style={styles.debugEventDot}>{(index + 1).toString()}</Text>
              <Text style={styles.debugEventText}>{event}</Text>
            </View>
          ))}
        </View>
      ) : null}

      {realtime?.spokenResponse ? (
        <View style={styles.assistantCard}>
          <View style={styles.assistantIcon}>
            <MessageCircle color={colors.accentStrong} size={18} strokeWidth={2.4} />
          </View>
          <Text style={styles.assistantText}>{realtime.spokenResponse}</Text>
        </View>
      ) : null}

      {realtime?.errorMessage ? (
        <View style={styles.transcriptCard}>
          <Text style={styles.cardLabel}>Voice failed</Text>
          <Text style={styles.transcriptText}>{realtime.errorMessage}</Text>
        </View>
      ) : null}

      {isReviewing ? (
        <>
          <View style={styles.assistantCard}>
            <View style={styles.assistantIcon}>
              <MessageCircle color={colors.accentStrong} size={18} strokeWidth={2.4} />
            </View>
            <Text style={styles.assistantText}>{preview.assistantSummary}</Text>
          </View>

          <View style={styles.planCard}>
            <Text style={styles.cardLabel}>Action plan</Text>
            <Text style={styles.planTitle}>{preview.actionPreview.summary}</Text>
            {preview.actionPreview.steps.map((step, index) => (
              <View key={step} style={styles.planStep}>
                <Text style={styles.stepNumber}>{(index + 1).toString()}</Text>
                <Text style={styles.stepText}>{step}</Text>
              </View>
            ))}
            <Text style={styles.riskLabel}>{preview.actionPreview.riskLabel}</Text>
            <View style={styles.actionRow}>
              <Pressable
                accessibilityRole="button"
                disabled
                style={[styles.secondaryButton, styles.disabledAction]}
              >
                <Text style={styles.secondaryButtonText}>Cancel</Text>
              </Pressable>
              <Pressable
                accessibilityRole="button"
                disabled
                style={[styles.primaryButton, styles.disabledAction]}
              >
                <Text style={styles.primaryButtonText}>Approve</Text>
              </Pressable>
            </View>
          </View>
        </>
      ) : null}

      {stage === 'ready' ? (
        null
      ) : (
        <Pressable
          accessibilityRole="button"
          onPress={onReset}
          style={styles.resetButton}
        >
          <Text style={styles.resetButtonText}>Reset</Text>
        </Pressable>
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  shell: {
    flex: 1,
    backgroundColor: colors.background
  },
  content: {
    padding: spacing.lg,
    paddingBottom: spacing.xl
  },
  centerState: {
    alignItems: 'center',
    flex: 1,
    justifyContent: 'center',
    padding: spacing.lg
  },
  stateText: {
    color: colors.textMuted,
    fontSize: 16,
    lineHeight: 23,
    marginTop: spacing.md,
    textAlign: 'center'
  },
  errorTitle: {
    color: colors.text,
    fontSize: 24,
    fontWeight: '800',
    letterSpacing: 0
  },
  title: {
    color: colors.text,
    fontSize: 30,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 36
  },
  contextLine: {
    alignItems: 'center',
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.sm,
    marginBottom: spacing.lg,
    marginTop: spacing.xs
  },
  contextText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0
  },
  voicePanel: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    padding: spacing.lg
  },
  micButton: {
    alignItems: 'center',
    backgroundColor: colors.brandCharcoal,
    borderRadius: 44,
    height: 88,
    justifyContent: 'center',
    width: 88
  },
  micButtonActive: {
    backgroundColor: colors.action
  },
  voiceStatus: {
    color: colors.text,
    fontSize: 21,
    fontWeight: '900',
    letterSpacing: 0,
    marginTop: spacing.md
  },
  voiceSubstatus: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0,
    marginTop: spacing.xs
  },
  transcriptCard: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.md,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  cardLabel: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.xs,
    textTransform: 'uppercase'
  },
  transcriptText: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '800',
    letterSpacing: 0,
    lineHeight: 24
  },
  transcriptDetailCard: {
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: radius.sm,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  fullTranscriptText: {
    color: colors.text,
    fontSize: 16,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 24
  },
  debugPanel: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  debugEventRow: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: spacing.sm,
    marginTop: spacing.xs
  },
  debugEventDot: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 21,
    textAlign: 'center',
    width: 18
  },
  debugEventText: {
    color: colors.text,
    flex: 1,
    fontSize: 14,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 21
  },
  assistantCard: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  assistantIcon: {
    alignItems: 'center',
    backgroundColor: colors.surfaceMuted,
    borderRadius: 16,
    height: 32,
    justifyContent: 'center',
    width: 32
  },
  assistantText: {
    color: colors.text,
    flex: 1,
    fontSize: 15,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 21
  },
  planCard: {
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    marginBottom: spacing.md,
    padding: spacing.md
  },
  planTitle: {
    color: colors.text,
    fontSize: 20,
    fontWeight: '900',
    letterSpacing: 0,
    marginBottom: spacing.md
  },
  planStep: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.sm
  },
  stepNumber: {
    color: colors.accentStrong,
    fontSize: 13,
    fontWeight: '900',
    letterSpacing: 0,
    width: 18
  },
  stepText: {
    color: colors.text,
    flex: 1,
    fontSize: 15,
    fontWeight: '700',
    letterSpacing: 0,
    lineHeight: 21
  },
  riskLabel: {
    color: colors.warning,
    fontSize: 13,
    fontWeight: '800',
    letterSpacing: 0,
    marginTop: spacing.xs
  },
  actionRow: {
    flexDirection: 'row',
    gap: spacing.sm,
    marginTop: spacing.md
  },
  primaryButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    flex: 1,
    justifyContent: 'center',
    minHeight: 46
  },
  primaryButtonText: {
    color: colors.onAction,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  secondaryButton: {
    alignItems: 'center',
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    flex: 1,
    justifyContent: 'center',
    minHeight: 46
  },
  secondaryButtonText: {
    color: colors.accentStrong,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  },
  disabledAction: {
    opacity: 0.5
  },
  previewButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    minHeight: 48
  },
  previewButtonText: {
    color: colors.onAction,
    fontSize: 16,
    fontWeight: '900',
    letterSpacing: 0
  },
  resetButton: {
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 44
  },
  resetButtonText: {
    color: colors.action,
    fontSize: 15,
    fontWeight: '900',
    letterSpacing: 0
  }
});
