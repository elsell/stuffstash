import { expect, it } from 'vitest';
import { createWebPerformanceSession } from './webPerformanceSession';

function runtime() {
  let now = 0;
  const tasks: { run: () => void; active: boolean; delay: number }[] = [];
  return {
    clock: { now: () => now }, advance: (amount: number) => { now += amount; },
    scheduler: { schedule(run: () => void, delay: number) {
      const task = { run, delay, active: true }; tasks.push(task);
      return () => { task.active = false; };
    } },
    tasks,
    fire(delay: number) {
      const task = tasks.find(value => value.active && value.delay === delay);
      if (!task) throw new Error('No scheduled task');
      task.active = false; task.run();
    }
  };
}

it('sends product latency with authentication and disposes without recursive delivery', async () => {
  const time = runtime();
  const requests: Request[] = [];
  let delivered!: () => void;
  const delivery = new Promise<void>(resolve => { delivered = resolve; });
  const session = createWebPerformanceSession({
    enabled: true, baseUrl: 'https://api.example.test', tokenProvider: () => 'private-session', ...time,
    fetch: async input => {
      const request = new Request(input); requests.push(request);
      if (new URL(request.url).pathname === '/client-telemetry') {
        expect(request.headers.get('Authorization')).toBe('Bearer private-session');
        expect(await request.json()).toEqual({ measurements: [{ platform: 'web', operation: 'request', surface: 'application', variant: 'none', outcome: 'success', durationMs: 25 }] });
        delivered();
        return Response.json({ data: { accepted: 1 }, meta: {} });
      }
      time.advance(25);
      return new Response('product');
    }
  });
  expect(await (await session.fetch('https://api.example.test/private')).text()).toBe('product');
  time.fire(5000);
  await delivery;
  session.dispose();
  expect(requests).toHaveLength(2);
  expect(requests[1].signal.aborted).toBe(true);
  await expect.poll(() => time.tasks.some(task => task.active)).toBe(false);
});

it('disabled sessions preserve fetch identity and never schedule image delivery', () => {
  const time = runtime();
  const fetchImpl: typeof fetch = async () => new Response();
  const session = createWebPerformanceSession({ enabled: false, baseUrl: 'https://api.example.test', tokenProvider: () => null, fetch: fetchImpl, ...time });
  expect(session.fetch).toBe(fetchImpl);
  session.observer.start({ operation: 'image', surface: 'gallery', variant: 'medium' })('success');
  session.dispose();
  expect(time.tasks).toHaveLength(0);
});

it('clears pending images on session disposal and ignores late image completions', () => {
  const time = runtime();
  const session = createWebPerformanceSession({ enabled: true, baseUrl: 'https://api.example.test', tokenProvider: () => null, fetch: async () => new Response(), ...time });
  const late = session.observer.start({ operation: 'image', surface: 'detail', variant: 'large' });
  session.observer.start({ operation: 'image', surface: 'list', variant: 'small' })('success');
  session.dispose();
  late('success');
  expect(time.tasks.some(task => task.active)).toBe(false);
});
