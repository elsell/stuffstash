export type ArchiveScope = { tenantId: string; inventoryId?: string };
export type ArchiveState = 'queued' | 'running' | 'awaiting_approval' | 'ready' | 'failed' | 'cancelled' | 'expired';
export type ArchiveJob = {
  id: string; kind: 'export' | 'restore'; state: ArchiveState; phase: string;
  createdAt: string; expiresAt: string; destinationInventoryId?: string;
  photos: boolean; otherFiles: boolean; failure?: string;
};
export type ArchivePreview = {
  inventoryName: string; assets: number; tags: number; customAssetTypes: number;
  customFields: number; photos: number; otherFiles: number; omittedAttachments: number;
  keyRemappings: { family: string; sourceKey: string; destinationKey: string }[];
};
export interface InventoryArchiveRepository {
  list(scope: ArchiveScope, after: string | undefined, signal: AbortSignal): Promise<{ jobs: ArchiveJob[]; nextCursor?: string }>;
  get(scope: ArchiveScope, id: string, signal: AbortSignal): Promise<ArchiveJob>;
  create(scope: Required<ArchiveScope>, key: string, selection: { photos: boolean; otherFiles: boolean }, signal: AbortSignal): Promise<ArchiveJob>;
  upload(tenantId: string, key: string, file: Blob, signal: AbortSignal): Promise<ArchiveJob>;
  preview(tenantId: string, id: string, signal: AbortSignal): Promise<ArchivePreview>;
  approve(tenantId: string, id: string, name: string, signal: AbortSignal): Promise<ArchiveJob>;
  cancel(scope: ArchiveScope, id: string, signal: AbortSignal): Promise<ArchiveJob>;
  retry(scope: ArchiveScope, id: string, signal: AbortSignal): Promise<ArchiveJob>;
  download(scope: ArchiveScope, id: string, signal: AbortSignal): Promise<ReadableStream<Uint8Array>>;
}
export interface ArchiveFileDelivery { save(content: () => Promise<ReadableStream<Uint8Array>>, signal: AbortSignal): Promise<void>; }
export type InventoryArchiveWorkspace = { repository: InventoryArchiveRepository; files: ArchiveFileDelivery };
export const inventoryArchiveContext = Symbol('inventoryArchive');
