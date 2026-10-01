import { t } from '$lib/presentation/localization';
export type SignInState = 'default' | 'expired' | 'rejected';
export type SignInFailure = 'configuration' | 'workspace' | 'start';

export interface SignInPresentation {
  title: string;
  description: string;
}

const presentations: Record<SignInState, SignInPresentation> = {
  default: {
    title: t('web.signInPresentation.signInToStuffStash'),
    description: t('web.signInPresentation.continueToYourSecureSignInPageYouLl')
  },
  expired: {
    title: t('web.signInPresentation.sessionExpired'),
    description: t('web.signInPresentation.yourSessionEndedSignInAgainToContinue')
  },
  rejected: {
    title: t('web.signInPresentation.weCouldnTOpenYourAccount'),
    description: t('web.signInPresentation.signInAgainIfTheProblemContinuesContactThe')
  }
};

const failureMessages: Record<SignInFailure, string> = {
  configuration: t('web.signInPresentation.stuffStashIsnTReadyToSignYouIn'),
  workspace: t('web.signInPresentation.stuffStashCouldnTLoadYourInventoryRefreshThe'),
  start: t('web.signInPresentation.theSecureSignInPageDidnTOpenTry')
};

export function signInPresentation(state: SignInState): SignInPresentation {
  return presentations[state];
}

export function signInFailureMessage(failure: SignInFailure): string {
  return failureMessages[failure];
}
