export function createTimeoutFetch(timeoutMs: number, fetchImpl: typeof fetch = fetch): typeof fetch {
  return async (input, init) => {
    const requestController = new AbortController();
    let abortCause: 'caller' | 'timeout' | undefined;
    const abortFromCaller = () => {
      if (abortCause) return;
      abortCause = 'caller';
      requestController.abort(init?.signal?.reason);
    };
    if (init?.signal?.aborted) {
      abortFromCaller();
    } else {
      init?.signal?.addEventListener('abort', abortFromCaller, { once: true });
    }
    const timeout = setTimeout(() => {
      if (abortCause) return;
      abortCause = 'timeout';
      requestController.abort();
    }, timeoutMs);

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
      init?.signal?.removeEventListener('abort', abortFromCaller);
    }
  };
}
