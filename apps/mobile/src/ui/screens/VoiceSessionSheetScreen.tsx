import { useState } from 'react';
import { router } from 'expo-router';
import { Check, ChevronDown, ChevronUp, MessageCircle, Mic, Pause, RotateCcw, X } from 'lucide-react-native';
import {
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView, useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors, radius, spacing } from '../theme/tokens';
import { useVoiceInteractionState, VoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { buildVoiceSessionPresentation } from '../navigation/VoiceSessionPresentation';
import { buildVoiceSessionSheetBodyPresentation } from './VoiceSessionSheetPresentation';

export function VoiceSessionSheetScreen() {
  const {
    approveRealtimeActionPlan,
    cancelRealtime,
    cancelRealtimeActionPlan,
    diagnosticsEnabled,
    reset,
    startRealtime,
    state,
    stopRealtime
  } = useVoiceInteractionState();
  const [diagnosticsExpanded, setDiagnosticsExpanded] = useState(false);
  const safeAreaInsets = useSafeAreaInsets();

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
      diagnosticsEnabled={diagnosticsEnabled}
      onClose={() => {
        if (router.canGoBack()) {
          router.back();
          return;
        }

        router.replace('/');
      }}
      onCancelSession={() => {
        void cancelRealtime();
      }}
      onApproveActionPlan={(planId) => {
        void approveRealtimeActionPlan(planId);
      }}
      onCancelActionPlan={(planId) => {
        void cancelRealtimeActionPlan(planId);
      }}
      onReset={() => {
        reset();
        setDiagnosticsExpanded(false);
      }}
      onOpenProviderProfiles={() => router.push('/provider-profiles')}
      onSessionMic={() => {
        void handleSessionMic();
      }}
      onToggleDiagnostics={() => setDiagnosticsExpanded((current) => !current)}
      safeAreaBottom={safeAreaInsets.bottom}
      state={state}
    />
  );
}

function VoiceSessionSheet({
  diagnosticsExpanded,
  diagnosticsEnabled,
  onClose,
  onCancelSession,
  onApproveActionPlan,
  onCancelActionPlan,
  onOpenProviderProfiles,
  onReset,
  onSessionMic,
  onToggleDiagnostics,
  safeAreaBottom,
  state
}: {
  readonly diagnosticsExpanded: boolean;
  readonly diagnosticsEnabled: boolean;
  readonly onApproveActionPlan: (planId: string) => void;
  readonly onCancelActionPlan: (planId: string) => void;
  readonly onClose: () => void;
  readonly onCancelSession: () => void;
  readonly onOpenProviderProfiles: () => void;
  readonly onReset: () => void;
  readonly onSessionMic: () => void;
  readonly onToggleDiagnostics: () => void;
  readonly safeAreaBottom: number;
  readonly state: VoiceInteractionState;
}) {
  const readyState = state.status === 'ready' ? state : null;
  const session = buildVoiceSessionPresentation({
    diagnosticsEnabled,
    diagnosticsExpanded,
    inventoryName: readyState?.preview.inventoryName ?? readyState?.realtime?.inventoryName ?? 'Inventory',
    realtime: readyState?.realtime ?? null,
    stage: state.stage,
    tenantName: readyState?.preview.tenantName ?? readyState?.realtime?.tenantName ?? 'Tenant'
  });
  const body = buildVoiceSessionSheetBodyPresentation(state, session, diagnosticsEnabled);
  const bottomAction = session.bottomAction;

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
        <>
          <ScrollView
            contentContainerStyle={[
              styles.sessionContent,
              !body.hasBodyContent && styles.emptySessionContent
            ]}
          >
            {session.transcript ? (
              <View style={styles.sessionSection}>
                <Text style={styles.sectionLabel}>Transcript</Text>
                <Text selectable style={styles.transcriptText}>
                  {session.transcript}
                </Text>
              </View>
            ) : null}

            {session.progressSteps.length ? (
              <View style={styles.progressSection}>
                <Text style={styles.sectionLabel}>Progress</Text>
                {session.progressSteps.map((step, index) => (
                  <View key={`${step}-${index.toString()}`} style={styles.progressStepRow}>
                    <Text style={styles.progressStepIndex}>{(index + 1).toString()}</Text>
                    <Text style={styles.progressStepText}>{step}</Text>
                  </View>
                ))}
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

            {session.actionPlan ? (
              <View style={styles.actionPlanSection}>
                <View style={styles.actionPlanHeader}>
                  <View style={styles.actionPlanHeaderText}>
                    <Text style={styles.sectionLabel}>Review change</Text>
                    <Text style={styles.actionPlanTitle}>{session.actionPlan.confirmationSummary}</Text>
                  </View>
                  <View style={styles.actionPlanCountPill}>
                    <Text style={styles.actionPlanCountText}>{session.actionPlan.summary}</Text>
                  </View>
                </View>
                <View style={styles.actionPlanCommandList}>
                  {session.actionPlan.commands.map((command, index) => (
                    <View key={`${command.id ?? command.title}-${index.toString()}`} style={styles.actionPlanRow}>
                      <View style={[
                        styles.actionPlanStepMarker,
                        command.tone === 'create' && styles.actionPlanCreateMarker,
                        command.tone === 'use' && styles.actionPlanUseMarker
                      ]}>
                        {command.tone === 'use' ? (
                          <Check color={colors.accentStrong} size={15} strokeWidth={2.8} />
                        ) : (
                          <Text style={styles.actionPlanStepText}>{(index + 1).toString()}</Text>
                        )}
                      </View>
                      <View style={styles.actionPlanCommandTextGroup}>
                        <Text style={styles.actionPlanText}>{command.title}</Text>
                        <Text style={styles.actionPlanCommandMeta}>{command.subtitle}</Text>
                        {command.placement ? (
                          <Text style={styles.actionPlanPlacement}>{command.placement}</Text>
                        ) : null}
                      </View>
                    </View>
                  ))}
                </View>
                {session.actionPlan.risks.length ? (
                  <View style={styles.actionPlanRisks}>
                    {session.actionPlan.risks.map((risk, index) => (
                      <Text key={`${risk}-${index.toString()}`} style={styles.actionPlanRisk}>
                        {risk}
                      </Text>
                    ))}
                  </View>
                ) : null}
                {session.actionPlan.status === 'approved' ? (
                  <Text style={styles.actionPlanStatus}>Approved. Applying change.</Text>
                ) : null}
                {session.actionPlan.status === 'cancelled' ? (
                  <Text style={styles.actionPlanStatus}>Cancelled. No change was made.</Text>
                ) : null}
                {session.actionPlan.status === 'executed' ? (
                  <Text style={styles.actionPlanStatus}>Applied.</Text>
                ) : null}
                {session.actionPlan.status === 'failed' ? (
                  <Text style={styles.actionPlanStatus}>Could not apply this change.</Text>
                ) : null}
              </View>
            ) : null}

            {state.realtime?.errorMessage ? (
              <View style={styles.errorSection}>
                <Text style={styles.sectionLabel}>Voice failed</Text>
                <Text style={styles.errorText}>{state.realtime.errorMessage}</Text>
                {session.recoveryAction?.target === 'provider_profiles' ? (
                  <Pressable
                    accessibilityRole="button"
                    onPress={onOpenProviderProfiles}
                    style={styles.recoveryButton}
                  >
                    <Text style={styles.recoveryButtonText}>{session.recoveryAction.label}</Text>
                  </Pressable>
                ) : null}
              </View>
            ) : null}

            {diagnosticsEnabled && state.realtime?.debugEvents.length ? (
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
                    <Text selectable style={styles.diagnosticText}>{event}</Text>
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

          <View style={[styles.bottomActionBar, { paddingBottom: spacing.md + safeAreaBottom }]}>
            <View style={[
              styles.bottomActionContent,
              bottomAction.kind === 'review_decision' && styles.reviewBottomActionContent
            ]}>
              <View style={styles.progressGroup}>
                <Text style={styles.progressTitle}>{session.progressLabel}</Text>
                <Text
                  numberOfLines={bottomAction.kind === 'review_decision' ? 1 : 2}
                  style={styles.progressHint}
                >
                  {hintForStage(state.stage)}
                </Text>
              </View>
              {bottomAction.kind === 'review_decision' ? (
                <View style={styles.reviewActionGroup}>
                  <Pressable
                    accessibilityLabel="Cancel voice change"
                    accessibilityRole="button"
                    onPress={() => onCancelActionPlan(bottomAction.planId)}
                    style={styles.cancelPlanButton}
                  >
                    <X color={colors.textMuted} size={17} strokeWidth={2.4} />
                    <Text style={styles.cancelPlanButtonText}>Cancel</Text>
                  </Pressable>
                  <Pressable
                    accessibilityLabel="Approve voice change"
                    accessibilityRole="button"
                    onPress={() => onApproveActionPlan(bottomAction.planId)}
                    style={styles.approvePlanButton}
                  >
                    <Check color={colors.onAction} size={18} strokeWidth={2.6} />
                    <Text style={styles.approvePlanButtonText}>Approve</Text>
                  </Pressable>
                </View>
              ) : bottomAction.kind === 'session_controls' ? (
                <>
                  {bottomAction.canCancel ? (
                    <Pressable
                      accessibilityLabel="Cancel voice session"
                      accessibilityRole="button"
                      onPress={onCancelSession}
                      style={styles.cancelSessionButton}
                    >
                      <Text style={styles.cancelSessionButtonText}>Cancel</Text>
                    </Pressable>
                  ) : null}
                  <Pressable
                    accessibilityLabel={bottomAction.mic.accessibilityLabel}
                    accessibilityRole="button"
                    accessibilityState={{ disabled: bottomAction.mic.disabled, selected: bottomAction.mic.selected }}
                    disabled={bottomAction.mic.disabled}
                    onPress={onSessionMic}
                    style={[
                      styles.sessionMicButton,
                      bottomAction.mic.selected && styles.activeSessionMicButton,
                      bottomAction.mic.disabled && styles.disabledSessionMicButton
                    ]}
                  >
                    {bottomAction.mic.selected ? (
                      <Pause color={colors.onAction} size={32} strokeWidth={2.5} />
                    ) : (
                      <Mic color={colors.onAction} size={34} strokeWidth={2.5} />
                    )}
                  </Pressable>
                </>
              ) : null}
            </View>
          </View>
        </>
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
    case 'cancelled':
      return 'You can start again when you are ready.';
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
  bottomActionBar: {
    borderTopColor: colors.border,
    borderTopWidth: StyleSheet.hairlineWidth,
    paddingTop: spacing.md
  },
  bottomActionContent: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.md
  },
  actionPlanRisk: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '600',
    lineHeight: 18,
    marginTop: spacing.xs
  },
  actionPlanRow: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: spacing.md,
    paddingVertical: spacing.sm
  },
  actionPlanCommandList: {
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: StyleSheet.hairlineWidth,
    marginTop: spacing.sm,
    paddingHorizontal: spacing.sm
  },
  actionPlanCommandMeta: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '800',
    lineHeight: 18,
    marginTop: 2
  },
  actionPlanCommandTextGroup: {
    flex: 1
  },
  actionPlanCountPill: {
    alignSelf: 'flex-start',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: StyleSheet.hairlineWidth,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  actionPlanCountText: {
    color: colors.text,
    fontSize: 12,
    fontWeight: '900'
  },
  actionPlanCreateMarker: {
    backgroundColor: colors.accent
  },
  actionPlanHeader: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: spacing.md,
    justifyContent: 'space-between'
  },
  actionPlanHeaderText: {
    flex: 1
  },
  actionPlanPlacement: {
    color: colors.text,
    fontSize: 13,
    fontWeight: '700',
    lineHeight: 18,
    marginTop: 2
  },
  actionPlanRisks: {
    marginTop: spacing.sm
  },
  actionPlanSection: {
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.md,
    padding: spacing.md
  },
  actionPlanStatus: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: '700',
    lineHeight: 18,
    marginTop: spacing.sm
  },
  actionPlanText: {
    color: colors.text,
    flex: 1,
    fontSize: 15,
    fontWeight: '700',
    lineHeight: 20
  },
  actionPlanStepMarker: {
    alignItems: 'center',
    backgroundColor: colors.textMuted,
    borderRadius: 14,
    height: 28,
    justifyContent: 'center',
    marginTop: 1,
    width: 28
  },
  actionPlanStepText: {
    color: colors.onAction,
    fontSize: 12,
    fontWeight: '900'
  },
  actionPlanUseMarker: {
    backgroundColor: colors.surface,
    borderColor: colors.accentStrong,
    borderWidth: StyleSheet.hairlineWidth
  },
  actionPlanTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: '800',
    lineHeight: 22
  },
  approvePlanButton: {
    alignItems: 'center',
    backgroundColor: colors.accent,
    borderRadius: radius.md,
    flex: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  approvePlanButtonText: {
    color: colors.onAction,
    fontSize: 14,
    fontWeight: '900'
  },
  cancelPlanButton: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: StyleSheet.hairlineWidth,
    flex: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  cancelPlanButtonText: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '900'
  },
  cancelSessionButton: {
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: StyleSheet.hairlineWidth,
    justifyContent: 'center',
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  cancelSessionButtonText: {
    color: colors.text,
    fontSize: 14,
    fontWeight: '900'
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
  emptySessionContent: {
    flexGrow: 1
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
  progressSection: {
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: StyleSheet.hairlineWidth,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm
  },
  progressStepIndex: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '900',
    width: 24
  },
  progressStepRow: {
    alignItems: 'flex-start',
    borderTopColor: colors.border,
    borderTopWidth: StyleSheet.hairlineWidth,
    flexDirection: 'row',
    gap: spacing.sm,
    paddingVertical: spacing.sm
  },
  progressStepText: {
    color: colors.text,
    flex: 1,
    fontSize: 14,
    fontWeight: '700',
    lineHeight: 20
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
  recoveryButton: {
    alignItems: 'center',
    alignSelf: 'flex-start',
    backgroundColor: colors.action,
    borderRadius: radius.md,
    justifyContent: 'center',
    marginTop: spacing.md,
    minHeight: 44,
    paddingHorizontal: spacing.md
  },
  recoveryButtonText: {
    color: colors.onAction,
    fontSize: 14,
    fontWeight: '900'
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
  reviewActionGroup: {
    flexDirection: 'row',
    gap: spacing.sm
  },
  reviewBottomActionContent: {
    alignItems: 'stretch',
    flexDirection: 'column',
    gap: spacing.sm
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
    paddingBottom: spacing.md
  },
  sessionMicButton: {
    alignItems: 'center',
    backgroundColor: colors.action,
    borderRadius: 31,
    height: 62,
    justifyContent: 'center',
    shadowColor: '#000000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.18,
    shadowRadius: 12,
    width: 62
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
    marginBottom: spacing.sm
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
