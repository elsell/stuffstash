import type { ArchiveClient, ArchiveJob as JobDTO } from '@stuff-stash/api-client';
import type { ArchiveJob, ArchiveState, InventoryArchiveRepository, ArchiveScope } from '$lib/ports/inventoryArchive';

export class ApiInventoryArchiveRepository implements InventoryArchiveRepository {
  constructor(private readonly client: ArchiveClient) {}
  async list(scope: ArchiveScope, after: string | undefined, signal: AbortSignal) {
    const page = await this.client.list(scope, after, signal);
    return { jobs: (page.data ?? []).map(job), nextCursor: page.meta.pagination?.nextCursor ?? undefined };
  }
  async get(scope: ArchiveScope, id: string, signal: AbortSignal) { return job(await this.client.get(scope, id, signal)); }
  async create(scope: Required<ArchiveScope>, key: string, selection: { photos: boolean; otherFiles: boolean }, signal: AbortSignal) { return job(await this.client.create(scope, key, selection, signal)); }
  async upload(tenantId: string, key: string, file: Blob, signal: AbortSignal) { return job(await this.client.upload(tenantId, key, file, signal)); }
  async preview(tenantId: string, id: string, signal: AbortSignal) { const p = await this.client.preview(tenantId, id, signal); return { ...p, keyRemappings: p.keyRemappings ?? [] }; }
  async approve(tenantId: string, id: string, name: string, signal: AbortSignal) { return job(await this.client.approve(tenantId, id, name, signal)); }
  async cancel(scope: ArchiveScope, id: string, signal: AbortSignal) { return job(await this.client.cancel(scope, id, signal)); }
  async retry(scope: ArchiveScope, id: string, signal: AbortSignal) { return job(await this.client.retry(scope, id, signal)); }
  download(scope: ArchiveScope, id: string, signal: AbortSignal) { return this.client.download(scope, id, signal); }
}
function job(dto: JobDTO): ArchiveJob {
  const states: ArchiveState[] = ['queued', 'running', 'awaiting_approval', 'ready', 'failed', 'cancelled', 'expired'];
  if (!states.includes(dto.state as ArchiveState) || (dto.kind !== 'export' && dto.kind !== 'restore')) throw new Error('Unsupported archive job response.');
  return { id: dto.id, kind: dto.kind, state: dto.state as ArchiveState, phase: dto.phase, createdAt: dto.createdAt, expiresAt: dto.expiresAt, destinationInventoryId: dto.destinationInventoryId, photos: dto.photos, otherFiles: dto.otherFiles, failure: dto.failure };
}
