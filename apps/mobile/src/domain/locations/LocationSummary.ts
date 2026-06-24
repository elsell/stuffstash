import type { AssetId } from '../assets/AssetSummary';
import type { AssetPhoto } from '../assets/AssetSummary';
import type { InventoryId } from '../inventories/InventorySummary';

export type LocationSummary = {
  readonly id: AssetId;
  readonly inventoryId: InventoryId;
  readonly title: string;
  readonly description: string;
  readonly containedAssetCount: number;
  readonly recentAssetTitles: readonly string[];
  readonly hasPhoto: boolean;
  readonly photo?: AssetPhoto;
};
