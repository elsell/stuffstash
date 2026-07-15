import { mount, tick, unmount } from 'svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import BitsScrollLockHarness from './BitsScrollLockHarness.svelte';

let component: ReturnType<typeof mount> | null = null;

afterEach(() => {
  if (component) unmount(component);
  component = null;
  document.body.innerHTML = '';
});

describe.sequential('transient surface test lifecycle', () => {
  it('leaves a fake-clock scroll-lock cleanup for the centralized lifecycle', async () => {
    vi.useFakeTimers();
    component = mount(BitsScrollLockHarness, { target: document.body });
    await tick();

    expect(document.body.style.overflow).toBe('hidden');
  });

  it('starts the next test with a clean body and working real timers', async () => {
    expect(vi.isFakeTimers()).toBe(false);
    expect(document.body.getAttribute('style')).toBeNull();
    expect(document.body.style.overflow).toBe('');
    expect(document.body.style.pointerEvents).toBe('');
    expect(document.body.style.getPropertyValue('--scrollbar-width')).toBe('');

    let fired = false;
    await new Promise<void>((resolve) => {
      window.setTimeout(() => {
        fired = true;
        resolve();
      }, 0);
    });
    expect(fired).toBe(true);
  });

  it('preserves a preexisting body style for the next transient-surface test', () => {
    document.body.setAttribute('style', '--test-preserved: keep;');
  });

  it('restores the exact preexisting body style after real primitive cleanup', async () => {
    vi.useFakeTimers();
    component = mount(BitsScrollLockHarness, { target: document.body });
    await tick();

    expect(document.body.style.overflow).toBe('hidden');
    expect(document.body.style.getPropertyValue('--test-preserved')).toBe('keep');
  });

  it('starts after preexisting-style cleanup with real timers and the exact captured attribute', () => {
    expect(vi.isFakeTimers()).toBe(false);
    expect(document.body.getAttribute('style')).toBe('--test-preserved: keep;');
    document.body.removeAttribute('style');
  });
});
