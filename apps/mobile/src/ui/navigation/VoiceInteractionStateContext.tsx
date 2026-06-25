import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from 'react';
import {
  VoiceInteractionPreviewQuery,
  VoiceInteractionPreviewViewModel
} from '../../application/voice/VoiceInteractionPreviewQuery';
import {
  RealtimeVoiceSessionController,
  VoiceRealtimeState
} from '../../application/voice/RealtimeVoiceSession';

export type VoiceInteractionStage = 'ready' | 'listening' | 'review' | 'processing' | 'speaking' | 'completed' | 'failed';

type VoiceInteractionState =
  | { readonly status: 'loading'; readonly stage: VoiceInteractionStage }
  | { readonly status: 'error'; readonly stage: VoiceInteractionStage; readonly message: string }
  | {
      readonly status: 'ready';
      readonly stage: VoiceInteractionStage;
      readonly preview: VoiceInteractionPreviewViewModel;
      readonly realtime: VoiceRealtimeState | null;
    };

type VoiceInteractionStateContextValue = {
  readonly state: VoiceInteractionState;
  readonly setStage: (stage: VoiceInteractionStage) => void;
  readonly startRealtime: () => Promise<void>;
  readonly stopRealtime: () => Promise<void>;
  readonly reset: () => void;
};

const VoiceInteractionStateContext = createContext<VoiceInteractionStateContextValue | null>(null);

type VoiceInteractionStateProviderProps = {
  readonly children: ReactNode;
  readonly previewQuery: VoiceInteractionPreviewQuery;
  readonly realtimeController: RealtimeVoiceSessionController;
};

export function VoiceInteractionStateProvider({
  children,
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
      state,
      setStage,
      startRealtime: async () => {
        const next = await realtimeController.start();
        setRealtime(next);
        setStage('listening');
      },
      stopRealtime: async () => {
        setStage('processing');
        const states = await realtimeController.stop();
        const finalState = states[states.length - 1] ?? null;
        setRealtime(finalState);
        setStage(finalState?.status ?? 'failed');
      },
      reset: () => {
        setRealtime(null);
        setStage('ready');
      }
    };
  }, [previewState, realtime, realtimeController, stage]);

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
