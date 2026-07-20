import type { CustomizationKind, CustomizationScope } from '../../domain/customization/Customization';

export type CustomizationEvent = {
  readonly name: 'settings.opened' | 'settings.level_selected' | 'customization.collection_load_failed' | 'customization.mutation_requested' | 'customization.mutation_succeeded' | 'customization.mutation_failed' | 'customization.permission_denied';
  readonly resource?: CustomizationKind;
  readonly scope?: CustomizationScope;
  readonly action?: 'create' | 'update' | 'archive' | 'restore' | 'delete';
};

export interface CustomizationObservability {
  record(event: CustomizationEvent): void;
}

export const noCustomizationObservability: CustomizationObservability = { record: () => undefined };

export class BufferedCustomizationObservability implements CustomizationObservability {
  private readonly buffer: CustomizationEvent[] = [];
  constructor(private readonly capacity = 100, private readonly sink?: (event: CustomizationEvent) => void) {}

  record(event: CustomizationEvent): void {
    this.buffer.push(event);
    if (this.buffer.length > this.capacity) this.buffer.splice(0, this.buffer.length - this.capacity);
    this.sink?.(event);
  }

  events(): readonly CustomizationEvent[] { return [...this.buffer]; }
}
