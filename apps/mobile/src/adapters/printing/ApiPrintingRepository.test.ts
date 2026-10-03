import { expect, it } from 'vitest';
import { LabelsClient, PrintingClient } from '@stuff-stash/api-client';
import { ApiPrintingRepository } from './ApiPrintingRepository';

class JobHTTPFake {
  job = { id: 'job', printerId: 'printer', status: 'uncertain', revision: 3, copies: 1,
    attempts: [{ id: 'old', completedCopies: 0, idleConfirmedAt: '2026-10-03T12:00:00Z' }, { id: 'latest', completedCopies: 0, idleConfirmedAt: undefined as string | undefined }],
    resolution: undefined as { reportedOutcome: string } | undefined };
  fetch: typeof fetch = async (input, init) => {
    const request = new Request(input, init); const path = new URL(request.url).pathname;
    const denied = (status: number) => Response.json({ error: { code: 'denied', message: 'Denied' } }, { status });
    if (request.headers.get('Authorization') !== 'Bearer editor') return denied(403);
    if (!path.startsWith('/tenants/tenant/inventories/inventory/print-jobs/job')) return denied(404);
    if (request.method === 'POST') {
      const body = await request.json();
      if (path !== '/tenants/tenant/inventories/inventory/print-jobs/job/resolution' || body.revision !== this.job.revision || body.acknowledgeUncertainty !== true || !this.job.attempts.at(-1)?.idleConfirmedAt) return denied(409);
      this.job = { ...this.job, revision: this.job.revision + 1, status: 'failed', resolution: { reportedOutcome: body.reportedOutcome } };
    }
    return Response.json({ data: this.job, meta: {} });
  };
}
it('uses only latest idle evidence and preserves uncertainty after an authorized revisioned resolution', async () => {
  const fake = new JobHTTPFake(); let token = 'editor';
  const options = { baseUrl: 'https://stash.example', tokenProvider: () => token, fetch: fake.fetch };
  const repository = new ApiPrintingRepository(new PrintingClient(options), new LabelsClient(options));
  const scope = { tenantId: 'tenant', inventoryId: 'inventory' }; const signal = new AbortController().signal;
  let job = await repository.job(scope, 'job', signal); expect(job.idleConfirmed).toBe(false);
  await expect(repository.resolve(scope, job, 'printed')).rejects.toMatchObject({ status: 409 });
  fake.job.attempts[1].idleConfirmedAt = '2026-10-03T12:05:00Z';
  job = await repository.job(scope, 'job', signal); expect(job.idleConfirmed).toBe(true);
  token = 'viewer'; await expect(repository.resolve(scope, job, 'printed')).rejects.toMatchObject({ status: 403 });
  token = 'editor'; await expect(repository.resolve({ ...scope, inventoryId: 'foreign' }, job, 'printed')).rejects.toMatchObject({ status: 404 });
  const resolved = await repository.resolve(scope, job, 'printed');
  expect(resolved.status).toBe('failed'); expect(resolved.completedCopies).toBe(0); expect(resolved.resolution?.reportedOutcome).toBe('printed');
  await expect(repository.resolve(scope, job, 'unknown')).rejects.toMatchObject({ status: 409 });
});
