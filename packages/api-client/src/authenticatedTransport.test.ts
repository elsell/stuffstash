import { describe, expect, it } from 'vitest';
import { createAuthenticatedTransport } from './authenticatedTransport';

describe('authenticated generated transport', () => {
  it('uses the current session token for each request and preserves scoped paths', async () => {
    const requests: Request[] = [];
    let token: string | null = 'first-session';
    const client = createAuthenticatedTransport({
      baseUrl: 'https://api.example.test/', tokenProvider: () => token,
      fetch: async (input, init) => {
        requests.push(new Request(input, init));
        return Response.json({ data: null, meta: {} });
      }
    });
    const read = () => client.GET('/tenants/{tenantId}/conversation-workflow-selection', {
      params: { path: { tenantId: 'home' } }
    });
    await read(); token = 'second-session'; await read(); token = null; await read();
    expect(requests.map(request => request.headers.get('Authorization')))
      .toEqual(['Bearer first-session', 'Bearer second-session', null]);
    expect(requests.every(request => request.url === 'https://api.example.test/tenants/home/conversation-workflow-selection')).toBe(true);
    expect(requests.every(request => request.redirect === 'error')).toBe(true);
  });

  it('does not dispatch after cancellation while session resolution is pending', async () => {
    let resolveToken!: (token: string) => void;
    const token = new Promise<string>(resolve => { resolveToken = resolve; });
    let dispatched = false;
    const client = createAuthenticatedTransport({
      baseUrl: 'https://api.example.test', tokenProvider: () => token,
      fetch: async () => { dispatched = true; return Response.json({ data: null }); }
    });
    const controller = new AbortController();
    const pending = client.GET('/tenants/{tenantId}/conversation-workflow-selection', {
      signal: controller.signal, params: { path: { tenantId: 'home' } }
    });
    controller.abort();
    await expect(pending).rejects.toMatchObject({ name: 'AbortError' });
    resolveToken('session');
    await Promise.resolve();
    expect(dispatched).toBe(false);
  });

  it('rejects a destination override before reading credentials or dispatching', async () => {
    let credentialsRead = false;
    let dispatched = false;
    const client = createAuthenticatedTransport({
      baseUrl: 'https://api.example.test',
      tokenProvider: () => { credentialsRead = true; return 'private'; },
      fetch: async () => { dispatched = true; return Response.json({}); }
    });
    await expect(client.GET('/tenants/{tenantId}/conversation-workflow-selection', {
      baseUrl: 'https://other.example.test', params: { path: { tenantId: 'home' } }
    })).rejects.toThrow('API destination');
    expect(credentialsRead).toBe(false);
    expect(dispatched).toBe(false);
  });
});
