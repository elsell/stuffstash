import type { AssetId, AssetLocationTrailSegment } from '../assets/AssetSummary';
import type { AssetPhoto } from '../assets/AssetSummary';
import type { InventoryId } from '../inventories/InventorySummary';

export type LocationSummary = {
  readonly id: AssetId;
  readonly inventoryId: InventoryId;
  readonly title: string;
  readonly description: string;
  readonly parentLocationTrail?: readonly AssetLocationTrailSegment[];
  readonly parentLocationTrailIncomplete?: boolean;
  readonly containedAssetCount: number;
  readonly recentAssetTitles: readonly string[];
  readonly hasPhoto: boolean;
  readonly photo?: AssetPhoto;
};
