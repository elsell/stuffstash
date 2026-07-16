import { customizationKeyIsValid, customizationKeyValidationMessage, suggestedCustomizationKey, type CustomizationScope } from '../../domain/customization/Customization';
import type { CreateCustomAssetTypeInput, CustomizationContext, CustomizationRepository, DefinitionAddress, UpdateCustomAssetTypeInput } from './CustomizationRepository';
import { CustomizationValidationError } from './CustomizationErrors';
import type { CustomizationObservability } from './CustomizationObservability';

export class ManageCustomAssetTypes {
  private saving = false;
  constructor(private readonly repository: CustomizationRepository, private readonly observability: CustomizationObservability) {}

  async create(context: CustomizationContext, scope: CustomizationScope, input: CreateCustomAssetTypeInput) {
    const key = input.key.trim() || suggestedCustomizationKey(input.displayName);
    if (!customizationKeyIsValid(key)) throw new CustomizationValidationError(customizationKeyValidationMessage);
    if (!input.displayName.trim()) throw new CustomizationValidationError('Asset type name is required.');
    if (input.description.trim().length > 1000) throw new CustomizationValidationError('Description must be 1,000 characters or fewer.');
    return this.singleFlight(scope, 'create', () => this.repository.createAssetType(context, scope, {
      key,
      displayName: input.displayName.trim(),
      description: input.description.trim()
    }));
  }

  update(address: DefinitionAddress, input: UpdateCustomAssetTypeInput) {
    return this.singleFlight(address.scope, 'update', () => this.repository.updateAssetType(address, input));
  }
  archive(address: DefinitionAddress) { return this.singleFlight(address.scope, 'archive', () => this.repository.archiveAssetType(address)); }
  restore(address: DefinitionAddress) { return this.singleFlight(address.scope, 'restore', () => this.repository.restoreAssetType(address)); }
  delete(address: DefinitionAddress) { return this.singleFlight(address.scope, 'delete', () => this.repository.deleteAssetType(address)); }

  private async singleFlight<T>(scope: CustomizationScope, action: 'create' | 'update' | 'archive' | 'restore' | 'delete', operation: () => Promise<T>): Promise<T> {
    if (this.saving) throw new CustomizationValidationError('This asset type change is already being saved.');
    this.saving = true;
    this.observability.record({ name: 'customization.mutation_requested', resource: 'asset-type', scope, action });
    try { const result = await operation(); this.observability.record({ name: 'customization.mutation_succeeded', resource: 'asset-type', scope, action }); return result; }
    catch (error) { this.observability.record({ name: 'customization.mutation_failed', resource: 'asset-type', scope, action }); throw error; }
    finally { this.saving = false; }
  }
}
