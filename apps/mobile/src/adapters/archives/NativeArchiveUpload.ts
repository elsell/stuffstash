import { assertReadActive } from '../../application/shared/ReadRequest';
export interface ArchiveUploadModule {
  upload(id: string, url: string, fileURI: string, headers: Record<string, string>): Promise<{ status: number; body: string }>;
  cancel(id: string): Promise<void>;
}

/** Native module owns file streaming, deadlines, redirect refusal and response size. */
export class NativeArchiveUpload {
  constructor(private readonly native: ArchiveUploadModule, private readonly newID: () => string) {}
  async send(uri: string, request: Request): Promise<Response> {
    assertReadActive(request.signal);
    if (request.method !== 'POST') throw new Error('Invalid archive upload method.');
    const id = this.newID();
    const headers: Record<string, string> = {};
    request.headers.forEach((value, key) => { headers[key] = value; });
    const abort = () => { void this.native.cancel(id).catch(() => {}); };
    // Both native calls enqueue on the same serial queue. Cancellation cannot
    // overtake registration even if it arrives before upload's native dispatch.
    const operation = this.native.upload(id, request.url, uri, headers);
    request.signal?.addEventListener('abort', abort, { once: true });
    if (request.signal?.aborted) abort();
    try {
      const result = await operation;
      assertReadActive(request.signal);
      if (result.status >= 300 && result.status < 400) throw new Error('Archive redirects are not permitted.');
      if (new TextEncoder().encode(result.body).byteLength > 65536) throw new Error('Archive response exceeded its size limit.');
      return new Response(result.body, { status: result.status, headers: { 'Content-Type': 'application/json' } });
    } catch (error) { assertReadActive(request.signal); throw error; }
    finally { request.signal?.removeEventListener('abort', abort); }
  }
}
