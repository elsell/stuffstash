import { tick } from 'svelte';
import { mount, unmount } from 'svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import AuthSignInScreen from './AuthSignInScreen.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('AuthSignInScreen', () => {
  it('starts the configured sign-in flow without showing demo workspace data', async () => {
    const onSignIn = vi.fn();
    component = mount(AuthSignInScreen, {
      target: document.body,
      props: { onSignIn }
    });

    expect(document.body.textContent).toContain('Stuff Stash');
    expect(document.body.textContent).toContain('Sign in to Stuff Stash');
    expect(document.body.textContent).toContain('Continue to your secure sign-in page. You’ll return here when you’re done.');
    expect(document.body.textContent).not.toContain('Local demo data');

    const shell = document.body.querySelector<HTMLElement>('.auth-shell');
    const panel = document.body.querySelector<HTMLElement>('.auth-panel');
    expect(shell).not.toBeNull();
    expect(panel).not.toBeNull();
    expect(panel!.dataset.authTrack).toBe('readable');
    expect(getComputedStyle(buttonContaining('Continue to sign in')).minHeight).toBe('48px');
    buttonContaining('Continue to sign in').click();
    await flush();

    expect(onSignIn).toHaveBeenCalledTimes(1);
  });

  it('surfaces configuration errors and disables sign-in when auth is not ready', async () => {
    const onSignIn = vi.fn();
    component = mount(AuthSignInScreen, {
      target: document.body,
      props: {
        canSignIn: false,
        error: 'Stuff Stash isn’t ready to sign you in. Reload the page to try again.',
        onSignIn
      }
    });

    expect(document.body.textContent).toContain('Stuff Stash isn’t ready to sign you in. Reload the page to try again.');
    expect(buttonContaining('Continue to sign in').disabled).toBe(true);
  });

  it('can show a session-expired sign-in prompt', () => {
    component = mount(AuthSignInScreen, {
      target: document.body,
      props: {
        title: 'Session expired.',
        description: 'Sign in again to continue.',
        onSignIn: vi.fn()
      }
    });

    expect(document.body.textContent).toContain('Session expired.');
    expect(document.body.textContent).toContain('Sign in again to continue.');
  });
});

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
}

function buttonContaining(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) {
    throw new Error(`Missing button containing ${text}`);
  }
  return button;
}
