import { createContext, ReactNode, useContext, useEffect, useMemo, useRef, useState } from 'react';
import {
  VoiceInteractionPreviewQuery,
  VoiceInteractionPreviewViewModel
} from '../../application/voice/VoiceInteractionPreviewQuery';
import {
  RealtimeVoiceSessionController,
  VoiceRealtimeFailureCode,
  VoiceRealtimeCancelledError,
  VoiceRealtimeState,
  VoicePhotoAttachmentStatus,
  type VoiceActionPlanPhotoDrafts
} from '../../application/voice/RealtimeVoiceSession';
import { redactUnsafeVoiceStructuredText } from '../../application/voice/VoiceTextSafety';

export type VoiceInteractionStage = 'ready' | 'listening' | 'review' | 'processing' | 'speaking' | 'completed' | 'cancelled' | 'failed';

export type VoiceInteractionState =
  | { readonly status: 'loading'; readonly stage: VoiceInteractionStage }
  | { readonly status: 'error'; readonly stage: VoiceInteractionStage; readonly message: string }
  | {
      readonly status: 'ready';
      readonly stage: VoiceInteractionStage;
      readonly preview: VoiceInteractionPreviewViewModel;
      readonly realtime: VoiceRealtimeState | null;
    };

type VoiceInteractionStateContextValue = {
  readonly diagnosticsEnabled: boolean;
  readonly state: VoiceInteractionState;
  readonly setStage: (stage: VoiceInteractionStage) => void;
  readonly startRealtime: () => Promise<void>;
  readonly stopRealtime: () => Promise<void>;
  readonly approveRealtimeActionPlan: (planId: string, photoDrafts?: VoiceActionPlanPhotoDrafts) => Promise<void>;
  readonly cancelRealtimeActionPlan: (planId: string) => Promise<void>;
  readonly retryRealtimeActionPlanPhotos: (planId: string) => Promise<void>;
  readonly cancelRealtime: () => Promise<void>;
  readonly reset: () => void;
};

type VoiceFailureContext = {
  readonly tenantName?: string;
  readonly inventoryName?: string;
};

const VoiceInteractionStateContext = createContext<VoiceInteractionStateContextValue | null>(null);

type VoiceInteractionStateProviderProps = {
  readonly children: ReactNode;
  readonly diagnosticsEnabled?: boolean;
  readonly previewQuery: VoiceInteractionPreviewQuery;
  readonly realtimeController: RealtimeVoiceSessionController;
};

export function VoiceInteractionStateProvider({
  children,
  diagnosticsEnabled = false,
  previewQuery,
  realtimeController
}: VoiceInteractionStateProviderProps) {
  const [stage, setStage] = useState<VoiceInteractionStage>('ready');
  const [realtime, setRealtime] = useState<VoiceRealtimeState | null>(null);
  const sessionGeneration = useRef(0);
  const [previewState, setPreviewState] = useState<
    | { readonly status: 'loading' }
    | { readonly status: 'error'; readonly message: string }
    | { readonly status: 'ready'; readonly preview: VoiceInteractionPreviewViewModel }
  >({ status: 'loading' });

  useEffect(() => {
    let isCurrent = true;

    previewQuery
      .execute()
      .then((preview) => {
        if (isCurrent) {
          setPreviewState({ status: 'ready', preview });
        }
      })
      .catch((error: unknown) => {
        if (isCurrent) {
          setPreviewState({
            status: 'error',
            message: readableError(error, 'Voice preview is not available.')
          });
        }
      });

    return () => {
      isCurrent = false;
    };
  }, [previewQuery]);

  useEffect(() => {
    if (stage !== 'listening') {
      return;
    }

    const interval = setInterval(() => {
      const recordingLevel = realtimeController.recordingLevel();
      setRealtime((current) => applyRecordingLevelToRealtime(current, recordingLevel));
    }, 100);

    return () => {
      clearInterval(interval);
    };
  }, [realtimeController, stage]);

  useEffect(() => {
    if (
      stage !== 'completed' ||
      realtime?.responseKind !== 'clarification' ||
      realtime.clarificationFollowUpAvailable !== true
    ) {
      return;
    }

    const interval = setInterval(() => {
      setRealtime((current) => refreshClarificationFollowUpAvailability(
        current,
        realtimeController.canSendFollowUpAudio()
      ));
    }, 500);

    return () => {
      clearInterval(interval);
    };
  }, [realtime?.clarificationFollowUpAvailable, realtime?.responseKind, realtimeController, stage]);

  const value = useMemo<VoiceInteractionStateContextValue>(() => {
    const state: VoiceInteractionState =
      previewState.status === 'ready'
        ? { status: 'ready', stage, preview: previewState.preview, realtime }
        : previewState.status === 'error'
          ? { status: 'error', stage, message: previewState.message }
          : { status: 'loading', stage };

    return {
      diagnosticsEnabled,
      state,
      setStage,
      startRealtime: async () => {
        const generation = sessionGeneration.current + 1;
        sessionGeneration.current = generation;
        try {
          const next = realtime?.status === 'completed' &&
            realtime.responseKind === 'clarification' &&
            realtime.clarificationFollowUpAvailable === true &&
            realtimeController.canSendFollowUpAudio()
            ? await realtimeController.startFollowUp()
            : await realtimeController.start();
          if (sessionGeneration.current !== generation) {
            return;
          }
          setRealtime(next);
          setStage('listening');
        } catch (error) {
          if (sessionGeneration.current !== generation) {
            return;
          }
          setRealtime(buildFailedVoiceRealtimeState(error, voiceFailureContext(realtime, previewState)));
          setStage('failed');
        }
      },
      stopRealtime: async () => {
        const generation = sessionGeneration.current;
        setStage('processing');
        try {
          const shouldSendFollowUp = realtime?.responseKind === 'clarification' &&
            realtime.clarificationFollowUpAvailable === true &&
            realtimeController.canSendFollowUpAudio();
          const states = await (shouldSendFollowUp ? realtimeController.stopFollowUp : realtimeController.stop).call(realtimeController, (nextState: VoiceRealtimeState) => {
            if (sessionGeneration.current !== generation) {
              return;
            }
            setRealtime(nextState);
            setStage(nextState.status);
          });
          const finalState = states[states.length - 1] ?? null;
          if (sessionGeneration.current !== generation) {
            return;
          }
          setRealtime(finalState);
          setStage(finalState?.status ?? 'failed');
        } catch (error) {
          if (sessionGeneration.current !== generation) {
            return;
          }
          if (isVoiceCancelledError(error)) {
            const cancelled = await realtimeController.cancel();
            if (sessionGeneration.current !== generation) {
              return;
            }
            setRealtime(cancelled);
            setStage('cancelled');
            return;
          }
          setRealtime(buildFailedVoiceRealtimeState(error, voiceFailureContext(realtime, previewState)));
          setStage('failed');
        }
      },
      approveRealtimeActionPlan: async (planId: string, photoDrafts?: VoiceActionPlanPhotoDrafts) => {
        setRealtime((current) => markReviewDecisionPending(current, 'Approving change'));
        try {
          await realtimeController.approveActionPlan(planId, photoDrafts);
        } catch (error) {
          setRealtime(buildFailedVoiceRealtimeState(error, voiceFailureContext(realtime, previewState)));
          setStage('failed');
        }
      },
      cancelRealtimeActionPlan: async (planId: string) => {
        setRealtime((current) => markReviewDecisionPending(current, 'Cancelling change'));
        try {
          await realtimeController.cancelActionPlan(planId);
        } catch (error) {
          setRealtime(buildFailedVoiceRealtimeState(error, voiceFailureContext(realtime, previewState)));
          setStage('failed');
        }
      },
      retryRealtimeActionPlanPhotos: async (planId: string) => {
        setRealtime((current) => markPhotoRetryInProgress(current));
        try {
          const photoAttachmentStatus = await realtimeController.retryPhotoAttachments(planId);
          setRealtime((current) => markPhotoRetryResult(current, photoAttachmentStatus));
        } catch (error) {
          setRealtime((current) => markPhotoRetryFailure(current, error));
        }
      },
      cancelRealtime: async () => {
        const generation = sessionGeneration.current + 1;
        sessionGeneration.current = generation;
        const cancelled = await realtimeController.cancel();
        if (sessionGeneration.current !== generation) {
          return;
        }
        setRealtime(cancelled);
        setStage('cancelled');
      },
      reset: () => {
        sessionGeneration.current++;
        setRealtime(null);
        setStage('ready');
      }
    };
  }, [diagnosticsEnabled, previewState, realtime, realtimeController, stage]);

  return (
    <VoiceInteractionStateContext.Provider value={value}>
      {children}
    </VoiceInteractionStateContext.Provider>
  );
}

export function markReviewDecisionPending(state: VoiceRealtimeState | null, progressLabel: string): VoiceRealtimeState | null {
  if (!state?.actionPlan || state.actionPlan.status !== 'proposed' || state.reviewDecisionPending) {
    return state;
  }

  return {
    ...state,
    progressLabel,
    reviewDecisionPending: true
  };
}

export function markPhotoRetryInProgress(state: VoiceRealtimeState | null): VoiceRealtimeState | null {
  return state ? { ...state, progressLabel: 'Adding photos' } : state;
}

export function markPhotoRetryResult(
  state: VoiceRealtimeState | null,
  photoAttachmentStatus: VoicePhotoAttachmentStatus
): VoiceRealtimeState | null {
  return state ? {
    ...state,
    progressLabel: photoAttachmentStatus.status === 'attached' ? 'Photos updated' : 'Photo upload failed',
    photoAttachmentStatus
  } : state;
}

export function markPhotoRetryFailure(state: VoiceRealtimeState | null, error: unknown): VoiceRealtimeState {
  return state ? {
    ...state,
    progressLabel: 'Photo upload failed',
    photoAttachmentStatus: {
      status: 'failed',
      message: 'Photos could not be attached. Try again.',
      canRetry: true
    }
  } : buildFailedVoiceRealtimeState(error, realtimeStateContext(state));
}

export function useVoiceInteractionState(): VoiceInteractionStateContextValue {
  const value = useContext(VoiceInteractionStateContext);

  if (value === null) {
    throw new Error('Voice interaction state is not available.');
  }

  return value;
}

export function applyRecordingLevelToRealtime(
  current: VoiceRealtimeState | null,
  recordingLevel: number
): VoiceRealtimeState | null {
  return current?.status === 'listening'
    ? { ...current, recordingLevel }
    : current;
}

export function refreshClarificationFollowUpAvailability(
  current: VoiceRealtimeState | null,
  canSendFollowUpAudio: boolean
): VoiceRealtimeState | null {
  if (
    current?.status !== 'completed' ||
    current.responseKind !== 'clarification' ||
    current.clarificationFollowUpAvailable !== true ||
    canSendFollowUpAudio
  ) {
    return current;
  }

  return {
    ...current,
    clarificationFollowUpAvailable: false
  };
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

function isVoiceCancelledError(error: unknown): boolean {
  return error instanceof VoiceRealtimeCancelledError ||
    (isObject(error) && error.code === 'voice_cancelled');
}

export function buildFailedVoiceRealtimeState(error: unknown, context: VoiceFailureContext = {}): VoiceRealtimeState {
  const readinessFailure = providerReadinessFailure(error);
  const failureCode: VoiceRealtimeFailureCode = readinessFailure
    ? 'provider_readiness'
    : 'voice_failed';

  return {
    status: 'failed',
    tenantName: safeContextLabel(context.tenantName),
    inventoryName: safeContextLabel(context.inventoryName),
    progressLabel: 'Voice failed',
    debugEvents: [],
    failureCode,
    errorMessage: readinessFailure?.message ?? 'Voice failed safely.'
  };
}

function voiceFailureContext(
  realtime: VoiceRealtimeState | null,
  previewState: { readonly status: 'loading' } | { readonly status: 'error'; readonly message: string } | { readonly status: 'ready'; readonly preview: VoiceInteractionPreviewViewModel }
): VoiceFailureContext {
  const realtimeContext = realtimeStateContext(realtime);
  if (realtimeContext.tenantName || realtimeContext.inventoryName) {
    return realtimeContext;
  }
  return previewState.status === 'ready'
    ? { tenantName: previewState.preview.tenantName, inventoryName: previewState.preview.inventoryName }
    : {};
}

function realtimeStateContext(state: VoiceRealtimeState | null): VoiceFailureContext {
  return state
    ? { tenantName: state.tenantName, inventoryName: state.inventoryName }
    : {};
}

function safeContextLabel(value: string | undefined): string {
  return redactUnsafeVoiceStructuredText(value ?? '').replace(/\s+/g, ' ').trim().slice(0, 120);
}

function providerReadinessFailure(error: unknown): { readonly message: string } | null {
  if (!isObject(error) || error.code !== 'provider_readiness') {
    return null;
  }

  const missingCapabilities = Array.isArray(error.missingCapabilities)
    ? error.missingCapabilities.filter(isVoiceProviderCapability)
    : [];

  return {
    message: missingCapabilities.length > 0
      ? `Voice provider profiles are not ready: ${missingCapabilities.join(', ')}.`
      : 'Voice provider profiles are not ready.'
  };
}

function isObject(value: unknown): value is {
  readonly code?: unknown;
  readonly missingCapabilities?: unknown;
} {
  return typeof value === 'object' && value !== null;
}

function isVoiceProviderCapability(value: unknown): value is string {
  return value === 'speech_to_text' || value === 'language_inference' || value === 'text_to_speech';
}
