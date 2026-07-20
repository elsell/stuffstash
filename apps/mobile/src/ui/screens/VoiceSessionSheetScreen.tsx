import { useEffect, useRef, useState } from 'react';
import { router } from 'expo-router';
import { Check, ChevronDown, ChevronUp, MapPin, MessageCircle, Mic, Pencil, RotateCcw, SendHorizontal, X } from 'lucide-react-native';
import {
  ActivityIndicator,
  Modal,
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
import { AppTextInput, appKeyboardDismissMode } from '../components/AppTextInput';
import { useVoiceInteractionState, VoiceInteractionState } from '../navigation/VoiceInteractionStateContext';
import { buildVoiceSessionPresentation } from '../navigation/VoiceSessionPresentation';
import { useAppServices } from '../navigation/AppServicesContext';
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
import type { ParentLookupResult } from '../../application/add/ParentLookupQuery';
import type { VoiceResponseArtifact } from '../../application/voice/RealtimeVoiceSession';
import type { VoiceSessionActionPlanCommand } from '../navigation/VoiceSessionPresentation';
import { assetDetailHref } from './AssetDetailNavigation';
import { VoiceResponseEntityText } from './VoiceResponseEntityText';
import {
  voicePlanCommandEdits,
  type VoicePlanCommandDrafts,
  type VoicePlanParentDraft
} from './VoicePlanEdits';

export function VoiceSessionSheetScreen() {
  const { parentLookupQuery, photoSelectionQuery } = useAppServices();
  const {
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
  const [diagnosticsExpanded, setDiagnosticsExpanded] = useState(false);
  const [photoDrafts, setPhotoDrafts] = useState<VoicePlanPhotoDrafts>({});
  const [commandDraftState, setCommandDraftState] = useState<{ readonly planId?: string; readonly drafts: VoicePlanCommandDrafts }>({ drafts: {} });
  const [parentPickerCommandId, setParentPickerCommandId] = useState<string | null>(null);
  const [parentQuery, setParentQuery] = useState('');
  const [parentMatches, setParentMatches] = useState<readonly ParentLookupResult[]>([]);
  const safeAreaInsets = useSafeAreaInsets();
  const activePlanId = state.status === 'ready' ? state.realtime?.actionPlan?.planId : undefined;
  const activePlanStatus = state.status === 'ready' ? state.realtime?.actionPlan?.status : undefined;
  const activePlanIdRef = useRef(activePlanId);
  const activePlanStatusRef = useRef(activePlanStatus);
  const parentPickerCommandIdRef = useRef<string | null>(null);
  const parentRequestGeneration = useRef(0);
  const commandDrafts = commandDraftState.planId === activePlanId && activePlanStatus === 'proposed' ? commandDraftState.drafts : {};

  useEffect(() => {
    activePlanIdRef.current = activePlanId;
  }, [activePlanId]);

  useEffect(() => {
    activePlanStatusRef.current = activePlanStatus;
  }, [activePlanStatus]);

  useEffect(() => {
    setPhotoDrafts({});
    setCommandDraftState({ planId: activePlanId, drafts: {} });
    setParentPickerCommandId(null);
    parentPickerCommandIdRef.current = null;
    parentRequestGeneration.current += 1;
    setParentMatches([]);
  }, [activePlanId, activePlanStatus]);

  async function loadParentMatches(query: string): Promise<void> {
    const generation = parentRequestGeneration.current + 1;
    parentRequestGeneration.current = generation;
    const planId = activePlanId;
    const commandId = parentPickerCommandIdRef.current;
    const matches = await parentLookupQuery.execute(query);
    if (parentRequestGeneration.current === generation && activePlanIdRef.current === planId && parentPickerCommandIdRef.current === commandId) {
      setParentMatches(matches);
    }
  }

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
        const planIdAtOpen = activePlanId;
        const planStatusAtOpen = activePlanStatus;
        showVoicePlanPhotoSourceChooser({
          onCamera: async () => {
            const photos = await photoSelectionQuery.captureFromCamera(existingCount);
            setPhotoDrafts((current) => (
              activePlanIdRef.current === planIdAtOpen &&
              activePlanStatusRef.current === planStatusAtOpen &&
              activePlanStatusRef.current === 'proposed'
                ? appendVoicePlanPhotoDrafts(current, commandKey, photos)
                : current
            ));
          },
          onLibrary: async () => {
            const photos = await photoSelectionQuery.selectFromLibrary(existingCount);
            setPhotoDrafts((current) => (
              activePlanIdRef.current === planIdAtOpen &&
              activePlanStatusRef.current === planStatusAtOpen &&
              activePlanStatusRef.current === 'proposed'
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
        setParentPickerCommandId(commandId);
        parentPickerCommandIdRef.current = commandId;
        setParentQuery('');
        void loadParentMatches('');
      }}
      onCloseParentPicker={() => {
        parentRequestGeneration.current += 1;
        parentPickerCommandIdRef.current = null;
        setParentPickerCommandId(null);
        setParentMatches([]);
      }}
      onChangeParentQuery={(query) => {
        setParentQuery(query);
        void loadParentMatches(query);
      }}
      onSelectParent={(commandId, parent) => {
        setCommandDraftState((current) => ({ planId: activePlanId, drafts: { ...(current.planId === activePlanId ? current.drafts : {}), [commandId]: { ...(current.planId === activePlanId ? current.drafts[commandId] : {}), parent } } }));
        parentPickerCommandIdRef.current = null;
        setParentPickerCommandId(null);
      }}
      parentMatches={parentMatches}
      parentPickerCommandId={parentPickerCommandId}
      parentQuery={parentQuery}
      onReset={() => {
        reset();
        setDiagnosticsExpanded(false);
        setPhotoDrafts({});
        setCommandDraftState({ drafts: {} });
      }}
      onOpenProviderProfiles={() => router.push('/settings/voice')}
      onOpenResponseArtifact={(artifact) => router.push(assetDetailHref(artifact.assetId))}
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
  onCloseParentPicker,
  onChangeParentQuery,
  onSelectParent,
  parentMatches,
  parentPickerCommandId,
  parentQuery,
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
  readonly onCloseParentPicker: () => void;
  readonly onChangeParentQuery: (query: string) => void;
  readonly onSelectParent: (commandId: string, parent: VoicePlanParentDraft) => void;
  readonly parentMatches: readonly ParentLookupResult[];
  readonly parentPickerCommandId: string | null;
  readonly parentQuery: string;
  readonly safeAreaBottom: number;
  readonly state: VoiceInteractionState;
}) {
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
          <X color={palette.textMuted} size={21} strokeWidth={2.4} />
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
            keyboardDismissMode={appKeyboardDismissMode()}
            keyboardShouldPersistTaps="handled"
          >
            {session.transcript ? (
              <View style={styles.sessionSection}>
                <Text style={styles.sectionLabel}>Transcript</Text>
                <Text selectable style={styles.transcriptText}>
                  {session.transcript}
                </Text>
              </View>
            ) : null}

            {actionPlan ? (
              <View style={styles.actionPlanSection}>
                <View style={styles.actionPlanHeader}>
                  <View style={styles.actionPlanHeaderText}>
                    <Text style={styles.sectionLabel}>Review change</Text>
                    <Text style={styles.actionPlanTitle}>{actionPlan.confirmationSummary}</Text>
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
                            {command.tone === 'use' ? (
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
                              <Text style={styles.actionPlanText}>{command.title}</Text>
                            )}
                            <Text style={styles.actionPlanCommandMeta}>{command.subtitle}</Text>
                            {!command.editable && command.placement ? (
                              <Text style={styles.actionPlanPlacement}>{command.placement}</Text>
                            ) : null}
                          </View>
                        </View>
                        {actionPlan.status === 'proposed' && command.photoDraftEligible ? (
                          <VoicePlanPhotoDraftStrip
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
                {actionPlan.status === 'approved' ? (
                  <Text style={styles.actionPlanStatus}>Approved. Applying change.</Text>
                ) : null}
                {actionPlan.status === 'cancelled' ? (
                  <Text style={styles.actionPlanStatus}>Cancelled. No change was made.</Text>
                ) : null}
                {actionPlan.status === 'executed' ? (
                  <View style={styles.actionPlanStatusGroup}>
                    <Text style={styles.actionPlanStatus}>
                      {state.realtime?.photoAttachmentStatus?.message ?? 'Applied.'}
                    </Text>
                    {state.realtime?.photoAttachmentStatus?.canRetry ? (
                      <Pressable
                        accessibilityLabel="Retry attaching voice photos"
                        accessibilityRole="button"
                        onPress={() => onRetryPhotos(actionPlan.planId)}
                        style={styles.retryPhotosButton}
                      >
                        <Text style={styles.retryPhotosButtonText}>Retry photos</Text>
                      </Pressable>
                    ) : null}
                  </View>
                ) : null}
                {actionPlan.status === 'failed' ? (
                  <Text style={styles.actionPlanStatus}>Could not apply this change.</Text>
                ) : null}
              </View>
            ) : null}

            {!actionPlan && session.progressTrace.length ? (
              <View style={styles.progressTraceSection}>
                <Text style={styles.sectionLabel}>Progress</Text>
                <View style={styles.progressTraceList}>
                  {session.progressTrace.map((step, index) => (
                    <View key={`${step}-${index.toString()}`} style={styles.progressTraceRow}>
                      <View style={styles.progressTraceMarker} />
                      <Text style={styles.progressTraceText}>{step}</Text>
                    </View>
                  ))}
                </View>
              </View>
            ) : null}

            {session.response ? (
              <View style={styles.responseSection}>
                <View style={styles.responseIcon}>
                  <MessageCircle color={palette.accentStrong} size={18} strokeWidth={2.4} />
                </View>
                <VoiceResponseEntityText
                  enabled={state.stage === 'completed' || state.stage === 'failed'}
                  onOpen={onOpenResponseArtifact}
                  references={session.responseArtifacts}
                  text={session.response}
                />
              </View>
            ) : null}

            {state.realtime?.errorMessage ? (
              <View accessibilityLiveRegion="assertive" style={styles.errorSection}>
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

            {session.canReset ? (
              <Pressable accessibilityRole="button" onPress={onReset} style={styles.resetButton}>
                <RotateCcw color={palette.textMuted} size={17} strokeWidth={2.4} />
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
                <Text accessibilityLiveRegion="polite" style={styles.progressTitle}>{session.progressLabel}</Text>
                <Text
                  numberOfLines={bottomAction.kind === 'review_decision' ? 1 : 2}
                  style={styles.progressHint}
                >
                  {session.bottomHint}
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
                    <X color={palette.textMuted} size={17} strokeWidth={2.4} />
                    <Text style={styles.cancelPlanButtonText}>Cancel</Text>
                  </Pressable>
                  <Pressable
                    accessibilityLabel="Approve voice change"
                    accessibilityRole="button"
                    onPress={() => onApproveActionPlan(bottomAction.planId)}
                    style={styles.approvePlanButton}
                  >
                    <Check color={palette.onAction} size={18} strokeWidth={2.6} />
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
                      bottomAction.mic.icon === 'send' && styles.sendSessionMicButton,
                      bottomAction.mic.icon === 'busy' && styles.busySessionMicButton,
                      bottomAction.mic.disabled && styles.disabledSessionMicButton
                    ]}
                  >
                    {bottomAction.mic.icon === 'send' ? (
                      <View style={styles.sendButtonContent}>
                        <VoiceLevelMeter
                          level={session.activity.kind === 'listening' ? session.activity.level : 0}
                          size="regular"
                        />
                        <SendHorizontal color={palette.onAction} size={27} strokeWidth={2.6} />
                      </View>
                    ) : bottomAction.mic.icon === 'busy' ? (
                      <ActivityIndicator color={palette.warning} size="small" />
                    ) : (
                      <Mic color={palette.onAction} size={34} strokeWidth={2.5} />
                    )}
                  </Pressable>
                </>
              ) : null}
            </View>
          </View>
          <ParentPicker
            commands={actionPlan?.commands ?? []}
            commandDrafts={commandDrafts}
            commandId={parentPickerCommandId}
            matches={parentMatches}
            onChangeQuery={onChangeParentQuery}
            onClose={onCloseParentPicker}
            onSelect={onSelectParent}
            query={parentQuery}
          />
        </>
      )}
    </SafeAreaView>
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
  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState(draft?.title ?? command.title);
  const title = draft?.title ?? command.title;
  const placement = draft?.parent?.label ?? command.placement?.replace(/^Inside (?:new )?/, '') ?? 'Inventory root';

  if (editing) {
    return (
      <View style={styles.inlineNameEditor}>
        <AppTextInput
          accessibilityLabel="Proposed item name"
          autoFocus
          maxLength={200}
          onChangeText={setValue}
          onSubmitEditing={() => {
            if (value.trim()) {
              onChangeTitle(value.trim());
              setEditing(false);
            }
          }}
          returnKeyType="done"
          selectTextOnFocus
          style={styles.inlineNameInput}
          value={value}
        />
        <Pressable
          accessibilityLabel="Save proposed name"
          accessibilityRole="button"
          disabled={!value.trim()}
          onPress={() => {
            if (value.trim()) {
              onChangeTitle(value.trim());
              setEditing(false);
            }
          }}
          style={styles.inlineEditorIconButton}
        >
          <Check color={palette.accentStrong} size={18} strokeWidth={2.6} />
        </Pressable>
        <Pressable
          accessibilityLabel="Cancel editing proposed name"
          accessibilityRole="button"
          onPress={() => {
            setValue(title);
            setEditing(false);
          }}
          style={styles.inlineEditorIconButton}
        >
          <X color={palette.textMuted} size={18} strokeWidth={2.4} />
        </Pressable>
      </View>
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

function ParentPicker({
  commands,
  commandDrafts,
  commandId,
  matches,
  onChangeQuery,
  onClose,
  onSelect,
  query
}: {
  readonly commands: readonly VoiceSessionActionPlanCommand[];
  readonly commandDrafts: VoicePlanCommandDrafts;
  readonly commandId: string | null;
  readonly matches: readonly ParentLookupResult[];
  readonly onChangeQuery: (query: string) => void;
  readonly onClose: () => void;
  readonly onSelect: (commandId: string, parent: VoicePlanParentDraft) => void;
  readonly query: string;
}) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  const currentIndex = commands.findIndex((command) => command.id === commandId);
  const proposedParents = currentIndex < 0
    ? []
    : commands.slice(0, currentIndex).filter((command) => command.editable && command.id);
  return (
    <Modal animationType="slide" onRequestClose={onClose} presentationStyle="pageSheet" visible={commandId !== null}>
      <SafeAreaView style={styles.parentPickerSheet}>
        <View style={styles.parentPickerHeader}>
          <View>
            <Text style={styles.parentPickerTitle}>Containing location</Text>
            <Text style={styles.parentPickerSubtitle}>Choose where this new thing belongs</Text>
          </View>
          <Pressable accessibilityLabel="Close location selector" accessibilityRole="button" onPress={onClose} style={styles.iconButton}>
            <X color={palette.textMuted} size={21} strokeWidth={2.4} />
          </Pressable>
        </View>
        <AppTextInput
          accessibilityLabel="Search containing locations"
          autoCapitalize="none"
          onChangeText={onChangeQuery}
          placeholder="Search locations, containers, and items"
          placeholderTextColor={palette.textMuted}
          style={styles.parentSearchInput}
          value={query}
        />
        <ScrollView contentContainerStyle={styles.parentPickerList} keyboardDismissMode={appKeyboardDismissMode()} keyboardShouldPersistTaps="handled">
          <ParentOption
            label="Inventory root"
            meta="No containing location"
            onPress={() => commandId && onSelect(commandId, { kind: 'root', label: 'Inventory root' })}
          />
          {proposedParents.map((command) => (
            <ParentOption
              key={command.id}
              label={command.id ? commandDrafts[command.id]?.title ?? command.title : command.title}
              meta="Created by this plan"
              onPress={() => commandId && command.id && onSelect(commandId, { kind: 'command', id: command.id, label: commandDrafts[command.id]?.title ?? command.title })}
            />
          ))}
          {matches.map((match) => (
            <ParentOption
              disabled={match.canSelectAsParent === false}
              key={match.id}
              label={match.title}
              meta={match.disabledReason ?? (match.willPromoteToContainer ? `${match.pathLabel} · Will become a container` : match.pathLabel)}
              onPress={() => commandId && onSelect(commandId, { kind: 'asset', id: match.id, label: match.pathLabel })}
            />
          ))}
        </ScrollView>
      </SafeAreaView>
    </Modal>
  );
}

function ParentOption({ disabled = false, label, meta, onPress }: { readonly disabled?: boolean; readonly label: string; readonly meta: string; readonly onPress: () => void }) {
  const palette = useAppearancePalette();
  const styles = createStyles(palette);
  return (
    <Pressable accessibilityLabel={`Select ${label}`} accessibilityRole="button" accessibilityState={{ disabled }} disabled={disabled} onPress={onPress} style={[styles.parentOption, disabled && styles.parentOptionDisabled]}>
      <MapPin color={palette.accentStrong} size={18} strokeWidth={2.3} />
      <View style={styles.parentOptionText}>
        <Text style={styles.parentOptionTitle}>{label}</Text>
        <Text numberOfLines={2} style={styles.parentOptionMeta}>{meta}</Text>
      </View>
      <ChevronDown color={palette.textMuted} size={17} strokeWidth={2.2} style={styles.parentOptionChevron} />
    </Pressable>
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

function SessionErrorState({ message }: { readonly message: string }) {
  const styles = createStyles(useAppearancePalette());
  return (
    <View style={styles.centerState}>
      <Text style={styles.errorTitle}>Voice unavailable</Text>
      <Text style={styles.centerStateText}>{message}</Text>
    </View>
  );
}

function createStyles(colors: MobileColorPalette) {
  return StyleSheet.create({
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
  retryPhotosButton: {
    alignItems: 'center',
    borderColor: colors.border,
    borderRadius: radius.sm,
    borderWidth: StyleSheet.hairlineWidth,
    justifyContent: 'center',
    minHeight: 34,
    paddingHorizontal: spacing.md
  },
  retryPhotosButtonText: {
    color: colors.accentStrong,
    fontSize: 12,
    fontWeight: '900'
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
  inlineEditorIconButton: {
    alignItems: 'center',
    height: 44,
    justifyContent: 'center',
    width: 36
  },
  inlineNameEditor: {
    alignItems: 'center',
    flexDirection: 'row',
    gap: 2
  },
  inlineNameInput: {
    backgroundColor: colors.surface,
    borderColor: colors.accent,
    borderRadius: radius.sm,
    borderWidth: 2,
    color: colors.text,
    flex: 1,
    fontSize: 16,
    fontWeight: '700',
    minHeight: 44,
    paddingHorizontal: spacing.sm
  },
  parentOption: {
    alignItems: 'center',
    borderBottomColor: colors.border,
    borderBottomWidth: StyleSheet.hairlineWidth,
    flexDirection: 'row',
    gap: spacing.sm,
    minHeight: 64,
    paddingVertical: spacing.sm
  },
  parentOptionChevron: {
    transform: [{ rotate: '-90deg' }]
  },
  parentOptionDisabled: {
    backgroundColor: colors.surfaceMuted
  },
  parentOptionMeta: {
    color: colors.textMuted,
    fontSize: 13,
    lineHeight: 18,
    marginTop: 2
  },
  parentOptionText: {
    flex: 1,
    minWidth: 0
  },
  parentOptionTitle: {
    color: colors.text,
    fontSize: 15,
    fontWeight: '800'
  },
  parentPickerHeader: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between'
  },
  parentPickerList: {
    paddingBottom: spacing.xl
  },
  parentPickerSheet: {
    backgroundColor: colors.surface,
    flex: 1,
    padding: spacing.lg
  },
  parentPickerSubtitle: {
    color: colors.textMuted,
    fontSize: 14,
    marginTop: 2
  },
  parentPickerTitle: {
    color: colors.text,
    fontSize: 22,
    fontWeight: '900'
  },
  parentSearchInput: {
    backgroundColor: colors.surfaceMuted,
    borderColor: colors.border,
    borderRadius: radius.md,
    borderWidth: 1,
    color: colors.text,
    fontSize: 16,
    marginBottom: spacing.sm,
    marginTop: spacing.md,
    minHeight: 48,
    paddingHorizontal: spacing.md
  },
  transcriptText: {
    color: colors.text,
    fontSize: 16,
    lineHeight: 23,
    marginTop: spacing.xs
  }
  });
}
