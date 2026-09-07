export function createTimeoutFetch(timeoutMs: number, fetchImpl: typeof fetch = fetch): typeof fetch {
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
      callerSignal?.removeEventListener('abort', abortFromCaller);
    }
  };
}
