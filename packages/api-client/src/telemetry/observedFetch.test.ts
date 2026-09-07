import { expect, it } from 'vitest';
import { createObservedFetch } from './observedFetch';
import type { PerformanceContext, PerformanceOutcome } from './performanceReporter';

const context: PerformanceContext = { platform: 'web', operation: 'request', surface: 'detail', variant: 'none' };

it('preserves request identity and streaming response without consuming its body', async () => {
  const request = new Request('https://api.example.test/private?token=secret');
  const init: RequestInit = { headers: { 'X-Test': 'private' } };
  const response = new Response(new ReadableStream({ start(controller) { controller.enqueue(new TextEncoder().encode('body')); } }));
  const outcomes: PerformanceOutcome[] = [];
  const contexts: PerformanceContext[] = [];
  const observed = createObservedFetch(async (input, options) => {
    expect(input).toBe(request);
    expect(options).toBe(init);
    expect(contexts).toEqual([context]);
    return response;
  }, { start(value) { contexts.push(value); return outcome => { outcomes.push(outcome); }; } }, context);
  expect(await observed(request, init)).toBe(response);
  expect(response.bodyUsed).toBe(false);
  expect(outcomes).toEqual(['success']);
  await response.body?.cancel();
});

it('reports HTTP failure while returning the original error response', async () => {
  const response = new Response(null, { status: 503 });
  const outcomes: PerformanceOutcome[] = [];
  const observed = createObservedFetch(async () => response, { start: () => outcome => { outcomes.push(outcome); } }, context);
  expect(await observed('https://api.example.test')).toBe(response);
  expect(outcomes).toEqual(['failure']);
});

it.each(['request', 'init'] as const)('preserves cancellation and original rejection from %s signal', async source => {
  const controller = new AbortController();
  const reason = new Error('private cancellation reason');
  const outcomes: PerformanceOutcome[] = [];
  const observed = createObservedFetch(async () => { controller.abort(reason); throw reason; }, {
    start: () => outcome => { outcomes.push(outcome); }
  }, context);
  const request = new Request('https://api.example.test', source === 'request' ? { signal: controller.signal } : undefined);
  await expect(observed(request, source === 'init' ? { signal: controller.signal } : undefined)).rejects.toBe(reason);
  expect(outcomes).toEqual(['cancelled']);
});

it('reports transport failure without replacing its error', async () => {
  const reason = new Error('private transport message');
  const outcomes: PerformanceOutcome[] = [];
  const observed = createObservedFetch(async () => { throw reason; }, { start: () => outcome => { outcomes.push(outcome); } }, context);
  await expect(observed('https://api.example.test')).rejects.toBe(reason);
  expect(outcomes).toEqual(['failure']);
});

it.each(['start', 'finish'] as const)('isolates observer %s failure from requests', async stage => {
  const response = new Response('unchanged');
  const observer = { start() {
    if (stage === 'start') throw new Error('observer failed');
    return () => { throw new Error('observer failed'); };
  } };
  expect(await createObservedFetch(async () => response, observer, context)('https://api.example.test')).toBe(response);
  const reason = new Error('original request failure');
  await expect(createObservedFetch(async () => { throw reason; }, observer, context)('https://api.example.test')).rejects.toBe(reason);
});

it('honors an explicitly detached request signal when classifying a failure', async () => {
  const controller = new AbortController();
  controller.abort();
  const request = new Request('https://api.example.test', { signal: controller.signal });
  const reason = new Error('independent transport failure');
  const outcomes: PerformanceOutcome[] = [];
  const observed = createObservedFetch(async (_input, init) => {
    expect(init?.signal).toBeNull();
    throw reason;
  }, { start: () => outcome => { outcomes.push(outcome); } }, context);
  await expect(observed(request, { signal: null })).rejects.toBe(reason);
  expect(outcomes).toEqual(['failure']);
});
