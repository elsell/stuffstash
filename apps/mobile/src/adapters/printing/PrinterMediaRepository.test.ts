import { expect, it } from 'vitest';
import { LabelsClient, PrintingClient } from '@stuff-stash/api-client';
import { PrintingFake } from '../../test-support/PrintingFake';
import { ApiPrintingRepository } from './ApiPrintingRepository';

class PrinterHTTPFake {
  printer = { ...new PrintingFake().printer, retired: true };
  fetch: typeof fetch = async (input, init) => {
    const request = new Request(input, init); const path = new URL(request.url).pathname;
    const deny = (status: number) => Response.json({ error: { code: 'denied', message: 'Denied' } }, { status });
    const principal = request.headers.get('Authorization'); if (!principal) return deny(401);
    if (!path.startsWith('/tenants/tenant/inventories/inventory/')) return deny(404);
    if (request.method === 'GET' && path.endsWith('/printer-profiles')) {
      return Response.json({ data: [{ adapterId: this.printer.adapterId, media: [{ presetId: 'brother-ql800-29x90', version: 1, name: '29 × 90 mm', widthMicrometers: 29000, heightMicrometers: 89800 }] },
        { adapterId: 'different-adapter', media: [{ presetId: 'other-stock', version: 1, name: 'Different stock' }] }], meta: {} });
    }
    if (path !== '/tenants/tenant/inventories/inventory/printers/printer' || request.method !== 'PATCH') return deny(404);
    if (principal !== 'Bearer owner') return deny(403);
    const body = await request.json();
    if (body.revision !== this.printer.revision) return deny(409);
    if (body.presetId !== 'brother-ql800-29x90' || body.presetVersion !== 1) return deny(422);
    this.printer = { ...this.printer, name: body.name ?? this.printer.name, retired: body.retired ?? this.printer.retired, revision: this.printer.revision + 1 };
    const media = this.printer.media;
    return Response.json({ data: { ...this.printer, media: { ...media, name: this.printer.mediaName, marginsMicrometers: media.margins, resolutionDpi: media.resolutionDPI } }, meta: {} });
  };
}
it('uses adapter-compatible catalog media and preserves scope, revision, and other registration fields', async () => {
  const fake = new PrinterHTTPFake(); let token: string | null = 'owner';
  const options = { baseUrl: 'https://stash.example', tokenProvider: () => token, fetch: fake.fetch };
  const repository = new ApiPrintingRepository(new PrintingClient(options), new LabelsClient(options));
  const scope = { tenantId: 'tenant', inventoryId: 'inventory' }; const printer = fake.printer;
  const presets = await repository.mediaPresets(scope, printer, new AbortController().signal);
  expect(presets).toEqual([{ id: 'brother-ql800-29x90', version: 1, name: '29 × 90 mm', widthMicrometers: 29000, heightMicrometers: 89800 }]);
  const saved = await repository.configurePrinter(scope, printer, presets[0]);
  expect(saved.revision).toBe(2); expect(saved.retired).toBe(true); expect(saved.name).toBe(printer.name);
  expect(saved.adapterId).toBe(printer.adapterId); expect(saved.media.presetId).toBe(presets[0].id);
  await expect(repository.configurePrinter(scope, printer, presets[0])).rejects.toMatchObject({ status: 409 });
  for (token of [null, 'viewer', 'connector']) await expect(repository.configurePrinter(scope, saved, presets[0])).rejects.toThrow();
  token = 'owner';
  for (const wrong of [{ ...scope, inventoryId: 'foreign' }, { ...scope, tenantId: 'foreign' }]) await expect(repository.configurePrinter(wrong, saved, presets[0])).rejects.toMatchObject({ status: 404 });
  await expect(repository.configurePrinter(scope, saved, { ...presets[0], id: 'unsupported' })).rejects.toMatchObject({ status: 422 });
  expect(fake.printer.revision).toBe(2);
});
