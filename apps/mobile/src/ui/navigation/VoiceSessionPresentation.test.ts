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
    }).canReset).toBe(false);
    expect(buildVoiceSessionPresentation({
      diagnosticsEnabled: false,
      diagnosticsExpanded: false,
      inventoryName: 'Home',
      realtime: null,
      stage: 'completed',
      tenantName: 'Main tenant'
    }).canReset).toBe(true);
  });
});
