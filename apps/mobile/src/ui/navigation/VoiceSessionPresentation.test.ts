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
      accessibilityLabel: 'Stop listening',
      primaryAction: 'stop',
      subtitle: 'Location context'
    });
  });

  it('expands the session surface for in-progress and completed sessions', () => {
    expect(buildVoiceAccessoryPresentation({ pathname: '/assets/asset-1', stage: 'processing', status: 'ready' }).primaryAction).toBe('expand');
    expect(buildVoiceAccessoryPresentation({ pathname: '/assets/asset-1', stage: 'completed', status: 'ready' }).primaryAction).toBe('expand');
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
        debugEvents: [{ label: 'Inventory list', status: 'Completed' }]
      },
      stage: 'processing',
      tenantName: 'Main tenant'
    });

    expect(session.diagnostics).toEqual(['Inventory list: Completed']);
  });

  it('does not expose reset while an active session needs true cancellation semantics', () => {
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: null,
      stage: 'processing',
      tenantName: 'Main tenant'
    })).toMatchObject({ canCancel: true, canReset: false });
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: null,
      stage: 'completed',
      tenantName: 'Main tenant'
    })).toMatchObject({ canCancel: false, canReset: true });
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
      canCancel: false,
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
      canApproveActionPlan: true,
      canCancelActionPlan: true,
      canReset: false
    });
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
      canApproveActionPlan: false,
      canCancelActionPlan: false,
      canReset: false,
      progressLabel: 'Approving change'
    });
  });

  it('offers provider setup recovery only for provider-readiness failures', () => {
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
  });
});
