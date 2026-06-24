import { StuffStashAPIError, StuffStashClient } from '@stuff-stash/api-client';
import type { RuntimeConfig } from '$lib/runtimeConfig';
import type { TokenProvider } from '@stuff-stash/api-client';
import type {
  AddAssetDraft,
  Asset,
  AssetAttachment,
  AssetLifecycleFilter,
  SearchRequest,
  SearchResult,
  SelectedPhoto,
  UpdateAssetDraft,
  WorkspaceData
} from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';
import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
import { mapAsset, mapAttachment, mapCapability, mapInventory, mapPrincipal, mapSearchResult, mapTenant } from './inventoryMapper';

export class StuffStashInventoryRepository implements InventoryRepository {
  private readonly client: StuffStashClient;
  private readonly uploadFetch: typeof fetch;
  private readonly config: RuntimeConfig;
  private selectedTenantId = readSessionValue('stuffstash.selectedTenantId');
  private selectedInventoryId = readSessionValue('stuffstash.selectedInventoryId');

  constructor(
    config: RuntimeConfig,
    tokenProvider: TokenProvider,
    private readonly observer: WorkspaceObserver,
    fetchImpl?: typeof fetch
  ) {
    this.config = config;
    this.uploadFetch = fetchImpl ?? fetch;
    this.client = new StuffStashClient({
      baseUrl: config.apiBaseUrl,
      tokenProvider,
      fetch: fetchImpl
    });
  }

  async loadWorkspace(): Promise<WorkspaceData> {
    this.observer.record('workspace.load_started');
    try {
      const principal = mapPrincipal(await this.client.me());
      const tenants = (await this.client.listMyTenants()).items.map(mapTenant);
      const selectedTenant = tenants.find((tenant) => tenant.id === this.selectedTenantId) ?? tenants[0] ?? null;
      if (!selectedTenant) {
        this.observer.record('workspace.loaded', { empty: true });
        return {
          context: {
            principal,
            tenants,
            inventories: [],
            selectedTenantId: '',
            selectedInventoryId: '',
            assetLifecycleState: 'active',
            mediaUploadPolicy: this.config.mediaUploadPolicy,
            capability: 'viewer'
          },
          assets: []
        };
      }
      return await this.loadTenantWorkspace(principal, tenants, selectedTenant.id, this.selectedInventoryId);
    } catch (error) {
      this.observer.record('workspace.load_failed');
      throw safeError(error);
    }
  }

  async createTenantWithInventory(input: { tenantName: string; inventoryName: string }): Promise<WorkspaceData> {
    const tenant = mapTenant(await this.client.createTenant(input.tenantName));
    const inventory = mapInventory(await this.client.createInventory(tenant.id, input.inventoryName));
    this.selectedTenantId = tenant.id;
    this.selectedInventoryId = inventory.id;
    this.rememberSelection();
    return {
      context: {
        principal: mapPrincipal(await this.client.me()),
        tenants: [tenant],
        inventories: [inventory],
        selectedTenantId: tenant.id,
        selectedInventoryId: inventory.id,
        assetLifecycleState: 'active',
        mediaUploadPolicy: this.config.mediaUploadPolicy,
        capability: mapCapability(inventory)
      },
      assets: []
    };
  }

  async createInventory(tenantId: string, inventoryName: string): Promise<WorkspaceData> {
    const principal = mapPrincipal(await this.client.me());
    const tenants = (await this.client.listMyTenants()).items.map(mapTenant);
    const inventory = mapInventory(await this.client.createInventory(tenantId, inventoryName));
    return this.loadTenantWorkspace(principal, tenants, tenantId, inventory.id);
  }

  async selectInventory(tenantId: string, inventoryId: string): Promise<WorkspaceData> {
    const principal = mapPrincipal(await this.client.me());
    const tenants = (await this.client.listMyTenants()).items.map(mapTenant);
    return this.loadTenantWorkspace(principal, tenants, tenantId, inventoryId, 'active');
  }

  async selectTenant(tenantId: string): Promise<WorkspaceData> {
    const principal = mapPrincipal(await this.client.me());
    const tenants = (await this.client.listMyTenants()).items.map(mapTenant);
    return this.loadTenantWorkspace(principal, tenants, tenantId, '', 'active');
  }

  async selectAssetLifecycle(
    tenantId: string,
    inventoryId: string,
    lifecycleState: AssetLifecycleFilter
  ): Promise<WorkspaceData> {
    const principal = mapPrincipal(await this.client.me());
    const tenants = (await this.client.listMyTenants()).items.map(mapTenant);
    return this.loadTenantWorkspace(principal, tenants, tenantId, inventoryId, lifecycleState);
  }

  async createAsset(tenantId: string, inventoryId: string, draft: AddAssetDraft): Promise<Asset> {
    this.observer.record('workspace.asset_create_started', { kind: draft.kind });
    try {
      const asset = mapAsset(
        await this.client.createAsset(tenantId, inventoryId, {
          kind: draft.kind,
          title: draft.title,
          description: draft.description,
          parentAssetId: draft.parentAssetId
        })
      );
      this.observer.record('workspace.asset_created', { kind: asset.kind });
      return asset;
    } catch (error) {
      this.observer.record('workspace.asset_create_failed', { kind: draft.kind });
      throw safeError(error);
    }
  }

  async getAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    this.observer.record('workspace.asset_detail_load_started');
    try {
      const asset = mapAsset(await this.client.getAsset(tenantId, inventoryId, assetId));
      this.observer.record('workspace.asset_detail_loaded', { kind: asset.kind });
      return asset;
    } catch (error) {
      this.observer.record('workspace.asset_detail_load_failed');
      throw safeError(error);
    }
  }

  async updateAsset(tenantId: string, inventoryId: string, assetId: string, draft: UpdateAssetDraft): Promise<Asset> {
    this.observer.record('workspace.asset_update_started');
    try {
      const asset = mapAsset(
        await this.client.updateAsset(tenantId, inventoryId, assetId, {
          title: draft.title,
          description: draft.description,
          parentAssetId: draft.parentAssetId
        })
      );
      this.observer.record('workspace.asset_updated', { kind: asset.kind });
      return asset;
    } catch (error) {
      this.observer.record('workspace.asset_update_failed');
      throw safeError(error);
    }
  }

  async archiveAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    this.observer.record('workspace.asset_archive_started');
    try {
      const asset = mapAsset(await this.client.archiveAsset(tenantId, inventoryId, assetId));
      this.observer.record('workspace.asset_archived', { kind: asset.kind });
      return asset;
    } catch (error) {
      this.observer.record('workspace.asset_archive_failed');
      throw safeError(error);
    }
  }

  async restoreAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    this.observer.record('workspace.asset_restore_started');
    try {
      const asset = mapAsset(await this.client.restoreAsset(tenantId, inventoryId, assetId));
      this.observer.record('workspace.asset_restored', { kind: asset.kind });
      return asset;
    } catch (error) {
      this.observer.record('workspace.asset_restore_failed');
      throw safeError(error);
    }
  }

  async deleteAsset(tenantId: string, inventoryId: string, assetId: string): Promise<void> {
    this.observer.record('workspace.asset_delete_started');
    try {
      await this.client.deleteAsset(tenantId, inventoryId, assetId);
      this.observer.record('workspace.asset_deleted');
    } catch (error) {
      this.observer.record('workspace.asset_delete_failed');
      throw safeError(error);
    }
  }

  async listAssetAttachments(tenantId: string, inventoryId: string, assetId: string): Promise<AssetAttachment[]> {
    this.observer.record('workspace.asset_attachments_load_started');
    try {
      const page = await this.client.listAssetAttachments(tenantId, inventoryId, assetId, 50);
      const attachments = await Promise.all(
        page.items.map(async (attachment) => {
          const thumbnail = attachment.contentType.startsWith('image/')
            ? await this.client.assetAttachmentThumbnailReference(tenantId, inventoryId, assetId, attachment.id)
            : undefined;
          return mapAttachment(attachment, await this.thumbnailObjectUrl(thumbnail), thumbnail?.headers);
        })
      );
      this.observer.record('workspace.asset_attachments_loaded', { attachmentCount: attachments.length });
      return attachments;
    } catch (error) {
      this.observer.record('workspace.asset_attachments_load_failed');
      throw safeError(error);
    }
  }

  async uploadAssetPhoto(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    photo: SelectedPhoto
  ): Promise<AssetAttachment> {
    this.observer.record('workspace.asset_attachment_upload_started');
    try {
      const upload = await this.client.initiateAssetAttachmentDirectUpload(tenantId, inventoryId, assetId, {
        fileName: photo.name,
        contentType: photo.contentType,
        sizeBytes: photo.sizeBytes
      });
      await this.uploadToDirectTarget(upload, photo.file);
      const attachment = await this.client.completeAssetAttachmentDirectUpload(tenantId, inventoryId, assetId, upload.uploadId);
      const thumbnail = await this.client.assetAttachmentThumbnailReference(tenantId, inventoryId, assetId, attachment.id);
      this.observer.record('workspace.asset_attachment_uploaded');
      return mapAttachment(attachment, await this.thumbnailObjectUrl(thumbnail), thumbnail.headers);
    } catch (error) {
      this.observer.record('workspace.asset_attachment_upload_failed');
      throw safeError(error);
    }
  }

  async archiveAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<AssetAttachment> {
    const attachment = await this.client.archiveAssetAttachment(tenantId, inventoryId, assetId, attachmentId);
    return mapAttachment(attachment);
  }

  async restoreAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<AssetAttachment> {
    const attachment = await this.client.restoreAssetAttachment(tenantId, inventoryId, assetId, attachmentId);
    return mapAttachment(attachment);
  }

  async deleteAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<void> {
    await this.client.deleteAssetAttachment(tenantId, inventoryId, assetId, attachmentId);
  }

  async searchAssets(request: SearchRequest): Promise<SearchResult[]> {
    this.observer.record('workspace.search_started');
    try {
      const page = await this.client.searchAssets(request.tenantId, request.query, {
        lifecycleState: request.lifecycleState,
        mode: request.mode
      });
      const items = page.items.filter((result) => result.inventory.id === request.inventoryId);
      this.observer.record('workspace.search_completed', { resultCount: items.length });
      return items.map(mapSearchResult);
    } catch (error) {
      this.observer.record('workspace.search_failed');
      throw safeError(error);
    }
  }

  private async loadTenantWorkspace(
    principal: ReturnType<typeof mapPrincipal>,
    tenants: ReturnType<typeof mapTenant>[],
    tenantId: string,
    inventoryId: string,
    lifecycleState: AssetLifecycleFilter = 'active'
  ): Promise<WorkspaceData> {
    this.selectedTenantId = tenantId;
    const inventories = (await this.client.listInventories(tenantId)).items.map(mapInventory);
    const selectedInventory = inventories.find((inventory) => inventory.id === inventoryId) ?? inventories[0] ?? null;
    this.selectedInventoryId = selectedInventory?.id ?? '';
    this.rememberSelection();
    const assets = selectedInventory
      ? (await this.client.listAssets(tenantId, selectedInventory.id, 100, undefined, lifecycleState)).items.map(mapAsset)
      : [];
    this.observer.record('workspace.loaded', {
      tenantCount: tenants.length,
      inventoryCount: inventories.length,
      assetCount: assets.length
    });
    return {
      context: {
        principal,
        tenants,
        inventories,
        selectedTenantId: tenantId,
        selectedInventoryId: this.selectedInventoryId,
        assetLifecycleState: lifecycleState,
        mediaUploadPolicy: this.config.mediaUploadPolicy,
        capability: mapCapability(selectedInventory)
      },
      assets
    };
  }

  private rememberSelection(): void {
    writeSessionValue('stuffstash.selectedTenantId', this.selectedTenantId);
    writeSessionValue('stuffstash.selectedInventoryId', this.selectedInventoryId);
  }

  private async uploadToDirectTarget(
    upload: Awaited<ReturnType<StuffStashClient['initiateAssetAttachmentDirectUpload']>>,
    file: File
  ): Promise<void> {
    const method = upload.method.toUpperCase();
    const target = new URL(upload.url);
    if (target.protocol !== 'http:' && target.protocol !== 'https:') {
      throw new Error('Direct upload target is not available in this browser.');
    }
    const init: RequestInit = { method, headers: upload.headers };
    if (method === 'POST' && Object.keys(upload.formFields).length > 0) {
      const body = new FormData();
      for (const [key, value] of Object.entries(upload.formFields)) {
        body.append(key, value);
      }
      body.append('file', file);
      init.body = body;
      init.headers = upload.headers;
    } else {
      init.body = file;
    }
    const response = await this.uploadFetch(upload.url, init);
    if (!response.ok) {
      throw new Error('Upload failed.');
    }
  }

  private async thumbnailObjectUrl(
    thumbnail: { uri: string; headers: Record<string, string> } | undefined
  ): Promise<string | undefined> {
    if (!thumbnail) {
      return undefined;
    }
    if (Object.keys(thumbnail.headers).length === 0) {
      return thumbnail.uri;
    }
    const response = await this.uploadFetch(thumbnail.uri, { headers: thumbnail.headers });
    if (!response.ok) {
      return undefined;
    }
    return URL.createObjectURL(await response.blob());
  }
}

function safeError(error: unknown): Error {
  if (error instanceof StuffStashAPIError) {
    return new Error(error.message);
  }
  if (error instanceof Error) {
    return error;
  }
  return new Error('Request failed.');
}

function readSessionValue(key: string): string {
  if (typeof sessionStorage === 'undefined') {
    return '';
  }
  return sessionStorage.getItem(key) ?? '';
}

function writeSessionValue(key: string, value: string): void {
  if (typeof sessionStorage === 'undefined') {
    return;
  }
  if (value) {
    sessionStorage.setItem(key, value);
  } else {
    sessionStorage.removeItem(key);
  }
}
