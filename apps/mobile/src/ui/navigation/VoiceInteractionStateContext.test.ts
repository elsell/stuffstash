import { describe, expect, it } from 'vitest';
import { VoiceProviderReadinessError } from '../../application/providerProfiles/ProviderProfileVoiceReadinessCheck';
import {
  applyRecordingLevelToRealtime,
  buildFailedVoiceRealtimeState,
  markReviewDecisionPending,
  markPhotoRetryFailure,
  markPhotoRetryInProgress,
  markPhotoRetryResult,
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

  it('keeps known inventory context on locally built voice failures', () => {
    const state = buildFailedVoiceRealtimeState(
      new VoiceProviderReadinessError(['language_inference']),
      { tenantName: 'Main tenant', inventoryName: 'Home inventory' }
    );

    expect(state).toMatchObject({
      status: 'failed',
      tenantName: 'Main tenant',
      inventoryName: 'Home inventory',
      failureCode: 'provider_readiness',
      errorMessage: 'Voice provider profiles are not ready: language_inference.'
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

  it('marks only active proposed review plans as decision pending', () => {
    expect(markReviewDecisionPending({
      status: 'review',
      tenantName: 'Home tenant',
      inventoryName: 'Home',
      progressLabel: 'Review needed',
      debugEvents: [],
      actionPlan: {
        planId: 'plan-1',
        status: 'proposed',
        confirmationSummary: 'Create item water bottle?',
        commands: [{ kind: 'create_asset', summary: 'Create item water bottle' }],
        risks: []
      }
    }, 'Approving change')).toMatchObject({
      status: 'review',
      progressLabel: 'Approving change',
      reviewDecisionPending: true
    });

    const executed = {
      status: 'completed' as const,
      tenantName: 'Home tenant',
      inventoryName: 'Home',
      progressLabel: 'Change applied',
      debugEvents: [],
      actionPlan: {
        planId: 'plan-1',
        status: 'executed' as const,
        confirmationSummary: 'Create item water bottle?',
        commands: [{ kind: 'create_asset', summary: 'Create item water bottle' }],
        risks: []
      }
    };
    expect(markReviewDecisionPending(executed, 'Approving change')).toBe(executed);
  });

  it('marks voice photo retry progress without changing the session outcome', () => {
    expect(markPhotoRetryInProgress(completedVoiceState(), 'plan-1')).toMatchObject({
      status: 'completed',
      progressLabel: 'Adding photos'
    });
    const newerState = completedVoiceState('plan-2');
    expect(markPhotoRetryInProgress(newerState, 'plan-1')).toBe(newerState);
    expect(markPhotoRetryInProgress(null, 'plan-1')).toBeNull();
  });

  it('maps voice photo retry results onto the existing session state', () => {
    expect(markPhotoRetryResult(completedVoiceState(), 'plan-1', {
      status: 'attached',
      message: 'Photos attached.'
    })).toMatchObject({
      status: 'completed',
      progressLabel: 'Photos updated',
      photoAttachmentStatus: {
        status: 'attached',
        message: 'Photos attached.'
      }
    });

    expect(markPhotoRetryResult(completedVoiceState(), 'plan-1', {
      status: 'partial_failed',
      message: 'One photo failed.',
      canRetry: true
    })).toMatchObject({
      status: 'completed',
      progressLabel: 'Photo upload failed',
      photoAttachmentStatus: {
        status: 'partial_failed',
        message: 'One photo failed.',
        canRetry: true
      }
    });

    const newerState = completedVoiceState('plan-2');
    expect(markPhotoRetryResult(newerState, 'plan-1', {
      status: 'attached',
      message: 'Photos attached.'
    })).toBe(newerState);
  });

  it('keeps inventory action success visible when a voice photo retry throws', () => {
    expect(markPhotoRetryFailure(completedVoiceState(), 'plan-1')).toMatchObject({
      status: 'completed',
      progressLabel: 'Photo upload failed',
      photoAttachmentStatus: {
        status: 'failed',
        message: 'Photos could not be attached. Try again.',
        canRetry: true
      }
    });

    const newerState = completedVoiceState('plan-2');
    expect(markPhotoRetryFailure(newerState, 'plan-1')).toBe(newerState);
    expect(markPhotoRetryFailure(null, 'plan-1')).toBeNull();
  });
});

function completedVoiceState(planId = 'plan-1') {
  return {
    status: 'completed' as const,
    tenantName: 'Home tenant',
    inventoryName: 'Home',
    progressLabel: 'Photo upload failed',
    spokenResponse: 'The approved change was applied.',
    debugEvents: [],
    photoAttachmentStatus: {
      status: 'failed' as const,
      message: 'Photos could not be attached.',
      canRetry: true
    },
    actionPlan: {
      planId,
      status: 'executed' as const,
      confirmationSummary: 'Create item water bottle?',
      commands: [{ kind: 'create_asset', summary: 'Create item water bottle' }],
      risks: []
    }
  };
}
