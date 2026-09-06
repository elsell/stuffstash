import { createRequire } from 'node:module';
import { expect, it } from 'vitest';
import { createAuthenticatedTransport } from '@stuff-stash/api-client';

const require = createRequire(import.meta.url);
const nativeRequire = createRequire(require.resolve('react-native/package.json'));
// Exercise the actual locked Request polyfill used by the native fetch runtime.
globalThis.Request = nativeRequire('whatwg-fetch').Request;

it('sends an authenticated measurement with the native Request and AbortSignal', async () => {
  const requests: Request[] = [];
  const client = createAuthenticatedTransport({ baseUrl: 'https://api.example.test', tokenProvider: () => 'session', fetch: async input => {
    const request = input as Request; requests.push(request);
    expect(request.headers.get('Authorization')).toBe('Bearer session');
    return Response.json({ data: { accepted: 1 }, meta: {} });
  } });
  const response = await client.POST('/client-telemetry', { signal: new AbortController().signal, body: { measurements: [{ platform: 'ios', operation: 'image', surface: 'detail', variant: 'large', outcome: 'success', durationMs: 10 }] } });
  expect(response.data?.data.accepted).toBe(1);
  expect(requests).toHaveLength(1);
});

it('does not send after native cancellation during token resolution', async () => {
  let started!: () => void;
  const resolving = new Promise<void>(resolve => { started = resolve; });
  let resolveToken!: (token: string) => void;
  const token = new Promise<string>(resolve => { resolveToken = resolve; });
  let requests = 0;
  const client = createAuthenticatedTransport({ baseUrl: 'https://api.example.test', tokenProvider: () => { started(); return token; }, fetch: async () => { requests++; return new Response(); } });
  const controller = new AbortController();
  const pending = client.POST('/client-telemetry', { signal: controller.signal, body: { measurements: [{ platform: 'ios', operation: 'image', surface: 'detail', variant: 'large', outcome: 'success', durationMs: 10 }] } });
  await resolving;
  controller.abort();
  await expect(pending).rejects.toMatchObject({ name: 'AbortError' });
  resolveToken('late-session');
  await Promise.resolve();
  expect(requests).toBe(0);
});
