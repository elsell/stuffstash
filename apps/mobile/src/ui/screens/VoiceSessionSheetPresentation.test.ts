import { describe, expect, it } from 'vitest';
import { buildVoiceSessionSheetBodyPresentation } from './VoiceSessionSheetPresentation';
import type { VoiceInteractionState } from '../navigation/VoiceInteractionStateContext';

describe('VoiceSessionSheetPresentation', () => {
  it('keeps the ready state body empty so the bottom action area owns the interaction', () => {
    expect(buildVoiceSessionSheetBodyPresentation(readyVoiceState(null), {}, false)).toEqual({
      hasBodyContent: false
    });
  });

  it('uses the body for transcript, response, progress, errors, and diagnostics', () => {
    expect(
      buildVoiceSessionSheetBodyPresentation(readyVoiceState(null), {
        transcript: 'Where is my water bottle?'
      }, false).hasBodyContent
    ).toBe(true);
    expect(
      buildVoiceSessionSheetBodyPresentation(readyVoiceState(null), {
        response: 'Your water bottle is in the Office.'
      }, false).hasBodyContent
    ).toBe(true);
    expect(
      buildVoiceSessionSheetBodyPresentation(readyVoiceState(null), {
        progressSteps: ['Sending audio']
      }, false).hasBodyContent
    ).toBe(true);
    expect(
      buildVoiceSessionSheetBodyPresentation(
        readyVoiceState({
          status: 'failed',
          tenantName: 'Home tenant',
          inventoryName: 'Home',
          progressLabel: 'Voice failed',
          debugEvents: [],
          errorMessage: 'Voice is not configured.'
        }),
        {},
        false
      ).hasBodyContent
    ).toBe(true);
  });

  it('requires explicit diagnostics enablement before diagnostics occupy body content', () => {
    expect(
      buildVoiceSessionSheetBodyPresentation(
        readyVoiceState({
          status: 'completed',
          tenantName: 'Home tenant',
          inventoryName: 'Home',
          progressLabel: 'Done',
          debugEvents: [{ label: 'Inventory search', status: 'Completed' }]
        }),
        {},
        false
      ).hasBodyContent
    ).toBe(false);
    expect(
      buildVoiceSessionSheetBodyPresentation(
        readyVoiceState({
          status: 'completed',
          tenantName: 'Home tenant',
          inventoryName: 'Home',
          progressLabel: 'Done',
          debugEvents: [{ label: 'Inventory search', status: 'Completed' }]
        }),
        {},
        true
      ).hasBodyContent
    ).toBe(true);
  });
});

function readyVoiceState(realtime: Extract<VoiceInteractionState, { readonly status: 'ready' }>['realtime']): VoiceInteractionState {
  return {
    status: 'ready',
    stage: realtime?.status ?? 'ready',
    preview: {
      tenantName: 'Home tenant',
      inventoryName: 'Home',
      sampleUtterance: 'Where is my water bottle?',
      assistantSummary: 'Ready',
      actionPreview: {
        summary: 'No action',
        steps: [],
        riskLabel: 'Read only'
      }
    },
    realtime
  };
}
