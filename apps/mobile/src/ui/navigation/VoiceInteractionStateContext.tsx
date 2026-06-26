import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from 'react';
import {
  VoiceInteractionPreviewQuery,
  VoiceInteractionPreviewViewModel
} from '../../application/voice/VoiceInteractionPreviewQuery';
import {
  RealtimeVoiceSessionController,
  VoiceRealtimeFailureCode,
  VoiceRealtimeState
} from '../../application/voice/RealtimeVoiceSession';

export type VoiceInteractionStage = 'ready' | 'listening' | 'review' | 'processing' | 'speaking' | 'completed' | 'failed';

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
  readonly reset: () => void;
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
        try {
          const next = await realtimeController.start();
          setRealtime(next);
          setStage('listening');
        } catch (error) {
          setRealtime(buildFailedVoiceRealtimeState(error));
          setStage('failed');
        }
      },
      stopRealtime: async () => {
        setStage('processing');
        try {
          const states = await realtimeController.stop((nextState) => {
            setRealtime(nextState);
            setStage(nextState.status);
          });
          const finalState = states[states.length - 1] ?? null;
          setRealtime(finalState);
          setStage(finalState?.status ?? 'failed');
        } catch (error) {
          setRealtime(buildFailedVoiceRealtimeState(error));
          setStage('failed');
        }
      },
      reset: () => {
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

export function useVoiceInteractionState(): VoiceInteractionStateContextValue {
  const value = useContext(VoiceInteractionStateContext);

  if (value === null) {
    throw new Error('Voice interaction state is not available.');
  }

  return value;
}

function readableError(error: unknown, fallback: string): string {
  return error instanceof Error ? error.message : fallback;
}

export function buildFailedVoiceRealtimeState(error: unknown): VoiceRealtimeState {
  const readinessFailure = providerReadinessFailure(error);
  const failureCode: VoiceRealtimeFailureCode = readinessFailure
    ? 'provider_readiness'
    : 'voice_failed';

  return {
    status: 'failed',
    tenantName: '',
    inventoryName: '',
    progressLabel: 'Voice failed',
    debugEvents: [],
    failureCode,
    errorMessage: readinessFailure?.message ?? 'Voice failed safely.'
  };
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
