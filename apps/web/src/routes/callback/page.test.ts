import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import CallbackPage from './+page.svelte';

const navigation = vi.hoisted(() => ({ goto: vi.fn(async () => {}) }));
const runtime = vi.hoisted(() => ({ loadRuntimeConfig: vi.fn(async () => runtimeConfig()) }));
const auth = vi.hoisted(() => ({ completeSignIn: vi.fn(async () => '/inventory-home') }));

vi.mock('$app/navigation', () => navigation);
vi.mock('$lib/runtimeConfig', () => runtime);
vi.mock('$lib/auth', () => auth);

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) unmount(component);
  component = null;
  document.body.innerHTML = '';
  document.title = '';
  navigation.goto.mockReset().mockResolvedValue(undefined);
  runtime.loadRuntimeConfig.mockReset().mockResolvedValue(runtimeConfig());
  auth.completeSignIn.mockReset().mockResolvedValue('/inventory-home');
});

describe('sign-in callback page', () => {
  it('renders the provider-neutral pending state while callback completion is unresolved', async () => {
    runtime.loadRuntimeConfig.mockReturnValueOnce(new Promise(() => {}));

    component = mount(CallbackPage, { target: document.body });
    await flush();

    expect(document.title).toBe('Signing in · Stuff Stash');
    expect(document.body.textContent).toContain('Finishing secure sign-in');
    expect(document.body.textContent).toContain('Stuff Stash is confirming your session.');
    expect(document.body.querySelector('[aria-live="polite"]')?.textContent).toContain('Confirming session');
    expect(document.body.textContent).not.toMatch(/Dex|OIDC|client ID/i);
  });

  it('completes sign-in and navigates to the safe restored destination', async () => {
    component = mount(CallbackPage, { target: document.body });
    await flush();

    expect(auth.completeSignIn).toHaveBeenCalledOnce();
    expect(navigation.goto).toHaveBeenCalledWith('/inventory-home');
  });

  it('presents and announces safe recovery without exposing raw callback diagnostics', async () => {
    const observed: CustomEvent[] = [];
    const observe = (event: Event) => observed.push(event as CustomEvent);
    window.addEventListener('stuffstash:auth-observability', observe);
    auth.completeSignIn.mockRejectedValueOnce(
      new Error('invalid OIDC state from Dex client stuff-stash-web at https://issuer.example.test')
    );

    try {
      component = mount(CallbackPage, { target: document.body });
      await flush();

      const action = document.body.querySelector<HTMLElement>('.callback-action');
      const panel = document.body.querySelector<HTMLElement>('.auth-panel');
      const alert = document.body.querySelector<HTMLElement>('[role="alert"]');
      expect(document.title).toBe('Sign-in failed · Stuff Stash');
      expect(panel?.dataset.authTrack).toBe('readable');
      expect(getComputedStyle(action!).minHeight).toBe('48px');
      expect(alert?.textContent).toContain('We couldn’t finish signing you in.');
      expect(alert?.textContent).toContain('Stuff Stash couldn’t confirm your session.');
      expect(document.body.textContent).toContain('Return to sign in');
      expect(document.body.textContent).not.toMatch(/OIDC|Dex|stuff-stash-web|issuer\.example/i);
      expect(observed.at(-1)?.detail).toEqual({
        eventName: 'auth.callback_failed',
        attributes: { failureType: 'Error', reason: 'callback_completion' }
      });
    } finally {
      window.removeEventListener('stuffstash:auth-observability', observe);
    }
  });
});

function runtimeConfig() {
  return {
    apiBaseUrl: 'https://api.example.test',
    oidcIssuer: 'https://identity.example.test',
    oidcClientId: 'web-client',
    oidcRedirectUri: 'https://app.example.test/callback',
    mediaUploadPolicy: {
      supportedContentTypes: ['image/jpeg'] as const,
      maxBytes: 1_000_000
    }
  };
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
