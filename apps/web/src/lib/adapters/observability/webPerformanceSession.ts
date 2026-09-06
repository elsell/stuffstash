import { createApiPerformanceReporter, createObservedFetch, type TokenProvider } from '@stuff-stash/api-client';
import { noopPerformanceObserver, type PerformanceObserver } from '$lib/ports/performanceObserver';

export interface WebPerformanceSession {
  fetch: typeof fetch;
  observer: PerformanceObserver;
  dispose(): void;
}
interface Options {
  enabled: boolean;
  baseUrl: string;
  tokenProvider: TokenProvider;
  fetch?: typeof fetch;
  clock?: { now(): number };
  scheduler?: { schedule(callback: () => void, delayMs: number): () => void };
}

export function createWebPerformanceSession(options: Options): WebPerformanceSession {
  const fetchImpl = options.fetch ?? fetch;
  if (!options.enabled) return { fetch: fetchImpl, observer: noopPerformanceObserver, dispose() {} };
  const reporter = createApiPerformanceReporter({
    connection: { baseUrl: options.baseUrl, tokenProvider: options.tokenProvider, fetch: fetchImpl },
    clock: options.clock ?? { now: () => performance.now() },
    scheduler: options.scheduler ?? { schedule(callback, delay) {
      const timer = setTimeout(callback, delay);
      return () => clearTimeout(timer);
    } }
  });
  return {
    fetch: createObservedFetch(fetchImpl, reporter, { platform: 'web', operation: 'request', surface: 'application', variant: 'none' }),
    observer: { start: context => reporter.start({ ...context, platform: 'web' }) },
    dispose: () => reporter.dispose()
  };
}
