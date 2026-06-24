import {
  type AssetLifecycleFilter,
  canCreateInventory,
  canEditInventory,
  type AddAssetDraft,
  type Asset,
  type AssetAttachment,
  type CreatedInventoryAccessInvitation,
  defaultMediaUploadPolicy,
  type Inventory,
  type InventoryAccessGrant,
  type InventoryAccessInvitation,
  type InventoryAccessRelationship,
  type InvitationStatusFilter,
  type SearchRequest,
  type SearchResult,
  type SelectedPhoto,
  type UpdateAssetDraft,
  type WorkspaceData
} from '$lib/domain/inventory';
import type { InventoryRepository, WorkspaceSeed } from '$lib/ports/inventoryRepository';
import type { InventoryAccessPage, InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
import { filterAssets } from '$lib/application/workspace';

export class SeededInventoryRepository implements InventoryRepository, InventoryAccessRepository {
  private seed: WorkspaceSeed;
  private attachments: AssetAttachment[] = [];
  private grants: InventoryAccessGrant[] = [];
  private invitations: InventoryAccessInvitation[] = [];
  private selectedTenantId: string;
  private selectedInventoryId: string;
  private selectedLifecycleState: AssetLifecycleFilter = 'active';

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
        permissions: ['view', 'create_asset', 'edit_asset', 'share', 'configure']
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
        permissions: ['view', 'create_asset', 'edit_asset', 'share', 'configure']
      }
    };
    this.seed = {
      ...this.seed,
      inventories: [inventory, ...this.seed.inventories]
    };
    this.selectedTenantId = tenantId;
    this.selectedInventoryId = inventory.id;
    this.selectedLifecycleState = 'active';
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
    const asset: Asset = {
      id: `asset-${Date.now()}`,
      tenantId,
      inventoryId,
      kind: draft.kind,
      title: draft.title,
      description: draft.description,
      parentAssetId: draft.parentAssetId,
      lifecycleState: 'active',
      photo: draft.photos[0]
        ? {
            id: draft.photos[0].id,
            url: draft.photos[0].previewUrl,
            alt: draft.title
          }
        : undefined,
      updatedAt: new Date().toISOString()
    };
    this.seed = { ...this.seed, assets: [asset, ...this.seed.assets] };
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
    const updated: Asset = {
      ...asset,
      title: draft.title,
      description: draft.description,
      parentAssetId: draft.parentAssetId,
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

  async archiveAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    return this.setAssetLifecycle(tenantId, inventoryId, assetId, 'archived');
  }

  async restoreAsset(tenantId: string, inventoryId: string, assetId: string): Promise<Asset> {
    return this.setAssetLifecycle(tenantId, inventoryId, assetId, 'active');
  }

  async deleteAsset(tenantId: string, inventoryId: string, assetId: string): Promise<void> {
    const asset = await this.getAsset(tenantId, inventoryId, assetId);
    this.seed = {
      ...this.seed,
      assets: this.seed.assets.filter((candidate) => candidate !== asset)
    };
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

  async uploadAssetPhoto(tenantId: string, inventoryId: string, assetId: string, photo: SelectedPhoto): Promise<AssetAttachment> {
    await this.getAsset(tenantId, inventoryId, assetId);
    const attachment: AssetAttachment = {
      id: `attachment-${Date.now()}`,
      tenantId,
      inventoryId,
      assetId,
      fileName: photo.name,
      contentType: photo.contentType,
      sizeBytes: photo.sizeBytes,
      lifecycleState: 'active',
      thumbnailUrl: photo.previewUrl
    };
    this.attachments = [attachment, ...this.attachments];
    return attachment;
  }

  async archiveAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<AssetAttachment> {
    return this.setAttachmentLifecycle(tenantId, inventoryId, assetId, attachmentId, 'archived');
  }

  async restoreAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<AssetAttachment> {
    return this.setAttachmentLifecycle(tenantId, inventoryId, assetId, attachmentId, 'active');
  }

  async deleteAssetAttachment(
    tenantId: string,
    inventoryId: string,
    assetId: string,
    attachmentId: string
  ): Promise<void> {
    const attachment = this.findAttachment(tenantId, inventoryId, assetId, attachmentId);
    this.attachments = this.attachments.filter((candidate) => candidate !== attachment);
  }

  async searchAssets(request: SearchRequest): Promise<SearchResult[]> {
    const inventory = this.seed.inventories.find((candidate) => candidate.id === this.selectedInventoryId);
    const searchableAssets = this.seed.assets.filter(
      (asset) =>
        asset.tenantId === request.tenantId &&
        asset.inventoryId === request.inventoryId &&
        (request.lifecycleState === 'all' || asset.lifecycleState === request.lifecycleState)
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
    return grant;
  }

  async revokeInventoryAccess(
    tenantId: string,
    inventoryId: string,
    principalId: string,
    relationship: InventoryAccessRelationship
  ): Promise<void> {
    this.grants = this.grants.filter(
      (grant) =>
        !(
          grant.tenantId === tenantId &&
          grant.inventoryId === inventoryId &&
          grant.principalId === principalId &&
          grant.relationship === relationship
        )
    );
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
    return updated;
  }

  async cancelInventoryAccessInvitation(tenantId: string, inventoryId: string, invitationId: string): Promise<void> {
    const invitation = this.findInvitation(tenantId, inventoryId, invitationId);
    const updated: InventoryAccessInvitation = { ...invitation, status: 'cancelled' };
    this.invitations = this.invitations.map((candidate) => (candidate === invitation ? updated : candidate));
  }

  async deleteInventoryAccessInvitation(tenantId: string, inventoryId: string, invitationId: string): Promise<void> {
    const invitation = this.findInvitation(tenantId, inventoryId, invitationId);
    this.invitations = this.invitations.filter((candidate) => candidate !== invitation);
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
        capability: capabilityForInventory(selectedInventory)
      },
      assets: this.workspaceAssets()
    };
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
