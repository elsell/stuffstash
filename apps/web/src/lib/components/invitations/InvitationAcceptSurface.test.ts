import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import InvitationAcceptSurface from './InvitationAcceptSurface.svelte';

const preview = {
  inventoryId: 'inventory-one', inventoryName: 'Workshop tools', relationship: 'viewer' as const,
  status: 'pending' as const, isExpired: false, expiresAt: '2026-07-21T12:00:00Z'
};
let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) unmount(component);
  component = null;
  document.body.innerHTML = '';
});

describe('InvitationAcceptSurface', () => {
  it('marks loading and accepting states as busy for assistive technology', () => {
    component = mount(InvitationAcceptSurface, { target: document.body, props: { state: 'loading' } });
    expect(document.body.querySelector('.invitation-card')?.getAttribute('aria-live')).toBe('polite');
    expect(document.body.querySelector('.invitation-card')?.getAttribute('aria-busy')).toBe('true');
    expect(document.body.textContent).toContain('Checking invitation…');
    expect(document.body.querySelector('button, a')).toBeNull();
    unmount(component);

    component = mount(InvitationAcceptSurface, { target: document.body, props: { state: 'ready', preview, busy: true } });
    expect(document.body.querySelector('.invitation-card')?.getAttribute('aria-busy')).toBe('true');
    expect(button('Accepting…').disabled).toBe(true);
  });

  it('shows the inventory and requires explicit acceptance', async () => {
    const onAccept = vi.fn();
    component = mount(InvitationAcceptSurface, { target: document.body, props: { state: 'ready', preview, onAccept } });
    expect(document.body.textContent).toContain('Join Workshop tools');
    expect(document.body.textContent).toContain('Can view');
    expect(button('Accept invitation').getBoundingClientRect).toBeDefined();
    button('Accept invitation').click();
    await tick();
    expect(onAccept).toHaveBeenCalledTimes(1);
  });

  it('keeps sign-in explicit and promises a preview before acceptance', async () => {
    const onSignIn = vi.fn();
    component = mount(InvitationAcceptSurface, { target: document.body, props: { state: 'signed_out', onSignIn } });
    expect(document.body.textContent).toContain('Sign in to view the inventory and access level before accepting.');
    button('Continue to sign in').click();
    await tick();
    expect(onSignIn).toHaveBeenCalledTimes(1);
  });

  it.each([
    ['expired', 'This invitation expired'],
    ['revoked', 'This invitation was revoked'],
    ['cancelled', 'This invitation was cancelled'],
    ['invalid', 'This invitation link is invalid']
  ] as const)('shows the %s terminal state', (state, heading) => {
    component = mount(InvitationAcceptSurface, { target: document.body, props: { state } });
    expect(document.body.textContent).toContain(heading);
    expect(document.body.querySelector('button')).toBeNull();
  });

  it('explains revoked and cancelled invitations with distinct recovery copy', () => {
    component = mount(InvitationAcceptSurface, { target: document.body, props: { state: 'revoked' } });
    expect(document.body.textContent).toContain('The inventory owner revoked this invitation.');
    unmount(component);

    component = mount(InvitationAcceptSurface, { target: document.body, props: { state: 'cancelled' } });
    expect(document.body.textContent).toContain('The inventory owner cancelled this invitation.');
  });

  it('offers account switching for an email mismatch', async () => {
    const onSwitchAccount = vi.fn();
    component = mount(InvitationAcceptSurface, { target: document.body, props: { state: 'email_mismatch', onSwitchAccount } });
    expect(document.body.textContent).toContain('This invitation is for another account');
    button('Switch account').click();
    await tick();
    expect(onSwitchAccount).toHaveBeenCalledTimes(1);
  });

  it('offers retry without implying access changed after a service failure', async () => {
    const onRetry = vi.fn();
    component = mount(InvitationAcceptSurface, { target: document.body, props: { state: 'unavailable', onRetry } });
    expect(document.body.textContent).toContain('Your access has not changed.');
    button('Try again').click();
    await tick();
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it.each(['accepted', 'success'] as const)('offers direct inventory entry for %s state', (state) => {
    component = mount(InvitationAcceptSurface, { target: document.body, props: { state, preview, openInventoryHref: '/tenants/t/inventories/i' } });
    expect(document.body.querySelector<HTMLAnchorElement>('a[href="/tenants/t/inventories/i"]')?.textContent).toContain('Open inventory');
  });
});

function button(label: string): HTMLButtonElement {
  const value = Array.from(document.body.querySelectorAll('button')).find((item) => item.textContent?.includes(label));
  if (!value) throw new Error(`Missing button: ${label}`);
  return value;
}
