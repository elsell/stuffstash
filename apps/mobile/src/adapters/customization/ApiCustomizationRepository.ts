import { StuffStashAPIError, type StuffStashClient } from '@stuff-stash/api-client';
import { CustomizationFailure, type CustomizationFailureKind } from '../../application/customization/CustomizationErrors';
import type { AssetTagDefinition, CustomAssetTypeDefinition, CustomFieldDefinition, CustomizationLifecycle, CustomizationScope } from '../../domain/customization/Customization';
import type {
  CreateCustomAssetTypeInput,
  CreateCustomFieldInput,
  CustomizationContext,
  CustomizationPage,
  CustomizationRepository,
  DefinitionAddress,
  UpdateCustomAssetTypeInput,
  UpdateCustomFieldInput
} from '../../application/customization/CustomizationRepository';

type Page<T> = { readonly items: readonly T[]; readonly pagination: { readonly nextCursor?: string | null } };
type ApiTag = { readonly id: string; readonly key: string; readonly displayName: string; readonly color?: string };
type ApiField = Omit<CustomFieldDefinition, 'kind' | 'lifecycle' | 'inventoryId'> & { readonly inventoryId: string | null; readonly lifecycleState: 'active' | 'archived' };
type ApiAssetType = Omit<CustomAssetTypeDefinition, 'kind' | 'lifecycle' | 'inventoryId'> & { readonly inventoryId: string | null; readonly lifecycleState: 'active' | 'archived' };

type ApiClient = Pick<StuffStashClient,
  | 'listAssetTags' | 'createAssetTag' | 'updateAssetTag' | 'archiveAssetTag'
  | 'listTenantCustomFieldDefinitions' | 'listInventoryCustomFieldDefinitions'
  | 'createTenantCustomFieldDefinition' | 'createInventoryCustomFieldDefinition'
  | 'updateTenantCustomFieldDefinition' | 'updateInventoryCustomFieldDefinition'
  | 'archiveTenantCustomFieldDefinition' | 'archiveInventoryCustomFieldDefinition'
  | 'restoreTenantCustomFieldDefinition' | 'restoreInventoryCustomFieldDefinition'
  | 'deleteTenantCustomFieldDefinition' | 'deleteInventoryCustomFieldDefinition'
  | 'listTenantCustomAssetTypes' | 'listInventoryCustomAssetTypes'
  | 'createTenantCustomAssetType' | 'createInventoryCustomAssetType'
  | 'updateTenantCustomAssetType' | 'updateInventoryCustomAssetType'
  | 'archiveTenantCustomAssetType' | 'archiveInventoryCustomAssetType'
  | 'restoreTenantCustomAssetType' | 'restoreInventoryCustomAssetType'
  | 'deleteTenantCustomAssetType' | 'deleteInventoryCustomAssetType'
>;

export class ApiCustomizationRepository implements CustomizationRepository {
  private readonly api: ApiClient;
  constructor(client: StuffStashClient) { this.api = client as ApiClient; }

  async listTags(context: CustomizationContext, cursor?: string): Promise<CustomizationPage<AssetTagDefinition>> {
    const page = await this.safe(() => this.api.listAssetTags(context.tenantId, context.inventoryId, 50, cursor));
    return mapPage(page, (item) => ({ kind: 'tag', ...item }));
  }
  async createTag(context: CustomizationContext, input: { readonly displayName: string; readonly color?: string }) {
    return { kind: 'tag' as const, ...await this.safe(() => this.api.createAssetTag(context.tenantId, context.inventoryId, input)) };
  }
  async updateTag(context: CustomizationContext, id: string, input: { readonly displayName?: string; readonly color?: string }) {
    return { kind: 'tag' as const, ...await this.safe(() => this.api.updateAssetTag(context.tenantId, context.inventoryId, id, input)) };
  }
  async archiveTag(context: CustomizationContext, id: string) { await this.safe(() => this.api.archiveAssetTag(context.tenantId, context.inventoryId, id)); }

  async listFields(context: CustomizationContext, scope: CustomizationScope, lifecycle: CustomizationLifecycle, cursor?: string) {
    const page = scope === 'tenant'
      ? await this.safe(() => this.api.listTenantCustomFieldDefinitions(context.tenantId, 50, cursor, lifecycle))
      : await this.safe(() => this.api.listInventoryCustomFieldDefinitions(context.tenantId, context.inventoryId, 50, cursor, lifecycle));
    return mapPage(page as Page<ApiField>, mapField);
  }
  async createField(context: CustomizationContext, scope: CustomizationScope, input: CreateCustomFieldInput) {
    const body = { ...input, enumOptions: [...input.enumOptions], customAssetTypeIds: [...input.customAssetTypeIds] };
    const item = scope === 'tenant'
      ? await this.safe(() => this.api.createTenantCustomFieldDefinition(context.tenantId, body))
      : await this.safe(() => this.api.createInventoryCustomFieldDefinition(context.tenantId, context.inventoryId, body));
    return mapField(item as ApiField);
  }
  async updateField(address: DefinitionAddress, input: UpdateCustomFieldInput) {
    const body = { ...input, enumOptions: input.enumOptions ? [...input.enumOptions] : undefined, customAssetTypeIds: input.customAssetTypeIds ? [...input.customAssetTypeIds] : undefined };
    const item = address.scope === 'tenant'
      ? await this.safe(() => this.api.updateTenantCustomFieldDefinition(address.tenantId, address.id, body))
      : await this.safe(() => this.api.updateInventoryCustomFieldDefinition(address.tenantId, requiredInventory(address), address.id, body));
    return mapField(item as ApiField);
  }
  async archiveField(address: DefinitionAddress) { await this.fieldLifecycle(address, 'archive'); }
  async restoreField(address: DefinitionAddress) { await this.fieldLifecycle(address, 'restore'); }
  async deleteField(address: DefinitionAddress) { await this.fieldLifecycle(address, 'delete'); }

  async listAssetTypes(context: CustomizationContext, scope: CustomizationScope, lifecycle: CustomizationLifecycle, cursor?: string) {
    const page = scope === 'tenant'
      ? await this.safe(() => this.api.listTenantCustomAssetTypes(context.tenantId, 50, cursor, lifecycle))
      : await this.safe(() => this.api.listInventoryCustomAssetTypes(context.tenantId, context.inventoryId, 50, cursor, lifecycle));
    return mapPage(page as Page<ApiAssetType>, mapAssetType);
  }
  async createAssetType(context: CustomizationContext, scope: CustomizationScope, input: CreateCustomAssetTypeInput) {
    const item = scope === 'tenant'
      ? await this.safe(() => this.api.createTenantCustomAssetType(context.tenantId, input))
      : await this.safe(() => this.api.createInventoryCustomAssetType(context.tenantId, context.inventoryId, input));
    return mapAssetType(item as ApiAssetType);
  }
  async updateAssetType(address: DefinitionAddress, input: UpdateCustomAssetTypeInput) {
    const item = address.scope === 'tenant'
      ? await this.safe(() => this.api.updateTenantCustomAssetType(address.tenantId, address.id, input))
      : await this.safe(() => this.api.updateInventoryCustomAssetType(address.tenantId, requiredInventory(address), address.id, input));
    return mapAssetType(item as ApiAssetType);
  }
  async archiveAssetType(address: DefinitionAddress) { await this.assetTypeLifecycle(address, 'archive'); }
  async restoreAssetType(address: DefinitionAddress) { await this.assetTypeLifecycle(address, 'restore'); }
  async deleteAssetType(address: DefinitionAddress) { await this.assetTypeLifecycle(address, 'delete'); }

  private async fieldLifecycle(address: DefinitionAddress, action: string) {
    if (action === 'archive') {
      if (address.scope === 'tenant') await this.safe(() => this.api.archiveTenantCustomFieldDefinition(address.tenantId, address.id));
      else await this.safe(() => this.api.archiveInventoryCustomFieldDefinition(address.tenantId, requiredInventory(address), address.id));
      return;
    }
    if (action === 'restore') {
      if (address.scope === 'tenant') await this.safe(() => this.api.restoreTenantCustomFieldDefinition(address.tenantId, address.id));
      else await this.safe(() => this.api.restoreInventoryCustomFieldDefinition(address.tenantId, requiredInventory(address), address.id));
    } else if (address.scope === 'tenant') await this.safe(() => this.api.deleteTenantCustomFieldDefinition(address.tenantId, address.id));
    else await this.safe(() => this.api.deleteInventoryCustomFieldDefinition(address.tenantId, requiredInventory(address), address.id));
  }
  private async assetTypeLifecycle(address: DefinitionAddress, action: string) {
    if (action === 'archive') {
      if (address.scope === 'tenant') await this.safe(() => this.api.archiveTenantCustomAssetType(address.tenantId, address.id));
      else await this.safe(() => this.api.archiveInventoryCustomAssetType(address.tenantId, requiredInventory(address), address.id));
      return;
    }
    if (action === 'restore') {
      if (address.scope === 'tenant') await this.safe(() => this.api.restoreTenantCustomAssetType(address.tenantId, address.id));
      else await this.safe(() => this.api.restoreInventoryCustomAssetType(address.tenantId, requiredInventory(address), address.id));
    } else if (address.scope === 'tenant') await this.safe(() => this.api.deleteTenantCustomAssetType(address.tenantId, address.id));
    else await this.safe(() => this.api.deleteInventoryCustomAssetType(address.tenantId, requiredInventory(address), address.id));
  }

  private async safe<T>(operation: () => Promise<T>): Promise<T> {
    try { return await operation(); }
    catch (error) {
      if (error instanceof CustomizationFailure) throw error;
      throw new CustomizationFailure(mapFailure(error));
    }
  }
}

function mapPage<A, B>(page: Page<A>, mapper: (item: A) => B): CustomizationPage<B> {
  return { items: page.items.map(mapper), nextCursor: page.pagination.nextCursor ?? undefined };
}
function mapField(item: ApiField): CustomFieldDefinition { return { ...item, kind: 'field', inventoryId: item.inventoryId ?? undefined, lifecycle: item.lifecycleState }; }
function mapAssetType(item: ApiAssetType): CustomAssetTypeDefinition { return { ...item, kind: 'asset-type', inventoryId: item.inventoryId ?? undefined, lifecycle: item.lifecycleState }; }
function requiredInventory(address: DefinitionAddress): string { if (!address.inventoryId) throw new Error('Inventory scope is required.'); return address.inventoryId; }
function mapFailure(error: unknown): CustomizationFailureKind {
  if (!(error instanceof StuffStashAPIError)) return 'unavailable';
  if (error.status === 401 || error.status === 403) return 'permission-denied';
  if (error.status === 404) return 'not-found';
  if (error.status === 409 || error.code.includes('conflict') || error.code.includes('in_use')) return 'conflict';
  if (error.status === 400 || error.status === 422) return 'invalid';
  return 'unavailable';
}
