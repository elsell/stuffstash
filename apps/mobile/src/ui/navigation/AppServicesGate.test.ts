import { describe, expect, it } from 'vitest';
import { ConnectionProfile } from '../../application/onboarding/ConnectionProfile';
import { MobileComposition } from '../../bootstrap/mobileComposition';
import {
  appServicesStateAfterAuthenticationRequired,
  appServicesStateAfterServerChange,
  appServicesStateAfterSignOut,
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

  it('returns to instance onboarding after startup errors and an explicit server change', () => {
    expect(appServicesStateAfterStartupError()).toEqual({
      status: 'onboarding',
      onboardingState: { step: 'instance' }
    });
    expect(appServicesStateAfterServerChange()).toEqual({
      status: 'onboarding',
      onboardingState: { step: 'instance' }
    });
  });

  it('returns to sign-in while preserving the server and tenant hint after sign out', () => {
    const profile = {
      apiBaseUrl: 'http://localhost:8080',
      tenantId: 'tenant-home'
    };

    expect(appServicesStateAfterSignOut(profile)).toEqual({
      status: 'onboarding',
      onboardingState: { step: 'signIn', profile }
    });
  });

  it('returns to sign-in while preserving the connection profile after auth loss', () => {
    const profile = {
      apiBaseUrl: 'http://localhost:8080',
      tenantId: 'tenant-home'
    };

    expect(appServicesStateAfterAuthenticationRequired(profile)).toEqual({
      status: 'onboarding',
      onboardingState: { step: 'signIn', profile }
    });
  });
});

function createComposition(profile: ConnectionProfile): MobileComposition {
  return { profile } as unknown as MobileComposition;
}
