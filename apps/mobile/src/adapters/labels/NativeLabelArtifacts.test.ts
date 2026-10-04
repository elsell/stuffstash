import { expect, it } from 'vitest';
import { LabelsClient, PrintingClient } from '@stuff-stash/api-client';
import { ApiLabelRepository } from './ApiLabelRepository';
import { ApiPrintingRepository } from '../printing/ApiPrintingRepository';

const png = new Uint8Array([137, 80, 78, 71, 13, 10, 26, 10]);
const pdf = new TextEncoder().encode('%PDF-1.4\n');
// Expo FetchResponse.blob constructs new Blob([ArrayBuffer]); RN rejects it.
// Preserve the real response byte reader and body consumption behavior.
class ExpoArtifactResponse extends Response {
  override async blob(): Promise<Blob> {
    await this.arrayBuffer();
    throw new Error("Creating blobs from 'ArrayBuffer' and 'ArrayBufferView' are not supported");
  }
}
class LabelArtifactServer {
  private artifacts = new Map<string, 'png' | 'pdf'>();
  fetch: typeof fetch = async (input, init) => {
    const request = new Request(input, init); const path = new URL(request.url).pathname;
    if (request.headers.get('Authorization') !== 'Bearer editor') return Response.json({ error: { code: 'unauthorized', message: 'Denied' } }, { status: 401 });
    if (!path.startsWith('/tenants/tenant/inventories/inventory/')) return Response.json({ error: { code: 'not_found', message: 'Denied' } }, { status: 404 });
    if (request.method === 'POST' && path.endsWith('/label')) return Response.json({ data: { labelId: 'label' } });
    if (request.method === 'POST' && path.endsWith('/label-renders')) {
      const body = await request.json(); const id = String(this.artifacts.size + 1);
      this.artifacts.set(id, body.format);
      return Response.json({ data: { id, widthPixels: 306, heightPixels: 991, displayRotation: 90, selectionFingerprint: 'selection' } });
    }
    const id = path.match(/\/label-renders\/([^/]+)\/content$/)?.[1];
    const format = id && this.artifacts.get(id);
    if (!format) return new Response('', { status: 404 });
    return new ExpoArtifactResponse(format === 'png' ? png : pdf, { headers: { 'Content-Type': format === 'png' ? 'image/png' : 'application/pdf' } });
  };
}
const scope = { tenantId: 'tenant', inventoryId: 'inventory' };
const template = { id: 'qr-title', version: 1, name: 'QR and title', showReference: true, supportsReference: true };
const media = { presetId: 'brother-ql800-29x90', version: 1, widthMicrometers: 29000, heightMicrometers: 89800,
  margins: { left: 1000, right: 1000, top: 1000, bottom: 1000 }, resolutionDPI: 300, rasterWidth: 306, rasterHeight: 991,
  orientation: 'portrait', colorMode: 'monochrome', cutPolicy: 'cut', displayRotation: 90 };
it('reads PNG, PDF and printer previews through Expo-compatible bytes without Blob conversion', async () => {
  const server = new LabelArtifactServer(); const options = { baseUrl: 'https://stash.example', tokenProvider: () => 'editor', fetch: server.fetch };
  const labels = new LabelsClient(options); const repository = new ApiLabelRepository(labels); const signal = new AbortController().signal;
  const selection = { media, template, showReference: true };
  expect((await repository.render(scope, 'asset', selection, 'png', signal)).bytes).toEqual(png);
  expect((await repository.render(scope, 'asset', selection, 'pdf', signal)).bytes).toEqual(pdf);
  const printing = new ApiPrintingRepository(new PrintingClient(options), labels);
  const printer = { id: 'printer', adapterId: 'brother', name: 'Brother', retired: false, readiness: 'ready', revision: 1, mediaFingerprint: 'media', mediaName: '29 × 90 mm', media };
  const preview = await printing.preview(scope, 'asset', printer, template, signal);
  expect(preview.file.bytes).toEqual(png); expect(preview.fingerprint).toBe('selection');
  await expect(repository.render({ ...scope, inventoryId: 'foreign' }, 'asset', selection, 'png', signal)).rejects.toMatchObject({ status: 404 });
});
