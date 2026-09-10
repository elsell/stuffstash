import { retainFailedConversation } from './VoiceConversationFailure';
import { appendConversationExchange, canCancelConversation, canSubmitConversation } from './VoiceConversationHistory';
import type { VoicePlanPhotoDrafts } from '../screens/VoicePlanPhotoDraftState';
import type { VoicePlanCommandDrafts } from '../screens/VoicePlanEdits';
import type { Dispatch, SetStateAction } from 'react';
import { useMobileInventoryServerQuery } from '../serverState/useMobileInventoryServerQuery';
import { mobileQueryKeys } from '../../adapters/serverState/MobileQueryClient';
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
  type VoiceActionPlanCommandEdit,
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

type TitleEditor = { readonly commandId: string; readonly value: string } | null;
type ConversationDraftState = { readonly planId?: string; readonly drafts: VoicePlanCommandDrafts };
type VoiceInteractionStateContextValue = {
  readonly titleEditor: TitleEditor;
  readonly setTitleEditor: Dispatch<SetStateAction<TitleEditor>>;
  readonly history: readonly VoiceRealtimeState[];
  readonly composerText: string;
  readonly setComposerText: (text: string) => void;
  readonly photoDrafts: VoicePlanPhotoDrafts;
  readonly setPhotoDrafts: Dispatch<SetStateAction<VoicePlanPhotoDrafts>>;
  readonly commandDraftState: ConversationDraftState;
  readonly setCommandDraftState: Dispatch<SetStateAction<ConversationDraftState>>;
  readonly sendText: () => Promise<void>;
  readonly pauseMedia: () => Promise<void>;
  readonly scrollOffset: React.MutableRefObject<number>;
  readonly railOffsets: React.MutableRefObject<Record<string, number>>;
  readonly diagnosticsEnabled: boolean;
  readonly state: VoiceInteractionState;
  readonly setStage: (stage: VoiceInteractionStage) => void;
  readonly startRealtime: () => Promise<void>;
  readonly stopRealtime: () => Promise<void>;
  readonly approveRealtimeActionPlan: (planId: string, photoDrafts?: VoiceActionPlanPhotoDrafts, edits?: readonly VoiceActionPlanCommandEdit[]) => Promise<void>;
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

export function VoiceInteractionStateProvider(props: VoiceInteractionStateProviderProps) {
  const preview = useMobileInventoryServerQuery({ key: mobileQueryKeys.voiceContext, query: signal => props.previewQuery.execute({ signal }) });
  const previewState: PreviewState = preview.data ? { status: 'ready', preview: preview.data } : preview.isError ? { status: 'error', message: readableError(preview.error, 'Voice preview is not available.') } : { status: 'loading' };
  return <ScopedVoiceInteractionStateProvider scopeKey={JSON.stringify(preview.resourceKey)} {...props} previewState={previewState} />;
}

type PreviewState = { readonly status: 'loading' } | { readonly status: 'error'; readonly message: string } | { readonly status: 'ready'; readonly preview: VoiceInteractionPreviewViewModel };

function ScopedVoiceInteractionStateProvider({ children, diagnosticsEnabled = false, realtimeController, previewState, scopeKey }: VoiceInteractionStateProviderProps & { readonly previewState: PreviewState; readonly scopeKey: string }) {
  const [titleEditor, setTitleEditor] = useState<TitleEditor>(null);
  const [history, setHistory] = useState<readonly VoiceRealtimeState[]>([]);
  const [composerText, setComposerText] = useState('');
  const [photoDrafts, setPhotoDrafts] = useState<VoicePlanPhotoDrafts>({});
  const [commandDraftState, setCommandDraftState] = useState<ConversationDraftState>({ drafts: {} });
  const scrollOffset = useRef(0);
  const railOffsets = useRef<Record<string, number>>({});
  const requestPending = useRef(false);
  const interactionLifetime = useRef(0);
  const photoRetries = useRef(new Set<string>());
  const [stage, setStage] = useState<VoiceInteractionStage>('ready');
  const [realtime, setRealtime] = useState<VoiceRealtimeState | null>(null);
  const sessionGeneration = useRef(0);
  const [stateOwner, setStateOwner] = useState(scopeKey);
  useEffect(() => {
    setStateOwner(scopeKey);
    setTitleEditor(null); setHistory([]); setComposerText(''); setPhotoDrafts({}); setCommandDraftState({ drafts: {} });
    scrollOffset.current = 0; railOffsets.current = {}; requestPending.current = false;
    setStage('ready');
    setRealtime(null);
    return () => { interactionLifetime.current++; photoRetries.current.clear(); sessionGeneration.current++; void realtimeController.dispose(); };
  }, [realtimeController, scopeKey]);

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
      (realtime?.responseKind !== 'clarification' && realtime?.responseKind !== 'answer') ||
      realtime.followUpAvailable !== true
    ) {
      return;
    }

    const interval = setInterval(() => {
      setRealtime((current) => refreshVoiceFollowUpAvailability(
        current,
        realtimeController.canSendFollowUpAudio()
      ));
    }, 500);

    return () => {
      clearInterval(interval);
    };
  }, [realtime?.followUpAvailable, realtime?.responseKind, realtimeController, stage]);

  const value = useMemo<VoiceInteractionStateContextValue>(() => {
    const state: VoiceInteractionState =
      previewState.status === 'ready'
        ? { status: 'ready', stage: stateOwner === scopeKey ? stage : 'ready', preview: previewState.preview, realtime: stateOwner === scopeKey ? realtime : null }
        : previewState.status === 'error'
          ? { status: 'error', stage, message: previewState.message }
          : { status: 'loading', stage };

    return {
      titleEditor, setTitleEditor,
      history: stateOwner === scopeKey ? history : [], composerText: stateOwner === scopeKey ? composerText : '', setComposerText,
      photoDrafts, setPhotoDrafts, commandDraftState, setCommandDraftState, scrollOffset, railOffsets,
      pauseMedia: async () => {
        if (stage === 'listening' || (canSubmitConversation(stage) && requestPending.current)) {
          sessionGeneration.current++;
          requestPending.current = false;
          setStage('ready'); setRealtime(null);
        }
        await realtimeController.pauseMedia();
      },
      sendText: async () => {
        if (!canSubmitConversation(stage) || requestPending.current || !composerText.trim()) return;
        requestPending.current = true;
        const text = composerText;
        const generation = ++sessionGeneration.current;
        setHistory(current => appendConversationExchange(current, realtime, commandDraftState.drafts));
        setComposerText(''); setRealtime(null); setStage('processing');
        let inputAccepted = false;
        try {
          await realtimeController.sendText(text, next => {
            inputAccepted ||= next.inputAccepted === true || !!next.actionPlan || !!next.spokenResponse;
            if (sessionGeneration.current !== generation) return;
            setRealtime(next); setStage(next.status);
          });
        } catch (error) {
          if (sessionGeneration.current === generation) {
            if (!inputAccepted) setComposerText(text);
            setRealtime(current => retainFailedConversation(current, buildFailedVoiceRealtimeState(error, voiceFailureContext(current, previewState)))); setStage('failed');
          }
        } finally { if (sessionGeneration.current === generation) requestPending.current = false; }
      },
      diagnosticsEnabled,
      state,
      setStage,
      startRealtime: async () => {
        if (!canSubmitConversation(stage) || requestPending.current) return;
        requestPending.current = true;
        setHistory(current => appendConversationExchange(current, realtime, commandDraftState.drafts));
        setRealtime(null);
        const generation = sessionGeneration.current + 1;
        sessionGeneration.current = generation;
        try {
          const next = realtime?.status === 'completed' &&
            (realtime.responseKind === 'clarification' || realtime.responseKind === 'answer') &&
            realtime.followUpAvailable === true &&
            realtimeController.canSendFollowUpAudio()
            ? await realtimeController.startFollowUp()
            : await realtimeController.start();
          if (sessionGeneration.current !== generation) {
            return;
          }
          setRealtime(next);
          setStage('listening');
          requestPending.current = false;
        } catch (error) {
          if (sessionGeneration.current !== generation) {
            return;
          }
          setRealtime(current => retainFailedConversation(current, buildFailedVoiceRealtimeState(error, voiceFailureContext(current, previewState))));
          setStage('failed');
          requestPending.current = false;
        }
      },
      stopRealtime: async () => {
        if (stage !== 'listening' || requestPending.current) return;
        requestPending.current = true;
        const generation = sessionGeneration.current;
        setStage('processing');
        try {
          const shouldSendFollowUp = (realtime?.responseKind === 'clarification' || realtime?.responseKind === 'answer') &&
            realtime.followUpAvailable === true &&
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
          requestPending.current = false;
        } catch (error) {
          if (sessionGeneration.current !== generation) {
            return;
          }
          if (isVoiceCancelledError(error)) {
            requestPending.current = false;
        const cancelled = await realtimeController.cancel();
            if (sessionGeneration.current !== generation) {
              return;
            }
            setRealtime(cancelled);
            setStage('cancelled');
            return;
          }
          setRealtime(current => retainFailedConversation(current, buildFailedVoiceRealtimeState(error, voiceFailureContext(current, previewState))));
          setStage('failed');
          requestPending.current = false;
        }
      },
      approveRealtimeActionPlan: async (planId: string, photoDrafts?: VoiceActionPlanPhotoDrafts, edits?: readonly VoiceActionPlanCommandEdit[]) => {
        const lifetime = interactionLifetime.current;
        setRealtime((current) => markReviewDecisionPending(current, 'Approving change'));
        try {
          await realtimeController.approveActionPlan(planId, photoDrafts, edits);
        } catch (error) {
          if (interactionLifetime.current !== lifetime) return;
          if (isObject(error) && error.code === 'review_validation_failed') {
            setRealtime(current => current ? { ...current, status: 'review', reviewDecisionPending: false, progressLabel: 'Check review details', errorMessage: 'Check the staged photos and edited fields, then approve again.' } : current);
            setStage('review');
            return;
          }
          setRealtime(current => retainFailedConversation(current, buildFailedVoiceRealtimeState(error, voiceFailureContext(current, previewState))));
          setStage('failed');
          requestPending.current = false;
        }
      },
      cancelRealtimeActionPlan: async (planId: string) => {
        const lifetime = interactionLifetime.current;
        setRealtime((current) => markReviewDecisionPending(current, 'Cancelling change'));
        try {
          await realtimeController.cancelActionPlan(planId);
        } catch (error) {
          if (interactionLifetime.current !== lifetime) return;
          setRealtime(current => retainFailedConversation(current, buildFailedVoiceRealtimeState(error, voiceFailureContext(current, previewState))));
          setStage('failed');
          requestPending.current = false;
        }
      },
      retryRealtimeActionPlanPhotos: async (planId: string) => {
        if (photoRetries.current.has(planId)) return;
        photoRetries.current.add(planId);
        const generation = interactionLifetime.current;
        setRealtime((current) => markPhotoRetryInProgress(current, planId));
        setHistory(current => current.map(exchange => markPhotoRetryInProgress(exchange, planId)!));
        try {
          const photoAttachmentStatus = await realtimeController.retryPhotoAttachments(planId, progress => {
            if (interactionLifetime.current !== generation) return;
            setRealtime(current => markPhotoRetryResult(current, planId, progress));
            setHistory(current => current.map(exchange => markPhotoRetryResult(exchange, planId, progress)!));
          });
          if (interactionLifetime.current !== generation) return;
          setRealtime((current) => markPhotoRetryResult(current, planId, photoAttachmentStatus));
          setHistory(current => current.map(exchange => markPhotoRetryResult(exchange, planId, photoAttachmentStatus)!));
        } catch {
          if (interactionLifetime.current !== generation) return;
          setRealtime((current) => markPhotoRetryFailure(current, planId));
          setHistory(current => current.map(exchange => markPhotoRetryFailure(exchange, planId)!));
        } finally {
          if (interactionLifetime.current === generation) photoRetries.current.delete(planId);
        }
      },
      cancelRealtime: async () => {
        if (!canCancelConversation(stage, realtime)) return;
        const generation = sessionGeneration.current + 1;
        sessionGeneration.current = generation;
        requestPending.current = false;
        const cancelled = await realtimeController.cancel();
        if (sessionGeneration.current !== generation) {
          return;
        }
        setRealtime(cancelled);
        setStage('cancelled');
      },
      reset: () => {
        interactionLifetime.current++; photoRetries.current.clear();
        sessionGeneration.current++;
        void realtimeController.dispose();
        requestPending.current = false;
        setTitleEditor(null); setHistory([]); setComposerText(''); setPhotoDrafts({}); setCommandDraftState({ drafts: {} });
        scrollOffset.current = 0; railOffsets.current = {};
        setRealtime(null);
        setStage('ready');
      }
    };
  }, [titleEditor, history, composerText, photoDrafts, commandDraftState, diagnosticsEnabled, previewState, realtime, realtimeController, stage, stateOwner, scopeKey]);

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

export function markPhotoRetryInProgress(state: VoiceRealtimeState | null, planId: string): VoiceRealtimeState | null {
  return voiceStateMatchesActionPlan(state, planId) ? { ...state, progressLabel: 'Adding photos', ...(state.photoAttachmentStatus ? { photoAttachmentStatus: { ...state.photoAttachmentStatus, message: 'Adding photos…', canRetry: false } } : {}) } : state;
}

export function markPhotoRetryResult(
  state: VoiceRealtimeState | null,
  planId: string,
  photoAttachmentStatus: VoicePhotoAttachmentStatus
): VoiceRealtimeState | null {
  return voiceStateMatchesActionPlan(state, planId) ? {
    ...state,
    progressLabel: photoAttachmentStatus.status === 'uploading' ? 'Adding photos' : photoAttachmentStatus.status === 'attached' ? 'Photos updated' : 'Photo upload failed',
    photoAttachmentStatus
  } : state;
}

export function markPhotoRetryFailure(state: VoiceRealtimeState | null, planId: string): VoiceRealtimeState | null {
  return voiceStateMatchesActionPlan(state, planId) ? {
    ...state,
    progressLabel: 'Photo upload failed',
    photoAttachmentStatus: {
      status: 'failed',
      message: 'Photos could not be attached. Try again.',
      canRetry: true
    }
  } : state;
}

function voiceStateMatchesActionPlan(state: VoiceRealtimeState | null, planId: string): state is VoiceRealtimeState {
  return Boolean(state?.actionPlan && state.actionPlan.planId === planId);
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

export function refreshVoiceFollowUpAvailability(
  current: VoiceRealtimeState | null,
  canSendFollowUpAudio: boolean
): VoiceRealtimeState | null {
  if (
    current?.status !== 'completed' ||
    (current.responseKind !== 'clarification' && current.responseKind !== 'answer') ||
    current.followUpAvailable !== true ||
    canSendFollowUpAudio
  ) {
    return current;
  }

  return {
    ...current,
    followUpAvailable: false
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
    errorMessage: readinessFailure?.message ?? (isObject(error) && error.code === 'connection_interrupted' ? 'The connection was interrupted. Try again when you are connected.' : 'Could not finish this request. Try again or start a new conversation.')
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
