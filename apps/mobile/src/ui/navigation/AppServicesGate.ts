import type { OnboardingStartState } from '../../application/onboarding/OnboardingCommand';
import type { ConnectionProfile } from '../../application/onboarding/ConnectionProfile';
import type { MobileComposition } from '../../bootstrap/mobileComposition';

export type AppServicesGateState<C = MobileComposition> =
  | { readonly status: 'loading' }
  | { readonly status: 'onboarding'; readonly onboardingState: OnboardingStartState }
  | { readonly status: 'ready'; readonly composition: C };

export type MobileCompositionFactory = (profile: ConnectionProfile) => MobileComposition;

export function appServicesStateFromOnboardingStart<C>(
  startState: OnboardingStartState,
  createComposition: (profile: ConnectionProfile) => C
): AppServicesGateState<C> {
  if (startState.step === 'complete' && startState.profile) {
    return { status: 'ready', composition: createComposition(startState.profile) };
  }

  return { status: 'onboarding', onboardingState: startState };
}

export function appServicesStateAfterStartupError(): Extract<AppServicesGateState, { status: 'onboarding' }> {
  return { status: 'onboarding', onboardingState: { step: 'instance' } };
}

export function appServicesStateAfterReset(): Extract<AppServicesGateState, { status: 'onboarding' }> {
  return { status: 'onboarding', onboardingState: { step: 'instance' } };
}

export function appServicesStateAfterServerChange(): Extract<AppServicesGateState, { status: 'onboarding' }> {
  return { status: 'onboarding', onboardingState: { step: 'instance' } };
}

export function appServicesStateAfterSignOut(profile: ConnectionProfile): Extract<AppServicesGateState, { status: 'onboarding' }> {
  return { status: 'onboarding', onboardingState: { step: 'signIn', profile } };
}

export function appServicesStateAfterAuthenticationRequired(
  profile: ConnectionProfile
): Extract<AppServicesGateState, { status: 'onboarding' }> {
  return { status: 'onboarding', onboardingState: { step: 'signIn', profile } };
}
