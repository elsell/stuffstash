import { describe, expect, it } from 'vitest';
import { createTimeoutFetch } from '../adapters/network/TimeoutFetch';

describe('createTimeoutFetch', () => {
  it('keeps the timeout active when a caller cancellation signal is present', async () => {
    const fetchWithTimeout = createTimeoutFetch(5, abortablePendingFetch());

    await expect(fetchWithTimeout('https://api.example.test', {
      signal: new AbortController().signal
    })).rejects.toThrow('Network request timed out');
  });

  it('preserves caller cancellation instead of reporting a timeout', async () => {
    const caller = new AbortController();
    const fetchWithTimeout = createTimeoutFetch(1_000, abortablePendingFetch());
    const pending = fetchWithTimeout('https://api.example.test', { signal: caller.signal });

    caller.abort();

    await expect(pending).rejects.toMatchObject({ name: 'AbortError' });
  });

  it('does not relabel a caller cancellation when fetch rejects after the timeout deadline', async () => {
    const caller = new AbortController();
    const fetchWithTimeout = createTimeoutFetch(5, delayedAbortablePendingFetch(15));
    const pending = fetchWithTimeout('https://api.example.test', { signal: caller.signal });

    caller.abort();

    await expect(pending).rejects.toMatchObject({ name: 'AbortError' });
  });
});

function abortablePendingFetch(): typeof fetch {
  return (async (_input, init) => new Promise<Response>((_resolve, reject) => {
    const signal = init?.signal;
    if (signal?.aborted) {
      reject(new DOMException('Aborted', 'AbortError'));
      return;
    }
    signal?.addEventListener('abort', () => {
      reject(new DOMException('Aborted', 'AbortError'));
    }, { once: true });
  })) as typeof fetch;
}

function delayedAbortablePendingFetch(rejectionDelayMs: number): typeof fetch {
  return (async (_input, init) => new Promise<Response>((_resolve, reject) => {
    init?.signal?.addEventListener('abort', () => {
      setTimeout(() => reject(new DOMException('Aborted', 'AbortError')), rejectionDelayMs);
    }, { once: true });
  })) as typeof fetch;
}


it('preserves cancellation carried on a generated Request', async () => {
  const caller = new AbortController();
  const reason = new DOMException('Caller cancelled', 'AbortError');
  const request = new Request('https://api.example.test', { signal: caller.signal });
  let downstream: AbortSignal | null | undefined;
  const fetchWithTimeout = createTimeoutFetch(1000, async (_input, init) => {
    downstream = init?.signal;
    caller.abort(reason);
    expect(downstream?.aborted).toBe(true);
    throw reason;
  });
  await expect(fetchWithTimeout(request)).rejects.toBe(reason);
});

it('allows an explicitly detached Request signal', async () => {
  const caller = new AbortController(); caller.abort();
  const response = new Response();
  const fetchWithTimeout = createTimeoutFetch(1000, async (_input, init) => {
    expect(init?.signal?.aborted).toBe(false);
    return response;
  });
  expect(await fetchWithTimeout(new Request('https://api.example.test', { signal: caller.signal }), { signal: null })).toBe(response);
});
