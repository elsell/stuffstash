import type { PerformanceContext, PerformanceOutcome } from './performanceReporter';

export interface RequestPerformanceObserver {
  start(context: PerformanceContext): (outcome: PerformanceOutcome) => void;
}

/** Measures response headers only; never reads or consumes product response bodies. */
export function createObservedFetch(
  fetchImpl: typeof fetch,
  observer: RequestPerformanceObserver,
  context: PerformanceContext
): typeof fetch {
  const { platform, operation, surface, variant } = context;
  return async (input, init) => {
    let finish: ((outcome: PerformanceOutcome) => void) | undefined;
    try {
      finish = observer.start({ platform, operation, surface, variant });
    } catch {
      // An optional measurement sink cannot prevent a product request.
    }
    const complete = (outcome: PerformanceOutcome) => {
      try { finish?.(outcome); } catch {
        // Preserve both successful responses and original transport errors.
      }
    };
    try {
      const response = await fetchImpl(input, init);
      complete(response.ok ? 'success' : 'failure');
      return response;
    } catch (error) {
      const signal = init?.signal ?? (input instanceof Request ? input.signal : undefined);
      complete(signal?.aborted ? 'cancelled' : 'failure');
      throw error;
    }
  };
}
