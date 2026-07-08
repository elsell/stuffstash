import type {
  VoiceActionPlanCommand,
  VoiceRealtimeState,
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
      accessibilityLabel: 'Open voice status',
      primaryAction: 'expand',
      subtitle: context,
      title: 'Voice loading',
      tone: 'attention'
    };
  }

  if (status === 'error') {
    return {
      accessibilityLabel: 'Open voice error',
      primaryAction: 'expand',
      subtitle: context,
      title: 'Voice unavailable',
      tone: 'failed'
    };
  }

  if (stage === 'listening') {
    return {
      accessibilityLabel: 'Send voice request',
      primaryAction: 'stop',
      subtitle: context,
      title: 'Listening',
      tone: 'active'
    };
  }

  if (stage === 'processing') {
    return {
      accessibilityLabel: 'Open voice session',
      primaryAction: 'expand',
      subtitle: context,
      title: accessoryProgressTitle(realtime),
      tone: 'attention'
    };
  }

  if (stage === 'speaking') {
    return {
      accessibilityLabel: 'Open voice response',
      primaryAction: 'expand',
      subtitle: context,
      title: 'Speaking',
      tone: 'attention'
    };
  }

  if (stage === 'completed') {
    if (hasAvailableClarificationFollowUp(realtime)) {
      return {
        accessibilityLabel: 'Open voice follow-up',
        primaryAction: 'expand',
        subtitle: safeAccessorySubtitle(realtime?.spokenResponse) ?? context,
        title: 'Needs detail',
        tone: 'attention'
      };
    }
    return {
      accessibilityLabel: 'Open voice answer',
      primaryAction: 'expand',
      subtitle: safeAccessorySubtitle(realtime?.spokenResponse) ?? context,
      title: 'Answer ready',
      tone: 'ready'
    };
  }

  if (stage === 'cancelled') {
    return {
      accessibilityLabel: 'Open cancelled voice session',
      primaryAction: 'expand',
      subtitle: context,
      title: 'Voice cancelled',
      tone: 'attention'
    };
  }

  if (stage === 'failed') {
    return {
      accessibilityLabel: 'Open voice error',
      primaryAction: 'expand',
      subtitle: safeFailureAccessorySubtitle(realtime, diagnosticsEnabled) ?? context,
      title: safeFailureAccessoryTitle(realtime),
      tone: 'failed'
    };
  }

  if (stage === 'review') {
    return {
      accessibilityLabel: 'Review voice plan',
      primaryAction: 'expand',
      subtitle: context,
      title: accessoryProgressTitle(realtime),
      tone: 'attention'
    };
  }

  return {
    accessibilityLabel: 'Start voice interaction',
    primaryAction: 'start',
    subtitle: context,
    title: 'Ask Stuff Stash',
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
      return 'Review needed';
    case 'speaking':
      return 'Speaking';
    case 'processing':
      return accessoryPhaseTitle(realtime.conversationPhase);
    case 'listening':
      return 'Listening';
    default:
      return 'Checking inventory';
  }
}

const voiceAccessoryPhaseTitles = {
  understanding: 'Understanding request',
  exploring: 'Checking inventory',
  planning: 'Preparing plan',
  reviewing: 'Preparing review',
  answering: 'Preparing answer',
  recovering: 'Recovering safely'
} satisfies Record<NonNullable<VoiceRealtimeState['conversationPhase']>, string>;

function accessoryPhaseTitle(phase: VoiceRealtimeState['conversationPhase']): string {
  return phase ? voiceAccessoryPhaseTitles[phase] : 'Checking inventory';
}

function safeFailureAccessorySubtitle(realtime: VoiceRealtimeState | null | undefined, diagnosticsEnabled: boolean): string | undefined {
  const code = realtime?.failureCode;
  if (
    code === 'provider_readiness' ||
    code === 'speech_to_text_failed' ||
    code === 'text_to_speech_failed'
  ) {
    return 'Check Voice providers and try again.';
  }
  if (code === 'language_inference_failed') {
    return diagnosticsEnabled ? 'Open diagnostics or check Voice providers.' : 'Check Voice providers and try again.';
  }
  return realtime?.status === 'failed' ? 'Open for details.' : undefined;
}

function safeFailureAccessoryTitle(realtime: VoiceRealtimeState | null | undefined): string {
  switch (realtime?.failureCode) {
    case 'speech_to_text_failed':
      return 'Speech input failed';
    case 'language_inference_failed':
      return 'Agent brain failed';
    case 'text_to_speech_failed':
      return 'Speech output failed';
    case 'clarification_turn_limit':
      return 'Voice needs a fresh start';
    case 'provider_readiness':
      return 'Voice providers needed';
    default:
      return 'Voice failed';
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
  readonly title: string;
  readonly transcript?: string;
};

export type VoiceSessionActivityPresentation =
  | { readonly kind: 'idle' }
  | { readonly kind: 'listening'; readonly label: string; readonly level: number }
  | { readonly kind: 'busy'; readonly label: string };

export type VoiceSessionActionPlanCommand = {
  readonly id?: string;
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
  const progressLabel = realtime?.progressLabel ?? progressForStage(stage);
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
    progressSteps: realtime?.progressSteps ?? [],
    progressTrace: progressTraceForState(stage, realtime),
    recoveryAction: isProviderRecoveryFailure(realtime?.failureCode)
      ? { label: 'Voice providers', target: 'provider_profiles' }
      : undefined,
    response: realtime?.spokenResponse,
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
    return { kind: 'listening', label: 'Listening', level: boundedLevel(recordingLevel) };
  }
  if (stage === 'processing' || stage === 'speaking') {
    return { kind: 'busy', label: progressLabel };
  }
  return { kind: 'idle' };
}

function progressTraceForState(stage: VoiceInteractionStage, realtime: VoiceRealtimeState | null): readonly string[] {
  if (realtime?.actionPlan) {
    return [];
  }
  if (stage !== 'processing' && stage !== 'speaking' && stage !== 'review') {
    return [];
  }
  const steps = realtime?.progressSteps ?? [];
  const bounded = uniqueProgressSteps(steps).slice(-5);
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

function boundedLevel(value: number | undefined): number {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    return 0;
  }
  return Math.max(0, Math.min(1, value));
}

function summarizeActionPlanCommands(commands: readonly VoiceActionPlanCommand[]): string {
  const creates = commands.filter((command) => command.operation === 'create' || command.kind === 'create_asset' || command.kind === 'create_location').length;
  if (creates === 0) {
    return `${commands.length} ${commands.length === 1 ? 'change' : 'changes'}`;
  }
  return `${creates} new ${creates === 1 ? 'thing' : 'things'}`;
}

function formatActionPlanCommands(commands: readonly VoiceActionPlanCommand[]): readonly VoiceSessionActionPlanCommand[] {
  const titlesByID = new Map<string, string>();
  for (const command of commands) {
    if (command.id) {
      titlesByID.set(command.id, command.title || command.summary);
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
    title: command.parentTitle ?? 'Existing place',
    subtitle: `Use existing ${parentKind}`,
    photoDraftEligible: false,
    tone: 'use'
  };
}

function formatActionPlanCommand(command: VoiceActionPlanCommand, titlesByID: ReadonlyMap<string, string>): VoiceSessionActionPlanCommand {
  const tone = command.operation === 'create' || command.kind === 'create_asset' || command.kind === 'create_location'
    ? 'create'
    : 'update';
  const assetKind = friendlyAssetKind(command.assetKind || command.kind);
  const title = tone === 'create'
    ? command.title || command.summary
    : command.title || neutralExistingAssetTitle(command.assetKind);
  return {
    id: command.id,
    title,
    subtitle: tone === 'create' ? `Create ${assetKind}` : command.summary,
    placement: placementLabel(command, titlesByID),
    photoDraftEligible: isPhotoDraftEligible(command, title),
    tone
  };
}

function neutralExistingAssetTitle(value: string | undefined): string {
  switch (value) {
    case 'item':
      return 'Selected item';
    case 'container':
      return 'Selected container';
    case 'location':
      return 'Selected location';
    default:
      return 'Selected asset';
  }
}

function isPhotoDraftEligible(command: VoiceActionPlanCommand, title: string): boolean {
  if (!command.id) {
    return false;
  }
  const assetKind = command.assetKind || command.kind;
  if (command.kind === 'move_asset' || command.operation === 'move') {
    const hasVerifiedTitle = Boolean(command.title && command.title === title);
    return hasVerifiedTitle && (assetKind === 'item' || assetKind === 'container' || assetKind === 'location');
  }
  if (command.kind === 'create_asset' || command.kind === 'create_location' || command.operation === 'create') {
    return assetKind === 'item' || assetKind === 'container' || assetKind === 'location' || command.kind === 'create_asset' || command.kind === 'create_location';
  }
  return false;
}

function placementLabel(command: VoiceActionPlanCommand, titlesByID: ReadonlyMap<string, string>): string | undefined {
  if (command.parentCommandId) {
    return `Inside new ${titlesByID.get(command.parentCommandId) ?? 'container'}`;
  }
  if (command.parentAssetId) {
    return `Inside ${command.parentTitle ?? 'existing place'}`;
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
  return code === 'provider_readiness' ||
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
      canCancel: stage === 'listening' || isWorking,
      mic: {
        accessibilityLabel: stage === 'listening'
          ? 'Send voice request'
          : stage === 'ready' && !realtime
            ? 'Start voice interaction'
            : stage === 'completed' && hasAvailableClarificationFollowUp(realtime)
              ? 'Answer follow-up'
            : isWorking
              ? 'Voice request in progress'
              : 'Start another voice interaction',
        disabled: isWorking,
        icon: stage === 'listening' ? 'send' : isWorking ? 'busy' : 'mic',
        selected: stage === 'listening'
      }
    };
  }

  return { kind: 'none' };
}

function titleForState(stage: VoiceInteractionStage, realtime: VoiceRealtimeState | null): string {
  if (stage === 'completed' && hasAvailableClarificationFollowUp(realtime)) {
    return 'Needs detail';
  }
  return titleForStage(stage);
}

function bottomHintForState(stage: VoiceInteractionStage, realtime: VoiceRealtimeState | null): string {
  if (stage === 'completed' && hasAvailableClarificationFollowUp(realtime)) {
    return 'Answer the follow-up to keep this conversation going.';
  }
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

function hasAvailableClarificationFollowUp(realtime: VoiceRealtimeState | null | undefined): boolean {
  return realtime?.responseKind === 'clarification' && realtime.clarificationFollowUpAvailable === true;
}

export function formatSafeDiagnosticEvent(event: VoiceSafeDiagnosticEvent): string {
  const summary = `${event.label}: ${event.status}`;
  return event.detail ? `${summary}\n${event.detail}` : summary;
}

function describeVoiceContext(pathname: string): string {
  if (pathname.startsWith('/assets/')) {
    return 'Asset context';
  }

  if (pathname.startsWith('/locations/')) {
    return 'Location context';
  }

  if (pathname === '/search') {
    return 'Search context';
  }

  if (pathname === '/add') {
    return 'Add context';
  }

  return 'Current inventory';
}

function titleForStage(stage: VoiceInteractionStage): string {
  switch (stage) {
    case 'listening':
      return 'Listening';
    case 'processing':
      return 'Checking inventory';
    case 'speaking':
      return 'Speaking';
    case 'completed':
      return 'Answer ready';
    case 'cancelled':
      return 'Cancelled';
    case 'failed':
      return 'Could not finish';
    case 'review':
      return 'Review needed';
    case 'ready':
      return 'Ask Stuff Stash';
  }
}

function progressForStage(stage: VoiceInteractionStage): string {
  switch (stage) {
    case 'listening':
      return 'Tap the mic when you are done.';
    case 'processing':
      return 'Looking through your inventory.';
    case 'speaking':
      return 'Playing the response.';
    case 'completed':
      return 'Response complete.';
    case 'cancelled':
      return 'Session cancelled.';
    case 'failed':
      return 'Voice failed safely.';
    case 'review':
      return 'Review the suggested action.';
    case 'ready':
      return 'Tap the mic and ask about this inventory.';
  }
}
