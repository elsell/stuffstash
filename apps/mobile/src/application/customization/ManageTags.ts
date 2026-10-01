import { t } from '../../presentation/localization';
import { normalizeTagColor } from '../../domain/customization/Customization';
import type { CustomizationContext, CustomizationRepository } from './CustomizationRepository';
import { CustomizationValidationError } from './CustomizationErrors';
import type { CustomizationObservability } from './CustomizationObservability';

export class ManageTags {
  private saving = false;
  constructor(private readonly repository: CustomizationRepository, private readonly observability: CustomizationObservability) {}

  async create(context: CustomizationContext, input: { readonly displayName: string; readonly color?: string }) {
    const color = validatedColor(input.color);
    return this.singleFlight('create', () => this.repository.createTag(context, {
      displayName: requiredName(input.displayName),
      color
    }));
  }

  async update(context: CustomizationContext, id: string, input: { readonly displayName: string; readonly color?: string }) {
    const color = validatedColor(input.color);
    return this.singleFlight('update', () => this.repository.updateTag(context, id, {
      displayName: requiredName(input.displayName),
      color: color ?? ''
    }));
  }

  async archive(context: CustomizationContext, id: string) {
    return this.singleFlight('archive', () => this.repository.archiveTag(context, id));
  }

  private async singleFlight<T>(action: 'create' | 'update' | 'archive', operation: () => Promise<T>): Promise<T> {
    if (this.saving) throw new CustomizationValidationError(t('mobile.ManageTags.thisTagChangeIsAlreadyBeingSaved'));
    this.saving = true;
    this.observability.record({ name: 'customization.mutation_requested', resource: 'tag', scope: 'inventory', action });
    try {
      const result = await operation();
      this.observability.record({ name: 'customization.mutation_succeeded', resource: 'tag', scope: 'inventory', action });
      return result;
    } catch (error) {
      this.observability.record({ name: 'customization.mutation_failed', resource: 'tag', scope: 'inventory', action });
      throw error;
    } finally { this.saving = false; }
  }
}

function requiredName(value: string): string {
  const trimmed = value.trim();
  if (!trimmed) throw new CustomizationValidationError(t('customization.tagRequired'));
  if (trimmed.length > 120) throw new CustomizationValidationError(t('customization.tagTooLong'));
  return trimmed;
}

function validatedColor(value: string | undefined): string | undefined {
  const normalized = normalizeTagColor(value);
  if (value?.trim() && !normalized) throw new CustomizationValidationError(t('mobile.ManageTags.enterASixDigitHexColorSuchAs2F80ED'));
  return normalized;
}
