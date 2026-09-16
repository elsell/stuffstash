import { useTaskPresentation } from '../navigation/useTaskPresentation';
import type { PhotoSelectionProvider } from '../../application/add/PhotoSelectionQuery';
import { useVoiceReferenceNavigation } from './useVoiceReferenceNavigation';
import { NativeCommandButton } from '../components/NativeCommandButton';
import { NativeSheetActions } from '../components/NativeSheetActions';
import { VoicePlanProgress } from './VoicePlanProgress';
import { VoicePlanNameEditor } from './VoicePlanNameEditor';
import { VoiceConversationHeader } from './VoiceConversationHeader';
import { voiceConversationReferences } from './VoiceConversationReferences';
import { VoiceConversationComposer } from './VoiceConversationComposer';
import { VoiceConversationExchange, VoiceResultRail } from './VoiceConversationExchange';
import { useCallback, useEffect, useRef, useState } from 'react';
import { router, useFocusEffect } from 'expo-router';
import { Check, ChevronDown, ChevronUp, MapPin, MessageCircle, Mic, Pencil, SendHorizontal } from 'lucide-react-native';
import {
  ActivityIndicator,
  Keyboard,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View
} from 'react-native';
import { SafeAreaView, useSafeAreaInsets } from 'react-native-safe-area-context';
import { useAppearancePalette } from '../theme/AppearanceContext';
import { radius, spacing, type MobileColorPalette } from '../theme/tokens';
import { VoiceLevelMeter } from '../components/VoiceLevelMeter';
import { appKeyboardDismissMode } from '../components/AppTextInput';
import { useVoiceInteractionState, VoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { buildVoiceSessionPresentation } from '../navigation/VoiceSessionPresentation';
import { useAppServices } from '../navigation/AppServicesContext';
import { VoicePreviewRecovery } from './VoicePreviewRecovery';
import { buildVoiceSessionSheetBodyPresentation } from './VoiceSessionSheetPresentation';
import {
  showVoicePlanPhotoSourceChooser,
  VoicePlanPhotoDraftStrip
} from './VoicePlanPhotoDrafts';
import {
  appendVoicePlanPhotoDrafts,
  removeVoicePlanPhotoDraft,
  type VoicePlanPhotoDrafts
} from './VoicePlanPhotoDraftState';
import type { VoiceResponseArtifact } from '../../application/voice/RealtimeVoiceSession';
import type { VoiceSessionActionPlanCommand } from '../navigation/VoiceSessionPresentation';
import { assetDetailHref } from './AssetDetailNavigation';
import { navigateAfterTransientDismissal } from '../navigation/TransientNavigation';
import { VoiceResponseEntityText } from './VoiceResponseEntityText';
import {
  voicePlanCommandEdits,
  type VoicePlanCommandDrafts
} from './VoicePlanEdits';

export function VoiceSessionSheetScreen() {
  const { photoSelectionQuery } = useAppServices();
  return <VoiceSessionWorkspace photoSelectionQuery={photoSelectionQuery} />;
}

export function VoiceSessionWorkspace({ photoSelectionQuery }: { readonly photoSelectionQuery: PhotoSelectionProvider }) {
  const {
    photoDrafts, setPhotoDrafts, commandDraftState, setCommandDraftState, setTitleEditor, pauseMedia, scopeIdentity,
    approveRealtimeActionPlan,
    cancelRealtime,
    cancelRealtimeActionPlan,
    diagnosticsEnabled,
    reset,
    retryRealtimeActionPlanPhotos,
    startRealtime,
    state,
    stopRealtime
  } = useVoiceInteractionState();
  const openResponseReference = useVoiceReferenceNavigation({ scopeIdentity, pauseMedia,
    onOpen: artifact => navigateAfterTransientDismissal(() => router.dismiss(), () => router.push(assetDetailHref(artifact.assetId))) });
  const pauseMediaRef = useRef(pauseMedia);
  pauseMediaRef.current = pauseMedia;
  useFocusEffect(useCallback(() => () => { void pauseMediaRef.current(); }, []));
  const [diagnosticsExpanded, setDiagnosticsExpanded] = useState(false);
  const safeAreaInsets = useSafeAreaInsets();
  const activePlanId = state.status === 'ready' ? state.realtime?.actionPlan?.planId : undefined;
  const activePlanStatus = state.status === 'ready' ? state.realtime?.actionPlan?.status : undefined;
  const beginPhotoPresentation = useTaskPresentation(photoSelectionQuery, JSON.stringify([activePlanId, activePlanStatus]));
  const commandDrafts = commandDraftState.planId === activePlanId ? commandDraftState.drafts : {};

  useEffect(() => {
    if (commandDraftState.planId !== activePlanId) {
      setTitleEditor(null);
      setPhotoDrafts({});
      setCommandDraftState({ planId: activePlanId, drafts: {} });
    }
  }, [activePlanId, activePlanStatus]);

  async function handleSessionMic(): Promise<void> {
    if (state.status !== 'ready') {
      return;
    }

    if (state.stage === 'listening') {
      await stopRealtime();
      return;
    }

    await startRealtime();
  }

  return (
    <VoiceSessionSheet
      diagnosticsExpanded={diagnosticsExpanded}
      diagnosticsEnabled={diagnosticsEnabled}
      onClose={() => {
        Keyboard.dismiss();
        void pauseMedia();
        if (router.canDismiss()) {
          router.dismiss();
          return;
        }

        router.replace('/');
      }}
      onCancelSession={() => {
        void cancelRealtime();
      }}
      onApproveActionPlan={(planId) => {
        void approveRealtimeActionPlan(planId, photoDrafts, voicePlanCommandEdits(commandDrafts));
      }}
      onCancelActionPlan={(planId) => {
        void cancelRealtimeActionPlan(planId);
      }}
      onRetryPhotos={(planId) => {
        void retryRealtimeActionPlanPhotos(planId);
      }}
      onAddPhotos={(commandKey) => {
        const existingCount = photoDrafts[commandKey]?.length ?? 0;
        if (!activePlanId || activePlanStatus !== 'proposed') return;
        const isCurrent = beginPhotoPresentation();
        if (!isCurrent()) return;
        showVoicePlanPhotoSourceChooser({
          isCurrent,
          onCamera: async () => {
            const photos = await photoSelectionQuery.captureFromCamera(existingCount);
            setPhotoDrafts((current) => (
              isCurrent()
                ? appendVoicePlanPhotoDrafts(current, commandKey, photos)
                : current
            ));
          },
          onLibrary: async () => {
            const photos = await photoSelectionQuery.selectFromLibrary(existingCount);
            setPhotoDrafts((current) => (
              isCurrent()
                ? appendVoicePlanPhotoDrafts(current, commandKey, photos)
                : current
            ));
          }
        });
      }}
      onRemovePhoto={(commandKey, photoId) => {
        setPhotoDrafts((current) => removeVoicePlanPhotoDraft(current, commandKey, photoId));
      }}
      commandDrafts={commandDrafts}
      onChangeCommandTitle={(commandId, title) => {
        setCommandDraftState((current) => ({ planId: activePlanId, drafts: { ...(current.planId === activePlanId ? current.drafts : {}), [commandId]: { ...(current.planId === activePlanId ? current.drafts[commandId] : {}), title } } }));
      }}
      onOpenParentPicker={(commandId) => {
        if (activePlanId && activePlanStatus === 'proposed') router.push({ pathname: '/voice-plan-location', params: { planId: activePlanId, commandId, scope: scopeIdentity } });
      }}
      onReset={() => {
        reset();
        setDiagnosticsExpanded(false);
        setPhotoDrafts({});
        setCommandDraftState({ drafts: {} });
      }}
      onOpenProviderProfiles={() => {
        navigateAfterTransientDismissal(
          () => router.dismiss(),
          () => router.push('/settings/voice')
        );
      }}
      onOpenResponseArtifact={openResponseReference}
      onSessionMic={() => {
        void handleSessionMic();
      }}
      onToggleDiagnostics={() => setDiagnosticsExpanded((current) => !current)}
      photoDrafts={photoDrafts}
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
  onAddPhotos,
  onRemovePhoto,
  onRetryPhotos,
  onOpenProviderProfiles,
  onOpenResponseArtifact,
  onReset,
  onSessionMic,
  onToggleDiagnostics,
  photoDrafts,
  commandDrafts,
  onChangeCommandTitle,
  onOpenParentPicker,
  safeAreaBottom,
  state
}: {
  readonly diagnosticsExpanded: boolean;
  readonly diagnosticsEnabled: boolean;
  readonly onApproveActionPlan: (planId: string) => void;
  readonly onCancelActionPlan: (planId: string) => void;
  readonly onAddPhotos: (commandKey: string) => void;
  readonly onRemovePhoto: (commandKey: string, photoId: string) => void;
  readonly onRetryPhotos: (planId: string) => void;
  readonly onClose: () => void;
  readonly onCancelSession: () => void;
  readonly onOpenProviderProfiles: () => void;
  readonly onOpenResponseArtifact: (artifact: VoiceResponseArtifact) => void;
  readonly onReset: () => void;
  readonly onSessionMic: () => void;
  readonly onToggleDiagnostics: () => void;
  readonly photoDrafts: VoicePlanPhotoDrafts;
  readonly commandDrafts: VoicePlanCommandDrafts;
  readonly onChangeCommandTitle: (commandId: string, title: string) => void;
  readonly onOpenParentPicker: (commandId: string) => void;
  readonly safeAreaBottom: number;
  readonly state: VoiceInteractionState;
}) {
  const { history, scrollOffset, titleEditor, retryPreview, scopeIdentity } = useVoiceInteractionState();
  const conversationScroll = useRef<ScrollView>(null);
  const followingLatest = useRef(scrollOffset.current === 0);
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  const readyState = state.status === 'ready' ? state : null;
  const session = buildVoiceSessionPresentation({
    diagnosticsEnabled,
    diagnosticsExpanded,
    inventoryName: readyState?.realtime?.inventoryName || readyState?.preview.inventoryName || 'Inventory',
    realtime: readyState?.realtime ?? null,
    stage: state.stage,
    tenantName: readyState?.realtime?.tenantName || readyState?.preview.tenantName || 'Tenant'
  });
  const body = buildVoiceSessionSheetBodyPresentation(state, session, diagnosticsEnabled);
  const bottomAction = session.bottomAction;
  const actionPlan = session.actionPlan;
  const references = voiceConversationReferences(readyState?.realtime ?? null);

  return (
    <KeyboardAvoidingView style={{ flex: 1 }} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
    <SafeAreaView style={styles.sheet} edges={['left', 'right']}>
      <VoiceConversationHeader realtime={readyState?.realtime ?? null} photoDrafts={photoDrafts}
        commandDrafts={commandDrafts} onReset={onReset} onClose={onClose} />
      <Text style={styles.sheetContext}>{session.contextLabel}</Text>

      {state.status === 'loading' ? (
        <SessionLoadingState />
      ) : state.status === 'error' ? (
        <VoicePreviewRecovery key={scopeIdentity} message={state.message} identity={scopeIdentity} onRetry={retryPreview} />
      ) : (
        <>
          <ScrollView
            style={styles.conversationViewport}
            nestedScrollEnabled
            ref={conversationScroll}
            onContentSizeChange={() => { if (followingLatest.current) conversationScroll.current?.scrollToEnd({ animated: false }); }}
            contentOffset={{ x: 0, y: scrollOffset.current }}
            onScroll={event => {
              const { contentOffset, contentSize, layoutMeasurement } = event.nativeEvent;
              scrollOffset.current = contentOffset.y;
              followingLatest.current = contentSize.height - layoutMeasurement.height - contentOffset.y < 60;
            }}
            scrollEventThrottle={100}
            contentContainerStyle={[
              styles.sessionContent,
              !body.hasBodyContent && styles.emptySessionContent
            ]}
            keyboardDismissMode={appKeyboardDismissMode()}
            keyboardShouldPersistTaps="handled"
          >
            {history.map((exchange, index) => <VoiceConversationExchange key={index} exchange={exchange} railKey={`history-${index}`} onOpen={onOpenResponseArtifact} />)}
            {!history.length && !session.transcript && !actionPlan ? <Text style={styles.progressHint}>Find something, add an item, or organize your belongings. Speak or type below.</Text> : null}
            {state.realtime?.startsNewContext && history.length ? <Text style={styles.progressHint}>New conversation context · earlier exchanges are shown for your reference.</Text> : null}
            {session.transcript ? (
              <View style={styles.sessionSection}>
                <Text style={styles.sectionLabel}>You</Text>
                <VoiceResponseEntityText enabled onOpen={onOpenResponseArtifact} showFallbackReferences={false} references={references} text={session.transcript} />
              </View>
            ) : null}

            {actionPlan ? (
              <View style={styles.actionPlanSection}>
                <View style={styles.actionPlanHeader}>
                  <View style={styles.actionPlanHeaderText}>
                    <Text style={styles.sectionLabel}>{actionPlan.status === 'executed' ? 'Saved' : actionPlan.status === 'approved' ? 'Saving changes' : 'Review change'}</Text>
                    <VoiceResponseEntityText enabled onOpen={onOpenResponseArtifact} references={references} text={actionPlan.confirmationSummary} />
                  </View>
                  <View style={styles.actionPlanCountPill}>
                    <Text style={styles.actionPlanCountText}>{actionPlan.summary}</Text>
                  </View>
                </View>
                <View style={styles.actionPlanCommandList}>
                  {actionPlan.commands.map((command, index) => {
                    const commandKey = command.id ?? `${command.title}-${index.toString()}`;
                    return (
                      <View key={commandKey} style={styles.actionPlanCommandBlock}>
                        <View style={styles.actionPlanRow}>
                          <View style={[
                            styles.actionPlanStepMarker,
                            command.tone === 'create' && styles.actionPlanCreateMarker,
                            command.tone === 'use' && styles.actionPlanUseMarker
                          ]}>
                            {command.tone === 'use' || actionPlan.status === 'executed' ? (
                              <Check color={palette.accentStrong} size={15} strokeWidth={2.8} />
                            ) : (
                              <Text style={styles.actionPlanStepText}>{(index + 1).toString()}</Text>
                            )}
                          </View>
                          <View style={styles.actionPlanCommandTextGroup}>
                            {command.editable && command.id && actionPlan.status === 'proposed' ? (
                              <EditablePlanCommandFields
                                command={command}
                                draft={commandDrafts[command.id]}
                                onChangeTitle={(title) => onChangeCommandTitle(command.id!, title)}
                                onOpenParent={() => onOpenParentPicker(command.id!)}
                              />
                            ) : (
                              <Text style={styles.actionPlanText}>{command.id ? commandDrafts[command.id]?.title ?? command.title : command.title}</Text>
                            )}
                            <Text style={styles.actionPlanCommandMeta}>{command.subtitle}</Text>
                            {command.expirationLabel ? <Text style={styles.actionPlanPlacement}>{command.expirationLabel}</Text> : null}
                            {!command.editable && command.placement ? (
                              <Text style={styles.actionPlanPlacement}>{command.placement}</Text>
                            ) : null}
                          </View>
                        </View>
                        {(actionPlan.status === 'proposed' || actionPlan.status === 'failed') && command.photoDraftEligible ? (
                          <VoicePlanPhotoDraftStrip
                            readOnly={actionPlan.status !== 'proposed'}
                            commandKey={commandKey}
                            onAddPhotos={onAddPhotos}
                            onRemovePhoto={onRemovePhoto}
                            photos={photoDrafts[commandKey] ?? []}
                          />
                        ) : null}
                      </View>
                    );
                  })}
                </View>
                {actionPlan.risks.length ? (
                  <View style={styles.actionPlanRisks}>
                    {actionPlan.risks.map((risk, index) => (
                      <Text key={`${risk}-${index.toString()}`} style={styles.actionPlanRisk}>
                        {risk}
                      </Text>
                    ))}
                  </View>
                ) : null}
                <VoicePlanProgress state={state.realtime} drafts={commandDrafts} />
                {actionPlan.status === 'cancelled' ? (
                  <Text style={styles.actionPlanStatus}>Cancelled. No change was made.</Text>
                ) : null}
                {actionPlan.status === 'executed' ? (
                  <View style={styles.actionPlanStatusGroup}>
                    {state.realtime?.photoAttachmentStatus?.canRetry ? (
                      <NativeCommandButton label="Retry photos" onPress={() => onRetryPhotos(actionPlan.planId)} />
                    ) : null}
                  </View>
                ) : null}
                {actionPlan.status === 'failed' ? (
                  <Text style={styles.actionPlanStatus}>Could not apply this change.</Text>
                ) : null}
              </View>
            ) : null}

            {session.isBusy && !actionPlan ? <View style={styles.progressTraceRow}><ActivityIndicator color={palette.action} /><Text accessibilityLiveRegion="polite" style={styles.progressHint}>{session.progressLabel}</Text></View> : null}

            {session.response ? (
              <View style={styles.responseSection}>
                <View style={styles.responseIcon}>
                  <MessageCircle color={palette.accentStrong} size={18} strokeWidth={2.4} />
                </View>
                <View style={styles.responseBody}><VoiceResponseEntityText
                  markdown
                  enabled
                  onOpen={onOpenResponseArtifact}
                  references={references}
                  text={session.response}
                /></View>
              </View>
            ) : null}

            {session.responseArtifacts.length ? <VoiceResultRail references={references} railKey="current" onOpen={onOpenResponseArtifact} /> : null}

            {state.realtime?.errorMessage ? (
              <View accessibilityLiveRegion="assertive" style={styles.errorSection}>
                <Text style={styles.sectionLabel}>{state.realtime.progressLabel}</Text>
                <Text style={styles.errorText}>{state.realtime.errorMessage}</Text>
                {session.recoveryAction?.target === 'provider_profiles' ? (
                  <NativeCommandButton label={session.recoveryAction.label} onPress={onOpenProviderProfiles} />
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
                    <ChevronUp color={palette.textMuted} size={18} strokeWidth={2.3} />
                  ) : (
                    <ChevronDown color={palette.textMuted} size={18} strokeWidth={2.3} />
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

          </ScrollView>

          <View style={[styles.bottomActionBar, { paddingBottom: spacing.md + safeAreaBottom }]}>
            <View style={[
              styles.bottomActionContent,
              bottomAction.kind === 'review_decision' && styles.reviewBottomActionContent
            ]}>
              {bottomAction.kind === 'review_decision' ? (
                <>
                {titleEditor && !titleEditor.value.trim() ? <Text accessibilityLiveRegion="polite" style={styles.progressHint}>Enter a name before approving.</Text> : null}
                <NativeSheetActions primaryLabel="Approve" primaryAccessibilityLabel="Approve voice change"
                  secondaryLabel="Cancel" secondaryAccessibilityLabel="Cancel voice change"
                  keyboardAvoidance="container" disabled={!!titleEditor && !titleEditor.value.trim()}
                  onApply={() => onApproveActionPlan(bottomAction.planId)}
                  onBack={() => onCancelActionPlan(bottomAction.planId)} />
                </>
              ) : <VoiceConversationComposer onMic={onSessionMic} />}

            </View>
          </View>
        </>
      )}
    </SafeAreaView>
    </KeyboardAvoidingView>
  );
}

function EditablePlanCommandFields({
  command,
  draft,
  onChangeTitle,
  onOpenParent
}: {
  readonly command: VoiceSessionActionPlanCommand;
  readonly draft?: VoicePlanCommandDrafts[string];
  readonly onChangeTitle: (title: string) => void;
  readonly onOpenParent: () => void;
}) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  const { titleEditor, setTitleEditor } = useVoiceInteractionState();
  const editing = titleEditor?.commandId === command.id;
  const value = editing ? titleEditor!.value : draft?.title ?? command.title;
  const setValue = (next: string) => { if (command.id) setTitleEditor({ commandId: command.id, value: next }); };
  const setEditing = (next: boolean) => { if (!next) setTitleEditor(null); };
  const title = draft?.title ?? command.title;
  const placement = draft?.parent?.label ?? command.placement?.replace(/^Inside (?:new )?/, '') ?? 'Inventory root';

  if (editing) {
    return (
      <VoicePlanNameEditor value={value} onChange={setValue}
        onSave={name => { onChangeTitle(name); setEditing(false); }}
        onCancel={() => setEditing(false)} />
    );
  }

  return (
    <View style={styles.editablePlanFields}>
      <Pressable
        accessibilityHint="Edits the name inline"
        accessibilityLabel={`Edit proposed name ${title}`}
        accessibilityRole="button"
        onPress={() => {
          setValue(title);
          setEditing(true);
        }}
        style={styles.editableNameButton}
      >
        <Text style={styles.actionPlanText}>{title}</Text>
        <Pencil color={palette.textMuted} size={16} strokeWidth={2.3} />
      </Pressable>
      <Pressable
        accessibilityHint="Opens the containing location selector"
        accessibilityLabel={`Change containing location, currently ${placement}`}
        accessibilityRole="button"
        onPress={onOpenParent}
        style={styles.editablePlacementButton}
      >
        <MapPin color={palette.accentStrong} size={15} strokeWidth={2.4} />
        <Text numberOfLines={2} style={styles.editablePlacementText}>{placement}</Text>
        <ChevronDown color={palette.textMuted} size={16} strokeWidth={2.3} />
      </Pressable>
    </View>
  );
}

function SessionLoadingState() {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  return (
    <View style={styles.centerState}>
      <ActivityIndicator color={palette.accent} />
      <Text style={styles.centerStateText}>Loading voice</Text>
    </View>
  );
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
  conversationViewport: { flex: 1, minHeight: 0 },
  responseBody: { flex: 1, minWidth: 0 },
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
    paddingVertical: spacing.xs
  },
  actionPlanCommandList: {
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: StyleSheet.hairlineWidth,
    marginTop: spacing.sm,
    paddingHorizontal: spacing.sm
  },
  actionPlanCommandBlock: {
    borderTopColor: colors.border,
    borderTopWidth: StyleSheet.hairlineWidth,
    paddingBottom: spacing.sm,
    paddingTop: spacing.sm
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
  actionPlanStatusGroup: {
    alignItems: 'flex-start',
    gap: spacing.sm,
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
    opacity: 1
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
  progressTraceList: {
    gap: spacing.sm,
    marginTop: spacing.sm
  },
  progressTraceMarker: {
    backgroundColor: colors.accent,
    borderRadius: 4,
    height: 8,
    marginTop: 6,
    width: 8
  },
  progressTraceRow: {
    alignItems: 'flex-start',
    flexDirection: 'row',
    gap: spacing.sm
  },
  progressTraceSection: {
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    padding: spacing.md
  },
  progressTraceText: {
    color: colors.text,
    flex: 1,
    fontSize: 14,
    fontWeight: '700',
    lineHeight: 20
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
    backgroundColor: colors.surface,
    borderRadius: radius.md,
    flexDirection: 'row',
    gap: spacing.sm,
    padding: spacing.md
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
  busySessionMicButton: {
    backgroundColor: colors.warningSurface,
    borderColor: colors.warningBorder,
    borderWidth: 1,
    shadowOpacity: 0.08
  },
  sendSessionMicButton: {
    backgroundColor: colors.action
  },
  sendButtonContent: {
    alignItems: 'center',
    gap: 2,
    justifyContent: 'center'
  },
  sessionSection: {
    alignSelf: 'flex-end',
    maxWidth: '94%',
    backgroundColor: colors.surfaceMuted,
    borderRadius: radius.lg,
    gap: spacing.xs,
    padding: spacing.sm
  },
  sheet: {
    backgroundColor: colors.surface,
    flex: 1,
    paddingHorizontal: spacing.md,
    paddingTop: spacing.sm
  },
  sheetContext: {
    color: colors.textMuted,
    fontSize: 14,
    fontWeight: '700',
    marginTop: 2
  },
  editableNameButton: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 44
  },
  editablePlacementButton: {
    alignItems: 'center',
    alignSelf: 'stretch',
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: 1,
    flexDirection: 'row',
    gap: spacing.xs,
    marginTop: spacing.xs,
    minHeight: 44,
    paddingHorizontal: spacing.sm,
    paddingVertical: spacing.xs
  },
  editablePlacementText: {
    color: colors.text,
    flex: 1,
    fontSize: 13,
    fontWeight: '700',
    lineHeight: 18
  },
  editablePlanFields: {
    alignItems: 'stretch'
  },
  transcriptText: {
    color: colors.text,
    fontSize: 16,
    lineHeight: 23,
    marginTop: spacing.xs
  }
  });
}
