import { describe, expect, it } from 'vitest';
import { ConnectionProfile } from '../../application/onboarding/ConnectionProfile';
import { MobileComposition } from '../../bootstrap/mobileComposition';
import {
  appServicesStateAfterReset,
  appServicesStateAfterStartupError,
  appServicesStateFromOnboardingStart
} from './AppServicesGate';

describe('AppServicesGate', () => {
  it('shows onboarding when startup finds no complete profile', () => {
    const state = appServicesStateFromOnboardingStart({ step: 'instance' }, createComposition);

    expect(state).toEqual({
      status: 'onboarding',
      onboardingState: { step: 'instance' }
    });
  });

  it('builds app services when onboarding is complete', () => {
    const profile = {
      apiBaseUrl: 'http://localhost:8080',
      devToken: 'dev:user-1',
      tenantId: 'tenant-home'
    };

    const state = appServicesStateFromOnboardingStart(
      { step: 'complete', profile, tenantName: 'Home' },
      createComposition
    );

    expect(state.status).toBe('ready');
    if (state.status !== 'ready') {
      throw new Error('expected ready state');
    }
    expect(state.composition).toEqual({ profile });
  });

  it('returns to instance onboarding after startup errors and reset', () => {
    expect(appServicesStateAfterStartupError()).toEqual({
      status: 'onboarding',
      onboardingState: { step: 'instance' }
    });
    expect(appServicesStateAfterReset()).toEqual({
      status: 'onboarding',
      onboardingState: { step: 'instance' }
    });
  });
});

function createComposition(profile: ConnectionProfile): MobileComposition {
  return { profile } as unknown as MobileComposition;
}
