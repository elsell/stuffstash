import type {
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
  pathname,
  status,
  stage
}: {
  readonly pathname: string;
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
      title: 'Checking inventory',
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
      subtitle: context,
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
      subtitle: context,
      title: 'Voice failed',
      tone: 'failed'
    };
  }

  if (stage === 'review') {
    return {
      accessibilityLabel: 'Review voice plan',
      primaryAction: 'expand',
      subtitle: context,
      title: 'Review needed',
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

export type VoiceSessionPresentation = {
  readonly actionPlan?: {
    readonly planId: string;
    readonly status: 'proposed' | 'approved' | 'cancelled' | 'executed' | 'failed';
    readonly confirmationSummary: string;
    readonly commands: readonly string[];
    readonly risks: readonly string[];
  };
  readonly bottomAction: VoiceSessionBottomAction;
  readonly canReset: boolean;
  readonly contextLabel: string;
  readonly diagnostics: readonly string[] | null;
  readonly isBusy: boolean;
  readonly progressLabel: string;
  readonly recoveryAction?: VoiceSessionRecoveryAction;
  readonly response?: string;
  readonly title: string;
  readonly transcript?: string;
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

  return {
    actionPlan: realtime?.actionPlan
      ? {
          planId: realtime.actionPlan.planId,
          status: realtime.actionPlan.status,
          confirmationSummary: realtime.actionPlan.confirmationSummary,
          commands: realtime.actionPlan.commands.map((command) => command.summary),
          risks: realtime.actionPlan.risks
        }
      : undefined,
    bottomAction,
    canReset: stage === 'completed' || stage === 'cancelled' || stage === 'failed' || (stage === 'review' && realtime?.actionPlan?.status !== 'proposed'),
    contextLabel: `${inventoryName} · ${tenantName}`,
    diagnostics,
    isBusy: stage === 'listening' || stage === 'processing' || stage === 'speaking',
    progressLabel,
    recoveryAction: isProviderRecoveryFailure(realtime?.failureCode)
      ? { label: 'Voice providers', target: 'provider_profiles' }
      : undefined,
    response: realtime?.spokenResponse,
    title,
    transcript: realtime?.transcript
  };
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
  return `${event.label}: ${event.status}`;
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
