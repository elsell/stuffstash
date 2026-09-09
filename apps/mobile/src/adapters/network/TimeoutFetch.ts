export function createTimeoutFetch(timeoutMs: number | ((input: RequestInfo | URL, init?: RequestInit) => number), fetchImpl: typeof fetch = fetch): typeof fetch {
  return async (input, init) => {
    const callerSignal = init?.signal !== undefined ? init.signal : (input instanceof Request ? input.signal : undefined);
    const requestController = new AbortController();
    let abortCause: 'caller' | 'timeout' | undefined;
    const abortFromCaller = () => {
      if (abortCause) return;
      abortCause = 'caller';
      requestController.abort(callerSignal?.reason);
    };
    if (callerSignal?.aborted) {
      abortFromCaller();
    } else {
      callerSignal?.addEventListener('abort', abortFromCaller, { once: true });
    }
    const timeout = setTimeout(() => {
      if (abortCause) return;
      abortCause = 'timeout';
      requestController.abort();
    }, typeof timeoutMs === 'number' ? timeoutMs : timeoutMs(input, init));

    try {
      return await fetchImpl(input, {
        ...init,
        signal: requestController.signal
      });
    } catch (error) {
      if (abortCause === 'timeout' && error instanceof Error && error.name === 'AbortError') {
        throw new Error('Network request timed out. Check that the API is reachable from this phone.');
      }
      throw error;
    } finally {
      clearTimeout(timeout);
      callerSignal?.removeEventListener('abort', abortFromCaller);
    }
  };
}

// Image validation reads and decodes uploaded bytes; it needs a larger budget
// than ordinary inventory queries, particularly while thumbnails are generated.
export function mobileApiRequestTimeoutMs(input: RequestInfo | URL, init?: RequestInit): number {
  const method = (init?.method ?? (input instanceof Request ? input.method : 'GET')).toUpperCase();
  const url = new URL(input instanceof Request ? input.url : String(input));
  return method === 'POST' && /\/tenants\/[^/]+\/inventories\/[^/]+\/assets\/[^/]+\/attachments\/direct-uploads\/[^/]+\/complete$/.test(url.pathname)
    ? 60000 : 8000;
}
