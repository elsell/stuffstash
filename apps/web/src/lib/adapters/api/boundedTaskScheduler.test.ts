import { describe, expect, it } from 'vitest';
import { BoundedTaskScheduler } from './boundedTaskScheduler';

describe('BoundedTaskScheduler', () => {
  it('shares one concurrency ceiling across overlapping callers', async () => {
    const scheduler = new BoundedTaskScheduler(2);
    let active = 0;
    let peak = 0;
    const tasks = Array.from({ length: 8 }, (_, value) =>
      scheduler.schedule(async () => {
        active += 1;
        peak = Math.max(peak, active);
        await Promise.resolve();
        await Promise.resolve();
        active -= 1;
        return value;
      })
    );

    await expect(Promise.all(tasks)).resolves.toEqual([0, 1, 2, 3, 4, 5, 6, 7]);
    expect(peak).toBe(2);
  });

  it('rejects queued and future work after close while allowing active work to settle', async () => {
    const scheduler = new BoundedTaskScheduler(1);
    let releaseActive!: () => void;
    const active = scheduler.schedule(
      () => new Promise<string>((resolve) => (releaseActive = () => resolve('active finished')))
    );
    const queued = scheduler.schedule(async () => 'queued ran');

    await Promise.resolve();
    scheduler.close();

    await expect(queued).rejects.toThrow('Task scheduler is closed.');
    await expect(scheduler.schedule(async () => 'late ran')).rejects.toThrow('Task scheduler is closed.');
    releaseActive();
    await expect(active).resolves.toBe('active finished');
  });

  it('returns capacity after a task rejects', async () => {
    const scheduler = new BoundedTaskScheduler(1);

    await expect(scheduler.schedule(async () => Promise.reject(new Error('failed')))).rejects.toThrow('failed');
    await expect(scheduler.schedule(async () => 'recovered')).resolves.toBe('recovered');
  });
});
