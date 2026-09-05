import createClient, { type Client } from 'openapi-fetch';
import type { paths } from './generated/schema';
import type { StuffStashClientOptions } from './stuffStashClient';

/** Generated transport for infrastructure adapters; map envelopes at that boundary. */
export function createAuthenticatedTransport(options: StuffStashClientOptions): Client<paths> {
  const baseUrl = options.baseUrl.replace(/\/+$/, '');
  const client = createClient<paths>({ baseUrl, fetch: options.fetch, redirect: 'error' });
  client.use({
    async onRequest({ request, options: requestOptions }) {
      if (requestOptions.baseUrl.replace(/\/+$/, '') !== baseUrl) {
        throw new Error('API destination overrides are not permitted.');
      }
      request.signal.throwIfAborted();
      const token = await resolveSessionToken(options.tokenProvider, request.signal);
      request.signal.throwIfAborted();
      request.headers.delete('Authorization');
      if (token) request.headers.set('Authorization', `Bearer ${token}`);
      return new Request(request, { redirect: 'error' });
    }
  });
  return client;
}

function resolveSessionToken(
  provider: StuffStashClientOptions['tokenProvider'], signal: AbortSignal
): Promise<string | null> {
  return new Promise((resolve, reject) => {
    const abort = () => reject(signal.reason);
    signal.addEventListener('abort', abort, { once: true });
    // Recheck after subscription so cancellation cannot fall between the checks.
    if (signal.aborted) { abort(); signal.removeEventListener('abort', abort); return; }
    Promise.resolve().then(() => {
      signal.throwIfAborted();
      return provider();
    }).then(resolve, reject).finally(() => {
      signal.removeEventListener('abort', abort);
    });
  });
}
