import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import AccountMenu from './AccountMenu.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) unmount(component);
  component = null;
  document.body.innerHTML = '';
});

describe('AccountMenu', () => {
  it('uses one desktop account row with menu semantics and restores focus', async () => {
    const onSignOut = vi.fn();
    const onOpenSettings = vi.fn();
    component = mount(AccountMenu, {
      target: document.body,
      props: {
        userLabel: 'owner@example.com',
        settingsHref: '/tenants/tenant-one/inventories/inventory-one/settings',
        onOpenSettings,
        onSignOut,
        disablePortal: true
      }
    });

    const trigger = requiredButton('Account menu for owner@example.com');
    expect(trigger.getAttribute('aria-haspopup')).toBe('menu');
    expect(trigger.getAttribute('aria-expanded')).toBe('false');

    trigger.focus();
    trigger.click();
    await flush();
    expect(document.body.querySelector('[role="menu"]')).not.toBeNull();
    expect(document.body.textContent).toContain('Signed in as');
    expect(document.body.textContent).toContain('owner@example.com');
    const settings = requiredLink('Settings');
    expect(settings.getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings');
    expect(settings.getAttribute('role')).toBe('menuitem');
    expect(settings.classList.contains('account-menu-link')).toBe(true);
    expect(settings.className).toContain('focus:bg-accent');

    const menu = document.body.querySelector<HTMLElement>('[role="menu"]');
    menu?.focus();
    menu?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true, cancelable: true }));
    await flush();
    expect(document.activeElement).toBe(settings);
    settings.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }));
    await flush();
    expect(onOpenSettings).toHaveBeenCalledTimes(1);
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
    expect(document.activeElement).toBe(trigger);

    trigger.focus();
    trigger.click();
    await flush();
    expect(document.body.querySelector('[data-variant="destructive"]')?.textContent).toContain('Sign out');

    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    await flush();
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
    expect(document.activeElement).toBe(trigger);
  });

  it('uses a modal bottom sheet on mobile with Settings and separated destructive sign out', async () => {
    const onSignOut = vi.fn();
    const onOpenSettings = vi.fn();
    const onOpenChange = vi.fn();
    component = mount(AccountMenu, {
      target: document.body,
      props: {
        mobile: true,
        userLabel: 'owner@example.com',
        settingsHref: '/tenants/tenant-one/inventories/inventory-one/settings',
        onOpenSettings,
        onSignOut,
        onOpenChange,
        disablePortal: true
      }
    });

    const trigger = requiredButton('Open account menu');
    expect(getComputedStyle(trigger).minHeight).toBe('44px');
    trigger.click();
    await flush();

    const sheet = document.body.querySelector<HTMLElement>('[data-slot="sheet-content"]');
    expect(sheet?.getAttribute('role')).toBe('dialog');
    expect(sheet?.getAttribute('aria-modal')).toBe('true');
    expect(document.body.textContent).toContain('Account');
    expect(document.body.textContent).toContain('owner@example.com');
    expect(onOpenChange).toHaveBeenLastCalledWith(true);

    const settings = requiredLink('Settings');
    expect(settings.getAttribute('href')).toBe('/tenants/tenant-one/inventories/inventory-one/settings');
    settings.click();
    expect(onOpenSettings).toHaveBeenCalledTimes(1);

    trigger.click();
    await flush();

    const signOut = requiredButton('Sign out');
    expect(signOut.getAttribute('data-variant')).toBe('destructive');
    signOut.click();
    expect(onSignOut).toHaveBeenCalledTimes(1);
  });

  it('locks scroll, provides a 44px close target, and restores mobile focus after Escape', async () => {
    component = mount(AccountMenu, {
      target: document.body,
      props: {
        mobile: true,
        userLabel: 'Signed-in account',
        settingsHref: '/tenants/tenant-one/inventories/inventory-one/settings',
        onOpenSettings: () => {},
        onSignOut: () => {},
        disablePortal: true
      }
    });

    const trigger = requiredButton('Open account menu');
    trigger.focus();
    trigger.click();
    await flush();

    const close = requiredButton('Close');
    expect(close.classList.contains('size-11')).toBe(true);
    expect(document.body.style.overflow).toBe('hidden');

    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
    await flush();
    expect(document.body.querySelector('[data-slot="sheet-content"]')).toBeNull();
    expect(document.activeElement).toBe(trigger);
    await new Promise<void>((resolve) => window.setTimeout(resolve, 35));
    await tick();
    expect(document.body.style.overflow).not.toBe('hidden');
  });
});

function requiredButton(name: string): HTMLButtonElement {
  const button = Array.from(document.body.querySelectorAll<HTMLButtonElement>('button')).find(
    (candidate) => candidate.getAttribute('aria-label') === name || candidate.textContent?.trim() === name
  );
  if (!button) throw new Error(`Missing button ${name}`);
  return button;
}

function requiredLink(name: string): HTMLAnchorElement {
  const link = Array.from(document.body.querySelectorAll<HTMLAnchorElement>('a')).find(
    (candidate) => candidate.textContent?.trim() === name
  );
  if (!link) throw new Error(`Missing link ${name}`);
  return link;
}

async function flush(): Promise<void> {
  await Promise.resolve();
  await tick();
  await Promise.resolve();
  await tick();
}
