import { expect, it } from 'vitest';
import { createApiPerformanceReporter } from './apiPerformanceReporter';

it('uses generated authenticated transport for reporting and keeps delivery outside measurement', async () => {
  const requests: Request[] = [];
  const reporter = createApiPerformanceReporter({
    connection: { baseUrl: 'https://api.example.test', tokenProvider: () => 'test-session', fetch: async input => {
      const request = new Request(input);
      requests.push(request);
      return Response.json({ data: { accepted: 1 }, meta: {} });
    } },
    clock: { now: () => 0 }, scheduler: { schedule: () => () => {} }
  });
  reporter.record({platform:'web',operation:'image',surface:'gallery',variant:'medium',outcome:'success',durationMs:12});
  await reporter.flush();
  await reporter.flush();
  expect(requests).toHaveLength(1);
  expect(requests[0].url).toBe('https://api.example.test/client-telemetry');
  expect(requests[0].headers.get('Authorization')).toBe('Bearer test-session');
  expect(requests[0].redirect).toBe('error');
  expect(await requests[0].json()).toEqual({ measurements: [{platform:'web',operation:'image',surface:'gallery',variant:'medium',outcome:'success',durationMs:12}] });
  reporter.dispose();
});
