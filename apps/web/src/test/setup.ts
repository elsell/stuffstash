import { tick } from 'svelte';
import { afterEach, beforeEach, vi } from 'vitest';

const TRANSIENT_SURFACE_TEARDOWN_MS = 30;
let bodyStyleBeforeTest: string | null = null;

beforeEach(() => {
  bodyStyleBeforeTest = document.body.getAttribute('style');
});

afterEach(async () => {
  await tick();

  if (!bodyHasActiveScrollLock()) return;
  const usingFakeTimers = vi.isFakeTimers();

  try {
    if (usingFakeTimers) {
      await vi.advanceTimersByTimeAsync(TRANSIENT_SURFACE_TEARDOWN_MS);
    } else {
      await new Promise<void>((resolve) => {
        window.setTimeout(resolve, TRANSIENT_SURFACE_TEARDOWN_MS);
      });
    }

    await tick();
  } finally {
    if (usingFakeTimers) vi.useRealTimers();
    if (bodyStyleBeforeTest === null) document.body.removeAttribute('style');
    else document.body.setAttribute('style', bodyStyleBeforeTest);
  }
});

function bodyHasActiveScrollLock(): boolean {
  return document.body.style.overflow === 'hidden'
    || document.body.style.pointerEvents === 'none'
    || document.body.style.getPropertyValue('--scrollbar-width') !== '';
}
