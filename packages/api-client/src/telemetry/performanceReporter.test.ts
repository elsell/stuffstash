import { expect, it } from 'vitest';
import { PerformanceReporter, type PerformanceMeasurement, type PerformanceScheduler } from './performanceReporter';

class Scheduler implements PerformanceScheduler {
  tasks: { callback: () => void; delay: number; active: boolean }[] = [];
  schedule(callback: () => void, delay: number): () => void {
    const task = { callback, delay, active: true };
    this.tasks.push(task);
    return () => { task.active = false; };
  }
  fire(delay: number): void {
    const task = this.tasks.find(value => value.active && value.delay === delay);
    if (!task) throw new Error('No scheduled task');
    task.active = false;
    task.callback();
  }
}
const measurement: PerformanceMeasurement = { platform: 'web', operation: 'image', surface: 'gallery', variant: 'medium', outcome: 'success', durationMs: 12 };

it('bounds pending data, drops oldest and sends only known fields', async () => {
  const scheduler = new Scheduler();
  const batches: PerformanceMeasurement[][] = [];
  const reporter = new PerformanceReporter({ clock: { now: () => 0 }, scheduler, capacity: 2, batchSize: 2,
    send: async batch => { batches.push([...batch]); } });
  reporter.record(measurement);
  reporter.record({ ...measurement, durationMs: 20 });
  reporter.record({ ...measurement, durationMs: 30, secret: 'private' } as PerformanceMeasurement);
  await reporter.flush();
  expect(batches).toEqual([[{ ...measurement, durationMs: 20 }, { ...measurement, durationMs: 30 }]]);
  reporter.dispose();
});

it('records completion once with a bounded monotonic duration', async () => {
  let now = 10;
  const batches: PerformanceMeasurement[][] = [];
  const reporter = new PerformanceReporter({ clock: { now: () => now }, scheduler: new Scheduler(), send: async batch => { batches.push([...batch]); } });
  const finish = reporter.start({ platform: 'web', operation: 'image', surface: 'gallery', variant: 'medium' });
  now = 100;
  finish('success');
  finish('failure');
  await reporter.flush();
  expect(batches).toEqual([[{ ...measurement, durationMs: 90 }]]);
  reporter.dispose();
});

it('drops failed delivery without retrying or failing the caller', async () => {
  let sends = 0;
  const reporter = new PerformanceReporter({ clock: { now: () => 0 }, scheduler: new Scheduler(), send: async () => { sends++; throw new Error('offline'); } });
  reporter.record(measurement);
  await reporter.flush();
  await reporter.flush();
  expect(sends).toBe(1);
  reporter.dispose();
});

it('aborts a stalled delivery and clears queued events on dispose', async () => {
  const scheduler = new Scheduler();
  let signal: AbortSignal | undefined;
  const reporter = new PerformanceReporter({ clock: { now: () => 0 }, scheduler, send: async (_batch, value) => { signal = value; return new Promise(() => {}); } });
  reporter.record(measurement);
  const pending = reporter.flush();
  await Promise.resolve();
  scheduler.fire(10000);
  await pending;
  expect(signal?.aborted).toBe(true);
  reporter.record(measurement);
  reporter.dispose();
  reporter.record(measurement);
  await reporter.flush();
  expect(scheduler.tasks.filter(task => task.active)).toHaveLength(0);
});
