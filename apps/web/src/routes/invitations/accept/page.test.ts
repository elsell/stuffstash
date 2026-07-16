import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import type { AuthSession } from '$lib/auth';
import {
  InvitationFailure,
  type InvitationAcceptance,
  type InvitationLinkMaterial,
  type InvitationPreview
} from '$lib/domain/invitation';
import InvitationAcceptPage from './+page.svelte';

const auth = vi.hoisted(() => ({
  getStoredSession: vi.fn<() => AuthSession | null>(() => ({ idToken: 'id-token', expiresAt: Date.now() + 60_000 })),
  signOut: vi.fn(),
  startSignIn: vi.fn()
}));
const repository = vi.hoisted(() => ({
  preview: vi.fn<(material: InvitationLinkMaterial) => Promise<InvitationPreview>>(async () => ({
    inventoryId: 'inventory-one', inventoryName: 'Workshop tools', relationship: 'viewer' as const,
    status: 'pending' as const, isExpired: false, expiresAt: '2026-07-21T12:00:00Z'
  })),
  accept: vi.fn<(material: InvitationLinkMaterial) => Promise<InvitationAcceptance>>(async () => ({
    tenantId: 'tenant-one', inventoryId: 'inventory-one', status: 'accepted'
  }))
}));
const runtime = vi.hoisted(() => ({
  loadRuntimeConfig: vi.fn(async () => ({ apiBaseUrl: 'https://api.example.test' }))
}));

vi.mock('$lib/auth', () => auth);
vi.mock('$lib/runtimeConfig', () => runtime);
vi.mock('$lib/adapters/api/stuffStashInventoryInvitationRepository', () => ({
  StuffStashInventoryInvitationRepository: class {
    preview = repository.preview;
    accept = repository.accept;
  }
}));

const invitationPath = '/invitations/accept?tenant=tenant-one&inventory=inventory-one&invitation=invite-one#token=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';
let component: ReturnType<typeof mount> | null = null;

beforeEach(() => {
  history.replaceState({}, '', invitationPath);
  auth.getStoredSession.mockReturnValue({ idToken: 'id-token', expiresAt: Date.now() + 60_000 });
  runtime.loadRuntimeConfig.mockReset();
  runtime.loadRuntimeConfig.mockResolvedValue({ apiBaseUrl: 'https://api.example.test' });
  repository.preview.mockClear();
  repository.accept.mockClear();
});

afterEach(() => {
  if (component) unmount(component);
  component = null;
  document.body.innerHTML = '';
  vi.clearAllMocks();
});

describe('invitation acceptance route', () => {
  it('scrubs one-time link material before runtime configuration resolves', async () => {
    let resolveConfig!: (value: { apiBaseUrl: string }) => void;
    runtime.loadRuntimeConfig.mockReturnValueOnce(new Promise((resolve) => { resolveConfig = resolve; }));
    component = mount(InvitationAcceptPage, { target: document.body });
    await tick();

    expect(window.location.hash).toBe('');
    expect(document.body.querySelector('.invitation-card')?.getAttribute('aria-busy')).toBe('true');
    expect(repository.preview).not.toHaveBeenCalled();

    resolveConfig({ apiBaseUrl: 'https://api.example.test' });
    await settle();
    expect(repository.preview).toHaveBeenCalledTimes(1);
  });

  it('scrubs the fragment, retries from in-memory material, and accepts only after an explicit action', async () => {
    repository.preview.mockRejectedValueOnce(new Error('temporary'));
    component = mount(InvitationAcceptPage, { target: document.body });
    await settle();

    expect(window.location.hash).toBe('');
    expect(repository.preview).toHaveBeenCalledWith(expect.objectContaining({ invitationId: 'invite-one' }));
    expect(repository.accept).not.toHaveBeenCalled();
    button('Try again').click();
    await settle();
    expect(repository.preview).toHaveBeenCalledTimes(2);
    button('Accept invitation').click();
    await settle();

    expect(repository.accept).toHaveBeenCalledTimes(1);
    expect(repository.accept.mock.calls[0]![0].token).toBe('');
    expect(document.body.textContent).toContain('You joined Workshop tools');
    expect(document.body.querySelector('a[href="/tenants/tenant-one/inventories/inventory-one"]')).not.toBeNull();
  });

  it('hands a signed-out visitor to OIDC with the complete local invitation return path', async () => {
    auth.getStoredSession.mockReturnValue(null);
    component = mount(InvitationAcceptPage, { target: document.body });
    await settle();

    expect(window.location.hash).toBe('');
    expect(document.body.textContent).toContain('You’ve been invited');
    button('Continue to sign in').click();
    await settle();

    expect(auth.startSignIn).toHaveBeenCalledTimes(1);
    const [, location, storage, browserHistory, returnTo] = auth.startSignIn.mock.calls[0] ?? [];
    expect(location).toBe(window.location);
    expect(storage).toBe(window.sessionStorage);
    expect(browserHistory).toBe(window.history);
    expect(returnTo).toBe(invitationPath);
  });

  it('signs out a mismatched identity and preserves the invitation for a second OIDC identity', async () => {
    repository.preview.mockRejectedValueOnce(new InvitationFailure('email_mismatch'));
    component = mount(InvitationAcceptPage, { target: document.body });
    await settle();

    expect(document.body.textContent).toContain('This invitation is for another account');
    button('Switch account').click();
    await settle();

    expect(auth.signOut).toHaveBeenCalledTimes(1);
    expect(auth.startSignIn).toHaveBeenCalledTimes(1);
    const [, location, storage, browserHistory, returnTo] = auth.startSignIn.mock.calls[0] ?? [];
    expect(location).toBe(window.location);
    expect(storage).toBe(window.sessionStorage);
    expect(browserHistory).toBe(window.history);
    expect(returnTo).toBe(invitationPath);
  });

  it.each([
    [{ status: 'expired' as const, isExpired: true }, 'This invitation expired'],
    [{ status: 'revoked' as const, isExpired: false }, 'This invitation was revoked'],
    [{ status: 'cancelled' as const, isExpired: false }, 'This invitation was cancelled'],
    [{ status: 'accepted' as const, isExpired: false }, 'You already joined Workshop tools']
  ])('clears raw token references after terminal preview %#', async (terminal, heading) => {
    repository.preview.mockResolvedValueOnce({
      inventoryId: 'inventory-one', inventoryName: 'Workshop tools', relationship: 'viewer',
      expiresAt: '2026-07-21T12:00:00Z', ...terminal
    });
    component = mount(InvitationAcceptPage, { target: document.body });
    await settle();

    expect(document.body.textContent).toContain(heading);
    expect(repository.preview.mock.calls[0]![0].token).toBe('');
    if (terminal.status === 'accepted') {
      expect(document.body.querySelector('a[href="/tenants/tenant-one/inventories/inventory-one"]')).not.toBeNull();
    }
  });

  it('clears raw token references after an invalid preview while retaining them for retryable failures', async () => {
    repository.preview.mockRejectedValueOnce(new InvitationFailure('invalid'));
    component = mount(InvitationAcceptPage, { target: document.body });
    await settle();
    expect(repository.preview.mock.calls[0]![0].token).toBe('');
    unmount(component);
    component = null;

    history.replaceState({}, '', invitationPath);
    repository.preview.mockRejectedValueOnce(new InvitationFailure('unavailable'));
    component = mount(InvitationAcceptPage, { target: document.body });
    await settle();
    expect(repository.preview.mock.calls.at(-1)![0].token).toBe('AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA');
    expect(document.body.textContent).toContain('Invitation could not be checked');
  });

  it('reuses scrubbed in-memory material when runtime configuration initially fails', async () => {
    runtime.loadRuntimeConfig.mockRejectedValueOnce(new Error('temporary config failure'));
    component = mount(InvitationAcceptPage, { target: document.body });
    await settle();

    expect(window.location.hash).toBe('');
    expect(document.body.textContent).toContain('Invitation could not be checked');
    button('Try again').click();
    await settle();

    expect(runtime.loadRuntimeConfig).toHaveBeenCalledTimes(2);
    expect(repository.preview).toHaveBeenCalledTimes(1);
    expect(document.body.textContent).toContain('Join Workshop tools');
  });

});

async function settle(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}

function button(label: string): HTMLButtonElement {
  const value = Array.from(document.body.querySelectorAll('button')).find((item) => item.textContent?.includes(label));
  if (!value) throw new Error(`Missing button: ${label}`);
  return value;
}
