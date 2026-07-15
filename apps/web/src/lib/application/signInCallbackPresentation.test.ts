import { describe, expect, it } from 'vitest';
import { failedSignInCallbackPresentation, pendingSignInCallbackPresentation } from './signInCallbackPresentation';

describe('sign-in callback presentation', () => {
  it('uses provider-neutral progress copy', () => {
    expect(pendingSignInCallbackPresentation()).toEqual({
      title: 'Finishing secure sign-in…',
      description: 'Stuff Stash is confirming your session.'
    });
  });

  it('does not expose raw provider or protocol errors', () => {
    const presentation = failedSignInCallbackPresentation(
      new Error('invalid OIDC state from Dex client stuff-stash-web')
    );

    expect(presentation).toEqual({
      title: 'We couldn’t finish signing you in.',
      description: 'Stuff Stash couldn’t confirm your session. Return to sign in and try again.',
      actionLabel: 'Return to sign in'
    });
    expect(JSON.stringify(presentation)).not.toContain('OIDC');
    expect(JSON.stringify(presentation)).not.toContain('Dex');
  });
});
