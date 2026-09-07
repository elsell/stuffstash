import { createApiPerformanceReporter, createObservedFetch, type TokenProvider } from '@stuff-stash/api-client';
import { noopPerformanceObserver, type PerformanceObserver } from '../../application/observability/PerformanceObserver';

export interface MobilePerformanceSession {
  fetch: typeof fetch;
  observer: PerformanceObserver;
  dispose(): void;
  acquire(): () => void;
}
interface Options {
  platform: 'ios' | 'android' | 'web';
  enabled: boolean;
  baseUrl: string;
  tokenProvider: TokenProvider;
  fetch?: typeof fetch;
  clock?: { now(): number };
  scheduler?: { schedule(callback: () => void, delayMs: number): () => void };
}

export function createMobilePerformanceSession(options: Options): MobilePerformanceSession {
  const fetchImpl = options.fetch ?? fetch;
  if (!options.enabled) return { fetch: fetchImpl, observer: noopPerformanceObserver, dispose() {}, acquire: () => () => {} };
  const reporter = createApiPerformanceReporter({
    connection: { baseUrl: options.baseUrl, tokenProvider: options.tokenProvider, fetch: fetchImpl },
    clock: options.clock ?? { now: () => performance.now() },
    scheduler: options.scheduler ?? { schedule(callback, delay) {
      const timer = setTimeout(callback, delay);
      return () => clearTimeout(timer);
    } }
  });
  let leases = 0;
  let disposed = false;
  const dispose = () => { disposed = true; reporter.dispose(); };
  return {
    acquire() {
      if (disposed) return () => {};
      leases++;
      let released = false;
      return () => {
        if (released) return;
        released = true;
        leases--;
        // React may synchronously reacquire this session during effect replay.
        void Promise.resolve().then(() => { if (leases === 0) dispose(); });
      };
    },
    fetch: createObservedFetch(fetchImpl, reporter, { platform: options.platform, operation: 'request', surface: 'application', variant: 'none' }),
    observer: { start: context => reporter.start({ ...context, platform: options.platform }) },
    dispose
  };
}
