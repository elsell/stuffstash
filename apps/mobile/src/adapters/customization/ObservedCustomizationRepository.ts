import type { CustomizationRepository } from '../../application/customization/CustomizationRepository';
import type { CustomizationMutation, CustomizationMutationObserver } from '../../application/customization/CustomizationMutationObserver';

export class ObservedCustomizationRepository implements CustomizationRepository {
  constructor(private readonly repository: CustomizationRepository, private readonly observer: CustomizationMutationObserver) {}
  listTags(...args: Parameters<CustomizationRepository['listTags']>) { return this.repository.listTags(...args); }
  listFields(...args: Parameters<CustomizationRepository['listFields']>) { return this.repository.listFields(...args); }
  listAssetTypes(...args: Parameters<CustomizationRepository['listAssetTypes']>) { return this.repository.listAssetTypes(...args); }
  createTag(...args: Parameters<CustomizationRepository['createTag']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'tag' }, () => this.repository.createTag(...args)); }
  updateTag(...args: Parameters<CustomizationRepository['updateTag']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'tag' }, () => this.repository.updateTag(...args)); }
  archiveTag(...args: Parameters<CustomizationRepository['archiveTag']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'tag' }, () => this.repository.archiveTag(...args)); }
  createField(...args: Parameters<CustomizationRepository['createField']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[1] === 'inventory' ? args[0].inventoryId : undefined, kind: 'field' }, () => this.repository.createField(...args)); }
  updateField(...args: Parameters<CustomizationRepository['updateField']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'field' }, () => this.repository.updateField(...args)); }
  archiveField(...args: Parameters<CustomizationRepository['archiveField']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'field' }, () => this.repository.archiveField(...args)); }
  restoreField(...args: Parameters<CustomizationRepository['restoreField']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'field' }, () => this.repository.restoreField(...args)); }
  deleteField(...args: Parameters<CustomizationRepository['deleteField']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'field' }, () => this.repository.deleteField(...args)); }
  createAssetType(...args: Parameters<CustomizationRepository['createAssetType']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[1] === 'inventory' ? args[0].inventoryId : undefined, kind: 'asset-type' }, () => this.repository.createAssetType(...args)); }
  updateAssetType(...args: Parameters<CustomizationRepository['updateAssetType']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'asset-type' }, () => this.repository.updateAssetType(...args)); }
  archiveAssetType(...args: Parameters<CustomizationRepository['archiveAssetType']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'asset-type' }, () => this.repository.archiveAssetType(...args)); }
  restoreAssetType(...args: Parameters<CustomizationRepository['restoreAssetType']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'asset-type' }, () => this.repository.restoreAssetType(...args)); }
  deleteAssetType(...args: Parameters<CustomizationRepository['deleteAssetType']>) { return this.changed({ tenantId: args[0].tenantId, inventoryId: args[0].inventoryId, kind: 'asset-type' }, () => this.repository.deleteAssetType(...args)); }

  private async changed<T>(mutation: CustomizationMutation, operation: () => Promise<T>): Promise<T> {
    const result = await operation();
    this.observer.onCustomizationChanged(mutation);
    return result;
  }
}
