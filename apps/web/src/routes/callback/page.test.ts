import { afterEach, describe, expect, it, vi } from 'vitest';
import { mount, unmount } from 'svelte';
import CallbackPage from './+page.svelte';

vi.mock('$app/navigation', () => ({ goto: vi.fn() }));
vi.mock('$lib/runtimeConfig', () => ({ loadRuntimeConfig: () => new Promise(() => {}) }));
vi.mock('$lib/auth', () => ({ completeSignIn: vi.fn() }));

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) {
    unmount(component);
    component = null;
  }
  document.body.innerHTML = '';
});

describe('sign-in callback page', () => {
  it('uses provider-neutral progress copy', () => {
    component = mount(CallbackPage, { target: document.body });

    expect(document.body.textContent).toContain('Finishing secure sign-in');
    expect(document.body.textContent).not.toContain('Dex');
    expect(document.body.textContent).not.toContain('OIDC');
  });
});
