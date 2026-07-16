export interface PendingSignInCallbackPresentation {
  title: string;
  description: string;
}

export interface FailedSignInCallbackPresentation extends PendingSignInCallbackPresentation {
  actionLabel: string;
}

export function pendingSignInCallbackPresentation(): PendingSignInCallbackPresentation {
  return {
    title: 'Finishing secure sign-in…',
    description: 'Stuff Stash is confirming your session.'
  };
}

export function failedSignInCallbackPresentation(_error: unknown): FailedSignInCallbackPresentation {
  return {
    title: 'We couldn’t finish signing you in.',
    description: 'Stuff Stash couldn’t confirm your session. Return to sign in and try again.',
    actionLabel: 'Return to sign in'
  };
}
