import { afterEach, describe, expect, it } from 'vitest';
import { mount, unmount } from 'svelte';
import TransientPortalForwarding from './transient-portal-forwarding.test.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('generated transient-surface portals', () => {
  it('forward child snippets to the Bits UI portal primitive', () => {
    component = mount(TransientPortalForwarding, { target: document.body });

    expect(document.querySelector('[data-testid="popover-portal-child"]')?.textContent).toBe('Popover content');
    expect(document.querySelector('[data-testid="sheet-portal-child"]')?.textContent).toBe('Sheet content');
  });
});
