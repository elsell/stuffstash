import type { CustomFieldDefinition, CustomizationScope } from '../../domain/customization/Customization';
import { customizationKeyIsValid, customizationKeyValidationMessage, suggestedCustomizationKey } from '../../domain/customization/Customization';
import type { CreateCustomFieldInput, CustomizationContext, CustomizationRepository, DefinitionAddress, UpdateCustomFieldInput } from './CustomizationRepository';
import { CustomizationValidationError } from './CustomizationErrors';
import type { CustomizationObservability } from './CustomizationObservability';

export class ManageCustomFields {
  private saving = false;
  constructor(private readonly repository: CustomizationRepository, private readonly observability: CustomizationObservability) {}

  async create(context: CustomizationContext, scope: CustomizationScope, input: CreateCustomFieldInput) {
    const key = input.key.trim() || suggestedCustomizationKey(input.displayName);
    validateKey(key);
    validateField(input);
    return this.singleFlight(scope, 'create', () => this.repository.createField(context, scope, { ...input, key, displayName: input.displayName.trim() }));
  }

  async update(address: DefinitionAddress, original: CustomFieldDefinition, input: UpdateCustomFieldInput) {
    if (input.enumOptions && original.enumOptions.some((option, index) => input.enumOptions?.[index] !== option)) {
      throw new CustomizationValidationError('Existing options cannot be renamed, removed, or reordered.');
    }
    if (original.applicability === 'all_assets' && input.applicability === 'custom_asset_types') {
      throw new CustomizationValidationError('A field that applies to all assets cannot be narrowed.');
    }
    return this.singleFlight(address.scope, 'update', () => this.repository.updateField(address, input));
  }

  archive(address: DefinitionAddress) { return this.singleFlight(address.scope, 'archive', () => this.repository.archiveField(address)); }
  restore(address: DefinitionAddress) { return this.singleFlight(address.scope, 'restore', () => this.repository.restoreField(address)); }
  delete(address: DefinitionAddress) { return this.singleFlight(address.scope, 'delete', () => this.repository.deleteField(address)); }

  private async singleFlight<T>(scope: CustomizationScope, action: 'create' | 'update' | 'archive' | 'restore' | 'delete', operation: () => Promise<T>): Promise<T> {
    if (this.saving) throw new CustomizationValidationError('This custom field change is already being saved.');
    this.saving = true;
    this.observability.record({ name: 'customization.mutation_requested', resource: 'field', scope, action });
    try { const result = await operation(); this.observability.record({ name: 'customization.mutation_succeeded', resource: 'field', scope, action }); return result; }
    catch (error) { this.observability.record({ name: 'customization.mutation_failed', resource: 'field', scope, action }); throw error; }
    finally { this.saving = false; }
  }
}

function validateField(input: CreateCustomFieldInput): void {
  if (!input.displayName.trim()) throw new CustomizationValidationError('Field name is required.');
  if (input.type === 'enum' && input.enumOptions.length === 0) throw new CustomizationValidationError('Add at least one enum option.');
  if (input.type !== 'enum' && input.enumOptions.length > 0) throw new CustomizationValidationError('Only enum fields can have options.');
  if (input.applicability === 'custom_asset_types' && input.customAssetTypeIds.length === 0) throw new CustomizationValidationError('Choose at least one asset type.');
}

function validateKey(key: string): void {
  if (!customizationKeyIsValid(key)) throw new CustomizationValidationError(customizationKeyValidationMessage);
}
