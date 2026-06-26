import { useState } from 'react';
import { router } from 'expo-router';
import { ChevronDown, ChevronUp, MessageCircle, Mic, Pause, RotateCcw, X } from 'lucide-react-native';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { colors, radius, spacing } from '../theme/tokens';
import { useVoiceInteractionState, VoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { buildVoiceSessionPresentation } from '../navigation/VoiceSessionPresentation';

export function VoiceSessionSheetScreen() {
  const { reset, startRealtime, state, stopRealtime } = useVoiceInteractionState();
  const [diagnosticsExpanded, setDiagnosticsExpanded] = useState(false);

  async function handleSessionMic(): Promise<void> {
    if (state.status !== 'ready') {
      return;
    }

    if (state.stage === 'listening') {
      await stopRealtime();
      return;
    }

    if (state.stage !== 'ready') {
      reset();
    }

    await startRealtime();
  }

  return (
    <VoiceSessionSheet
      diagnosticsExpanded={diagnosticsExpanded}
      onClose={() => {
        if (router.canGoBack()) {
          router.back();
          return;
        }

        router.replace('/');
      }}
      onReset={() => {
        reset();
        setDiagnosticsExpanded(false);
      }}
      onSessionMic={() => {
        void handleSessionMic();
      }}
      onToggleDiagnostics={() => setDiagnosticsExpanded((current) => !current)}
      state={state}
    />
  );
}

function VoiceSessionSheet({
  diagnosticsExpanded,
  onClose,
  onReset,
  onSessionMic,
  onToggleDiagnostics,
  state
}: {
  readonly diagnosticsExpanded: boolean;
  readonly onClose: () => void;
  readonly onReset: () => void;
  readonly onSessionMic: () => void;
  readonly onToggleDiagnostics: () => void;
  readonly state: VoiceInteractionState;
}) {
  const readyState = state.status === 'ready' ? state : null;
  const session = buildVoiceSessionPresentation({
    diagnosticsEnabled: Boolean(readyState?.realtime?.debugEvents.length),
    diagnosticsExpanded,
    inventoryName: readyState?.preview.inventoryName ?? readyState?.realtime?.inventoryName ?? 'Inventory',
    realtime: readyState?.realtime ?? null,
    stage: state.stage,
    tenantName: readyState?.preview.tenantName ?? readyState?.realtime?.tenantName ?? 'Tenant'
  });
  const canUseMic = state.status === 'ready' && state.stage !== 'processing' && state.stage !== 'speaking';
  const micAccessibilityLabel =
    state.stage === 'listening'
      ? 'Stop listening'
      : state.stage === 'ready' && !readyState?.realtime
        ? 'Start voice interaction'
        : 'Start another voice interaction';

  return (
    <SafeAreaView style={styles.sheet} edges={['left', 'right']}>
      <View style={styles.sheetHeader}>
        <View style={styles.sheetTitleGroup}>
          <Text style={styles.sheetTitle}>{session.title}</Text>
          <Text numberOfLines={1} style={styles.sheetContext}>
            {session.contextLabel}
          </Text>
        </View>
        <Pressable
          accessibilityLabel="Close voice session"
          accessibilityRole="button"
          onPress={onClose}
          style={styles.iconButton}
        >
          <X color={colors.textMuted} size={21} strokeWidth={2.4} />
        </Pressable>
      </View>

      {state.status === 'loading' ? (
        <SessionLoadingState />
      ) : state.status === 'error' ? (
        <SessionErrorState message={state.message} />
      ) : (
        <ScrollView contentContainerStyle={styles.sessionContent}>
          <View style={styles.primaryControlRow}>
            <Pressable
              accessibilityLabel={micAccessibilityLabel}
              accessibilityRole="button"
              accessibilityState={{ disabled: !canUseMic, selected: state.stage === 'listening' }}
              disabled={!canUseMic}
              onPress={onSessionMic}
              style={[
                styles.sessionMicButton,
                state.stage === 'listening' && styles.activeSessionMicButton,
                !canUseMic && styles.disabledSessionMicButton
              ]}
            >
              {state.stage === 'listening' ? (
                <Pause color={colors.onAction} size={34} strokeWidth={2.5} />
              ) : (
                <Mic color={colors.onAction} size={36} strokeWidth={2.5} />
              )}
            </Pressable>
            <View style={styles.progressGroup}>
              <Text style={styles.progressTitle}>{session.progressLabel}</Text>
              <Text style={styles.progressHint}>{hintForStage(state.stage)}</Text>
            </View>
          </View>

          {session.transcript ? (
            <View style={styles.sessionSection}>
              <Text style={styles.sectionLabel}>Transcript</Text>
              <Text selectable style={styles.transcriptText}>
                {session.transcript}
              </Text>
            </View>
          ) : null}

          {session.response ? (
            <View style={styles.responseSection}>
              <View style={styles.responseIcon}>
                <MessageCircle color={colors.accentStrong} size={18} strokeWidth={2.4} />
              </View>
              <Text style={styles.responseText}>{session.response}</Text>
            </View>
          ) : null}

          {state.realtime?.errorMessage ? (
            <View style={styles.errorSection}>
              <Text style={styles.sectionLabel}>Voice failed</Text>
              <Text style={styles.errorText}>{state.realtime.errorMessage}</Text>
            </View>
          ) : null}

          {state.realtime?.debugEvents.length ? (
            <View style={styles.diagnosticsSection}>
              <Pressable
                accessibilityLabel={diagnosticsExpanded ? 'Hide voice diagnostics' : 'Show voice diagnostics'}
                accessibilityRole="button"
                accessibilityState={{ expanded: diagnosticsExpanded }}
                onPress={onToggleDiagnostics}
                style={styles.diagnosticsHeader}
              >
                <Text style={styles.sectionLabel}>Diagnostics</Text>
                {diagnosticsExpanded ? (
                  <ChevronUp color={colors.textMuted} size={18} strokeWidth={2.3} />
                ) : (
                  <ChevronDown color={colors.textMuted} size={18} strokeWidth={2.3} />
                )}
              </Pressable>
              {session.diagnostics?.map((event, index) => (
                <View key={`${event}-${index.toString()}`} style={styles.diagnosticRow}>
                  <Text style={styles.diagnosticIndex}>{(index + 1).toString()}</Text>
                  <Text style={styles.diagnosticText}>{event}</Text>
                </View>
              ))}
            </View>
          ) : null}

          {session.canReset ? (
            <Pressable accessibilityRole="button" onPress={onReset} style={styles.resetButton}>
              <RotateCcw color={colors.textMuted} size={17} strokeWidth={2.4} />
              <Text style={styles.resetButtonText}>Reset session</Text>
            </Pressable>
          ) : null}
        </ScrollView>
      )}
    </SafeAreaView>
  );
}

function SessionLoadingState() {
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={colors.accent} />
      <Text style={styles.centerStateText}>Loading voice</Text>
    </View>
  );
}

function SessionErrorState({ message }: { readonly message: string }) {
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Voice unavailable</Text>
      <Text style={styles.centerStateText}>{message}</Text>
    </View>
  );
}

function hintForStage(stage: VoiceInteractionState['stage']): string {
  switch (stage) {
    case 'ready':
      return 'Ask a question about this inventory.';
    case 'completed':
      return 'You can ask another question or close this.';
    case 'failed':
      return 'Reset and try again when you are ready.';
    default:
      return 'Keep this open while Stuff Stash works.';
  }
}

const styles = StyleSheet.create({
  activeSessionMicButton: {
    backgroundColor: colors.success
  },
  centerState: {
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 220,
    padding: spacing.lg
  },
  centerStateText: {
    color: colors.textMuted,
    fontSize: 15,
    lineHeight: 22,
    marginTop: spacing.sm,
    textAlign: 'center'
  },
  diagnosticIndex: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '800',
    width: 24
  },
  diagnosticRow: {
    alignItems: 'flex-start',
    borderTopColor: colors.border,
    borderTopWidth: StyleSheet.hairlineWidth,
    flexDirection: 'row',
    gap: spacing.sm,
    paddingVertical: spacing.sm
  },
  diagnosticText: {
    color: colors.textMuted,
    flex: 1,
    fontSize: 13,
    lineHeight: 18
  },
  diagnosticsHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
    minHeight: 44
  },
  diagnosticsSection: {
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    paddingHorizontal: spacing.md
  },
  disabledSessionMicButton: {
    opacity: 0.58
  },
  errorSection: {
    backgroundColor: colors.warningSurface,
    borderRadius: radius.md,
    padding: spacing.md
  },
  errorText: {
    color: colors.warning,
    fontSize: 15,
    lineHeight: 22,
    marginTop: spacing.xs
  },
  errorTitle: {
    color: colors.text,
    fontSize: 22,
    fontWeight: '800',
    letterSpacing: 0
  },
  iconButton: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: 20,
    borderWidth: 1,
    height: 40,
    justifyContent: 'center',
    width: 40
  },
  primaryControlRow: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.md
  },
  progressGroup: {
    flex: 1,
    minWidth: 0
  },
  progressHint: {
    color: colors.textMuted,
    fontSize: 14,
    lineHeight: 20,
    marginTop: spacing.xs
  },
  progressTitle: {
    color: colors.text,
    fontSize: 18,
    fontWeight: '800',
    letterSpacing: 0,
    lineHeight: 23
  },
  resetButton: {
    alignItems: 'center',
    alignSelf: 'flex-start',
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  resetButtonText: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '800'
  },
  responseIcon: {
    alignItems: 'center',
    backgroundColor: colors.brandDustyBlueSoft,
    borderRadius: 18,
    height: 36,
    justifyContent: 'center',
    width: 36
  },
  responseSection: {
    alignItems: 'flex-start',
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    flexDirection: 'row',
    gap: spacing.sm,
    padding: spacing.md
  },
  responseText: {
    color: colors.text,
    flex: 1,
    fontSize: 17,
    fontWeight: '700',
    lineHeight: 24
  },
  sectionLabel: {
    color: colors.textMuted,
    fontSize: 12,
    fontWeight: '900',
    letterSpacing: 0,
    textTransform: 'uppercase'
  },
  sessionContent: {
    gap: spacing.md,
    paddingBottom: spacing.xl
  },
  sessionMicButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: 34,
    height: 68,
    justifyContent: 'center',
    shadowColor: '#000000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.18,
    shadowRadius: 12,
    width: 68
  },
  sessionSection: {
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    padding: spacing.md
  },
  sheet: {
    backgroundColor: colors.surface,
    flex: 1,
    paddingHorizontal: spacing.lg,
    paddingTop: spacing.lg
  },
  sheetContext: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    marginTop: 2
  },
  sheetHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.md,
    justifyContent: 'space-between',
    marginBottom: spacing.md
  },
  sheetTitle: {
    color: colors.text,
    fontSize: 24,
    fontWeight: '900',
    letterSpacing: 0,
    lineHeight: 29
  },
  sheetTitleGroup: {
    flex: 1,
    minWidth: 0
  },
  transcriptText: {
    color: colors.text,
    fontSize: 16,
    lineHeight: 23,
    marginTop: spacing.xs
  }
});
