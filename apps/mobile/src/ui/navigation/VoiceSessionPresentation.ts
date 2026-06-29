import type {
  VoiceActionPlanCommand,
  VoiceRealtimeState,
  VoiceSafeDiagnosticEvent
} from '../../application/voice/RealtimeVoiceSession';
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
      accessibilityLabel: 'Stop listening',
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
      tone: 'active'
    };
  }

  if (stage === 'completed') {
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
      title: 'Voice failed',
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
  const normalized = value?.replace(/\s+/g, ' ').trim();
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
      return 'Checking inventory';
    case 'listening':
      return 'Listening';
    default:
      return 'Checking inventory';
  }
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

export type VoiceSessionPresentation = {
  readonly actionPlan?: {
    readonly planId: string;
    readonly status: 'proposed' | 'approved' | 'cancelled' | 'executed' | 'failed';
    readonly confirmationSummary: string;
    readonly summary: string;
    readonly commands: readonly VoiceSessionActionPlanCommand[];
    readonly risks: readonly string[];
  };
  readonly bottomAction: VoiceSessionBottomAction;
  readonly canReset: boolean;
  readonly contextLabel: string;
  readonly diagnostics: readonly string[] | null;
  readonly isBusy: boolean;
  readonly progressLabel: string;
  readonly progressSteps: readonly string[];
  readonly recoveryAction?: VoiceSessionRecoveryAction;
  readonly response?: string;
  readonly title: string;
  readonly transcript?: string;
};

export type VoiceSessionActionPlanCommand = {
  readonly id?: string;
  readonly title: string;
  readonly subtitle: string;
  readonly placement?: string;
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
  const title = titleForStage(stage);
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
    bottomAction,
    canReset: stage === 'completed' || stage === 'cancelled' || stage === 'failed' || (stage === 'review' && realtime?.actionPlan?.status !== 'proposed'),
    contextLabel: `${inventoryName} · ${tenantName}`,
    diagnostics,
    isBusy: stage === 'listening' || stage === 'processing' || stage === 'speaking',
    progressLabel,
    progressSteps: realtime?.progressSteps ?? [],
    recoveryAction: isProviderRecoveryFailure(realtime?.failureCode)
      ? { label: 'Voice providers', target: 'provider_profiles' }
      : undefined,
    response: realtime?.spokenResponse,
    title,
    transcript: realtime?.transcript ?? activePartialTranscript
  };
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
    tone: 'use'
  };
}

function formatActionPlanCommand(command: VoiceActionPlanCommand, titlesByID: ReadonlyMap<string, string>): VoiceSessionActionPlanCommand {
  const title = command.title || command.summary;
  const assetKind = friendlyAssetKind(command.assetKind || command.kind);
  const tone = command.operation === 'create' || command.kind === 'create_asset' || command.kind === 'create_location'
    ? 'create'
    : 'update';
  return {
    id: command.id,
    title,
    subtitle: tone === 'create' ? `Create ${assetKind}` : command.summary,
    placement: placementLabel(command, titlesByID),
    tone
  };
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
          ? 'Stop listening'
          : stage === 'ready' && !realtime
            ? 'Start voice interaction'
            : 'Start another voice interaction',
        disabled: isWorking,
        selected: stage === 'listening'
      }
    };
  }

  return { kind: 'none' };
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
