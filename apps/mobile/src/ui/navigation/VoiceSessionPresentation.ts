import { t } from '../../presentation/localization';
import { formatExpirationChange } from '../presentation/ExpirationPresentation';
import { canCancelConversation } from './VoiceConversationHistory';
import type {
  VoiceActionPlanCommand,
  VoiceRealtimeState,
  VoiceResponseArtifact,
  VoiceSafeDiagnosticEvent
} from '../../application/voice/RealtimeVoiceSession';
import { redactUnsafeVoiceText } from '../../application/voice/VoiceTextSafety';
import type { VoiceInteractionStage } from './VoiceInteractionStateContext';

export type VoiceAccessoryPrimaryAction = 'expand' | 'start' | 'stop';

export type VoiceAccessoryPresentation = {
  readonly accessibilityLabel: string;
  readonly primaryAction: VoiceAccessoryPrimaryAction;
  readonly subtitle: string;
  readonly title: string;
  readonly tone: 'ready' | 'active' | 'attention' | 'failed';
};

export function buildVoiceAccessoryPresentation({
  diagnosticsEnabled = false,
  pathname,
  realtime,
  status,
  stage
}: {
  readonly diagnosticsEnabled?: boolean;
  readonly pathname: string;
  readonly realtime?: VoiceRealtimeState | null;
  readonly status?: 'error' | 'loading' | 'ready';
  readonly stage: VoiceInteractionStage;
}): VoiceAccessoryPresentation {
  const context = describeVoiceContext(pathname);

  if (status === 'loading') {
    return {
      accessibilityLabel: t('mobile.VoiceSessionPresentation.openVoiceStatus'),
      primaryAction: 'expand',
      subtitle: context,
      title: t('mobile.VoiceSessionPresentation.voiceLoading'),
      tone: 'attention'
    };
  }

  if (status === 'error') {
    return {
      accessibilityLabel: t('mobile.VoiceSessionPresentation.openVoiceError'),
      primaryAction: 'expand',
      subtitle: context,
      title: t('mobile.VoiceSessionPresentation.voiceUnavailable'),
      tone: 'failed'
    };
  }

  if (stage === 'listening') {
    return {
      accessibilityLabel: t('mobile.VoiceSessionPresentation.sendVoiceRequest'),
      primaryAction: 'stop',
      subtitle: context,
      title: t('mobile.VoiceSessionPresentation.listening'),
      tone: 'active'
    };
  }

  if (stage === 'processing') {
    return {
      accessibilityLabel: t('mobile.VoiceSessionPresentation.openVoiceSession'),
      primaryAction: 'expand',
      subtitle: context,
      title: accessoryProgressTitle(realtime),
      tone: 'attention'
    };
  }

  if (stage === 'speaking') {
    return {
      accessibilityLabel: t('mobile.VoiceSessionPresentation.openVoiceResponse'),
      primaryAction: 'expand',
      subtitle: context,
      title: t('mobile.VoiceSessionPresentation.speaking'),
      tone: 'attention'
    };
  }

  if (stage === 'completed') {
    if (hasAvailableVoiceFollowUp(realtime)) {
      return {
        accessibilityLabel: t('mobile.VoiceSessionPresentation.openVoiceFollowUp'),
        primaryAction: 'expand',
        subtitle: safeAccessorySubtitle(realtime?.spokenResponse) ?? context,
        title: realtime?.responseKind === 'clarification' ? t('mobile.VoiceSessionPresentation.needsDetail') : t('mobile.VoiceSessionPresentation.answerReady'),
        tone: realtime?.responseKind === 'clarification' ? 'attention' : 'ready'
      };
    }
    const terminal = completedVoicePresentation(realtime);
    return {
      accessibilityLabel: terminal.accessibilityLabel,
      primaryAction: 'expand',
      subtitle: safeAccessorySubtitle(realtime?.spokenResponse) ?? context,
      title: terminal.title,
      tone: terminal.tone
    };
  }

  if (stage === 'cancelled') {
    return {
      accessibilityLabel: t('mobile.VoiceSessionPresentation.openCancelledVoiceSession'),
      primaryAction: 'expand',
      subtitle: context,
      title: t('mobile.VoiceSessionPresentation.voiceCancelled'),
      tone: 'attention'
    };
  }

  if (stage === 'failed') {
    return {
      accessibilityLabel: t('mobile.VoiceSessionPresentation.openVoiceError'),
      primaryAction: 'expand',
      subtitle: safeFailureAccessorySubtitle(realtime, diagnosticsEnabled) ?? context,
      title: safeFailureAccessoryTitle(realtime),
      tone: 'failed'
    };
  }

  if (stage === 'review') {
    return {
      accessibilityLabel: t('mobile.VoiceSessionPresentation.reviewVoicePlan'),
      primaryAction: 'expand',
      subtitle: context,
      title: accessoryProgressTitle(realtime),
      tone: 'attention'
    };
  }

  return {
    accessibilityLabel: t('mobile.VoiceSessionPresentation.startVoiceInteraction'),
    primaryAction: 'start',
    subtitle: context,
    title: t('mobile.VoiceSessionPresentation.askStuffStash'),
    tone: 'ready'
  };
}

function safeAccessorySubtitle(value: string | undefined): string | undefined {
  const normalized = redactUnsafeVoiceText(value ?? '').replace(/\s+/g, ' ').trim();
  if (!normalized) {
    return undefined;
  }
  return normalized.length <= 96 ? normalized : `${normalized.slice(0, 95).trim()}...`;
}

function accessoryProgressTitle(realtime: VoiceRealtimeState | null | undefined): string {
  switch (realtime?.status) {
    case 'review':
      return t('mobile.VoiceSessionPresentation.reviewNeeded');
    case 'speaking':
      return t('mobile.VoiceSessionPresentation.speaking');
    case 'processing':
      return accessoryPhaseTitle(realtime.conversationPhase);
    case 'listening':
      return t('mobile.VoiceSessionPresentation.listening');
    default:
      return t('mobile.VoiceSessionPresentation.checkingInventory');
  }
}

const voiceAccessoryPhaseTitles = {
  understanding: t('mobile.voice.phase.understanding'),
  exploring: t('mobile.voice.phase.exploring'),
  planning: t('mobile.voice.phase.planning'),
  reviewing: t('mobile.voice.phase.reviewing'),
  answering: t('mobile.voice.phase.answering'),
  recovering: t('mobile.voice.phase.recovering')
} satisfies Record<NonNullable<VoiceRealtimeState['conversationPhase']>, string>;

function accessoryPhaseTitle(phase: VoiceRealtimeState['conversationPhase']): string {
  return phase ? voiceAccessoryPhaseTitles[phase] : t('mobile.VoiceSessionPresentation.checkingInventory');
}

function safeFailureAccessorySubtitle(realtime: VoiceRealtimeState | null | undefined, diagnosticsEnabled: boolean): string | undefined {
  const code = realtime?.failureCode;
  if (code === 'provider_billing_disabled') {
    return t('mobile.VoiceSessionPresentation.askYourProviderAdministratorToRestoreGoogleCloudBilling');
  }
  if (
    code === 'provider_readiness' ||
    code === 'speech_to_text_failed' ||
    code === 'text_to_speech_failed'
  ) {
    return t('mobile.VoiceSessionPresentation.checkVoiceProvidersAndTryAgain');
  }
  if (code === 'language_inference_failed') {
    return diagnosticsEnabled ? t('mobile.VoiceSessionPresentation.openDiagnosticsOrCheckVoiceProviders') : t('mobile.VoiceSessionPresentation.checkVoiceProvidersAndTryAgain');
  }
  if (code === 'clarification_turn_limit') {
    return t('mobile.VoiceSessionPresentation.startAFreshVoiceRequest');
  }
  return realtime?.status === 'failed' ? t('mobile.VoiceSessionPresentation.openForDetails') : undefined;
}

function safeFailureAccessoryTitle(realtime: VoiceRealtimeState | null | undefined): string {
  switch (realtime?.failureCode) {
    case 'provider_billing_disabled':
      return t('mobile.VoiceSessionPresentation.providerBillingIsDisabled');
    case 'speech_to_text_failed':
      return t('mobile.VoiceSessionPresentation.speechInputFailed');
    case 'language_inference_failed':
      return t('mobile.VoiceSessionPresentation.agentBrainFailed');
    case 'text_to_speech_failed':
      return t('mobile.VoiceSessionPresentation.speechOutputFailed');
    case 'clarification_turn_limit':
      return t('mobile.VoiceSessionPresentation.voiceNeedsAFreshStart');
    case 'provider_readiness':
      return t('mobile.VoiceSessionPresentation.voiceProvidersNeeded');
    default:
      return t('mobile.VoiceSessionPresentation.voiceFailed');
  }
}

export type VoiceSessionPresentation = {
  readonly activity: VoiceSessionActivityPresentation;
  readonly actionPlan?: {
    readonly planId: string;
    readonly status: 'proposed' | 'approved' | 'cancelled' | 'executed' | 'failed';
    readonly confirmationSummary: string;
    readonly summary: string;
    readonly commands: readonly VoiceSessionActionPlanCommand[];
    readonly risks: readonly string[];
  };
  readonly bottomAction: VoiceSessionBottomAction;
  readonly bottomHint: string;
  readonly canReset: boolean;
  readonly contextLabel: string;
  readonly diagnostics: readonly string[] | null;
  readonly isBusy: boolean;
  readonly progressLabel: string;
  readonly progressSteps: readonly string[];
  readonly progressTrace: readonly string[];
  readonly recoveryAction?: VoiceSessionRecoveryAction;
  readonly response?: string;
  readonly responseArtifacts: readonly VoiceResponseArtifact[];
  readonly title: string;
  readonly transcript?: string;
};

export type VoiceSessionActivityPresentation =
  | { readonly kind: 'idle' }
  | { readonly kind: 'listening'; readonly label: string; readonly level: number }
  | { readonly kind: 'busy'; readonly label: string };

export type VoiceSessionActionPlanCommand = {
  readonly expirationLabel?: string;
  readonly changes?: readonly string[];
  readonly id?: string;
  readonly editable: boolean;
  readonly title: string;
  readonly subtitle: string;
  readonly placement?: string;
  readonly photoDraftEligible: boolean;
  readonly tone: 'create' | 'use' | 'update';
};

export type VoiceSessionBottomAction =
  | { readonly kind: 'review_decision'; readonly planId: string }
  | {
      readonly kind: 'session_controls';
      readonly canCancel: boolean;
      readonly mic: {
        readonly accessibilityLabel: string;
        readonly disabled: boolean;
        readonly icon: 'mic' | 'send' | 'busy';
        readonly selected: boolean;
      };
    }
  | { readonly kind: 'none' };

export type VoiceSessionRecoveryAction = {
  readonly label: string;
  readonly target: 'provider_profiles';
};

export function buildVoiceSessionPresentation({
  diagnosticsEnabled,
  diagnosticsExpanded,
  inventoryName,
  realtime,
  stage,
  tenantName
}: {
  readonly diagnosticsEnabled: boolean;
  readonly diagnosticsExpanded: boolean;
  readonly inventoryName: string;
  readonly realtime: VoiceRealtimeState | null;
  readonly stage: VoiceInteractionStage;
  readonly tenantName: string;
}): VoiceSessionPresentation {
  const title = titleForState(stage, realtime);
  const progressLabel = safeProgressPresentationText(realtime?.progressLabel ?? progressForStage(stage), 100) || progressForStage(stage);
  const progressSteps = safeProgressPresentationSteps(realtime?.progressSteps ?? []);
  const diagnostics =
    diagnosticsEnabled && diagnosticsExpanded
      ? (realtime?.debugEvents ?? []).map(formatSafeDiagnosticEvent)
      : null;
  const bottomAction = bottomActionForState(stage, realtime);
  const activePartialTranscript = stage === 'listening' || stage === 'processing' || stage === 'speaking' || stage === 'review'
    ? realtime?.partialTranscript
    : undefined;

  return {
    actionPlan: realtime?.actionPlan
      ? {
          planId: realtime.actionPlan.planId,
          status: realtime.actionPlan.status,
          confirmationSummary: realtime.actionPlan.confirmationSummary,
          summary: summarizeActionPlanCommands(realtime.actionPlan.commands),
          commands: formatActionPlanCommands(realtime.actionPlan.commands),
          risks: realtime.actionPlan.risks
        }
      : undefined,
    activity: activityForState(stage, progressLabel, realtime?.recordingLevel),
    bottomHint: bottomHintForState(stage, realtime),
    bottomAction,
    canReset: stage === 'completed' || stage === 'cancelled' || stage === 'failed' || (stage === 'review' && realtime?.actionPlan?.status !== 'proposed'),
    contextLabel: `${inventoryName} · ${tenantName}`,
    diagnostics,
    isBusy: stage === 'listening' || stage === 'processing' || stage === 'speaking',
    progressLabel,
    progressSteps,
    progressTrace: progressTraceForState(stage, progressSteps, realtime),
    recoveryAction: isProviderRecoveryFailure(realtime?.failureCode)
      ? { label: t('mobile.VoiceSessionPresentation.voiceProviders'), target: 'provider_profiles' }
      : undefined,
    response: realtime?.spokenResponse,
    responseArtifacts: realtime?.responseArtifacts ?? [],
    title,
    transcript: realtime?.transcript ?? activePartialTranscript
  };
}

function activityForState(
  stage: VoiceInteractionStage,
  progressLabel: string,
  recordingLevel: number | undefined
): VoiceSessionActivityPresentation {
  if (stage === 'listening') {
    return { kind: 'listening', label: t('mobile.VoiceSessionPresentation.listening'), level: boundedLevel(recordingLevel) };
  }
  if (stage === 'processing' || stage === 'speaking') {
    return { kind: 'busy', label: progressLabel };
  }
  return { kind: 'idle' };
}

function progressTraceForState(stage: VoiceInteractionStage, progressSteps: readonly string[], realtime: VoiceRealtimeState | null): readonly string[] {
  if (realtime?.actionPlan) {
    return [];
  }
  if (stage !== 'processing' && stage !== 'speaking' && stage !== 'review') {
    return [];
  }
  const bounded = uniqueProgressSteps(progressSteps).slice(-5);
  return bounded.length > 1 ? bounded : [];
}

function uniqueProgressSteps(steps: readonly string[]): readonly string[] {
  const unique: string[] = [];
  for (const step of steps) {
    const normalized = step.replace(/\s+/g, ' ').trim();
    if (!normalized || unique[unique.length - 1] === normalized) {
      continue;
    }
    unique.push(normalized.length <= 72 ? normalized : `${normalized.slice(0, 71).trim()}...`);
  }
  return unique;
}

function safeProgressPresentationSteps(steps: readonly string[]): readonly string[] {
  return steps
    .map((step) => safeProgressPresentationText(step, 100))
    .filter((step) => step.length > 0);
}

function safeProgressPresentationText(value: string, maxLength: number): string {
  if (unsafeProgressPresentationText(value)) {
    return t('mobile.VoiceSessionPresentation.workingSafely');
  }
  const normalized = redactUnsafeVoiceText(value)
    .replace(/\s+/g, ' ')
    .trim();
  if (normalized.length <= maxLength) {
    return normalized;
  }
  return normalized.slice(0, maxLength).trim();
}

function unsafeProgressPresentationText(value: string): boolean {
  return /\b(raw prompt|stack trace|raw query|raw transcript|raw provider response|raw model response|provider[-_ ]?session[-_ ]?id|authorization|credential|password|secret|token|bearer|api[-_ ]?key|asset[-_ ]?id|parent[-_ ]?asset[-_ ]?id|inventory[-_ ]?id|tenant[-_ ]?id|tool[-_ ]?call[-_ ]?id)\b/gi.test(value);
}

function boundedLevel(value: number | undefined): number {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    return 0;
  }
  return Math.max(0, Math.min(1, value));
}

function summarizeActionPlanCommands(commands: readonly VoiceActionPlanCommand[]): string {
  const creates = commands.filter((command) => command.operation === 'create' || command.kind === 'create_asset' || command.kind === 'create_location').length;
  if (creates === 0) {
    return t('mobile.voice.planChangeCount', { count: commands.length });
  }
  return t('mobile.voice.planCreateCount', { count: creates });
}

function formatActionPlanCommands(commands: readonly VoiceActionPlanCommand[]): readonly VoiceSessionActionPlanCommand[] {
  const titlesByID = new Map<string, string>();
  for (const command of commands) {
    if (command.id) {
      titlesByID.set(command.id, displayTitleForActionPlanCommand(command));
    }
  }
  const formatted: VoiceSessionActionPlanCommand[] = [];
  const usedExistingParents = new Set<string>();
  for (const command of commands) {
    if (command.parentAssetId && !usedExistingParents.has(command.parentAssetId)) {
      usedExistingParents.add(command.parentAssetId);
      formatted.push(formatExistingParentUseCommand(command));
    }
    formatted.push(formatActionPlanCommand(command, titlesByID));
  }
  return formatted;
}

function formatExistingParentUseCommand(command: VoiceActionPlanCommand): VoiceSessionActionPlanCommand {
  const parentKind = friendlyParentKind(command.parentKind);
  return {
    id: command.parentAssetId ? `use-${command.parentAssetId}` : undefined,
    title: command.parentTitle ?? t('mobile.VoiceSessionPresentation.existingPlace'),
    subtitle: t('mobile.VoiceSessionPresentation.useExisting', { parentKind: String(parentKind) }),
    photoDraftEligible: false,
    editable: false,
    tone: 'use'
  };
}

function formatActionPlanCommand(command: VoiceActionPlanCommand, titlesByID: ReadonlyMap<string, string>): VoiceSessionActionPlanCommand {
  const tone = command.operation === 'create' || command.kind === 'create_asset' || command.kind === 'create_location'
    ? 'create'
    : 'update';
  const assetKind = friendlyAssetKind(command.assetKind || command.kind);
  const title = displayTitleForActionPlanCommand(command);
  return {
    id: command.id,
    title,
    subtitle: tone === 'create' ? t('mobile.VoiceSessionPresentation.create', { assetKind: String(assetKind) }) : command.summary,
    placement: placementLabel(command, titlesByID),
    expirationLabel: formatExpirationChange(command.expiration, command.expirationCleared),
    changes: command.changes,
    photoDraftEligible: isPhotoDraftEligible(command, title),
    editable: tone === 'create' && Boolean(command.id),
    tone
  };
}

function displayTitleForActionPlanCommand(command: VoiceActionPlanCommand): string {
  const tone = command.operation === 'create' || command.kind === 'create_asset' || command.kind === 'create_location'
    ? 'create'
    : 'update';
  const verifiedTitle = command.title?.trim() ?? '';
  return tone === 'create'
    ? verifiedTitle || command.summary
    : verifiedTitle || neutralExistingAssetTitle(command.assetKind);
}

function neutralExistingAssetTitle(value: string | undefined): string {
  switch (value) {
    case 'item':
      return t('mobile.VoiceSessionPresentation.selectedItem');
    case 'container':
      return t('mobile.VoiceSessionPresentation.selectedContainer');
    case 'location':
      return t('mobile.VoiceSessionPresentation.selectedLocation');
    default:
      return t('mobile.VoiceSessionPresentation.selectedAsset');
  }
}

function isPhotoDraftEligible(command: VoiceActionPlanCommand, title: string): boolean {
  if (!command.id) {
    return false;
  }
  const assetKind = command.assetKind || command.kind;
  if (command.kind === 'move_asset' || command.operation === 'move') {
    const hasVerifiedTitle = Boolean(command.title?.trim() && command.title.trim() === title);
    return hasVerifiedTitle && (assetKind === 'item' || assetKind === 'container' || assetKind === 'location');
  }
  if (command.kind === 'create_asset' || command.kind === 'create_location' || command.operation === 'create') {
    return assetKind === 'item' || assetKind === 'container' || assetKind === 'location' || command.kind === 'create_asset' || command.kind === 'create_location';
  }
  return false;
}

function placementLabel(command: VoiceActionPlanCommand, titlesByID: ReadonlyMap<string, string>): string | undefined {
  if (command.parentCommandId) {
    return t('mobile.VoiceSessionPresentation.insideNew', { value: String(titlesByID.get(command.parentCommandId) ?? 'container') });
  }
  if (command.parentAssetId) {
    return t('mobile.VoiceSessionPresentation.inside', { value: String(command.parentTitle ?? 'existing place') });
  }
  return undefined;
}

function friendlyAssetKind(value: string | undefined): string {
  switch (value) {
    case 'create_location':
    case 'location':
      return 'location';
    case 'container':
      return 'container';
    case 'item':
    case 'create_asset':
      return 'item';
    default:
      return 'item';
  }
}

function friendlyParentKind(value: string | undefined): string {
  switch (value) {
    case 'location':
      return 'location';
    case 'container':
      return 'container';
    default:
      return 'place';
  }
}

function isProviderRecoveryFailure(code: VoiceRealtimeState['failureCode']): boolean {
  return code === 'provider_billing_disabled' ||
    code === 'provider_readiness' ||
    code === 'speech_to_text_failed' ||
    code === 'language_inference_failed' ||
    code === 'text_to_speech_failed';
}

function bottomActionForState(stage: VoiceInteractionStage, realtime: VoiceRealtimeState | null): VoiceSessionBottomAction {
  if (realtime?.actionPlan?.status === 'proposed' && !realtime.reviewDecisionPending) {
    return { kind: 'review_decision', planId: realtime.actionPlan.planId };
  }
  if ((realtime?.actionPlan?.status === 'approved' || realtime?.actionPlan?.status === 'proposed') && realtime.reviewDecisionPending) {
    return { kind: 'none' };
  }

  if (stage === 'ready' || stage === 'listening' || stage === 'processing' || stage === 'speaking' || stage === 'completed' || stage === 'cancelled' || stage === 'failed') {
    const isWorking = stage === 'processing' || stage === 'speaking';
    return {
      kind: 'session_controls',
      canCancel: canCancelConversation(stage, realtime),
      mic: {
        accessibilityLabel: stage === 'listening'
          ? t('mobile.VoiceSessionPresentation.sendVoiceRequest')
          : stage === 'ready' && !realtime
            ? t('mobile.VoiceSessionPresentation.startVoiceInteraction')
            : stage === 'completed' && hasAvailableVoiceFollowUp(realtime)
              ? realtime?.responseKind === 'clarification' ? t('mobile.VoiceSessionPresentation.answerFollowUp') : t('mobile.VoiceSessionPresentation.askAFollowUp')
            : isWorking
              ? t('mobile.VoiceSessionPresentation.voiceRequestInProgress')
              : t('mobile.VoiceSessionPresentation.startAnotherVoiceInteraction'),
        disabled: isWorking,
        icon: stage === 'listening' ? 'send' : isWorking ? 'busy' : 'mic',
        selected: stage === 'listening'
      }
    };
  }

  return { kind: 'none' };
}

function titleForState(stage: VoiceInteractionStage, realtime: VoiceRealtimeState | null): string {
  if (stage === 'completed' && hasAvailableVoiceFollowUp(realtime)) {
    return realtime?.responseKind === 'clarification' ? t('mobile.VoiceSessionPresentation.needsDetail') : t('mobile.VoiceSessionPresentation.answerReady');
  }
  if (stage === 'completed') {
    return completedVoicePresentation(realtime).title;
  }
  return titleForStage(stage);
}

function bottomHintForState(stage: VoiceInteractionStage, realtime: VoiceRealtimeState | null): string {
  if (stage === 'completed' && hasAvailableVoiceFollowUp(realtime)) {
    return realtime?.responseKind === 'clarification' ? t('mobile.VoiceSessionPresentation.answerTheFollowUpToKeepThisConversationGoing') : t('mobile.VoiceSessionPresentation.askAFollowUpToKeepThisConversationGoing');
  }
  if (stage === 'completed') {
    return completedVoicePresentation(realtime).bottomHint;
  }
  switch (stage) {
    case 'ready':
      return t('mobile.VoiceSessionPresentation.askAQuestionAboutThisInventory');
    case 'cancelled':
      return t('mobile.VoiceSessionPresentation.youCanStartAgainWhenYouAreReady');
    case 'failed':
      return t('mobile.VoiceSessionPresentation.resetAndTryAgainWhenYouAreReady');
    default:
      return t('mobile.VoiceSessionPresentation.keepThisOpenWhileStuffStashWorks');
  }
}

function completedVoicePresentation(realtime: VoiceRealtimeState | null | undefined): {
  readonly accessibilityLabel: string;
  readonly bottomHint: string;
  readonly title: string;
  readonly tone: VoiceAccessoryPresentation['tone'];
} {
  switch (realtime?.actionPlan?.status) {
    case 'executed':
      return {
        accessibilityLabel: t('mobile.VoiceSessionPresentation.openAppliedVoiceChange'),
        bottomHint: t('mobile.VoiceSessionPresentation.theReviewedChangeWasApplied'),
        title: t('mobile.VoiceSessionPresentation.changeApplied'),
        tone: 'ready'
      };
    case 'cancelled':
      return {
        accessibilityLabel: t('mobile.VoiceSessionPresentation.openCancelledVoiceChange'),
        bottomHint: t('mobile.VoiceSessionPresentation.noChangeWasMade'),
        title: t('mobile.VoiceSessionPresentation.changeCancelled'),
        tone: 'attention'
      };
  }
  switch (realtime?.responseKind) {
    case 'unsupported_action':
      return {
        accessibilityLabel: t('mobile.VoiceSessionPresentation.openUnsupportedVoiceResult'),
        bottomHint: t('mobile.VoiceSessionPresentation.tryAnotherWayToMakeThisChange'),
        title: t('mobile.VoiceSessionPresentation.voiceActionUnavailable'),
        tone: 'attention'
      };
    case 'safe_failure':
      return {
        accessibilityLabel: t('mobile.VoiceSessionPresentation.openSafeVoiceResult'),
        bottomHint: t('mobile.VoiceSessionPresentation.startAFreshRequestOrCloseThis'),
        title: t('mobile.VoiceSessionPresentation.couldNotFinishSafely'),
        tone: 'attention'
      };
    default:
      return {
        accessibilityLabel: t('mobile.VoiceSessionPresentation.openVoiceAnswer'),
        bottomHint: t('mobile.VoiceSessionPresentation.youCanAskAnotherQuestionOrCloseThis'),
        title: t('mobile.VoiceSessionPresentation.answerReady'),
        tone: 'ready'
      };
  }
}

function hasAvailableVoiceFollowUp(realtime: VoiceRealtimeState | null | undefined): boolean {
  return (realtime?.responseKind === 'clarification' || realtime?.responseKind === 'answer') && realtime.followUpAvailable === true;
}

export function formatSafeDiagnosticEvent(event: VoiceSafeDiagnosticEvent): string {
  const summary = `${event.label}: ${event.status}`;
  return event.detail ? `${summary}\n${event.detail}` : summary;
}

function describeVoiceContext(pathname: string): string {
  if (pathname.startsWith('/assets/')) {
    return t('mobile.VoiceSessionPresentation.assetContext');
  }

  if (pathname.startsWith('/locations/')) {
    return t('mobile.VoiceSessionPresentation.locationContext');
  }

  if (pathname === '/search') {
    return t('mobile.VoiceSessionPresentation.searchContext');
  }

  if (pathname === '/add') {
    return t('mobile.VoiceSessionPresentation.addContext');
  }

  return t('mobile.VoiceSessionPresentation.currentInventory');
}

function titleForStage(stage: VoiceInteractionStage): string {
  switch (stage) {
    case 'listening':
      return t('mobile.VoiceSessionPresentation.listening');
    case 'processing':
      return t('mobile.VoiceSessionPresentation.checkingInventory');
    case 'speaking':
      return t('mobile.VoiceSessionPresentation.speaking');
    case 'completed':
      return t('mobile.VoiceSessionPresentation.answerReady');
    case 'cancelled':
      return t('mobile.VoiceSessionPresentation.cancelled');
    case 'failed':
      return t('mobile.VoiceSessionPresentation.couldNotFinish');
    case 'review':
      return t('mobile.VoiceSessionPresentation.reviewNeeded');
    case 'ready':
      return t('mobile.VoiceSessionPresentation.askStuffStash');
  }
}

function progressForStage(stage: VoiceInteractionStage): string {
  switch (stage) {
    case 'listening':
      return t('mobile.VoiceSessionPresentation.tapTheMicWhenYouAreDone');
    case 'processing':
      return t('mobile.VoiceSessionPresentation.lookingThroughYourInventory');
    case 'speaking':
      return t('mobile.VoiceSessionPresentation.playingTheResponse');
    case 'completed':
      return t('mobile.VoiceSessionPresentation.responseComplete');
    case 'cancelled':
      return t('mobile.VoiceSessionPresentation.sessionCancelled');
    case 'failed':
      return t('mobile.VoiceSessionPresentation.voiceFailedSafely');
    case 'review':
      return t('mobile.VoiceSessionPresentation.reviewTheSuggestedAction');
    case 'ready':
      return t('mobile.VoiceSessionPresentation.tapTheMicAndAskAboutThisInventory');
  }
}
