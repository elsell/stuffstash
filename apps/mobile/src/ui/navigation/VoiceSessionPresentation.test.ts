import { describe, expect, it } from 'vitest';
import {
  buildVoiceAccessoryPresentation,
  buildVoiceSessionPresentation
} from './VoiceSessionPresentation';

describe('VoiceSessionPresentation', () => {
  it('starts listening from the collapsed ready bubble without navigating away', () => {
    expect(buildVoiceAccessoryPresentation({ pathname: '/', stage: 'ready', status: 'ready' })).toMatchObject({
      accessibilityLabel: 'Start voice interaction',
      primaryAction: 'start',
      subtitle: 'Current inventory'
    });
  });

  it('stops recording from the collapsed listening bubble', () => {
    expect(buildVoiceAccessoryPresentation({ pathname: '/locations/location-1', stage: 'listening', status: 'ready' })).toMatchObject({
      accessibilityLabel: 'Send voice request',
      primaryAction: 'stop',
      subtitle: 'Location context'
    });
  });

  it('expands the session surface for in-progress and completed sessions', () => {
    expect(buildVoiceAccessoryPresentation({ pathname: '/assets/asset-1', stage: 'processing', status: 'ready' }).primaryAction).toBe('expand');
    expect(buildVoiceAccessoryPresentation({ pathname: '/assets/asset-1', stage: 'completed', status: 'ready' }).primaryAction).toBe('expand');
  });

  it('keeps post-send collapsed states in the working tone', () => {
    expect(buildVoiceAccessoryPresentation({ pathname: '/', stage: 'processing', status: 'ready' })).toMatchObject({
      title: 'Checking inventory',
      tone: 'attention'
    });
    expect(buildVoiceAccessoryPresentation({ pathname: '/', stage: 'speaking', status: 'ready' })).toMatchObject({
      title: 'Speaking',
      tone: 'attention'
    });
  });

  it('summarizes active realtime progress in the collapsed accessory', () => {
    expect(buildVoiceAccessoryPresentation({
      pathname: '/',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        partialTranscript: 'Where is my secret',
        progressLabel: 'Searching visible inventory',
        debugEvents: []
      },
      stage: 'processing',
      status: 'ready'
    })).toMatchObject({
      title: 'Checking inventory',
      subtitle: 'Current inventory',
      tone: 'attention'
    });
  });

  it('uses safe graph-loop phase labels in the collapsed accessory', () => {
    expect(buildVoiceAccessoryPresentation({
      pathname: '/',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        conversationPhase: 'planning',
        progressLabel: 'Preparing a safe plan.',
        debugEvents: []
      },
      stage: 'processing',
      status: 'ready'
    })).toMatchObject({
      title: 'Preparing plan',
      subtitle: 'Current inventory',
      tone: 'attention'
    });

    expect(buildVoiceAccessoryPresentation({
      pathname: '/',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        conversationPhase: 'recovering',
        progressLabel: 'Recovering safely.',
        debugEvents: []
      },
      stage: 'processing',
      status: 'ready'
    })).toMatchObject({
      title: 'Recovering safely'
    });
  });

  it('summarizes terminal realtime answers and failures in the collapsed accessory', () => {
    expect(buildVoiceAccessoryPresentation({
      pathname: '/locations/location-1',
      realtime: {
        status: 'completed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Done',
        spokenResponse: 'Your water bottle is in the Office.',
        debugEvents: []
      },
      stage: 'completed',
      status: 'ready'
    })).toMatchObject({
      title: 'Answer ready',
      subtitle: 'Your water bottle is in the Office.'
    });

    const unsafeAnswer = buildVoiceAccessoryPresentation({
      pathname: '/locations/location-1',
      realtime: {
        status: 'completed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Done',
        spokenResponse: 'Your water bottle is in the Office. {"assetId":"water-bottle-1"} bearer secret-token stack trace',
        debugEvents: []
      },
      stage: 'completed',
      status: 'ready'
    });
    expect(unsafeAnswer.subtitle).toBe('Your water bottle is in the Office. {assetId: [redacted]} bearer [redacted] [redacted]');
    expect(unsafeAnswer.subtitle).not.toContain('water-bottle-1');
    expect(unsafeAnswer.subtitle).not.toContain('secret-token');
    expect(unsafeAnswer.subtitle).not.toContain('stack trace');

    const unsafePayloadAnswer = buildVoiceAccessoryPresentation({
      pathname: '/locations/location-1',
      realtime: {
        status: 'completed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Done',
        spokenResponse: 'Raw prompt: hidden system instruction about the office',
        debugEvents: []
      },
      stage: 'completed',
      status: 'ready'
    });
    expect(unsafePayloadAnswer.subtitle).toBe('[redacted]');
    expect(unsafePayloadAnswer.subtitle).not.toContain('hidden system instruction');

    expect(buildVoiceAccessoryPresentation({
      pathname: '/locations/location-1',
      realtime: {
        status: 'completed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Needs detail',
        responseKind: 'clarification',
        clarificationFollowUpAvailable: true,
        spokenResponse: 'Which item should I update?',
        debugEvents: []
      },
      stage: 'completed',
      status: 'ready'
    })).toMatchObject({
      accessibilityLabel: 'Open voice follow-up',
      title: 'Needs detail',
      subtitle: 'Which item should I update?'
    });

    expect(buildVoiceAccessoryPresentation({
      pathname: '/locations/location-1',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Voice failed',
        failureCode: 'speech_to_text_failed',
        errorMessage: 'Speech-to-text provider failed. Check Voice providers and try again.',
        debugEvents: []
      },
      stage: 'failed',
      status: 'ready'
    })).toMatchObject({
      title: 'Speech input failed',
      subtitle: 'Check Voice providers and try again.'
    });

    expect(buildVoiceAccessoryPresentation({
      pathname: '/locations/location-1',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Voice failed',
        failureCode: 'language_inference_failed',
        errorMessage: 'Language model stopped while continuing this request. Check Voice providers and try again.',
        debugEvents: []
      },
      stage: 'failed',
      status: 'ready'
    })).toMatchObject({
      title: 'Agent brain failed',
      subtitle: 'Check Voice providers and try again.'
    });

    expect(buildVoiceAccessoryPresentation({
      diagnosticsEnabled: true,
      pathname: '/locations/location-1',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Voice failed',
        failureCode: 'language_inference_failed',
        errorMessage: 'Language model stopped while continuing this request. Check diagnostics or Voice providers and try again.',
        debugEvents: []
      },
      stage: 'failed',
      status: 'ready'
    })).toMatchObject({
      title: 'Agent brain failed',
      subtitle: 'Open diagnostics or check Voice providers.'
    });

    expect(buildVoiceAccessoryPresentation({
      pathname: '/locations/location-1',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Voice needs a fresh start',
        failureCode: 'clarification_turn_limit',
        errorMessage: 'That thread needs a fresh voice request. Start again with the missing detail included.',
        debugEvents: []
      },
      stage: 'failed',
      status: 'ready'
    })).toMatchObject({
      title: 'Voice needs a fresh start',
      subtitle: 'Start a fresh voice request.'
    });

    expect(buildVoiceAccessoryPresentation({
      pathname: '/locations/location-1',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Speech output failed',
        failureCode: 'text_to_speech_failed',
        errorMessage: 'Speech output failed after Stuff Stash prepared the answer. Check Voice providers and try again.',
        spokenResponse: 'Your water bottle is in the Office.',
        debugEvents: []
      },
      stage: 'failed',
      status: 'ready'
    })).toMatchObject({
      title: 'Speech output failed',
      subtitle: 'Check Voice providers and try again.'
    });

    expect(buildVoiceAccessoryPresentation({
      pathname: '/locations/location-1',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'raw prompt bearer secret stack trace provider session',
        failureCode: 'text_to_speech_failed',
        errorMessage: 'Speech output failed after Stuff Stash prepared the answer. Check Voice providers and try again.',
        debugEvents: []
      },
      stage: 'failed',
      status: 'ready'
    })).toMatchObject({
      title: 'Speech output failed'
    });
  });

  it('labels completed clarification responses as needing detail', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'completed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Needs detail',
        responseKind: 'clarification',
        clarificationFollowUpAvailable: true,
        spokenResponse: 'Which item should I update?',
        debugEvents: []
      },
      stage: 'completed',
      tenantName: 'Main tenant'
    });

    expect(session.title).toBe('Needs detail');
    expect(session.bottomHint).toBe('Answer the follow-up to keep this conversation going.');
    expect(session.bottomAction).toMatchObject({
      kind: 'session_controls',
      mic: { accessibilityLabel: 'Answer follow-up' }
    });
  });

  it('does not advertise same-session follow-up after clarification availability is gone', () => {
    const realtime = {
      status: 'completed' as const,
      tenantName: 'Main tenant',
      inventoryName: 'Home',
      progressLabel: 'Needs detail',
      responseKind: 'clarification',
      clarificationFollowUpAvailable: false,
      spokenResponse: 'Which item should I update?',
      debugEvents: []
    };

    expect(buildVoiceAccessoryPresentation({
      pathname: '/locations/location-1',
      realtime,
      stage: 'completed',
      status: 'ready'
    })).toMatchObject({
      accessibilityLabel: 'Open voice answer',
      title: 'Answer ready'
    });

    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime,
      stage: 'completed',
      tenantName: 'Main tenant'
    });

    expect(session.title).toBe('Answer ready');
    expect(session.bottomHint).toBe('You can ask another question or close this.');
    expect(session.bottomAction).toMatchObject({
      kind: 'session_controls',
      mic: { accessibilityLabel: 'Start another voice interaction' }
    });
  });

  it('does not leak raw realtime details into the collapsed accessory', () => {
    const unsafeText = 'raw prompt bearer secret stack trace provider id transcript';

    const processing = buildVoiceAccessoryPresentation({
      pathname: '/',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        partialTranscript: unsafeText,
        progressLabel: unsafeText,
        debugEvents: [{ label: 'Inventory lookup', status: 'Updated' }]
      },
      stage: 'processing',
      status: 'ready'
    });
    expect(`${processing.title} ${processing.subtitle}`).not.toContain('bearer secret');
    expect(processing.title).toBe('Checking inventory');

    const failed = buildVoiceAccessoryPresentation({
      pathname: '/',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Voice failed',
        errorMessage: unsafeText,
        debugEvents: []
      },
      stage: 'failed',
      status: 'ready'
    });
    expect(`${failed.title} ${failed.subtitle}`).not.toContain('bearer secret');
    expect(failed.subtitle).toBe('Open for details.');
  });

  it('does not promise to start voice while loading or unavailable', () => {
    expect(buildVoiceAccessoryPresentation({ pathname: '/', stage: 'ready', status: 'loading' })).toMatchObject({
      accessibilityLabel: 'Open voice status',
      primaryAction: 'expand',
      title: 'Voice loading'
    });
    expect(buildVoiceAccessoryPresentation({ pathname: '/', stage: 'ready', status: 'error' })).toMatchObject({
      accessibilityLabel: 'Open voice error',
      primaryAction: 'expand',
      title: 'Voice unavailable'
    });
  });

  it('keeps diagnostics collapsed and unavailable unless explicitly enabled', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: true,
      inventoryName: 'Home',
      realtime: {
        status: 'completed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        transcript: 'Where is my water bottle?',
        spokenResponse: 'Your water bottle is in the Office.',
        progressLabel: 'Done',
        debugEvents: [{ label: 'Inventory search', status: 'Completed' }]
      },
      stage: 'completed',
      tenantName: 'Main tenant'
    });

    expect(session.title).toBe('Answer ready');
    expect(session.transcript).toBe('Where is my water bottle?');
    expect(session.response).toBe('Your water bottle is in the Office.');
    expect(session.diagnostics).toBeNull();
  });

  it('shows safe diagnostics only when enabled and expanded', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: true,
      diagnosticsExpanded: true,
      inventoryName: 'Home',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Checking location',
        debugEvents: [{
          label: 'Inventory list',
          status: 'Completed',
          detail: '{\n  "count": 1\n}'
        }]
      },
      stage: 'processing',
      tenantName: 'Main tenant'
    });

    expect(session.diagnostics).toEqual(['Inventory list: Completed\n{\n  "count": 1\n}']);
  });

  it('exposes safe progress steps without requiring developer diagnostics', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Searching visible inventory',
        progressSteps: ['Sending audio', 'Connected', 'Searching visible inventory'],
        debugEvents: [{ label: 'Inventory search', status: 'Completed' }]
      },
      stage: 'processing',
      tenantName: 'Main tenant'
    });

    expect(session.progressSteps).toEqual(['Sending audio', 'Connected', 'Searching visible inventory']);
    expect(session.progressTrace).toEqual(['Sending audio', 'Connected', 'Searching visible inventory']);
    expect(session.diagnostics).toBeNull();
  });

  it('redacts unsafe progress labels and timeline entries at the presentation boundary', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'raw prompt bearer abc/def== stack trace provider session id: gemini-live-1',
        progressSteps: [
          'Sending audio',
          'Authorization: tok+en/with~punctuation bearer eyJhbGciOi.test.sig== {"parentAssetId":"kitchen-1"}',
          'providerSessionId: gemini-live-2',
          'provider_session_id: gemini-live-3',
          'Checking inventory'
        ],
        debugEvents: []
      },
      stage: 'processing',
      tenantName: 'Main tenant'
    });

    const visibleText = `${session.progressLabel} ${session.progressSteps.join(' ')} ${session.progressTrace.join(' ')}`;
    expect(visibleText).toContain('Working safely');
    expect(visibleText).not.toContain('Authorization');
    expect(visibleText).not.toContain('parentAssetId');
    expect(visibleText).not.toContain('providerSessionId');
    expect(visibleText).not.toContain('provider_session_id');
    expect(visibleText).not.toContain('bearer');
    expect(visibleText).not.toContain('raw prompt');
    expect(visibleText).not.toContain('abc/def');
    expect(visibleText).not.toContain('tok+en');
    expect(visibleText).not.toContain('eyJhbGciOi');
    expect(visibleText).not.toContain('kitchen-1');
    expect(visibleText).not.toContain('stack trace');
    expect(visibleText).not.toContain('gemini-live-1');
  });

  it('bounds active progress trace and hides it behind action-plan review content', () => {
    const active = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Recovering safely',
        progressSteps: [
          'Sending audio',
          'Connected',
          'Connected',
          'Understanding request',
          'Checking inventory',
          'Preparing plan',
          'Recovering safely after model contract failure with an overly long but safe status label that should be bounded in the sheet'
        ],
        debugEvents: []
      },
      stage: 'processing',
      tenantName: 'Main tenant'
    });
    expect(active.progressTrace).toEqual([
      'Connected',
      'Understanding request',
      'Checking inventory',
      'Preparing plan',
      'Recovering safely after model contract failure with an overly long but...'
    ]);

    const review = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'review',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Review needed',
        progressSteps: ['Sending audio', 'Checking inventory', 'Review needed'],
        debugEvents: [],
        actionPlan: {
          planId: 'plan-1',
          status: 'proposed',
          confirmationSummary: 'Create item water bottle?',
          commands: [{ kind: 'create_asset', summary: 'Create item water bottle' }],
          risks: []
        }
      },
      stage: 'review',
      tenantName: 'Main tenant'
    });
    expect(review.progressTrace).toEqual([]);
  });

  it('uses the partial transcript until the final transcript is available', () => {
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        partialTranscript: 'Where is',
        progressLabel: 'Transcribing',
        debugEvents: []
      },
      stage: 'processing',
      tenantName: 'Main tenant'
    }).transcript).toBe('Where is');

    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        partialTranscript: 'Where is',
        transcript: 'Where is the drill?',
        progressLabel: 'Thinking',
        debugEvents: []
      },
      stage: 'processing',
      tenantName: 'Main tenant'
    }).transcript).toBe('Where is the drill?');
  });

  it('does not show partial transcripts after the active session has failed or been cancelled', () => {
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        partialTranscript: 'Where is',
        progressLabel: 'Voice failed',
        debugEvents: []
      },
      stage: 'failed',
      tenantName: 'Main tenant'
    }).transcript).toBeUndefined();

    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'cancelled',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        partialTranscript: 'Where is',
        progressLabel: 'Cancelled',
        debugEvents: []
      },
      stage: 'cancelled',
      tenantName: 'Main tenant'
    }).transcript).toBeUndefined();
  });

  it('does not expose reset while an active session needs true cancellation semantics', () => {
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: null,
      stage: 'processing',
      tenantName: 'Main tenant'
    })).toMatchObject({
      bottomAction: {
        kind: 'session_controls',
        canCancel: true
      },
      canReset: false
    });
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: null,
      stage: 'completed',
      tenantName: 'Main tenant'
    })).toMatchObject({
      bottomAction: {
        kind: 'session_controls',
        canCancel: false
      },
      canReset: true
    });
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'cancelled',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Cancelled',
        debugEvents: []
      },
      stage: 'cancelled',
      tenantName: 'Main tenant'
    })).toMatchObject({
      bottomAction: {
        kind: 'session_controls',
        canCancel: false
      },
      canReset: true,
      title: 'Cancelled',
      progressLabel: 'Cancelled'
    });
  });

  it('does not expose reset while an action plan is still awaiting review', () => {
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'review',
        tenantName: 'Main tenant',
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
      },
      stage: 'review',
      tenantName: 'Main tenant'
    })).toMatchObject({
      bottomAction: {
        kind: 'review_decision',
        planId: 'plan-1'
      },
      canReset: false,
    });
  });

  it('formats dependent create plans for clear mobile review', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'review',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Review needed',
        debugEvents: [],
        actionPlan: {
          planId: 'plan-1',
          status: 'proposed',
          confirmationSummary: 'Create a box underneath the TV and add an Apple TV remote inside it?',
          commands: [
            {
              id: 'cmd-box',
              kind: 'create_asset',
              operation: 'create',
              title: 'Box underneath the TV',
              assetKind: 'container',
              parentAssetId: 'location-1',
              parentTitle: 'Living room',
              parentKind: 'location',
              summary: 'Create Box underneath the TV in Living room'
            },
            {
              id: 'cmd-remote',
              kind: 'create_asset',
              operation: 'create',
              title: 'Apple TV remote',
              assetKind: 'item',
              parentCommandId: 'cmd-box',
              summary: 'Create Apple TV remote inside Box underneath the TV'
            }
          ],
          risks: []
        }
      },
      stage: 'review',
      tenantName: 'Main tenant'
    });

    expect(session.actionPlan).toMatchObject({
      summary: '2 new things',
      commands: [
        {
          title: 'Living room',
          subtitle: 'Use existing location',
          editable: false,
          photoDraftEligible: false,
          tone: 'use'
        },
        {
          title: 'Box underneath the TV',
          subtitle: 'Create container',
          editable: true,
          placement: 'Inside Living room',
          photoDraftEligible: true,
          tone: 'create'
        },
        {
          title: 'Apple TV remote',
          subtitle: 'Create item',
          editable: true,
          placement: 'Inside new Box underneath the TV',
          photoDraftEligible: true,
          tone: 'create'
        }
      ]
    });
    expect(session.actionPlan?.commands.map((command) => command.title).join(' ')).not.toContain('location-1');
  });

  it('requires stable command ids before voice review rows can stage photos', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'review',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Review needed',
        debugEvents: [],
        actionPlan: {
          planId: 'plan-1',
          status: 'proposed',
          confirmationSummary: 'Create item water bottle?',
          commands: [{
            kind: 'create_asset',
            operation: 'create',
            title: 'Water bottle',
            assetKind: 'item',
            summary: 'Create Water bottle'
          }],
          risks: []
        }
      },
      stage: 'review',
      tenantName: 'Main tenant'
    });

    expect(session.actionPlan?.commands[0]).toMatchObject({
      title: 'Water bottle',
      photoDraftEligible: false,
      tone: 'create'
    });
  });

  it('uses normalized display titles for dependent create placement labels', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'review',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Review needed',
        debugEvents: [],
        actionPlan: {
          planId: 'plan-1',
          status: 'proposed',
          confirmationSummary: 'Create a box and put the remote inside it?',
          commands: [
            {
              id: 'cmd-box',
              kind: 'create_asset',
              operation: 'create',
              title: '   ',
              assetKind: 'container',
              summary: 'Box under the TV'
            },
            {
              id: 'cmd-remote',
              kind: 'create_asset',
              operation: 'create',
              title: 'Remote',
              assetKind: 'item',
              parentCommandId: 'cmd-box',
              summary: 'Create Remote inside Box under the TV'
            }
          ],
          risks: []
        }
      },
      stage: 'review',
      tenantName: 'Main tenant'
    });

    expect(session.actionPlan?.commands).toMatchObject([
      {
        title: 'Box under the TV',
        subtitle: 'Create container'
      },
      {
        title: 'Remote',
        placement: 'Inside new Box under the TV'
      }
    ]);
  });

  it('uses neutral row titles for existing-asset changes without verified asset titles', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'review',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Review needed',
        debugEvents: [],
        actionPlan: {
          planId: 'plan-1',
          status: 'proposed',
          confirmationSummary: 'Move the shed to the backyard?',
          commands: [{
            id: 'cmd-move-shed',
            kind: 'move_asset',
            operation: 'move',
            assetKind: 'item',
            summary: 'Move the shed to the backyard.'
          }],
          risks: []
        }
      },
      stage: 'review',
      tenantName: 'Main tenant'
    });

    expect(session.actionPlan?.commands[0]).toMatchObject({
      title: 'Selected item',
      subtitle: 'Move the shed to the backyard.',
      photoDraftEligible: false,
      tone: 'update'
    });
  });

  it('allows photo drafts on reviewed move rows with verified asset titles', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'review',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Review needed',
        debugEvents: [],
        actionPlan: {
          planId: 'plan-1',
          status: 'proposed',
          confirmationSummary: 'Move the shed to the backyard?',
          commands: [{
            id: 'cmd-move-shed',
            kind: 'move_asset',
            operation: 'move',
            title: 'Shed',
            assetKind: 'item',
            summary: 'Move Shed to Backyard'
          }],
          risks: []
        }
      },
      stage: 'review',
      tenantName: 'Main tenant'
    });

    expect(session.actionPlan?.commands[0]).toMatchObject({
      title: 'Shed',
      subtitle: 'Move Shed to Backyard',
      photoDraftEligible: true,
      tone: 'update'
    });
  });

  it('does not treat blank existing-asset command titles as verified display context', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'review',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Review needed',
        debugEvents: [],
        actionPlan: {
          planId: 'plan-1',
          status: 'proposed',
          confirmationSummary: 'Move the shed to the backyard?',
          commands: [{
            id: 'cmd-move-shed',
            kind: 'move_asset',
            operation: 'move',
            title: '   ',
            assetKind: 'item',
            summary: 'Move the shed to the backyard.'
          }],
          risks: []
        }
      },
      stage: 'review',
      tenantName: 'Main tenant'
    });

    expect(session.actionPlan?.commands[0]).toMatchObject({
      title: 'Selected item',
      subtitle: 'Move the shed to the backyard.',
      photoDraftEligible: false,
      tone: 'update'
    });
  });

  it('does not offer photo drafts on lifecycle or checkout review rows', () => {
    const session = buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'review',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Review needed',
        debugEvents: [],
        actionPlan: {
          planId: 'plan-1',
          status: 'proposed',
          confirmationSummary: 'Update the drill?',
          commands: [
            {
              id: 'cmd-archive-drill',
              kind: 'archive_asset',
              operation: 'archive',
              title: 'Drill',
              assetKind: 'item',
              summary: 'Archive Drill'
            },
            {
              id: 'cmd-return-drill',
              kind: 'return_asset',
              operation: 'return',
              title: 'Drill',
              assetKind: 'item',
              summary: 'Return Drill'
            }
          ],
          risks: []
        }
      },
      stage: 'review',
      tenantName: 'Main tenant'
    });

    expect(session.actionPlan?.commands.map((command) => ({
      title: command.title,
      photoDraftEligible: command.photoDraftEligible
    }))).toEqual([
      { title: 'Drill', photoDraftEligible: false },
      { title: 'Drill', photoDraftEligible: false }
    ]);
  });

  it('disables action plan decisions while a review decision is pending', () => {
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'review',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Approving change',
        debugEvents: [],
        reviewDecisionPending: true,
        actionPlan: {
          planId: 'plan-1',
          status: 'proposed',
          confirmationSummary: 'Create item water bottle?',
          commands: [{ kind: 'create_asset', summary: 'Create item water bottle' }],
          risks: []
        }
      },
      stage: 'review',
      tenantName: 'Main tenant'
    })).toMatchObject({
      bottomAction: {
        kind: 'none'
      },
      canReset: false,
      progressLabel: 'Approving change'
    });
  });

  it('describes bottom action controls for normal, working, review, and terminal states', () => {
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: null,
      stage: 'listening',
      tenantName: 'Main tenant'
    }).bottomAction).toEqual({
      kind: 'session_controls',
      canCancel: true,
      mic: {
        accessibilityLabel: 'Send voice request',
        disabled: false,
        icon: 'send',
        selected: true
      }
    });

    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: null,
      stage: 'processing',
      tenantName: 'Main tenant'
    }).bottomAction).toEqual({
      kind: 'session_controls',
      canCancel: true,
      mic: {
        accessibilityLabel: 'Voice request in progress',
        disabled: true,
        icon: 'busy',
        selected: false
      }
    });

    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Applying change',
        debugEvents: [],
        reviewDecisionPending: true,
        actionPlan: {
          planId: 'plan-1',
          status: 'approved',
          confirmationSummary: 'Create item water bottle?',
          commands: [{ kind: 'create_asset', summary: 'Create item water bottle' }],
          risks: []
        }
      },
      stage: 'processing',
      tenantName: 'Main tenant'
    }).bottomAction).toEqual({
      kind: 'none'
    });

    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'completed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Change applied',
        debugEvents: [],
        actionPlan: {
          planId: 'plan-1',
          status: 'executed',
          confirmationSummary: 'Create item water bottle?',
          commands: [{ kind: 'create_asset', summary: 'Create item water bottle' }],
          risks: []
        }
      },
      stage: 'completed',
      tenantName: 'Main tenant'
    }).bottomAction).toEqual({
      kind: 'session_controls',
      canCancel: false,
      mic: {
        accessibilityLabel: 'Start another voice interaction',
        disabled: false,
        icon: 'mic',
        selected: false
      }
    });
  });

  it('marks the active sheet chrome as listening or busy without requiring diagnostics', () => {
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: null,
      stage: 'listening',
      tenantName: 'Main tenant'
    })).toMatchObject({
      activity: {
        kind: 'listening',
        label: 'Listening',
        level: 0
      },
      isBusy: true
    });

    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'processing',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Searching visible inventory',
        debugEvents: []
      },
      stage: 'processing',
      tenantName: 'Main tenant'
    })).toMatchObject({
      activity: {
        kind: 'busy',
        label: 'Searching visible inventory'
      },
      bottomAction: {
        kind: 'session_controls',
        mic: {
          accessibilityLabel: 'Voice request in progress',
          disabled: true,
          icon: 'busy'
        }
      },
      isBusy: true
    });
  });

  it('normalizes recorder metering into the listening send control', () => {
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'listening',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Listening',
        recordingLevel: 0.72,
        debugEvents: []
      },
      stage: 'listening',
      tenantName: 'Main tenant'
    })).toMatchObject({
      activity: {
        kind: 'listening',
        level: 0.72
      }
    });
  });

  it('offers provider setup recovery only for provider-readiness and provider-stage failures', () => {
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Voice failed',
        debugEvents: [],
        failureCode: 'provider_readiness',
        errorMessage: 'Voice provider profiles are not ready: text_to_speech.'
      },
      stage: 'failed',
      tenantName: 'Main tenant'
    }).recoveryAction).toEqual({
      label: 'Voice providers',
      target: 'provider_profiles'
    });

    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Voice failed',
        debugEvents: [],
        failureCode: 'speech_to_text_failed',
        errorMessage: 'Speech-to-text provider failed. Check Voice providers and try again.'
      },
      stage: 'failed',
      tenantName: 'Main tenant'
    }).recoveryAction).toEqual({
      label: 'Voice providers',
      target: 'provider_profiles'
    });

    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Voice failed',
        debugEvents: [],
        errorMessage: 'Voice provider profiles are not ready: text_to_speech.'
      },
      stage: 'failed',
      tenantName: 'Main tenant'
    }).recoveryAction).toBeUndefined();

    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Voice failed',
        debugEvents: [],
        errorMessage: 'Voice socket closed before the session completed.'
      },
      stage: 'failed',
      tenantName: 'Main tenant'
    }).recoveryAction).toBeUndefined();

    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: {
        status: 'failed',
        tenantName: 'Main tenant',
        inventoryName: 'Home',
        progressLabel: 'Voice needs a fresh start',
        debugEvents: [],
        failureCode: 'clarification_turn_limit',
        errorMessage: 'That thread needs a fresh voice request. Start again with the missing detail included.'
      },
      stage: 'failed',
      tenantName: 'Main tenant'
    })).toMatchObject({
      bottomHint: 'Reset and try again when you are ready.',
      recoveryAction: undefined
    });
  });
});
