/** Bridges archive control requests to Expo while leaving responses streaming.
 * Restore file uploads use NativeArchiveUpload and never enter this adapter.
 */
export function createExpoArchiveFetch(nativeFetch: typeof fetch): typeof fetch {
  return async (input, init) => {
    const request = new Request(input, init);
    const headers: Record<string, string> = {};
    request.headers.forEach((value, name) => { headers[name] = value; });
    // React Native's Request has text(), but no public body stream. Only small
    // JSON control requests pass through this path, never uploaded ZIP files.
    const body = request.method === 'GET' || request.method === 'HEAD'
      ? undefined : await request.text();
    return nativeFetch(request.url, {
      method: request.method,
      headers,
      body,
      signal: request.signal,
      credentials: request.credentials,
      redirect: 'error'
    });
  };
}
