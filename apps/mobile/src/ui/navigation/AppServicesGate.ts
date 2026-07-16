import { OnboardingStartState } from '../../application/onboarding/OnboardingCommand';
import { ConnectionProfile } from '../../application/onboarding/ConnectionProfile';
import { MobileComposition } from '../../bootstrap/mobileComposition';

export type AppServicesGateState =
  | { readonly status: 'loading' }
  | { readonly status: 'onboarding'; readonly onboardingState: OnboardingStartState }
  | { readonly status: 'ready'; readonly composition: MobileComposition };

export type MobileCompositionFactory = (profile: ConnectionProfile) => MobileComposition;

export function appServicesStateFromOnboardingStart(
  startState: OnboardingStartState,
  createComposition: MobileCompositionFactory
): AppServicesGateState {
  if (startState.step === 'complete' && startState.profile) {
    return { status: 'ready', composition: createComposition(startState.profile) };
  }

  return { status: 'onboarding', onboardingState: startState };
}

export function appServicesStateAfterStartupError(): AppServicesGateState {
  return { status: 'onboarding', onboardingState: { step: 'instance' } };
}

export function appServicesStateAfterReset(): AppServicesGateState {
  return { status: 'onboarding', onboardingState: { step: 'instance' } };
}

export function appServicesStateAfterServerChange(): AppServicesGateState {
  return { status: 'onboarding', onboardingState: { step: 'instance' } };
}

export function appServicesStateAfterSignOut(profile: ConnectionProfile): AppServicesGateState {
  return { status: 'onboarding', onboardingState: { step: 'signIn', profile } };
}

export function appServicesStateAfterAuthenticationRequired(
  profile: ConnectionProfile
): AppServicesGateState {
  return { status: 'onboarding', onboardingState: { step: 'signIn', profile } };
}
