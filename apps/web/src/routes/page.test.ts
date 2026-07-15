import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import Page from './+page.svelte';
import { AuthenticationRequiredError } from '$lib/application/authenticationRequired';

const auth = vi.hoisted(() => ({
  getStoredSession: vi.fn<() => { idToken: string; expiresAt: number } | null>(() => null),
  hasRecentlyCompletedSignIn: vi.fn(() => false),
  signOut: vi.fn(),
  startSignIn: vi.fn(async () => {})
}));
const runtime = vi.hoisted(() => ({
  loadRuntimeConfig: vi.fn(async () => runtimeConfig())
}));
const inventory = vi.hoisted(() => ({
  loadWorkspace: vi.fn(async () => ({}))
}));

vi.mock('$lib/auth', () => auth);
vi.mock('$lib/runtimeConfig', () => runtime);
vi.mock('$lib/adapters/api/stuffStashInventoryRepository', () => ({
  StuffStashInventoryRepository: class {
    loadWorkspace = inventory.loadWorkspace;
  }
}));

let component: ReturnType<typeof mount> | null = null;
const observedAuthEvents: CustomEvent[] = [];
const observeAuthEvent = (event: Event) => observedAuthEvents.push(event as CustomEvent);

beforeEach(() => {
  observedAuthEvents.length = 0;
  window.addEventListener('stuffstash:auth-observability', observeAuthEvent);
});

afterEach(() => {
  window.removeEventListener('stuffstash:auth-observability', observeAuthEvent);
  if (component) unmount(component);
  component = null;
  document.body.innerHTML = '';
  auth.getStoredSession.mockReset().mockReturnValue(null);
  auth.hasRecentlyCompletedSignIn.mockReset().mockReturnValue(false);
  auth.signOut.mockReset();
  auth.startSignIn.mockReset().mockResolvedValue(undefined);
  runtime.loadRuntimeConfig.mockReset().mockResolvedValue(runtimeConfig());
  inventory.loadWorkspace.mockReset().mockResolvedValue({});
});

describe('root sign-in route', () => {
  it('composes the provider-neutral ordinary sign-in presentation', async () => {
    component = mount(Page, { target: document.body });
    await flush();

    expect(document.body.textContent).toContain('Sign in to Stuff Stash');
    expect(document.body.textContent).toContain('Continue to your secure sign-in page. You’ll return here when you’re done.');
    expect(document.body.textContent).not.toMatch(/Dex|OIDC|client ID|identity provider/i);
  });

  it('replaces raw configuration diagnostics with fixed recovery copy', async () => {
    runtime.loadRuntimeConfig.mockRejectedValueOnce(
      new Error('Missing web runtime configuration value: oidcClientId for Dex https://issuer.example.test')
    );

    component = mount(Page, { target: document.body });
    await flush();

    expect(document.body.textContent).toContain('Stuff Stash isn’t ready to sign you in. Reload the page to try again.');
    expect(document.body.textContent).not.toMatch(/oidcClientId|Dex|issuer\.example|runtime configuration/i);
    expect(buttonContaining('Continue to sign in').disabled).toBe(true);
    expect(lastAuthEvent()).toEqual({
      eventName: 'auth.runtime_configuration_failed',
      attributes: { failureType: 'Error', reason: 'runtime_configuration' }
    });
  });

  it('catches sign-in start failures and leaves a safe visible retry', async () => {
    auth.startSignIn.mockRejectedValueOnce(new Error('Dex issuer client ID rejected redirect_uri'));
    component = mount(Page, { target: document.body });
    await flush();

    buttonContaining('Continue to sign in').click();
    await flush();

    expect(document.body.textContent).toContain('The secure sign-in page didn’t open. Try again.');
    expect(document.body.textContent).not.toMatch(/Dex|issuer|client ID|redirect_uri/i);
    expect(buttonContaining('Continue to sign in').disabled).toBe(false);
    expect(lastAuthEvent()).toEqual({
      eventName: 'auth.sign_in_start_failed',
      attributes: { failureType: 'Error', reason: 'sign_in_navigation' }
    });
  });

  it('replaces raw workspace diagnostics with fixed recovery copy', async () => {
    auth.getStoredSession.mockReturnValue(session());
    inventory.loadWorkspace.mockRejectedValueOnce(
      new Error('GET https://api.example.test returned 503 from postgres inventory adapter')
    );

    component = mount(Page, { target: document.body });
    await flush();

    expect(document.body.textContent).toContain('Stuff Stash couldn’t load your inventory. Refresh the page to try again.');
    expect(document.body.textContent).not.toMatch(/api\.example|503|postgres|adapter/i);
    expect(lastAuthEvent()).toEqual({
      eventName: 'auth.workspace_load_failed',
      attributes: { failureType: 'Error', reason: 'workspace_transport' }
    });
  });

  it.each([
    [false, 'Session expired', 'Your session ended. Sign in again to continue.'],
    [true, 'We couldn’t open your account', 'Sign in again. If the problem continues, contact the person who manages this server.']
  ] as const)('composes safe authenticated-boundary recovery when callback recency is %s', async (recent, title, description) => {
    auth.getStoredSession.mockReturnValue(session());
    auth.hasRecentlyCompletedSignIn.mockReturnValue(recent);
    inventory.loadWorkspace.mockRejectedValueOnce(new AuthenticationRequiredError('Dex rejected OIDC client ID'));

    component = mount(Page, { target: document.body });
    await flush();

    expect(document.body.textContent).toContain(title);
    expect(document.body.textContent).toContain(description);
    expect(document.body.textContent).not.toMatch(/Dex|OIDC|client ID/i);
    expect(lastAuthEvent()).toEqual({
      eventName: 'auth.session_invalidated',
      attributes: {
        failureType: 'AuthenticationRequiredError',
        reason: recent ? 'post_callback_rejected' : 'session_expired'
      }
    });
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

function session() {
  return { idToken: 'header.payload.signature', expiresAt: Date.now() + 60_000 };
}

function lastAuthEvent() {
  return observedAuthEvents.at(-1)?.detail;
}

function buttonContaining(text: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find((candidate) =>
    candidate.textContent?.includes(text)
  );
  if (!button) throw new Error(`Missing button containing ${text}`);
  return button;
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
