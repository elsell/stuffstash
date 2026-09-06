export interface PerformanceContext {
  platform: 'ios' | 'android' | 'web';
  operation: 'request' | 'image';
  surface: 'home' | 'list' | 'detail' | 'gallery' | 'fullscreen' | 'upload';
  variant: 'none' | 'small' | 'medium' | 'large' | 'original';
}
export type PerformanceOutcome = 'success' | 'failure' | 'cancelled';
export interface PerformanceMeasurement extends PerformanceContext {
  outcome: PerformanceOutcome;
  durationMs: number;
}
export interface PerformanceScheduler {
  schedule(callback: () => void, delayMs: number): () => void;
}
export interface PerformanceReporterOptions {
  clock: { now(): number };
  scheduler: PerformanceScheduler;
  send(batch: readonly PerformanceMeasurement[], signal: AbortSignal): Promise<void>;
  capacity?: number;
  batchSize?: number;
}
interface Delivery { controller: AbortController; done: Promise<void> }

const flushIntervalMs = 5000;
const deliveryTimeoutMs = 10000;

/** Best-effort transport infrastructure. Own one instance per authenticated server session. */
export class PerformanceReporter {
  private readonly capacity: number;
  private readonly batchSize: number;
  private readonly queue: PerformanceMeasurement[] = [];
  private cancelFlush?: () => void;
  private active?: Delivery;
  private disposed = false;

  constructor(private readonly options: PerformanceReporterOptions) {
    this.capacity = options.capacity ?? 100;
    this.batchSize = options.batchSize ?? 20;
    if (!Number.isInteger(this.capacity) || this.capacity < 1 || this.capacity > 100 ||
      !Number.isInteger(this.batchSize) || this.batchSize < 1 || this.batchSize > 50 || this.batchSize > this.capacity) {
      throw new Error('Invalid performance buffer limits');
    }
  }

  start(context: PerformanceContext): (outcome: PerformanceOutcome) => void {
    const started = this.options.clock.now();
    const { platform, operation, surface, variant } = context;
    let finished = false;
    return outcome => {
      if (finished) return;
      finished = true;
      this.record({ platform, operation, surface, variant, outcome,
        durationMs: Math.min(60000, Math.max(0, this.options.clock.now() - started)) });
    };
  }

  record(value: PerformanceMeasurement): void {
    if (this.disposed || !valid(value)) return;
    const { platform, operation, surface, variant, outcome, durationMs } = value;
    this.queue.push({ platform, operation, surface, variant, outcome, durationMs });
    if (this.queue.length > this.capacity) this.queue.shift();
    this.scheduleFlush();
  }

  flush(): Promise<void> {
    if (this.disposed) return Promise.resolve();
    if (this.active) return this.active.done;
    if (!this.queue.length) return Promise.resolve();
    this.cancelFlush?.();
    this.cancelFlush = undefined;
    const batch = this.queue.splice(0, this.batchSize);
    const delivery: Delivery = { controller: new AbortController(), done: Promise.resolve() };
    this.active = delivery;
    delivery.done = this.deliver(batch, delivery);
    return delivery.done;
  }

  dispose(): void {
    this.disposed = true;
    this.queue.length = 0;
    this.cancelFlush?.();
    this.cancelFlush = undefined;
    this.active?.controller.abort();
  }

  private scheduleFlush(): void {
    if (this.disposed || this.cancelFlush || this.active || !this.queue.length) return;
    this.cancelFlush = this.options.scheduler.schedule(() => {
      this.cancelFlush = undefined;
      void this.flush();
    }, flushIntervalMs);
  }

  private async deliver(batch: PerformanceMeasurement[], delivery: Delivery): Promise<void> {
    const signal = delivery.controller.signal;
    const cancelDeadline = this.options.scheduler.schedule(() => delivery.controller.abort(), deliveryTimeoutMs);
    let removeAbort = () => {};
    try {
      const aborted = new Promise<never>((_resolve, reject) => {
        const abort = () => reject(new Error('Performance delivery cancelled'));
        signal.addEventListener('abort', abort, { once: true });
        removeAbort = () => signal.removeEventListener('abort', abort);
        if (signal.aborted) abort();
      });
      const sent = Promise.resolve().then(() => {
        signal.throwIfAborted();
        return this.options.send(batch, signal);
      });
      await Promise.race([sent, aborted]);
    } catch {
      // Drop this batch: retrying telemetry must not compete with product traffic.
    } finally {
      removeAbort();
      cancelDeadline();
      if (this.active === delivery) this.active = undefined;
      this.scheduleFlush();
    }
  }
}

function valid(value: PerformanceMeasurement): boolean {
  return Number.isFinite(value.durationMs) && value.durationMs >= 0 && value.durationMs <= 60000 &&
    ['ios', 'android', 'web'].includes(value.platform) && ['request', 'image'].includes(value.operation) &&
    ['home', 'list', 'detail', 'gallery', 'fullscreen', 'upload'].includes(value.surface) &&
    ['none', 'small', 'medium', 'large', 'original'].includes(value.variant) &&
    ['success', 'failure', 'cancelled'].includes(value.outcome);
}
