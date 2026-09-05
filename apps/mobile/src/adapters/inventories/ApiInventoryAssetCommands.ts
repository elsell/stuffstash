import {
  type InventoryMutationKind,
  type InventoryMutationObserver
} from '../../application/home/InventoryMutationObserver';
import {
  AddInventoryAssetPhotoInput,
  CreateInventoryAssetInput,
  CreateInventoryAssetPhotoInput,
  CreateInventoryAssetTagInput,
  UpdateInventoryAssetInput
} from '../../application/home/InventorySummaryRepository';
import { assetId, AssetSummary, type AssetTagSummary } from '../../domain/assets/AssetSummary';
import {
  inventoryId,
  tenantId
} from '../../domain/inventories/InventorySummary';
import {
  isDirectUploadTargetSupported,
  type DirectUploadTargetPolicy
} from '../uploads/DirectUploadPolicy';
import { ApiInventoryDirectory } from './ApiInventoryDirectory';

import { attachmentContentBase64, type DirectUploadTransport } from '../uploads/ExpoDirectUploadTransport';
import { ApiInventoryAssetTraversal } from './ApiInventoryAssetTraversal';
import type { InventoryApiClient } from './InventoryApiClient';
import { ancestorTrail, emptyInventorySummary, mapAsset, mapAssetTag, placementAssetsWithSelectedOverrides } from './InventoryAssetMapping';
export class ApiInventoryAssetCommands {
  constructor(private readonly client: InventoryApiClient, private readonly directory: ApiInventoryDirectory, private readonly traversal: ApiInventoryAssetTraversal, private readonly directUploadTransport: DirectUploadTransport, private readonly directUploadPolicy: DirectUploadTargetPolicy, private readonly mutationObserver: InventoryMutationObserver) {}
  async createAsset(input: CreateInventoryAssetInput): Promise<AssetSummary> {
    const selected = await this.directory.selectedForCommand();
    const inventory = emptyInventorySummary(selected.tenant, selected.inventory);
    const parent = input.parentAssetId ? await this.client.getAsset(selected.tenant.id, selected.inventory.id, input.parentAssetId) : undefined;
    const currentPlacementAssets = parent ? [parent, ...await this.traversal.loadAssetAncestors(parent)] : [];
    const asset = await this.client.createAsset(inventory.tenantId, inventory.id, {
      kind: input.kind,
      title: input.title,
      description: input.description,
      parentAssetId: input.parentAssetId,
      ...(input.tagIds !== undefined ? { tagIds: [...input.tagIds] } : {})
    });
    this.observeMutation(
      'asset_created',
      inventory.tenantId,
      inventory.id,
      asset.id,
      ancestorTrail(asset, currentPlacementAssets).map((ancestor) => ancestor.id),
      parent?.kind === 'item' ? parent.id : undefined
    );

    const placementAssets = placementAssetsWithSelectedOverrides(
      currentPlacementAssets,
      [asset]
    );
    return mapAsset(inventory.name, asset, placementAssets);
  }

  async createAssetTag(input: CreateInventoryAssetTagInput): Promise<AssetTagSummary> {
    const selected = await this.directory.selectedForCommand();
    const tag = await this.client.createAssetTag(selected.tenant.id, selected.inventory.id, {
      displayName: input.displayName,
      ...(input.color !== undefined ? { color: input.color } : {})
    });
    this.observeMutation('asset_tag_created', selected.tenant.id, selected.inventory.id);
    return mapAssetTag(tag);
  }

  async updateAsset(input: UpdateInventoryAssetInput): Promise<AssetSummary> {
    const selected = await this.directory.selectedForCommand();
    const inventory = emptyInventorySummary(selected.tenant, selected.inventory);
    const previous = await this.client.getAsset(selected.tenant.id, selected.inventory.id, input.assetId);
    const destination = { ...previous, parentAssetId: input.parentAssetId !== undefined ? input.parentAssetId : previous.parentAssetId };
    const placementAssets = [previous, ...await this.traversal.loadAncestorsForAssets([previous, destination])];
    const asset = await this.client.updateAsset(inventory.tenantId, inventory.id, input.assetId, {
      ...(input.title !== undefined ? { title: input.title } : {}),
      ...(input.description !== undefined ? { description: input.description } : {}),
      ...(input.parentAssetId !== undefined ? { parentAssetId: input.parentAssetId } : {}),
      ...(input.tagIds !== undefined ? { tagIds: [...input.tagIds] } : {})
    });
    this.observeMutation(
      'asset_updated',
      inventory.tenantId,
      inventory.id,
      asset.id,
      [
        ...ancestorTrail(previous, placementAssets),
        ...ancestorTrail(asset, placementAssets)
      ].map((ancestor) => ancestor.id)
    );

    const knownAssets = placementAssetsWithSelectedOverrides(
      placementAssets,
      [asset]
    );
    return mapAsset(inventory.name, asset, knownAssets);
  }

  async addAssetPhoto(
    assetIdValue: AssetSummary['id'],
    input: CreateInventoryAssetPhotoInput
  ): Promise<void> {
    const selected = await this.directory.selectedForCommand();
    await this.addInventoryAssetPhoto({
      tenantId: tenantId(selected.tenant.id),
      inventoryId: inventoryId(selected.inventory.id),
      assetId: assetIdValue,
      ...input
    });
  }

  async addInventoryAssetPhoto(input: AddInventoryAssetPhotoInput): Promise<void> {
    if (input.uri && input.sizeBytes && input.sizeBytes > 0) {
      const directUpload = input.directUpload ?? await this.client.initiateAssetAttachmentDirectUpload(input.tenantId, input.inventoryId, input.assetId, {
        fileName: input.fileName,
        contentType: input.contentType,
        sizeBytes: input.sizeBytes
      });
      if (!isDirectUploadTargetSupported(directUpload.url, this.directUploadPolicy)) {
        throw new Error('Unsupported direct attachment upload target.');
      }
      const uploaded = await this.directUploadTransport.upload({
        upload: directUpload,
        fileUri: input.uri,
        fileName: input.fileName,
        contentType: input.contentType
      });
      if (uploaded) {
        await this.client.completeAssetAttachmentDirectUpload(input.tenantId, input.inventoryId, input.assetId, directUpload.uploadId);
        this.observeMutation('asset_photo_changed', input.tenantId, input.inventoryId, input.assetId);
        return;
      }
    }

    await this.client.createAssetAttachment(input.tenantId, input.inventoryId, input.assetId, {
      fileName: input.fileName,
      contentType: input.contentType,
      contentBase64: await attachmentContentBase64(input)
    });
    this.observeMutation('asset_photo_changed', input.tenantId, input.inventoryId, input.assetId);
  }

  async deleteAssetPhoto(assetIdValue: AssetSummary['id'], photoId: string): Promise<void> {
    const selected = await this.directory.selectedForCommand();
    await this.client.deleteAssetAttachment(selected.tenant.id, selected.inventory.id, assetIdValue, photoId);
    this.observeMutation('asset_photo_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
  }

  async archiveAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    const selected = await this.directory.selectedForCommand();
    await this.client.archiveAsset(selected.tenant.id, selected.inventory.id, assetIdValue);
    this.observeMutation('asset_lifecycle_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
  }

  async restoreAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    const selected = await this.directory.selectedForCommand();
    await this.client.restoreAsset(selected.tenant.id, selected.inventory.id, assetIdValue);
    this.observeMutation('asset_lifecycle_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
  }

  async deleteAsset(assetIdValue: AssetSummary['id']): Promise<void> {
    const selected = await this.directory.selectedForCommand();
    await this.client.deleteAsset(selected.tenant.id, selected.inventory.id, assetIdValue);
    this.observeMutation('asset_lifecycle_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
  }

  async checkoutAsset(assetIdValue: AssetSummary['id'], input: { readonly details?: string } = {}) {
    const selected = await this.directory.selectedForCommand();
    const checkout = await this.client.checkoutAsset(selected.tenant.id, selected.inventory.id, assetIdValue, input);
    this.observeMutation('asset_checkout_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
    return {
      id: checkout.id,
      assetId: assetId(checkout.assetId),
      undoableOperationId: checkout.undoableOperationId
    };
  }

  async returnAsset(assetIdValue: AssetSummary['id'], input: { readonly details?: string } = {}) {
    const selected = await this.directory.selectedForCommand();
    const checkout = await this.client.returnAsset(selected.tenant.id, selected.inventory.id, assetIdValue, input);
    this.observeMutation('asset_checkout_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
    return {
      id: checkout.id,
      assetId: assetId(checkout.assetId),
      undoableOperationId: checkout.undoableOperationId
    };
  }

  async updateReturnedCheckoutDetails(assetIdValue: AssetSummary['id'], checkoutId: string, input: { readonly details?: string } = {}) {
    const selected = await this.directory.selectedForCommand();
    const checkout = await this.client.updateReturnedCheckoutDetails(selected.tenant.id, selected.inventory.id, assetIdValue, checkoutId, input);
    this.observeMutation('asset_checkout_changed', selected.tenant.id, selected.inventory.id, assetIdValue);
    return {
      id: checkout.id,
      assetId: assetId(checkout.assetId),
      undoableOperationId: checkout.undoableOperationId
    };
  }

  async undoInventoryOperation(operationId: string): Promise<void> {
    const selected = await this.directory.selectedForCommand();
    const asset = await this.client.applyUndoableOperation(selected.tenant.id, selected.inventory.id, operationId, 'undo');
    this.observeMutation('operation_reversed', selected.tenant.id, selected.inventory.id, asset.id, asset.parentAssetId ? [asset.parentAssetId] : []);
  }

  private observeMutation(
    kind: InventoryMutationKind,
    tenantID: string,
    inventoryID: string,
    assetID?: string,
    relatedAssetIDs: readonly string[] = [],
    promotedParentId?: string
  ): void {
    this.mutationObserver.onInventoryMutation({
      kind,
      tenantId: tenantID,
      inventoryId: inventoryID,
      ...(assetID ? { assetId: assetID } : {}),
      ...(promotedParentId ? { promotedParentId } : {}),
      ...(relatedAssetIDs.length > 0 ? { relatedAssetIds: [...new Set(relatedAssetIDs)] } : {})
    });
  }

}
