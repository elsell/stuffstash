import { customizationNameOrder, type CustomDefinition, type CustomizationLifecycle, type CustomizationScope } from '../../domain/customization/Customization';
import type { CustomizationContext, CustomizationRepository } from './CustomizationRepository';
import { noCustomizationObservability, type CustomizationObservability } from './CustomizationObservability';

const maximumPages = 50;

export type CustomizationCollection<T> = {
  readonly items: readonly T[];
  readonly complete: boolean;
};

export class CustomizationCollectionQuery {
  constructor(private readonly repository: CustomizationRepository, private readonly observability: CustomizationObservability = noCustomizationObservability) {}

  async tags(context: CustomizationContext): Promise<CustomizationCollection<Awaited<ReturnType<CustomizationRepository['listTags']>>['items'][number]>> {
    try { return await this.loadAll((cursor) => this.repository.listTags(context, cursor)); }
    catch (error) { this.observability.record({ name: 'customization.collection_load_failed', resource: 'tag', scope: 'inventory' }); throw error; }
  }

  async fields(context: CustomizationContext, scope: CustomizationScope, lifecycle: CustomizationLifecycle): Promise<CustomizationCollection<CustomDefinition & { kind: 'field' }>> {
    try { return await this.loadAll((cursor) => this.repository.listFields(context, scope, lifecycle, cursor)); }
    catch (error) { this.observability.record({ name: 'customization.collection_load_failed', resource: 'field', scope }); throw error; }
  }

  async assetTypes(context: CustomizationContext, scope: CustomizationScope, lifecycle: CustomizationLifecycle): Promise<CustomizationCollection<CustomDefinition & { kind: 'asset-type' }>> {
    try { return await this.loadAll((cursor) => this.repository.listAssetTypes(context, scope, lifecycle, cursor)); }
    catch (error) { this.observability.record({ name: 'customization.collection_load_failed', resource: 'asset-type', scope }); throw error; }
  }

  private async loadAll<T extends { readonly id: string; readonly displayName: string }>(
    load: (cursor?: string) => Promise<{ readonly items: readonly T[]; readonly nextCursor?: string }>
  ): Promise<CustomizationCollection<T>> {
    const items: T[] = [];
    const seen = new Set<string>();
    let cursor: string | undefined;
    for (let pageIndex = 0; pageIndex < maximumPages; pageIndex += 1) {
      const page = await load(cursor);
      for (const item of page.items) if (!seen.has(item.id)) { seen.add(item.id); items.push(item); }
      if (!page.nextCursor) return { items: customizationNameOrder(items), complete: true };
      if (page.nextCursor === cursor) return { items: customizationNameOrder(items), complete: false };
      cursor = page.nextCursor;
    }
    return { items: customizationNameOrder(items), complete: false };
  }
}
