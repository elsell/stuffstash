import type { AssetTagSummary } from '../../domain/assets/AssetSummary';
import type { ReadRequest } from '../shared/ReadRequest';

export type AddAssetContext = {
  readonly tenantId: string;
  readonly tenantName: string;
  readonly inventoryId: string;
  readonly inventoryName: string;
  readonly canAdd: boolean;
  readonly assetTags: readonly AssetTagSummary[];
};

export type AddAssetContextRepository = {
  getAddAssetContext(request?: ReadRequest): Promise<AddAssetContext>;
};

export class AddAssetContextQuery {
  constructor(private readonly context: AddAssetContextRepository) {}

  execute(request: ReadRequest = {}): Promise<AddAssetContext> {
    return this.context.getAddAssetContext(request);
  }
}
