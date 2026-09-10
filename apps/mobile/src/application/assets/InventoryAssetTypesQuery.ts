import { customizationNameOrder } from '../../domain/customization/Customization';
import type { CustomizationContextQuery } from '../customization/CustomizationContextQuery';
import type { CustomizationCollectionQuery } from '../customization/CustomizationQueries';
import { assertReadActive, type ReadRequest } from '../shared/ReadRequest';

export class InventoryAssetTypesQuery {
  constructor(
    private readonly context: Pick<CustomizationContextQuery, 'execute'>,
    private readonly collections: Pick<CustomizationCollectionQuery, 'assetTypes'>
  ) {}

  async execute(tenantId: string, inventoryId: string, request: ReadRequest = {}) {
    assertReadActive(request.signal);
    const context = await this.context.execute(request);
    assertReadActive(request.signal);
    if (context.tenantId !== tenantId || context.inventoryId !== inventoryId) {
      throw new Error('Inventory changed. Reopen the item to continue.');
    }
    const collection = await this.collections.assetTypes(context, 'inventory', 'active', request);
    assertReadActive(request.signal);
    if (!collection.complete) throw new Error('Could not load the complete asset type list. Try again.');
    return customizationNameOrder(collection.items.filter((type) =>
      type.lifecycle === 'active' && type.tenantId === tenantId
      && (type.scope === 'tenant' ? !type.inventoryId : type.inventoryId === inventoryId)
    ));
  }
}
