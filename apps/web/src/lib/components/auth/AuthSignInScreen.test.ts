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
    expect(document.body.textContent).toContain('Sign in to continue.');
    expect(document.body.textContent).not.toContain('Local demo data');

    buttonContaining('Sign in').click();
    await flush();

    expect(onSignIn).toHaveBeenCalledTimes(1);
  });

  it('surfaces configuration errors and disables sign-in when auth is not ready', async () => {
    const onSignIn = vi.fn();
    component = mount(AuthSignInScreen, {
      target: document.body,
      props: {
        canSignIn: false,
        error: 'Unable to load web runtime configuration.',
        onSignIn
      }
    });

    expect(document.body.textContent).toContain('Unable to load web runtime configuration.');
    expect(buttonContaining('Sign in').disabled).toBe(true);
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
