import { describe, expect, it } from 'vitest';
import { VoiceProviderReadinessError } from '../../application/providerProfiles/ProviderProfileVoiceReadinessCheck';
import {
  applyRecordingLevelToRealtime,
  buildFailedVoiceRealtimeState,
  refreshClarificationFollowUpAvailability
} from './VoiceInteractionStateContext';

describe('buildFailedVoiceRealtimeState', () => {
  it('keeps provider-readiness failures typed and safely summarized', () => {
    const state = buildFailedVoiceRealtimeState(
      new VoiceProviderReadinessError(['speech_to_text', 'text_to_speech'])
    );

    expect(state).toMatchObject({
      status: 'failed',
      progressLabel: 'Voice failed',
      failureCode: 'provider_readiness',
      errorMessage: 'Voice provider profiles are not ready: speech_to_text, text_to_speech.'
    });
  });

  it('does not display raw generic voice failure details', () => {
    const state = buildFailedVoiceRealtimeState(
      new Error('raw provider transport failure with endpoint and bearer token')
    );

    expect(state).toMatchObject({
      status: 'failed',
      failureCode: 'voice_failed',
      errorMessage: 'Voice failed safely.'
    });
  });

  it('redacts unrecognized provider-readiness capability values', () => {
    const state = buildFailedVoiceRealtimeState({
      code: 'provider_readiness',
      missingCapabilities: ['text_to_speech', 'secret_endpoint']
    });

    expect(state.errorMessage).toBe('Voice provider profiles are not ready: text_to_speech.');
  });

  it('applies live recorder levels only to active listening state', () => {
    expect(applyRecordingLevelToRealtime({
      status: 'listening',
      tenantName: 'Home tenant',
      inventoryName: 'Home',
      progressLabel: 'Listening',
      debugEvents: []
    }, 0.42)).toMatchObject({
      status: 'listening',
      recordingLevel: 0.42
    });

    expect(applyRecordingLevelToRealtime({
      status: 'processing',
      tenantName: 'Home tenant',
      inventoryName: 'Home',
      progressLabel: 'Sending audio',
      debugEvents: []
    }, 0.8)).not.toHaveProperty('recordingLevel');

    expect(applyRecordingLevelToRealtime(null, 0.8)).toBeNull();
  });

  it('removes completed clarification follow-up availability when the live transport is gone', () => {
    expect(refreshClarificationFollowUpAvailability({
      status: 'completed',
      tenantName: 'Home tenant',
      inventoryName: 'Home',
      progressLabel: 'Needs detail',
      responseKind: 'clarification',
      clarificationFollowUpAvailable: true,
      debugEvents: []
    }, false)).toMatchObject({
      status: 'completed',
      responseKind: 'clarification',
      clarificationFollowUpAvailable: false
    });

    const answerState = {
      status: 'completed' as const,
      tenantName: 'Home tenant',
      inventoryName: 'Home',
      progressLabel: 'Done',
      responseKind: 'answer',
      debugEvents: []
    };
    expect(refreshClarificationFollowUpAvailability(answerState, false)).toBe(answerState);
  });
});
