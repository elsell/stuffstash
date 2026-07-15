import { StuffStashAPIError, StuffStashClient } from '@stuff-stash/api-client';
import type { RuntimeConfig } from '$lib/runtimeConfig';
import type { Asset as ApiAsset, AssetPhotoVariant, TokenProvider } from '@stuff-stash/api-client';
import type {
  AddAssetDraft,
  Asset,
  AssetAttachment,
  AssetCheckout,
  AssetCheckoutDraft,
  AssetTag,
  AssetTagDraft,
  AssetLifecycleFilter,
  CheckedOutAsset,
  CreatedInventoryAccessInvitation,
  InventoryAccessRelationship,
  ImportJob,
  ImportJobCancellationMode,
  InvitationStatusFilter,
  ImportSourceRequest,
  SearchRequest,
  SearchResult,
  SelectedAttachment,
  SelectedPhoto,
  UpdateAssetDraft,
  UndoableOperationDirection,
  WorkspaceData
} from '$lib/domain/inventory';
import type { InventoryRepository } from '$lib/ports/inventoryRepository';
import type { BrowseAssetsPage, BrowseAssetsRequest, InventoryBrowseRepository } from '$lib/ports/inventoryBrowseRepository';
import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
import { fileToBase64 } from '$lib/application/fileEncoding';
import { AuthenticationRequiredError } from '$lib/application/authenticationRequired';
import { normalizeImportSourceRequest } from '$lib/application/workspaceImportRequest';
import { parseInvitationLink } from '$lib/application/invitationLink';
import type {
  CustomAssetTypeDraft,
  CustomFieldDefinitionDraft,
  InventoryCustomizationRepository
} from '$lib/ports/inventoryCustomizationRepository';
import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
import {
  mapAsset,
  mapAssetCheckout,
  mapAssetTag,
  mapAttachment,
  mapAuditRecord,
  mapCapability,
  mapCreatedInventoryAccessInvitation,
  mapCustomAssetType,
  mapCustomFieldDefinition,
  mapCheckedOutAsset,
  mapInventory,
  mapPrincipal,
  mapSearchResult,
  mapTenant
} from './inventoryMapper';
import { mapInventoryAccessGrant, mapInventoryAccessInvitation } from './inventoryMapper';
import { collectCursorPages } from './cursorPagination';

const workspacePageSize = 100;

export class StuffStashInventoryRepository
  implements InventoryRepository, InventoryBrowseRepository, InventoryAccessRepository, InventoryAuditRepository, InventoryCustomizationRepository
{
  private readonly client: StuffStashClient;
  private readonly uploadFetch: typeof fetch;
  private readonly config: RuntimeConfig;
  private readonly invitationOrigin: string;
  private readonly invitationAllowInsecureLocalHTTP: boolean;
  private selectedTenantId = readSessionValue('stuffstash.selectedTenantId');
  private selectedInventoryId = readSessionValue('stuffstash.selectedInventoryId');

  constructor(
    config: RuntimeConfig,
    tokenProvider: TokenProvider,
    private readonly observer: WorkspaceObserver,
    fetchImpl?: typeof fetch,
    invitationOrigin: string = typeof window === 'undefined' ? '' : window.location.origin
  ) {
    this.config = config;
    this.invitationOrigin = invitationOrigin;
    this.invitationAllowInsecureLocalHTTP = config.invitationAllowInsecureLocalHTTP;
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
            customAssetTypes: [],
            customFieldDefinitions: [],
            assetTags: [],
            capability: 'viewer'
          },
          assets: [],
          checkedOutAssets: []
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
        customAssetTypes: [],
        customFieldDefinitions: [],
        assetTags: [],
        capability: mapCapability(inventory)
      },
      assets: [],
      checkedOutAssets: []
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
      const asset = await this.mapAssetWithPrimaryPhoto(
        await this.client.createAsset(tenantId, inventoryId, {
          kind: draft.kind,
          title: draft.title,
          description: draft.description,
          parentAssetId: draft.parentAssetId,
          customAssetTypeId: draft.customAssetTypeId,
          customFields: draft.customFields,
          tagIds: draft.tagIds
        })
      );
      this.observer.record('workspace.asset_created', { kind: asset.kind });
      return asset;
    } catch (error) {
      this.observer.record('workspace.asset_create_failed', { kind: draft.kind });
      throw safeError(error);
    }
  }

  async createAssetTag(tenantId: string, inventoryId: string, draft: AssetTagDraft): Promise<AssetTag> {
    this.observer.record('workspace.asset_tag_create_started');
    try {
      const tag = mapAssetTag(await this.client.createAssetTag(tenantId, inventoryId, draft));
      this.observer.record('workspace.asset_tag_created');
      return tag;
    } catch (error) {
      this.observer.record('workspace.asset_tag_create_failed');
      throw safeError(error);
    }
  }

  async getAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    this.observer.record('workspace.asset_detail_load_started');
    try {
      const asset = await this.mapAssetWithPrimaryPhoto(await this.client.getAsset(tenantId, inventoryId, assetId), 'medium');
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
      const asset = await this.mapAssetWithPrimaryPhoto(
        await this.client.updateAsset(tenantId, inventoryId, assetId, {
          title: draft.title,
          description: draft.description,
          parentAssetId: draft.parentAssetId,
          customFields: draft.customFields,
          tagIds: draft.tagIds
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
      const asset = await this.mapAssetWithPrimaryPhoto(await this.client.archiveAsset(tenantId, inventoryId, assetId));
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
      const asset = await this.mapAssetWithPrimaryPhoto(await this.client.restoreAsset(tenantId, inventoryId, assetId));
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

  async checkoutAsset(tenantId: string, inventoryId: string, assetId: string, draft: AssetCheckoutDraft): Promise<AssetCheckout> {
    this.observer.record('workspace.asset_checkout_started');
    try {
      const checkout = mapAssetCheckout(await this.client.checkoutAsset(tenantId, inventoryId, assetId, draft));
      this.observer.record('workspace.asset_checked_out');
      return checkout;
    } catch (error) {
      this.observer.record('workspace.asset_checkout_failed');
      throw safeError(error);
    }
  }

  async returnAsset(tenantId: string, inventoryId: string, assetId: string, draft: AssetCheckoutDraft): Promise<AssetCheckout> {
    this.observer.record('workspace.asset_return_started');
    try {
      const checkout = mapAssetCheckout(await this.client.returnAsset(tenantId, inventoryId, assetId, draft));
      this.observer.record('workspace.asset_returned');
      return checkout;
    } catch (error) {
      this.observer.record('workspace.asset_return_failed');
      throw safeError(error);
    }
  }

  async applyAssetOperation(
    tenantId: string,
    inventoryId: string,
    operationId: string,
    direction: UndoableOperationDirection
  ): Promise<Asset> {
    this.observer.record('workspace.asset_operation_apply_started', { direction });
    try {
      const asset = await this.mapAssetWithPrimaryPhoto(
        await this.client.applyUndoableOperation(tenantId, inventoryId, operationId, direction)
      );
      this.observer.record('workspace.asset_operation_applied', { direction, kind: asset.kind });
      return asset;
    } catch (error) {
      this.observer.record('workspace.asset_operation_apply_failed', { direction });
      throw safeError(error);
    }
  }

  async listAssetCheckoutHistory(tenantId: string, inventoryId: string, assetId: string): Promise<AssetCheckout[]> {
    this.observer.record('workspace.asset_checkout_history_load_started');
    try {
      const page = await this.client.listAssetCheckoutHistory(tenantId, inventoryId, assetId, 50);
      const items = page.items.map(mapAssetCheckout);
      this.observer.record('workspace.asset_checkout_history_loaded', { recordCount: items.length });
      return items;
    } catch (error) {
      this.observer.record('workspace.asset_checkout_history_load_failed');
      throw safeError(error);
    }
  }

  async listCheckedOutAssets(tenantId: string, inventoryId: string): Promise<CheckedOutAsset[]> {
    this.observer.record('workspace.checked_out_assets_load_started');
    try {
      const page = await this.client.listCheckedOutAssets(tenantId, inventoryId, 50);
      const checkedOutAssets = await Promise.all(
        page.items.map(async (item) => {
          const mapped = mapCheckedOutAsset(item);
          return {
            ...mapped,
            asset: await this.mapAssetWithPrimaryPhoto(item.asset)
          };
        })
      );
      this.observer.record('workspace.checked_out_assets_loaded', { assetCount: checkedOutAssets.length });
      return checkedOutAssets;
    } catch (error) {
      this.observer.record('workspace.checked_out_assets_load_failed');
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

  async uploadAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachment: SelectedAttachment
  ): Promise<AssetAttachment> {
    this.observer.record('workspace.asset_attachment_upload_started');
    try {
      const upload = await this.client.initiateAssetAttachmentDirectUpload(tenantId, inventoryId, assetId, {
        fileName: attachment.name,
        contentType: attachment.contentType,
        sizeBytes: attachment.sizeBytes
      });
      await this.uploadToDirectTarget(upload, attachment.file);
      const uploaded = await this.client.completeAssetAttachmentDirectUpload(tenantId, inventoryId, assetId, upload.uploadId);
      const thumbnail = uploaded.contentType.startsWith('image/')
        ? await this.client.assetAttachmentThumbnailReference(tenantId, inventoryId, assetId, uploaded.id)
        : undefined;
      this.observer.record('workspace.asset_attachment_uploaded');
      return mapAttachment(uploaded, await this.thumbnailObjectUrl(thumbnail), thumbnail?.headers);
    } catch (error) {
      if (!isDirectUploadTargetUnavailable(error)) {
        this.observer.record('workspace.asset_attachment_upload_failed');
        throw safeError(error);
      }
      try {
        const uploaded = await this.client.createAssetAttachment(tenantId, inventoryId, assetId, {
          fileName: attachment.name,
          contentType: attachment.contentType,
          contentBase64: await fileToBase64(attachment.file)
        });
        const thumbnail = uploaded.contentType.startsWith('image/')
          ? await this.client.assetAttachmentThumbnailReference(tenantId, inventoryId, assetId, uploaded.id)
          : undefined;
        this.observer.record('workspace.asset_attachment_uploaded');
        return mapAttachment(uploaded, await this.thumbnailObjectUrl(thumbnail), thumbnail?.headers);
      } catch (fallbackError) {
        this.observer.record('workspace.asset_attachment_upload_failed');
        throw safeError(fallbackError);
      }
    }
  }

  async uploadAssetPhoto(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    photo: SelectedPhoto
  ): Promise<AssetAttachment> {
    try {
      return await this.uploadAssetAttachment(tenantId, inventoryId, assetId, photo);
    } catch (error) {
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
        inventoryId: request.inventoryId,
        tagIds: request.tagIds,
        lifecycleState: request.lifecycleState,
        mode: request.mode,
        checkoutState: request.checkoutState ?? 'any'
      });
      this.observer.record('workspace.search_completed', { resultCount: page.items.length });
      return Promise.all(
        page.items.map(async (result) => ({
          ...mapSearchResult(result),
          asset: await this.mapAssetWithPrimaryPhoto(result.asset)
        }))
      );
    } catch (error) {
      this.observer.record('workspace.search_failed');
      throw safeError(error);
    }
  }

  async browseAssets(request: BrowseAssetsRequest): Promise<BrowseAssetsPage> {
    this.observer.record('workspace.browse_started', { scope: request.scope });
    try {
      const page = request.query.trim() || request.tagIds.length > 0
        ? await this.searchBrowseAssets(request)
        : request.checkoutState === 'checked_out'
          ? await this.checkedOutBrowseAssets(request)
          : await this.listBrowseAssets(request);
      this.observer.record('workspace.browse_completed', { scope: request.scope, resultCount: page.assets.length });
      return page;
    } catch (error) {
      this.observer.record('workspace.browse_failed', { scope: request.scope });
      throw safeError(error);
    }
  }

  async hasAnyAssets(tenantId: string, inventoryId: string): Promise<boolean> {
    try {
      const page = await this.client.listAssets(tenantId, inventoryId, 1, undefined, 'all');
      return page.items.length > 0;
    } catch (error) {
      throw safeError(error);
    }
  }

  async loadActiveContainmentMap(tenantId: string, inventoryId: string): Promise<Asset[]> {
    this.observer.record('workspace.browse_started', { surface: 'map' });
    try {
      const assets = await collectCursorPages((cursor) => this.client.listAssets(tenantId, inventoryId, workspacePageSize, cursor, 'active', 'id_asc'));
      const mapped = assets.map(mapAsset);
      this.observer.record('workspace.browse_completed', { surface: 'map', resultCount: mapped.length });
      return mapped;
    } catch (error) {
      this.observer.record('workspace.browse_failed', { surface: 'map' });
      throw safeError(error);
    }
  }

  private async listBrowseAssets(request: BrowseAssetsRequest): Promise<BrowseAssetsPage> {
    const selected: Asset[] = [];
    let cursor = request.cursor;
    let hasMore = false;
    do {
      const page = await this.client.listAssets(
        request.tenantId, request.inventoryId, Math.max(1, request.limit - selected.length), cursor,
        request.lifecycleState, request.sort
      );
      selected.push(...page.items.map(mapAsset).filter((asset) => browseAssetMatches(asset, request)));
      cursor = page.pagination.nextCursor ?? undefined;
      hasMore = page.pagination.hasMore;
    } while (selected.length < request.limit && hasMore && cursor);
    return browsePage(selected.slice(0, request.limit), [], cursor ?? null, hasMore);
  }

  private async searchBrowseAssets(request: BrowseAssetsRequest): Promise<BrowseAssetsPage> {
    const selected: SearchResult[] = [];
    let cursor = request.cursor;
    let hasMore = false;
    do {
      const page = await this.client.searchAssets(request.tenantId, request.query, {
        limit: Math.max(1, request.limit - selected.length), cursor, inventoryId: request.inventoryId,
        tagIds: request.tagIds, lifecycleState: request.lifecycleState, mode: request.mode, checkoutState: request.checkoutState
      });
      selected.push(...page.items.map(mapSearchResult).filter((result) => browseAssetMatches(result.asset, request)));
      cursor = page.pagination.nextCursor ?? undefined;
      hasMore = page.pagination.hasMore;
    } while (selected.length < request.limit && hasMore && cursor);
    const pageResults = selected.slice(0, request.limit);
    return browsePage(pageResults.map((result) => result.asset), pageResults, cursor ?? null, hasMore);
  }

  private async checkedOutBrowseAssets(request: BrowseAssetsRequest): Promise<BrowseAssetsPage> {
    const checkedOut = await collectCursorPages((cursor) =>
      this.client.listCheckedOutAssets(request.tenantId, request.inventoryId, workspacePageSize, cursor)
    );
    const assets = checkedOut
      .map(mapCheckedOutAsset)
      .map((entry) => entry.asset)
      .filter((asset) => browseAssetMatches(asset, request))
      .sort((left, right) => request.sort === 'id_asc'
        ? left.id.localeCompare(right.id)
        : (Date.parse(right.updatedAt ?? '') || 0) - (Date.parse(left.updatedAt ?? '') || 0) || right.id.localeCompare(left.id));
    const offset = Number.parseInt(request.cursor ?? '0', 10) || 0;
    const pageAssets = assets.slice(offset, offset + request.limit);
    const nextOffset = offset + pageAssets.length;
    return browsePage(pageAssets, [], nextOffset < assets.length ? String(nextOffset) : null, nextOffset < assets.length);
  }

  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    this.observer.record('workspace.import_jobs_load_started');
    try {
      const jobs = await this.client.listImportJobs(tenantId, inventoryId);
      this.observer.record('workspace.import_jobs_loaded', { jobCount: jobs.length });
      return jobs;
    } catch (error) {
      this.observer.record('workspace.import_jobs_load_failed');
      throw safeError(error);
    }
  }

  async previewImportJob(tenantId: string, inventoryId: string, input: ImportSourceRequest): Promise<ImportJob> {
    const request = normalizeImportSourceRequest(input);
    this.observer.record('workspace.import_job_preview_started', { sourceType: request.sourceType });
    try {
      const job = await this.client.previewImportJob(tenantId, inventoryId, request, { requestId: importJobRequestId('preview') });
      this.observer.record('workspace.import_job_preview_completed', {
        sourceType: request.sourceType,
        assetCount: job.counts.assets,
        warningCount: job.counts.warnings,
        errorCount: job.counts.errors
      });
      return job;
    } catch (error) {
      this.observer.record('workspace.import_job_preview_failed', { sourceType: request.sourceType });
      throw safeError(error);
    }
  }

  async getImportJob(tenantId: string, inventoryId: string, jobId: string): Promise<ImportJob> {
    return this.client.getImportJob(tenantId, inventoryId, jobId);
  }

  async startImportJob(
    tenantId: string,
    inventoryId: string,
    jobId: string,
    input: ImportSourceRequest
  ): Promise<ImportJob> {
    const request = normalizeImportSourceRequest(input);
    this.observer.record('workspace.import_job_start_started', { sourceType: request.sourceType });
    try {
      const job = await this.client.startImportJob(tenantId, inventoryId, jobId, request, { requestId: importJobRequestId('start') });
      this.observer.record('workspace.import_job_started', { sourceType: request.sourceType, jobId });
      return job;
    } catch (error) {
      this.observer.record('workspace.import_job_start_failed', { sourceType: request.sourceType, jobId });
      throw safeError(error);
    }
  }

  async cancelImportJob(
    tenantId: string,
    inventoryId: string,
    jobId: string,
    mode: ImportJobCancellationMode
  ): Promise<ImportJob> {
    this.observer.record('workspace.import_job_cancel_started', { mode, jobId });
    try {
      const job = await this.client.cancelImportJob(tenantId, inventoryId, jobId, mode, { requestId: importJobRequestId('cancel') });
      this.observer.record('workspace.import_job_cancel_requested', { mode, jobId });
      return job;
    } catch (error) {
      this.observer.record('workspace.import_job_cancel_failed', { mode, jobId });
      throw safeError(error);
    }
  }

  async removeImportJobFromHistory(tenantId: string, inventoryId: string, jobId: string): Promise<void> {
    this.observer.record('workspace.import_job_history_remove_started', { jobId });
    try {
      await this.client.removeImportJobFromHistory(tenantId, inventoryId, jobId, { requestId: importJobRequestId('remove') });
      this.observer.record('workspace.import_job_history_removed', { jobId });
    } catch (error) {
      this.observer.record('workspace.import_job_history_remove_failed', { jobId });
      throw safeError(error);
    }
  }

  async listInventoryAccessGrants(tenantId: string, inventoryId: string, cursor?: string) {
    this.observer.record('workspace.access_grants_load_started');
    try {
      const page = await this.client.listInventoryAccessGrants(tenantId, inventoryId, 50, cursor);
      const items = page.items.map(mapInventoryAccessGrant);
      this.observer.record('workspace.access_grants_loaded', { grantCount: items.length });
      return { items, pagination: page.pagination };
    } catch (error) {
      this.observer.record('workspace.access_grants_load_failed');
      throw safeError(error);
    }
  }

  async grantInventoryAccess(
    tenantId: string,
    inventoryId: string,
    principalId: string,
    relationship: InventoryAccessRelationship
  ) {
    this.observer.record('workspace.access_grant_started', { relationship });
    try {
      const grant = mapInventoryAccessGrant(
        await this.client.grantInventoryAccess(tenantId, inventoryId, { principalId, relationship })
      );
      this.observer.record('workspace.access_granted', { relationship });
      return grant;
    } catch (error) {
      this.observer.record('workspace.access_grant_failed', { relationship });
      throw safeError(error);
    }
  }

  async revokeInventoryAccess(
    tenantId: string,
    inventoryId: string,
    principalId: string,
    relationship: InventoryAccessRelationship
  ): Promise<void> {
    this.observer.record('workspace.access_revoke_started', { relationship });
    try {
      await this.client.revokeInventoryAccess(tenantId, inventoryId, principalId, relationship);
      this.observer.record('workspace.access_revoked', { relationship });
    } catch (error) {
      this.observer.record('workspace.access_revoke_failed', { relationship });
      throw safeError(error);
    }
  }

  async listInventoryAccessInvitations(
    tenantId: string,
    inventoryId: string,
    status: InvitationStatusFilter = 'all',
    cursor?: string
  ) {
    this.observer.record('workspace.access_invitations_load_started', { status });
    try {
      const page = await this.client.listInventoryAccessInvitations(tenantId, inventoryId, { limit: 50, status, cursor });
      const items = page.items.map(mapInventoryAccessInvitation);
      this.observer.record('workspace.access_invitations_loaded', { invitationCount: items.length });
      return { items, pagination: page.pagination };
    } catch (error) {
      this.observer.record('workspace.access_invitations_load_failed', { status });
      throw safeError(error);
    }
  }

  async createInventoryAccessInvitation(
    tenantId: string,
    inventoryId: string,
    email: string,
    relationship: InventoryAccessRelationship
  ) {
    this.observer.record('workspace.access_invitation_create_started', { relationship });
    try {
      const invitation = mapCreatedInventoryAccessInvitation(
        await this.client.createInventoryAccessInvitation(tenantId, inventoryId, { email, relationship })
      );
      assertCanonicalCreatedInvitation(
        invitation,
        tenantId,
        inventoryId,
        this.invitationOrigin,
        this.invitationAllowInsecureLocalHTTP
      );
      this.observer.record('workspace.access_invitation_created', { relationship });
      return invitation;
    } catch (error) {
      this.observer.record('workspace.access_invitation_create_failed', { relationship });
      throw safeError(error);
    }
  }

  async updateInventoryAccessInvitationExpiration(
    tenantId: string,
    inventoryId: string,
    invitationId: string,
    expiresAt: string
  ) {
    this.observer.record('workspace.access_invitation_expiration_started');
    try {
      const invitation = mapInventoryAccessInvitation(
        await this.client.updateInventoryAccessInvitationExpiration(tenantId, inventoryId, invitationId, expiresAt)
      );
      this.observer.record('workspace.access_invitation_expiration_updated');
      return invitation;
    } catch (error) {
      this.observer.record('workspace.access_invitation_expiration_failed');
      throw safeError(error);
    }
  }

  async cancelInventoryAccessInvitation(tenantId: string, inventoryId: string, invitationId: string): Promise<void> {
    this.observer.record('workspace.access_invitation_cancel_started');
    try {
      await this.client.cancelInventoryAccessInvitation(tenantId, inventoryId, invitationId);
      this.observer.record('workspace.access_invitation_cancelled');
    } catch (error) {
      this.observer.record('workspace.access_invitation_cancel_failed');
      throw safeError(error);
    }
  }

  async deleteInventoryAccessInvitation(tenantId: string, inventoryId: string, invitationId: string): Promise<void> {
    this.observer.record('workspace.access_invitation_delete_started');
    try {
      await this.client.deleteInventoryAccessInvitation(tenantId, inventoryId, invitationId);
      this.observer.record('workspace.access_invitation_deleted');
    } catch (error) {
      this.observer.record('workspace.access_invitation_delete_failed');
      throw safeError(error);
    }
  }

  async listTenantAuditRecords(tenantId: string, cursor?: string, signal?: AbortSignal) {
    this.observer.record('workspace.audit_load_started', { scope: 'tenant' });
    try {
      const page = await this.client.listTenantAuditRecords(tenantId, 50, cursor, signal);
      const items = page.items.map(mapAuditRecord);
      this.observer.record('workspace.audit_loaded', { scope: 'tenant', recordCount: items.length });
      return { items, pagination: page.pagination };
    } catch (error) {
      if (isAbortError(error)) {
        throw error;
      }
      this.observer.record('workspace.audit_load_failed', { scope: 'tenant' });
      throw safeError(error);
    }
  }

  async listInventoryAuditRecords(tenantId: string, inventoryId: string, cursor?: string, signal?: AbortSignal) {
    this.observer.record('workspace.audit_load_started', { scope: 'inventory' });
    try {
      const page = await this.client.listInventoryAuditRecords(tenantId, inventoryId, 50, cursor, signal);
      const items = page.items.map(mapAuditRecord);
      this.observer.record('workspace.audit_loaded', { scope: 'inventory', recordCount: items.length });
      return { items, pagination: page.pagination };
    } catch (error) {
      if (isAbortError(error)) {
        throw error;
      }
      this.observer.record('workspace.audit_load_failed', { scope: 'inventory' });
      throw safeError(error);
    }
  }

  async listInventoryCustomAssetTypes(tenantId: string, inventoryId: string, cursor?: string) {
    const page = await this.client.listInventoryCustomAssetTypes(tenantId, inventoryId, 50, cursor);
    return { items: page.items.map(mapCustomAssetType), pagination: page.pagination };
  }

  async createCustomAssetType(tenantId: string, inventoryId: string, draft: CustomAssetTypeDraft) {
    const input = { key: draft.key, displayName: draft.displayName, description: draft.description };
    const assetType =
      draft.scope === 'tenant'
        ? await this.client.createTenantCustomAssetType(tenantId, input)
        : await this.client.createInventoryCustomAssetType(tenantId, inventoryId, input);
    return mapCustomAssetType(assetType);
  }

  async archiveCustomAssetType(tenantId: string, inventoryId: string, customAssetTypeId: string, scope: 'tenant' | 'inventory') {
    const assetType =
      scope === 'tenant'
        ? await this.client.archiveTenantCustomAssetType(tenantId, customAssetTypeId)
        : await this.client.archiveInventoryCustomAssetType(tenantId, inventoryId, customAssetTypeId);
    return mapCustomAssetType(assetType);
  }

  async listInventoryCustomFieldDefinitions(tenantId: string, inventoryId: string, cursor?: string) {
    const page = await this.client.listInventoryCustomFieldDefinitions(tenantId, inventoryId, 50, cursor);
    return { items: page.items.map(mapCustomFieldDefinition), pagination: page.pagination };
  }

  async createCustomFieldDefinition(tenantId: string, inventoryId: string, draft: CustomFieldDefinitionDraft) {
    const input = {
      key: draft.key,
      displayName: draft.displayName,
      type: draft.type,
      enumOptions: draft.enumOptions,
      applicability: draft.applicability,
      customAssetTypeIds: draft.customAssetTypeIds
    };
    const definition =
      draft.scope === 'tenant'
        ? await this.client.createTenantCustomFieldDefinition(tenantId, input)
        : await this.client.createInventoryCustomFieldDefinition(tenantId, inventoryId, input);
    return mapCustomFieldDefinition(definition);
  }

  async archiveCustomFieldDefinition(tenantId: string, inventoryId: string, definitionId: string, scope: 'tenant' | 'inventory') {
    const definition =
      scope === 'tenant'
        ? await this.client.archiveTenantCustomFieldDefinition(tenantId, definitionId)
        : await this.client.archiveInventoryCustomFieldDefinition(tenantId, inventoryId, definitionId);
    return mapCustomFieldDefinition(definition);
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
    const customAssetTypes = selectedInventory
      ? (await this.client.listInventoryCustomAssetTypes(tenantId, selectedInventory.id, 100)).items.map(mapCustomAssetType)
      : [];
    const customFieldDefinitions = selectedInventory
      ? (await this.client.listInventoryCustomFieldDefinitions(tenantId, selectedInventory.id, 100)).items.map(mapCustomFieldDefinition)
      : [];
    const assetTags = selectedInventory
      ? (await this.client.listAssetTags(tenantId, selectedInventory.id, 100)).items.map(mapAssetTag)
      : [];
    const assets = selectedInventory
      ? await Promise.all(
          (await this.client.listAssets(tenantId, selectedInventory.id, 100, undefined, lifecycleState)).items.map((asset) =>
            this.mapAssetWithPrimaryPhoto(asset)
          )
        )
      : [];
    const checkedOutAssets = selectedInventory ? await this.listCheckedOutAssets(tenantId, selectedInventory.id) : [];
    this.observer.record('workspace.loaded', {
      tenantCount: tenants.length,
      inventoryCount: inventories.length,
      assetCount: assets.length,
      checkedOutAssetCount: checkedOutAssets.length
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
        customAssetTypes,
        customFieldDefinitions,
        assetTags,
        capability: mapCapability(selectedInventory)
      },
      assets,
      checkedOutAssets
    };
  }

  private rememberSelection(): void {
    writeSessionValue('stuffstash.selectedTenantId', this.selectedTenantId);
    writeSessionValue('stuffstash.selectedInventoryId', this.selectedInventoryId);
  }

  private async mapAssetWithPrimaryPhoto(asset: ApiAsset, variant: AssetPhotoVariant = 'small'): Promise<Asset> {
    const mapped = mapAsset(asset);
    if (!asset.primaryPhoto) {
      return mapped;
    }
    try {
      const thumbnail = await this.client.assetAttachmentThumbnailReference(
        asset.tenantId,
        asset.inventoryId,
        asset.id,
        asset.primaryPhoto.id,
        variant
      );
      const thumbnailUrl = await this.thumbnailObjectUrl(thumbnail);
      if (!thumbnailUrl) {
        this.observer.record('workspace.asset_primary_photo_load_failed', { assetId: asset.id });
        return { ...mapped, photoUnavailable: true };
      }
      return {
        ...mapped,
        photo: {
          id: asset.primaryPhoto.id,
          assetId: asset.id,
          url: thumbnailUrl,
          alt: asset.title
        }
      };
    } catch {
      this.observer.record('workspace.asset_primary_photo_load_failed', { assetId: asset.id });
      return { ...mapped, photoUnavailable: true };
    }
  }

  private async uploadToDirectTarget(
    upload: Awaited<ReturnType<StuffStashClient['initiateAssetAttachmentDirectUpload']>>,
    file: File
  ): Promise<void> {
    const method = upload.method.toUpperCase();
    const target = new URL(upload.url);
    if (target.protocol !== 'http:' && target.protocol !== 'https:') {
      throw new DirectUploadTargetUnavailableError();
    }
    const init: RequestInit = { method, headers: upload.headers };
    if (method === 'POST' && Object.keys(upload.formFields).length > 0) {
      const body = new FormData();
      for (const [key, value] of Object.entries(upload.formFields)) {
        body.append(key, value);
      }
      body.append('file', await filePart(file), file.name);
      init.body = body;
      init.headers = withoutContentType(upload.headers);
    } else {
      init.body = file;
    }
    let response: Response;
    try {
      response = await this.uploadFetch(upload.url, init);
    } catch {
      throw new DirectUploadFailedError();
    }
    if (!response.ok) {
      throw new DirectUploadFailedError();
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

function browseAssetMatches(asset: Asset, request: BrowseAssetsRequest): boolean {
  if (request.lifecycleState !== 'all' && asset.lifecycleState !== request.lifecycleState) return false;
  if (request.scope === 'places' && (asset.kind !== 'location' || asset.parentAssetId !== null)) return false;
  if (request.scope === 'containers' && asset.kind !== 'container') return false;
  if (request.scope === 'items' && asset.kind !== 'item') return false;
  if (request.checkoutState === 'available' && asset.currentCheckout) return false;
  if (request.checkoutState === 'checked_out' && !asset.currentCheckout) return false;
  return true;
}

function browsePage(
  assets: Asset[], searchResults: SearchResult[], nextCursor: string | null, hasMore: boolean
): BrowseAssetsPage {
  return { assets, searchResults, nextCursor, hasMore };
}

function assertCanonicalCreatedInvitation(
  created: CreatedInventoryAccessInvitation,
  tenantId: string,
  inventoryId: string,
  expectedOrigin: string,
  allowInsecureLocalHTTP: boolean
): void {
  let url: URL;
  try {
    url = new URL(created.inviteUrl);
  } catch {
    throw new Error('Stuff Stash returned an invalid invitation link.');
  }
  const permittedHTTP = url.protocol === 'http:' && allowInsecureLocalHTTP && isLocalInvitationHost(url.hostname);
  if (url.origin !== expectedOrigin || (url.protocol !== 'https:' && !permittedHTTP)) {
    throw new Error('Stuff Stash returned an invalid invitation link.');
  }
  const material = parseInvitationLink(created.inviteUrl, url.origin);
  if (
    created.invitation.tenantId !== tenantId ||
    created.invitation.inventoryId !== inventoryId ||
    !material ||
    material.tenantId !== tenantId ||
    material.inventoryId !== inventoryId ||
    material.invitationId !== created.invitation.id
  ) {
    throw new Error('Stuff Stash returned an invalid invitation link.');
  }
}

function isLocalInvitationHost(hostname: string): boolean {
  if (hostname === 'localhost' || hostname === '[::1]' || hostname === '::1') return true;
  const octets = hostname.split('.');
  if (octets.length !== 4 || octets.some((value) => !/^\d{1,3}$/.test(value))) return false;
  const values = octets.map(Number);
  if (values.some((value) => value < 0 || value > 255)) return false;
  return values[0] === 127 ||
    values[0] === 10 ||
    (values[0] === 172 && values[1] >= 16 && values[1] <= 31) ||
    (values[0] === 192 && values[1] === 168);
}

function safeError(error: unknown): Error {
  if (error instanceof StuffStashAPIError) {
    if (error.status === 401) {
      return new AuthenticationRequiredError(error.message);
    }
    return new SafeAPIError(error.status, error.code, error.message);
  }
  if (error instanceof Error) {
    return error;
  }
  return new Error('Request failed.');
}

class SafeAPIError extends Error {
  readonly safeForUser: boolean;

  constructor(
    readonly status: number,
    readonly code: string,
    message: string
  ) {
    super(message);
    this.name = 'SafeAPIError';
    this.safeForUser = isSafeAPIStatus(status);
  }
}

function isSafeAPIStatus(status: number): boolean {
  return status === 400 || status === 422;
}

class DirectUploadTargetUnavailableError extends Error {
  constructor() {
    super('Direct upload target is not available in this browser.');
  }
}

class DirectUploadFailedError extends Error {
  safeForUser = true as const;

  constructor() {
    super('Direct upload to media storage failed.');
  }
}

function isDirectUploadTargetUnavailable(error: unknown): boolean {
  return error instanceof DirectUploadTargetUnavailableError;
}

function withoutContentType(headers: Record<string, string>): Record<string, string> {
  return Object.fromEntries(Object.entries(headers).filter(([key]) => key.toLowerCase() !== 'content-type'));
}

async function filePart(file: File): Promise<File> {
  return file;
}

function isAbortError(error: unknown): boolean {
  return error instanceof Error && error.name === 'AbortError';
}

function importJobRequestId(action: 'preview' | 'start' | 'cancel' | 'remove'): string {
  const suffix =
    typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
      ? crypto.randomUUID()
      : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
  return `web-import-${action}-${suffix}`.slice(0, 128);
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
