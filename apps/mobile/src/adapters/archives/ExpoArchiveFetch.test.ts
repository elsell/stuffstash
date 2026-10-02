import { createRequire } from 'node:module';
import { expect, it } from 'vitest';
import { ArchiveClient } from '@stuff-stash/api-client';
import { createExpoArchiveFetch } from './ExpoArchiveFetch';

// Use the same Request implementation installed for React Native, not Node's
// streaming Request, which concealed this incompatibility in existing tests.
const require = createRequire(import.meta.url);
const nativeRequire = createRequire(require.resolve('react-native/package.json'));
const native = nativeRequire('whatwg-fetch') as { Request: typeof Request; Headers: typeof Headers };

it('preserves archive JSON, credentials and cancellation across the Expo boundary', async () => {
  const previous = { Request: globalThis.Request, Headers: globalThis.Headers };
  globalThis.Request = native.Request;
  globalThis.Headers = native.Headers;
  try {
    const requests: { url: string; method: string; body: unknown; headers: Headers; redirect: unknown; signal: unknown }[] = [];
    const nativeFetch: typeof fetch = async (input, init) => {
      const request = typeof input === 'string' ? undefined : input as Request;
      requests.push({ url: request?.url ?? String(input), method: init?.method ?? request?.method ?? 'GET',
        body: init?.body ?? request?.body, headers: new Headers(init?.headers ?? request?.headers),
        redirect: init?.redirect ?? request?.redirect, signal: init?.signal ?? request?.signal });
      return new Response(JSON.stringify({ data: { id: 'job' }, meta: {} }), { status: 200 });
    };
    const client = new ArchiveClient({ baseUrl: 'https://api.example.test', tokenProvider: () => 'test-token', fetch: createExpoArchiveFetch(nativeFetch) });
    const controller = new AbortController();
    await client.create({ tenantId: 'tenant', inventoryId: 'inventory' }, 'request-key', { photos: true, otherFiles: false }, controller.signal);
    await client.approve('tenant', 'job', 'Restored home', controller.signal);
    expect(JSON.parse(String(requests[0].body))).toEqual({ inventoryId: 'inventory', photos: true, otherFiles: false });
    expect(JSON.parse(String(requests[1].body))).toEqual({ name: 'Restored home' });
    expect(requests[0].headers.get('Idempotency-Key')).toBe('request-key');
    for (const request of requests) {
      expect(request.headers.get('Authorization')).toBe('Bearer test-token');
      expect(request.method).toBe('POST');
      expect(request.redirect).toBe('error');
      expect(request.signal).toBe(controller.signal);
    }
  } finally { globalThis.Request = previous.Request; globalThis.Headers = previous.Headers; }
});

it('leaves download responses streaming without reading their bodies', async () => {
  let reads = 0;
  const stream = new ReadableStream<Uint8Array>({ pull() { reads++; } }, { highWaterMark: 0 });
  const response = new Response(stream);
  const nativeFetch: typeof fetch = async (input, init) => {
    expect(String(input)).toBe('https://api.example.test/archive');
    expect(init?.method).toBe('GET');
    expect(init?.body).toBeUndefined();
    expect(init?.redirect).toBe('error');
    return response;
  };
  const result = await createExpoArchiveFetch(nativeFetch)(new Request('https://api.example.test/archive'));
  expect(result).toBe(response);
  expect(result.body).toBe(stream);
  expect(reads).toBe(0);
  await stream.cancel();
});
