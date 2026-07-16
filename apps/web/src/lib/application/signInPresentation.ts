export type SignInState = 'default' | 'expired' | 'rejected';
export type SignInFailure = 'configuration' | 'workspace' | 'start';

export interface SignInPresentation {
  title: string;
  description: string;
}

const presentations: Record<SignInState, SignInPresentation> = {
  default: {
    title: 'Sign in to Stuff Stash',
    description: 'Continue to your secure sign-in page. You’ll return here when you’re done.'
  },
  expired: {
    title: 'Session expired',
    description: 'Your session ended. Sign in again to continue.'
  },
  rejected: {
    title: 'We couldn’t open your account',
    description: 'Sign in again. If the problem continues, contact the person who manages this server.'
  }
};

const failureMessages: Record<SignInFailure, string> = {
  configuration: 'Stuff Stash isn’t ready to sign you in. Reload the page to try again.',
  workspace: 'Stuff Stash couldn’t load your inventory. Refresh the page to try again.',
  start: 'The secure sign-in page didn’t open. Try again.'
};

export function signInPresentation(state: SignInState): SignInPresentation {
  return presentations[state];
}

export function signInFailureMessage(failure: SignInFailure): string {
  return failureMessages[failure];
}
