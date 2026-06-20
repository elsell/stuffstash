// @vitest-environment node

import { render } from 'svelte/server';
import { describe, expect, it } from 'vitest';
import SignInPanel from './SignInPanel.svelte';

describe('SignInPanel', () => {
  it('renders the unauthenticated sign-in state', () => {
    const { body } = render(SignInPanel, { props: { onSignIn: async () => {} } });

    expect(body).toContain('Sign in with local Dex');
    expect(body).toContain('Sign in');
  });
});
