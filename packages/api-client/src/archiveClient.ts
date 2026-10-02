import { createAuthenticatedTransport } from './authenticatedTransport';
import { StuffStashAPIError, type StuffStashClientOptions } from './stuffStashClient';
import type { components } from './generated/schema';

export type ArchiveJob = components['schemas']['ArchiveJob'];
export type ArchivePreview = components['schemas']['ArchivePreview'];
export type ArchiveScope = { tenantId: string; inventoryId?: string };
const jobs = '/tenants/{tenantId}/archive-jobs';
const jobPath = '/tenants/{tenantId}/archive-jobs/{jobId}';

/** Transport only: clients map job DTOs to their own presentation models. */
export class ArchiveClient {
  private readonly transport;
  constructor(options: StuffStashClientOptions) { this.transport = createAuthenticatedTransport(options); }

  async create(scope: Required<ArchiveScope>, requestKey: string, selection: { photos: boolean; otherFiles: boolean }, signal?: AbortSignal): Promise<ArchiveJob> {
    const result = await this.transport.POST(jobs, { params: { path: { tenantId: scope.tenantId }, header: { 'Idempotency-Key': requestKey } }, body: { inventoryId: scope.inventoryId, ...selection }, signal });
    return unwrap(result).data;
  }
  async list(scope: ArchiveScope, after?: string, signal?: AbortSignal) {
    const result = await this.transport.GET(jobs, { params: { path: { tenantId: scope.tenantId }, query: { inventoryId: scope.inventoryId, after, limit: 25 } }, signal });
    return unwrap(result);
  }
  async get(scope: ArchiveScope, jobId: string, signal?: AbortSignal): Promise<ArchiveJob> {
    return unwrap(await this.transport.GET(jobPath, { params: parameters(scope, jobId), signal })).data;
  }
  async cancel(scope: ArchiveScope, jobId: string, signal?: AbortSignal): Promise<ArchiveJob> {
    return unwrap(await this.transport.DELETE(jobPath, { params: parameters(scope, jobId), signal })).data;
  }
  async retry(scope: ArchiveScope, jobId: string, signal?: AbortSignal): Promise<ArchiveJob> {
    return unwrap(await this.transport.POST(`${jobPath}/retry`, { params: parameters(scope, jobId), signal })).data;
  }
  async preview(tenantId: string, jobId: string, signal?: AbortSignal): Promise<ArchivePreview> {
    return unwrap(await this.transport.GET(`${jobPath}/preview`, { params: parameters({ tenantId }, jobId), signal })).data;
  }
  async approve(tenantId: string, jobId: string, name: string, signal?: AbortSignal): Promise<ArchiveJob> {
    return unwrap(await this.transport.POST(`${jobPath}/approve`, { params: parameters({ tenantId }, jobId), body: { name }, signal })).data;
  }
  async upload(tenantId: string, requestKey: string, content: Blob, signal?: AbortSignal): Promise<ArchiveJob> {
    const result = await this.transport.POST('/tenants/{tenantId}/archive-restores', {
      params: { path: { tenantId }, header: { 'Idempotency-Key': requestKey } },
      headers: { 'Content-Type': 'application/zip' },
      // OpenAPI represents binary as string. The serializer supplies the actual
      // file body without base64 conversion or loading it into JSON.
      body: '', bodySerializer: () => content, signal
    });
    return unwrap(result).data;
  }
  async download(scope: ArchiveScope, jobId: string, signal?: AbortSignal): Promise<ReadableStream<Uint8Array>> {
    const result = await this.transport.GET(`${jobPath}/content`, { params: parameters(scope, jobId), parseAs: 'stream', signal });
    return unwrap(result);
  }
}

function parameters(scope: ArchiveScope, jobId: string) {
  return { path: { tenantId: scope.tenantId, jobId }, query: { inventoryId: scope.inventoryId } };
}
function unwrap<T>(result: { data?: T; error?: unknown; response: Response }): NonNullable<T> {
  if (!result.response.ok || result.error) {
    const envelope = result.error as { error?: { code?: string; message?: string } } | undefined;
    throw new StuffStashAPIError(result.response.status, envelope?.error?.code ?? 'archive_failed', envelope?.error?.message ?? 'Could not complete the archive request.');
  }
  if (result.data === undefined || result.data === null) throw new Error('The server did not return the archive response.');
  return result.data;
}
