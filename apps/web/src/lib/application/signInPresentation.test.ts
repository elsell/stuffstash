import { describe, expect, it } from 'vitest';
import { signInFailureMessage, signInPresentation } from './signInPresentation';

describe('sign-in presentation', () => {
  it.each([
    ['default', 'Sign in to Stuff Stash', 'Continue to your secure sign-in page. You’ll return here when you’re done.'],
    ['expired', 'Session expired', 'Your session ended. Sign in again to continue.'],
    ['rejected', 'We couldn’t open your account', 'Sign in again. If the problem continues, contact the person who manages this server.']
  ] as const)('keeps the %s state provider-neutral', (state, title, description) => {
    const presentation = signInPresentation(state);

    expect(presentation).toEqual({ title, description });
    expect(JSON.stringify(presentation)).not.toMatch(/Dex|OIDC|client ID|identity provider/i);
  });

  it.each([
    ['configuration', 'Stuff Stash isn’t ready to sign you in. Reload the page to try again.'],
    ['workspace', 'Stuff Stash couldn’t load your inventory. Refresh the page to try again.'],
    ['start', 'The secure sign-in page didn’t open. Try again.']
  ] as const)('uses a fixed calm %s failure without raw diagnostics', (failure, expected) => {
    const message = signInFailureMessage(failure);

    expect(message).toBe(expected);
    expect(message).not.toMatch(/Dex|OIDC|client ID|runtime configuration|https?:\/\//i);
  });
});
