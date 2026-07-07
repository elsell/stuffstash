import {
  type AssetLifecycleFilter,
  canCreateInventory,
  canEditInventory,
  type AddAssetDraft,
  type Asset,
  type AssetAttachment,
  type AssetCheckout,
  type AssetCheckoutDraft,
  type CheckedOutAsset,
  type AuditRecord,
  type CreatedInventoryAccessInvitation,
  type CustomAssetType,
  type CustomFieldDefinition,
  defaultMediaUploadPolicy,
  type Inventory,
  type InventoryAccessGrant,
  type InventoryAccessInvitation,
  type ImportJob,
  type ImportJobCancellationMode,
  type InventoryAccessRelationship,
  type InvitationStatusFilter,
  type ImportSourceRequest,
  type SearchRequest,
  type SearchResult,
  type SelectedAttachment,
  type SelectedPhoto,
  type UpdateAssetDraft,
  type WorkspaceData
} from '$lib/domain/inventory';
import type { InventoryRepository, WorkspaceSeed } from '$lib/ports/inventoryRepository';
import type { InventoryAccessPage, InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
import type { AuditRecordPage, InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
import type {
  CustomAssetTypeDraft,
  CustomFieldDefinitionDraft,
  InventoryCustomizationRepository
} from '$lib/ports/inventoryCustomizationRepository';
import { filterAssets } from '$lib/application/workspace';

type ScopedImportJob = {
  tenantId: string;
  inventoryId: string;
  job: ImportJob;
};

export class SeededInventoryRepository
  implements InventoryRepository, InventoryAccessRepository, InventoryAuditRepository, InventoryCustomizationRepository
{
  private seed: WorkspaceSeed;
  private attachments: AssetAttachment[] = [];
  private checkoutRecords: AssetCheckout[] = [];
  private auditRecords: AuditRecord[] = [];
  private grants: InventoryAccessGrant[] = [];
  private invitations: InventoryAccessInvitation[] = [];
  private importJobs: ScopedImportJob[] = [];
  private selectedTenantId: string;
  private selectedInventoryId: string;
  private selectedLifecycleState: AssetLifecycleFilter = 'active';
  private nextAssetSequence = 1;
  private nextImportJobSequence = 1;

  constructor(seed: WorkspaceSeed) {
    this.seed = seed;
    this.selectedTenantId = seed.tenants[0]?.id ?? '';
    this.selectedInventoryId = this.firstInventoryIdForTenant(this.selectedTenantId);
  }

  async loadWorkspace(): Promise<WorkspaceData> {
    return this.workspace();
  }

  async createTenantWithInventory(input: { tenantName: string; inventoryName: string }): Promise<WorkspaceData> {
    const tenant = {
      id: `tenant-${Date.now()}`,
      name: input.tenantName,
      access: {
        relationship: 'owner',
        permissions: ['view', 'create_inventory', 'configure']
      }
    };
    const inventory = {
      id: `inventory-${Date.now()}`,
      tenantId: tenant.id,
      name: input.inventoryName,
      access: {
        relationship: 'owner',
        permissions: ['view', 'create_asset', 'edit_asset', 'share', 'configure', 'view_import_job', 'create_import_job']
      }
    };
    this.seed = {
      ...this.seed,
      tenants: [tenant, ...this.seed.tenants],
      inventories: [inventory, ...this.seed.inventories]
    };
    this.selectedTenantId = tenant.id;
    this.selectedInventoryId = inventory.id;
    this.selectedLifecycleState = 'active';
    this.recordAudit({
      tenantId: tenant.id,
      inventoryId: null,
      action: 'tenant.created',
      targetType: 'tenant',
      targetId: tenant.id,
      metadata: { name: tenant.name }
    });
    this.recordInventoryAudit(inventory, 'inventory.created');
    return this.workspace();
  }

  async createInventory(tenantId: string, inventoryName: string): Promise<WorkspaceData> {
    const tenant = this.seed.tenants.find((candidate) => candidate.id === tenantId);
    if (!canCreateInventory(tenant)) {
      throw new Error('You do not have permission to create inventories in this tenant.');
    }
    const inventory: Inventory = {
      id: `inventory-${Date.now()}`,
      tenantId,
      name: inventoryName,
      access: {
        relationship: 'owner',
        permissions: ['view', 'create_asset', 'edit_asset', 'share', 'configure', 'view_import_job', 'create_import_job']
      }
    };
    this.seed = {
      ...this.seed,
      inventories: [inventory, ...this.seed.inventories]
    };
    this.selectedTenantId = tenantId;
    this.selectedInventoryId = inventory.id;
    this.selectedLifecycleState = 'active';
    this.recordInventoryAudit(inventory, 'inventory.created');
    return this.workspace();
  }

  async selectInventory(tenantId: string, inventoryId: string): Promise<WorkspaceData> {
    this.selectedTenantId = tenantId;
    this.selectedInventoryId = this.seed.inventories.some(
      (inventory) => inventory.tenantId === tenantId && inventory.id === inventoryId
    )
      ? inventoryId
      : this.firstInventoryIdForTenant(tenantId);
    this.selectedLifecycleState = 'active';
    return this.workspace();
  }

  async selectTenant(tenantId: string): Promise<WorkspaceData> {
    this.selectedTenantId = tenantId;
    this.selectedInventoryId = this.firstInventoryIdForTenant(tenantId);
    this.selectedLifecycleState = 'active';
    return this.workspace();
  }

  async selectAssetLifecycle(
    tenantId: string,
    inventoryId: string,
    lifecycleState: AssetLifecycleFilter
  ): Promise<WorkspaceData> {
    this.selectedTenantId = tenantId;
    this.selectedInventoryId = this.seed.inventories.some(
      (inventory) => inventory.tenantId === tenantId && inventory.id === inventoryId
    )
      ? inventoryId
      : this.firstInventoryIdForTenant(tenantId);
    this.selectedLifecycleState = lifecycleState;
    return this.workspace();
  }

  async createAsset(tenantId: string, inventoryId: string, draft: AddAssetDraft): Promise<Asset> {
    this.validateInventoryScope(tenantId, inventoryId);
    this.validateAssetParent(tenantId, inventoryId, null, draft.parentAssetId);
    const assetId = this.nextAssetId();
    const asset: Asset = {
      id: assetId,
      tenantId,
      inventoryId,
      kind: draft.kind,
      title: draft.title,
      description: draft.description,
      parentAssetId: draft.parentAssetId,
      lifecycleState: 'active',
      customAssetTypeId: draft.customAssetTypeId,
      customFields: draft.customFields ?? {},
      customAssetTypeLabel: this.customAssetTypeLabel(draft.customAssetTypeId),
      photo: draft.photos[0]
        ? {
            id: draft.photos[0].id,
            assetId,
            url: draft.photos[0].previewUrl,
            alt: draft.title
          }
        : undefined,
      updatedAt: new Date().toISOString()
    };
    this.seed = { ...this.seed, assets: [asset, ...this.seed.assets] };
    this.recordAssetAudit(asset, 'asset.created');
    return asset;
  }

  async getAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    const asset = this.seed.assets.find(
      (candidate) => candidate.tenantId === tenantId && candidate.inventoryId === inventoryId && candidate.id === assetId
    );
    if (!asset) {
      throw new Error('Asset not found.');
    }
    return asset;
  }

  async updateAsset(tenantId: string, inventoryId: string, assetId: string, draft: UpdateAssetDraft): Promise<Asset> {
    const asset = await this.getAsset(tenantId, inventoryId, assetId);
    this.validateAssetParent(tenantId, inventoryId, assetId, draft.parentAssetId);
    const updated: Asset = {
      ...asset,
      title: draft.title,
      description: draft.description,
      parentAssetId: draft.parentAssetId,
      customFields: draft.customFields ?? asset.customFields ?? {},
      updatedAt: new Date().toISOString()
    };
    this.seed = {
      ...this.seed,
      assets: this.seed.assets.map((candidate) =>
        candidate.tenantId === tenantId && candidate.inventoryId === inventoryId && candidate.id === assetId
          ? updated
          : candidate
      )
    };
    this.recordAssetAudit(updated, updated.parentAssetId !== asset.parentAssetId ? 'asset.moved' : 'asset.updated');
    return updated;
  }

  async archiveAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    const asset = await this.setAssetLifecycle(tenantId, inventoryId, assetId, 'archived');
    this.recordAssetAudit(asset, 'asset.archived');
    return asset;
  }

  async restoreAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    const asset = await this.setAssetLifecycle(tenantId, inventoryId, assetId, 'active');
    this.recordAssetAudit(asset, 'asset.restored');
    return asset;
  }

  async deleteAsset(tenantId: string, inventoryId: string, assetId: string): Promise<void> {
    const asset = await this.getAsset(tenantId, inventoryId, assetId);
    this.seed = {
      ...this.seed,
      assets: this.seed.assets.filter((candidate) => candidate !== asset)
    };
    this.recordAssetAudit(asset, 'asset.deleted');
  }

  async checkoutAsset(tenantId: string, inventoryId: string, assetId: string, draft: AssetCheckoutDraft): Promise<AssetCheckout> {
    const asset = await this.getAsset(tenantId, inventoryId, assetId);
    if (asset.lifecycleState !== 'active') {
      throw new Error('Archived assets cannot be checked out.');
    }
    if (asset.currentCheckout) {
      throw new Error('Asset is already checked out.');
    }
    const now = new Date().toISOString();
    const checkout: AssetCheckout = {
      id: `checkout-local-${this.checkoutRecords.length + 1}-${Date.now()}`,
      tenantId,
      inventoryId,
      assetId,
      state: 'open',
      checkedOutAt: now,
      checkedOutByPrincipalId: this.seed.principal.id,
      checkoutDetails: draft.details?.trim() || undefined,
      createdAt: now,
      updatedAt: now
    };
    this.checkoutRecords = [checkout, ...this.checkoutRecords];
    this.replaceAsset({
      ...asset,
      currentCheckout: {
        id: checkout.id,
        state: checkout.state,
        checkedOutAt: checkout.checkedOutAt,
        checkedOutByPrincipalId: checkout.checkedOutByPrincipalId
      },
      updatedAt: now
    });
    this.recordAssetAudit(asset, 'asset.checked_out');
    return checkout;
  }

  async returnAsset(tenantId: string, inventoryId: string, assetId: string, draft: AssetCheckoutDraft): Promise<AssetCheckout> {
    const asset = await this.getAsset(tenantId, inventoryId, assetId);
    const openCheckout = this.checkoutRecords.find(
      (record) =>
        record.tenantId === tenantId &&
        record.inventoryId === inventoryId &&
        record.assetId === assetId &&
        record.state === 'open'
    );
    if (!asset.currentCheckout || !openCheckout) {
      throw new Error('Asset is not checked out.');
    }
    const now = new Date().toISOString();
    const returned: AssetCheckout = {
      ...openCheckout,
      state: 'returned',
      returnedAt: now,
      returnedByPrincipalId: this.seed.principal.id,
      returnDetails: draft.details?.trim() || undefined,
      updatedAt: now
    };
    this.checkoutRecords = this.checkoutRecords.map((record) => (record === openCheckout ? returned : record));
    this.replaceAsset({
      ...asset,
      currentCheckout: undefined,
      updatedAt: now
    });
    this.recordAssetAudit(asset, 'asset.returned');
    return returned;
  }

  async listAssetCheckoutHistory(tenantId: string, inventoryId: string, assetId: string): Promise<AssetCheckout[]> {
    return this.checkoutRecords.filter(
      (record) => record.tenantId === tenantId && record.inventoryId === inventoryId && record.assetId === assetId
    );
  }

  async listCheckedOutAssets(tenantId: string, inventoryId: string): Promise<CheckedOutAsset[]> {
    return this.seed.assets
      .filter((asset) => asset.tenantId === tenantId && asset.inventoryId === inventoryId && asset.currentCheckout)
      .map((asset) => ({ asset, checkout: asset.currentCheckout! }));
  }

  async listAssetAttachments(tenantId: string, inventoryId: string, assetId: string): Promise<AssetAttachment[]> {
    return this.attachments.filter(
      (attachment) =>
        attachment.tenantId === tenantId &&
        attachment.inventoryId === inventoryId &&
        attachment.assetId === assetId &&
        attachment.lifecycleState === 'active'
    );
  }

  async uploadAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentInput: SelectedAttachment
  ): Promise<AssetAttachment> {
    await this.getAsset(tenantId, inventoryId, assetId);
    const attachment: AssetAttachment = {
      id: `attachment-${Date.now()}`,
      tenantId,
      inventoryId,
      assetId,
      fileName: attachmentInput.name,
      contentType: attachmentInput.contentType,
      sizeBytes: attachmentInput.sizeBytes,
      lifecycleState: 'active',
      thumbnailUrl: attachmentInput.contentType.startsWith('image/') ? attachmentInput.previewUrl : undefined
    };
    this.attachments = [attachment, ...this.attachments];
    this.recordAudit({
      tenantId,
      inventoryId,
      action: 'attachment.created',
      targetType: 'attachment',
      targetId: attachment.id,
      metadata: { assetId, fileName: attachment.fileName }
    });
    return attachment;
  }

  async uploadAssetPhoto(tenantId: string, inventoryId: string, assetId: string, photo: SelectedPhoto): Promise<AssetAttachment> {
    return this.uploadAssetAttachment(tenantId, inventoryId, assetId, photo);
  }

  async archiveAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<AssetAttachment> {
    const attachment = await this.setAttachmentLifecycle(tenantId, inventoryId, assetId, attachmentId, 'archived');
    this.recordAudit({
      tenantId,
      inventoryId,
      action: 'asset_photo.archived',
      targetType: 'attachment',
      targetId: attachment.id,
      metadata: { assetId, fileName: attachment.fileName }
    });
    return attachment;
  }

  async restoreAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<AssetAttachment> {
    const attachment = await this.setAttachmentLifecycle(tenantId, inventoryId, assetId, attachmentId, 'active');
    this.recordAudit({
      tenantId,
      inventoryId,
      action: 'asset_photo.restored',
      targetType: 'attachment',
      targetId: attachment.id,
      metadata: { assetId, fileName: attachment.fileName }
    });
    return attachment;
  }

  async deleteAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<void> {
    const attachment = this.findAttachment(tenantId, inventoryId, assetId, attachmentId);
    this.attachments = this.attachments.filter((candidate) => candidate !== attachment);
    this.recordAudit({
      tenantId,
      inventoryId,
      action: 'asset_photo.deleted',
      targetType: 'attachment',
      targetId: attachment.id,
      metadata: { assetId, fileName: attachment.fileName }
    });
  }

  async searchAssets(request: SearchRequest): Promise<SearchResult[]> {
    const inventory = this.seed.inventories.find((candidate) => candidate.id === this.selectedInventoryId);
    const searchableAssets = this.seed.assets.filter(
      (asset) =>
        asset.tenantId === request.tenantId &&
        asset.inventoryId === request.inventoryId &&
        (request.lifecycleState === 'all' || asset.lifecycleState === request.lifecycleState) &&
        (request.checkoutState === undefined ||
          request.checkoutState === 'any' ||
          (request.checkoutState === 'checked_out' && !!asset.currentCheckout) ||
          (request.checkoutState === 'available' && !asset.currentCheckout))
    );
    const matches = request.mode === 'exact' ? exactAssets(searchableAssets, request.query) : filterAssets(searchableAssets, request.query);
    return matches.map((asset) => ({
      type: 'asset',
      asset,
      inventory: {
        id: inventory?.id ?? this.selectedInventoryId,
        name: inventory?.name ?? 'Inventory'
      },
      matches: [{ field: 'title', value: asset.title }]
    }));
  }

  async listImportJobs(tenantId: string, inventoryId: string): Promise<ImportJob[]> {
    return this.importJobs.filter((entry) => entry.tenantId === tenantId && entry.inventoryId === inventoryId).map((entry) => entry.job);
  }

  async previewImportJob(tenantId: string, inventoryId: string, input: ImportSourceRequest): Promise<ImportJob> {
    const now = new Date().toISOString();
    const job: ImportJob = {
      id: `import-job-local-${this.nextImportJobSequence}`,
      status: 'previewed',
      source: {
        type: input.sourceType,
        name: input.sourceType === 'legacy_homebox_csv' ? 'Homebox CSV' : 'Homebox',
        baseUrl: input.baseUrl,
        imageImport: input.sourceType === 'legacy_homebox_csv' ? 'unavailable' : input.includeImages ? 'enabled' : 'disabled',
        allowPrivateNetwork: input.allowPrivateNetwork,
        allowInsecureTLS: input.allowInsecureTLS,
        fingerprint: `local-preview-${this.nextImportJobSequence}`
      },
      counts: {
        fields: 0,
        locations: 0,
        assets: 0,
        attachments: 0,
        warnings: 0,
        errors: 0,
        fieldsCreated: 0,
        fieldsExisting: 0,
        locationsCreated: 0,
        assetsCreated: 0,
        assetsSkipped: 0,
        attachmentsCreated: 0,
        attachmentsSkipped: 0,
        recordsDiscarded: 0,
        sourceLinksDiscarded: 0
      },
      preview: {
        fields: [],
        locations: [],
        assets: [],
        attachments: [],
        messages: [
          {
            code: 'local-import-preview',
            severity: 'warning',
            summary: 'Connect to the Stuff Stash API to preview real import records.'
          }
        ],
        fieldsTruncated: false,
        locationsTruncated: false,
        assetsTruncated: false,
        attachmentsTruncated: false,
        messagesTruncated: false
      },
      progress: { phase: 'ready', done: 0, total: 0, updatedAt: now },
      progressHistory: [{ phase: 'ready', done: 0, total: 0, updatedAt: now }],
      createdAt: now,
      updatedAt: now,
      resources: [],
      messages: [
        {
          code: 'local-import-preview',
          severity: 'warning',
          summary: 'Connect to the Stuff Stash API to preview real import records.'
        }
      ]
    };
    this.nextImportJobSequence += 1;
    this.importJobs.unshift({ tenantId, inventoryId, job });
    return job;
  }

  async getImportJob(tenantId: string, inventoryId: string, jobId: string): Promise<ImportJob> {
    const entry = this.importJobs.find(
      (candidate) => candidate.tenantId === tenantId && candidate.inventoryId === inventoryId && candidate.job.id === jobId
    );
    if (!entry) {
      throw new Error('Import job not found.');
    }
    return entry.job;
  }

  async startImportJob(
    tenantId: string,
    inventoryId: string,
    jobId: string,
    _input: ImportSourceRequest
  ): Promise<ImportJob> {
    const job = await this.getImportJob(tenantId, inventoryId, jobId);
    const now = new Date().toISOString();
    job.status = 'running';
    job.startedAt = now;
    job.updatedAt = now;
    job.progress = { phase: 'reading_source', done: 0, total: job.counts.assets, message: 'Queued locally', updatedAt: now };
    job.progressHistory = importJobProgressHistoryWith(job.progressHistory, job.progress);
    return job;
  }

  async cancelImportJob(
    tenantId: string,
    inventoryId: string,
    jobId: string,
    mode: ImportJobCancellationMode
  ): Promise<ImportJob> {
    const job = await this.getImportJob(tenantId, inventoryId, jobId);
    const now = new Date().toISOString();
    job.status = 'cancel_requested';
    job.cancellationMode = mode;
    job.updatedAt = now;
    job.progress = { ...job.progress, message: 'Cancellation requested', updatedAt: now };
    job.progressHistory = importJobProgressHistoryWith(job.progressHistory, job.progress);
    return job;
  }

  async removeImportJobFromHistory(tenantId: string, inventoryId: string, jobId: string): Promise<void> {
    await this.getImportJob(tenantId, inventoryId, jobId);
    this.importJobs = this.importJobs.filter(
      (candidate) => !(candidate.tenantId === tenantId && candidate.inventoryId === inventoryId && candidate.job.id === jobId)
    );
  }

  async listTenantAuditRecords(tenantId: string, cursor?: string, _signal?: AbortSignal): Promise<AuditRecordPage> {
    const records = this.auditRecords.filter((record) => record.tenantId === tenantId);
    return page(records, cursor);
  }

  async listInventoryAuditRecords(
    tenantId: string,
    inventoryId: string,
    cursor?: string,
    _signal?: AbortSignal
  ): Promise<AuditRecordPage> {
    const records = this.auditRecords.filter((record) => record.tenantId === tenantId && record.inventoryId === inventoryId);
    return page(records, cursor);
  }

  async listInventoryCustomAssetTypes(tenantId: string, inventoryId: string, cursor?: string): Promise<InventoryAccessPage<CustomAssetType>> {
    return page(this.effectiveCustomAssetTypes(tenantId, inventoryId), cursor);
  }

  async createCustomAssetType(tenantId: string, inventoryId: string, draft: CustomAssetTypeDraft): Promise<CustomAssetType> {
    const assetType: CustomAssetType = {
      id: `custom-asset-type-${Date.now()}`,
      tenantId,
      inventoryId: draft.scope === 'inventory' ? inventoryId : null,
      scope: draft.scope,
      key: draft.key,
      displayName: draft.displayName,
      description: draft.description,
      lifecycleState: 'active'
    };
    this.seed = { ...this.seed, customAssetTypes: [assetType, ...this.seed.customAssetTypes] };
    this.recordAudit({
      tenantId,
      inventoryId: assetType.inventoryId,
      action: 'custom_asset_type.created',
      targetType: 'custom_asset_type',
      targetId: assetType.id,
      metadata: { key: assetType.key, displayName: assetType.displayName }
    });
    return assetType;
  }

  async archiveCustomAssetType(
    tenantId: string,
    inventoryId: string,
    customAssetTypeId: string,
    scope: 'tenant' | 'inventory'
  ): Promise<CustomAssetType> {
    const assetType = this.findCustomAssetType(tenantId, inventoryId, customAssetTypeId, scope);
    const updated: CustomAssetType = { ...assetType, lifecycleState: 'archived' };
    this.seed = {
      ...this.seed,
      customAssetTypes: this.seed.customAssetTypes.map((candidate) => (candidate === assetType ? updated : candidate))
    };
    this.recordAudit({
      tenantId,
      inventoryId: updated.inventoryId,
      action: 'custom_asset_type.archived',
      targetType: 'custom_asset_type',
      targetId: updated.id,
      metadata: { key: updated.key }
    });
    return updated;
  }

  async listInventoryCustomFieldDefinitions(
    tenantId: string,
    inventoryId: string,
    cursor?: string
  ): Promise<InventoryAccessPage<CustomFieldDefinition>> {
    return page(this.effectiveCustomFieldDefinitions(tenantId, inventoryId), cursor);
  }

  async createCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    draft: CustomFieldDefinitionDraft
  ): Promise<CustomFieldDefinition> {
    this.validateCustomFieldTargets(tenantId, inventoryId, draft);
    const definition: CustomFieldDefinition = {
      id: `custom-field-${Date.now()}`,
      tenantId,
      inventoryId: draft.scope === 'inventory' ? inventoryId : null,
      scope: draft.scope,
      key: draft.key,
      displayName: draft.displayName,
      type: draft.type,
      enumOptions: draft.enumOptions,
      applicability: draft.applicability,
      customAssetTypeIds: draft.customAssetTypeIds,
      lifecycleState: 'active'
    };
    this.seed = { ...this.seed, customFieldDefinitions: [definition, ...this.seed.customFieldDefinitions] };
    this.recordAudit({
      tenantId,
      inventoryId: definition.inventoryId,
      action: 'custom_field_definition.created',
      targetType: 'custom_field_definition',
      targetId: definition.id,
      metadata: { key: definition.key, displayName: definition.displayName }
    });
    return definition;
  }

  async archiveCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    definitionId: string,
    scope: 'tenant' | 'inventory'
  ): Promise<CustomFieldDefinition> {
    const definition = this.findCustomFieldDefinition(tenantId, inventoryId, definitionId, scope);
    const updated: CustomFieldDefinition = { ...definition, lifecycleState: 'archived' };
    this.seed = {
      ...this.seed,
      customFieldDefinitions: this.seed.customFieldDefinitions.map((candidate) => (candidate === definition ? updated : candidate))
    };
    this.recordAudit({
      tenantId,
      inventoryId: updated.inventoryId,
      action: 'custom_field_definition.archived',
      targetType: 'custom_field_definition',
      targetId: updated.id,
      metadata: { key: updated.key }
    });
    return updated;
  }

  async listInventoryAccessGrants(
    tenantId: string,
    inventoryId: string,
    cursor?: string
  ): Promise<InventoryAccessPage<InventoryAccessGrant>> {
    return page(
      this.grants.filter((grant) => grant.tenantId === tenantId && grant.inventoryId === inventoryId),
      cursor
    );
  }

  async grantInventoryAccess(
    tenantId: string,
    inventoryId: string,
    principalId: string,
    relationship: InventoryAccessRelationship
  ): Promise<InventoryAccessGrant> {
    const existing = this.grants.find(
      (grant) =>
        grant.tenantId === tenantId &&
        grant.inventoryId === inventoryId &&
        grant.principalId === principalId &&
        grant.relationship === relationship
    );
    if (existing) {
      return existing;
    }
    const grant: InventoryAccessGrant = { tenantId, inventoryId, principalId, relationship };
    this.grants = [grant, ...this.grants];
    this.recordAudit({
      tenantId,
      inventoryId,
      action: 'inventory_access.granted',
      targetType: 'principal',
      targetId: principalId,
      metadata: { relationship }
    });
    return grant;
  }

  async revokeInventoryAccess(
    tenantId: string,
    inventoryId: string,
    principalId: string,
    relationship: InventoryAccessRelationship
  ): Promise<void> {
    const revoked = this.grants.some(
      (grant) =>
        grant.tenantId === tenantId &&
        grant.inventoryId === inventoryId &&
        grant.principalId === principalId &&
        grant.relationship === relationship
    );
    this.grants = this.grants.filter(
      (grant) =>
        !(
          grant.tenantId === tenantId &&
          grant.inventoryId === inventoryId &&
          grant.principalId === principalId &&
          grant.relationship === relationship
        )
    );
    if (revoked) {
      this.recordAudit({
        tenantId,
        inventoryId,
        action: 'inventory_access.revoked',
        targetType: 'principal',
        targetId: principalId,
        metadata: { relationship }
      });
    }
  }

  async listInventoryAccessInvitations(
    tenantId: string,
    inventoryId: string,
    status: InvitationStatusFilter = 'all',
    cursor?: string
  ): Promise<InventoryAccessPage<InventoryAccessInvitation>> {
    return page(
      this.invitations.filter(
        (invitation) =>
          invitation.tenantId === tenantId &&
          invitation.inventoryId === inventoryId &&
          (status === 'all' || invitation.status === status)
      ),
      cursor
    );
  }

  async createInventoryAccessInvitation(
    tenantId: string,
    inventoryId: string,
    email: string,
    relationship: InventoryAccessRelationship
  ): Promise<CreatedInventoryAccessInvitation> {
    const acceptanceToken = `local-demo-${Date.now()}`;
    const invitation: InventoryAccessInvitation = {
      id: `invitation-${Date.now()}`,
      tenantId,
      inventoryId,
      email: email.trim().toLowerCase(),
      relationship,
      status: 'pending',
      isExpired: false,
      expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
      inviterPrincipalId: this.seed.principal.id
    };
    this.invitations = [invitation, ...this.invitations];
    this.recordAudit({
      tenantId,
      inventoryId,
      action: 'inventory_access_invitation.created',
      targetType: 'access_invitation',
      targetId: invitation.id,
      metadata: { email: invitation.email, relationship }
    });
    return { invitation, acceptanceToken };
  }

  async updateInventoryAccessInvitationExpiration(
    tenantId: string,
    inventoryId: string,
    invitationId: string,
    expiresAt: string
  ): Promise<InventoryAccessInvitation> {
    const invitation = this.findInvitation(tenantId, inventoryId, invitationId);
    const updated: InventoryAccessInvitation = { ...invitation, expiresAt, isExpired: Date.parse(expiresAt) <= Date.now() };
    this.invitations = this.invitations.map((candidate) => (candidate === invitation ? updated : candidate));
    this.recordAudit({
      tenantId,
      inventoryId,
      action: 'inventory_access_invitation.expiration_updated',
      targetType: 'access_invitation',
      targetId: invitationId,
      metadata: { expiresAt }
    });
    return updated;
  }

  async cancelInventoryAccessInvitation(tenantId: string, inventoryId: string, invitationId: string): Promise<void> {
    const invitation = this.findInvitation(tenantId, inventoryId, invitationId);
    const updated: InventoryAccessInvitation = { ...invitation, status: 'cancelled' };
    this.invitations = this.invitations.map((candidate) => (candidate === invitation ? updated : candidate));
    this.recordAudit({
      tenantId,
      inventoryId,
      action: 'inventory_access_invitation.cancelled',
      targetType: 'access_invitation',
      targetId: invitationId,
      metadata: { email: invitation.email }
    });
  }

  async deleteInventoryAccessInvitation(tenantId: string, inventoryId: string, invitationId: string): Promise<void> {
    const invitation = this.findInvitation(tenantId, inventoryId, invitationId);
    this.invitations = this.invitations.filter((candidate) => candidate !== invitation);
    this.recordAudit({
      tenantId,
      inventoryId,
      action: 'inventory_access_invitation.deleted',
      targetType: 'access_invitation',
      targetId: invitationId,
      metadata: { email: invitation.email }
    });
  }

  private workspace(): WorkspaceData {
    const inventories = this.seed.inventories.filter((inventory) => inventory.tenantId === this.selectedTenantId);
    const selectedInventory = inventories.find((inventory) => inventory.id === this.selectedInventoryId) ?? null;
    return {
      context: {
        principal: this.seed.principal,
        tenants: this.seed.tenants,
        inventories,
        selectedTenantId: this.selectedTenantId,
        selectedInventoryId: this.selectedInventoryId,
        assetLifecycleState: this.selectedLifecycleState,
        mediaUploadPolicy: defaultMediaUploadPolicy,
        customAssetTypes: this.effectiveCustomAssetTypes(this.selectedTenantId, this.selectedInventoryId),
        customFieldDefinitions: this.effectiveCustomFieldDefinitions(this.selectedTenantId, this.selectedInventoryId),
        capability: capabilityForInventory(selectedInventory)
      },
      assets: this.workspaceAssets(),
      checkedOutAssets: this.listCheckedOutAssetsSync(this.selectedTenantId, this.selectedInventoryId)
    };
  }

  private replaceAsset(updated: Asset): void {
    this.seed = {
      ...this.seed,
      assets: this.seed.assets.map((candidate) =>
        candidate.tenantId === updated.tenantId && candidate.inventoryId === updated.inventoryId && candidate.id === updated.id
        ? updated
        : candidate
      )
    };
  }

  private listCheckedOutAssetsSync(tenantId: string, inventoryId: string): CheckedOutAsset[] {
    return this.seed.assets
      .filter((asset) => asset.tenantId === tenantId && asset.inventoryId === inventoryId && asset.currentCheckout)
      .map((asset) => ({ asset, checkout: asset.currentCheckout! }));
  }

  private recordAssetAudit(asset: Asset, action: string): void {
    this.recordAudit({
      tenantId: asset.tenantId,
      inventoryId: asset.inventoryId,
      action,
      targetType: 'asset',
      targetId: asset.id,
      metadata: { title: asset.title }
    });
  }

  private recordInventoryAudit(inventory: Inventory, action: string): void {
    this.recordAudit({
      tenantId: inventory.tenantId,
      inventoryId: inventory.id,
      action,
      targetType: 'inventory',
      targetId: inventory.id,
      metadata: { name: inventory.name }
    });
  }

  private recordAudit(event: {
    tenantId: string;
    inventoryId: string | null;
    action: string;
    targetType: string;
    targetId: string;
    metadata: Record<string, string>;
  }): void {
    this.auditRecords = [
      {
        id: `local-audit-${this.auditRecords.length + 1}-${Date.now()}`,
        tenantId: event.tenantId,
        inventoryId: event.inventoryId,
        principalId: this.seed.principal.id,
        action: event.action,
        source: 'local_demo',
        targetType: event.targetType,
        targetId: event.targetId,
        occurredAt: new Date().toISOString(),
        metadata: event.metadata
      },
      ...this.auditRecords
    ];
  }

  private workspaceAssets(): Asset[] {
    return this.seed.assets.filter(
      (asset) =>
        asset.tenantId === this.selectedTenantId &&
        asset.inventoryId === this.selectedInventoryId &&
        asset.lifecycleState === this.selectedLifecycleState
    );
  }

  private async setAssetLifecycle(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    lifecycleState: AssetLifecycleFilter
  ): Promise<Asset> {
    const asset = await this.getAsset(tenantId, inventoryId, assetId);
    const updated: Asset = {
      ...asset,
      lifecycleState,
      updatedAt: new Date().toISOString()
    };
    this.seed = {
      ...this.seed,
      assets: this.seed.assets.map((candidate) =>
        candidate.tenantId === tenantId && candidate.inventoryId === inventoryId && candidate.id === assetId
          ? updated
          : candidate
      )
    };
    return updated;
  }

  private setAttachmentLifecycle(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string,
    lifecycleState: AssetLifecycleFilter
  ): AssetAttachment {
    const attachment = this.findAttachment(tenantId, inventoryId, assetId, attachmentId);
    const updated: AssetAttachment = { ...attachment, lifecycleState };
    this.attachments = this.attachments.map((candidate) => (candidate === attachment ? updated : candidate));
    return updated;
  }

  private findAttachment(tenantId: string, inventoryId: string, assetId: string, attachmentId: string): AssetAttachment {
    const attachment = this.attachments.find(
      (candidate) =>
        candidate.tenantId === tenantId &&
        candidate.inventoryId === inventoryId &&
        candidate.assetId === assetId &&
        candidate.id === attachmentId
    );
    if (!attachment) {
      throw new Error('Attachment not found.');
    }
    return attachment;
  }

  private findInvitation(tenantId: string, inventoryId: string, invitationId: string): InventoryAccessInvitation {
    const invitation = this.invitations.find(
      (candidate) => candidate.tenantId === tenantId && candidate.inventoryId === inventoryId && candidate.id === invitationId
    );
    if (!invitation) {
      throw new Error('Invitation not found.');
    }
    return invitation;
  }

  private effectiveCustomAssetTypes(tenantId: string, inventoryId: string): CustomAssetType[] {
    return this.seed.customAssetTypes.filter(
      (assetType) =>
        assetType.tenantId === tenantId &&
        assetType.lifecycleState === 'active' &&
        (assetType.scope === 'tenant' || assetType.inventoryId === inventoryId)
    );
  }

  private effectiveCustomFieldDefinitions(tenantId: string, inventoryId: string): CustomFieldDefinition[] {
    return this.seed.customFieldDefinitions.filter(
      (definition) =>
        definition.tenantId === tenantId &&
        definition.lifecycleState === 'active' &&
        (definition.scope === 'tenant' || definition.inventoryId === inventoryId)
    );
  }

  private customAssetTypeLabel(customAssetTypeId: string | undefined): string | undefined {
    if (!customAssetTypeId) {
      return undefined;
    }
    return this.seed.customAssetTypes.find((assetType) => assetType.id === customAssetTypeId)?.displayName;
  }

  private findCustomAssetType(
    tenantId: string,
    inventoryId: string,
    customAssetTypeId: string,
    scope: 'tenant' | 'inventory'
  ): CustomAssetType {
    const assetType = this.seed.customAssetTypes.find(
      (candidate) =>
        candidate.id === customAssetTypeId &&
        candidate.tenantId === tenantId &&
        candidate.scope === scope &&
        (scope === 'tenant' || candidate.inventoryId === inventoryId)
    );
    if (!assetType) {
      throw new Error('Custom asset type not found.');
    }
    return assetType;
  }

  private findCustomFieldDefinition(
    tenantId: string,
    inventoryId: string,
    definitionId: string,
    scope: 'tenant' | 'inventory'
  ): CustomFieldDefinition {
    const definition = this.seed.customFieldDefinitions.find(
      (candidate) =>
        candidate.id === definitionId &&
        candidate.tenantId === tenantId &&
        candidate.scope === scope &&
        (scope === 'tenant' || candidate.inventoryId === inventoryId)
    );
    if (!definition) {
      throw new Error('Custom field definition not found.');
    }
    return definition;
  }

  private validateCustomFieldTargets(tenantId: string, inventoryId: string, draft: CustomFieldDefinitionDraft): void {
    if (draft.applicability === 'all_assets') {
      if (draft.customAssetTypeIds.length > 0) {
        throw new Error('All-asset fields cannot target custom asset types.');
      }
      return;
    }
    if (draft.customAssetTypeIds.length === 0) {
      throw new Error('Custom-type fields require at least one target.');
    }
    const seen = new Set<string>();
    for (const customAssetTypeId of draft.customAssetTypeIds) {
      if (seen.has(customAssetTypeId)) {
        throw new Error('Custom field targets must be unique.');
      }
      seen.add(customAssetTypeId);
      const assetType = this.seed.customAssetTypes.find(
        (candidate) =>
          candidate.id === customAssetTypeId &&
          candidate.tenantId === tenantId &&
          candidate.lifecycleState === 'active' &&
          (candidate.scope === 'tenant' || candidate.inventoryId === inventoryId)
      );
      if (!assetType || (draft.scope === 'tenant' && assetType.scope !== 'tenant')) {
        throw new Error('Custom field target is not available.');
      }
    }
  }

  private validateInventoryScope(tenantId: string, inventoryId: string): void {
    const inventory = this.seed.inventories.find(
      (candidate) => candidate.tenantId === tenantId && candidate.id === inventoryId
    );
    if (!inventory) {
      throw new Error('Inventory not found.');
    }
  }

  private validateAssetParent(
    tenantId: string,
    inventoryId: string,
    assetId: string | null,
    parentAssetId: string | null
  ): void {
    if (!parentAssetId) {
      return;
    }
    if (parentAssetId === assetId) {
      throw new Error('Asset cannot contain itself.');
    }
    const parent = this.seed.assets.find(
      (candidate) => candidate.tenantId === tenantId && candidate.inventoryId === inventoryId && candidate.id === parentAssetId
    );
    if (!parent) {
      throw new Error('Parent asset not found.');
    }
    if (parent.lifecycleState !== 'active') {
      throw new Error('Parent asset must be active.');
    }
    if (parent.kind !== 'container' && parent.kind !== 'location') {
      throw new Error('Parent asset must be a container or location.');
    }
    if (assetId && this.isDescendant(parent, assetId)) {
      throw new Error('Asset cannot be moved inside its own contents.');
    }
  }

  private isDescendant(candidate: Asset, ancestorAssetId: string): boolean {
    let parentAssetId = candidate.parentAssetId;
    while (parentAssetId) {
      if (parentAssetId === ancestorAssetId) {
        return true;
      }
      const parent = this.seed.assets.find(
        (asset) =>
          asset.tenantId === candidate.tenantId &&
          asset.inventoryId === candidate.inventoryId &&
          asset.id === parentAssetId
      );
      parentAssetId = parent?.parentAssetId ?? null;
    }
    return false;
  }

  private nextAssetId(): string {
    let id = '';
    do {
      id = `asset-local-${this.nextAssetSequence}`;
      this.nextAssetSequence += 1;
    } while (this.seed.assets.some((asset) => asset.id === id));
    return id;
  }

  private firstInventoryIdForTenant(tenantId: string): string {
    return this.seed.inventories.find((inventory) => inventory.tenantId === tenantId)?.id ?? '';
  }
}

function capabilityForInventory(inventory: Inventory | null): 'editor' | 'viewer' {
  if (canEditInventory(inventory)) {
    return 'editor';
  }
  return 'viewer';
}

function exactAssets(assets: Asset[], query: string): Asset[] {
  const normalized = query.trim().toLowerCase();
  if (!normalized) {
    return [];
  }
  return assets.filter(
    (asset) =>
      asset.title.toLowerCase() === normalized ||
      asset.description.toLowerCase() === normalized ||
      asset.customAssetTypeLabel?.toLowerCase() === normalized
  );
}

function importJobProgressHistoryWith(history: ImportJob['progressHistory'], progress: ImportJob['progress']): ImportJob['progressHistory'] {
  const existing = [...history];
  const last = existing.at(-1);
  if (last?.phase === progress.phase && last.message === progress.message) {
    existing[existing.length - 1] = progress;
    return existing;
  }
  return [...existing, progress];
}

function page<T>(items: T[], cursor?: string): InventoryAccessPage<T> {
  const limit = 50;
  const start = cursor ? Number.parseInt(cursor, 10) : 0;
  const safeStart = Number.isFinite(start) && start > 0 ? start : 0;
  const selected = items.slice(safeStart, safeStart + limit);
  const nextIndex = safeStart + selected.length;
  return {
    items: selected,
    pagination: {
      limit,
      nextCursor: nextIndex < items.length ? String(nextIndex) : null,
      hasMore: nextIndex < items.length
    }
  };
}
