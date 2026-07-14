import type { AssetCardViewModel } from '../assets/AssetViewModels';
import { toAssetCardViewModel } from '../assets/AssetViewModels';
import { createInventoryOverview } from '../../domain/inventories/InventorySummary';
import type { AccessRole } from '../../domain/inventories/InventorySummary';
import type { LocationSummary } from '../../domain/locations/LocationSummary';
import { HomeDashboardSnapshotRepository } from './InventorySummaryRepository';

export type HomeDashboardInventoryViewModel = {
  readonly id: string;
  readonly tenantId: string;
  readonly tenantName: string;
  readonly name: string;
  readonly roleLabel: string;
  readonly updatedAtLabel: string;
};

export type HomeDashboardTenantViewModel = {
  readonly id: string;
  readonly name: string;
};

export type HomeDashboardLocationViewModel = {
  readonly id: string;
  readonly title: string;
  readonly description: string;
  readonly containedAssetCountLabel: string;
  readonly recentAssetLabel: string;
  readonly photoLabel: string;
  readonly photo?: {
    readonly uri: string;
    readonly headers?: Readonly<Record<string, string>>;
  };
};

export type HomeDashboardViewModel = {
  readonly tenantId: string;
  readonly tenantName: string;
  readonly inventoryId: string;
  readonly inventoryName: string;
  readonly tenants: readonly HomeDashboardTenantViewModel[];
  readonly inventories: readonly HomeDashboardInventoryViewModel[];
  readonly canAdd: boolean;
  readonly topLocations: readonly HomeDashboardLocationViewModel[];
  readonly locations: readonly HomeDashboardLocationViewModel[];
  readonly recentAssets: readonly AssetCardViewModel[];
  readonly checkedOutAssets: readonly AssetCardViewModel[];
  readonly assetTags: readonly {
    readonly id: string;
    readonly key: string;
    readonly displayName: string;
    readonly color?: string;
  }[];
};

export class HomeDashboardQuery {
  constructor(private readonly inventories: HomeDashboardSnapshotRepository) {}

  async execute(): Promise<HomeDashboardViewModel> {
    const { checkedOutAssets, workspace } = await this.inventories.getHomeDashboardSnapshot();
    const inventory =
      workspace.inventories.find((item) => item.id === workspace.defaultInventoryId) ??
      workspace.inventories[0];

    if (!inventory) {
      throw new Error('Inventory workspace must include at least one inventory.');
    }

    const tenant = workspace.tenants.find((item) => item.id === inventory.tenantId);

    if (!tenant) {
      throw new Error('Selected inventory must belong to a tenant.');
    }

    const overview = createInventoryOverview(tenant, inventory, workspace.inventories);

    return {
      tenantId: tenant.id,
      tenantName: overview.tenantName,
      inventoryId: inventory.id,
      inventoryName: overview.inventoryName,
      tenants: workspace.tenants.map((item) => ({
        id: item.id,
        name: item.name
      })),
      inventories: overview.inventories.map((item) => ({
        id: item.id,
        tenantId: item.tenantId,
        tenantName:
          workspace.tenants.find((tenantOption) => tenantOption.id === item.tenantId)?.name ??
          'Unknown tenant',
        name: item.name,
        roleLabel: labelAccessRole(item.role),
        updatedAtLabel: item.updatedAtLabel
      })),
      canAdd: inventory.permissions.includes('create_asset'),
      topLocations: overview.locations.slice(0, 3).map(toLocationViewModel),
      locations: overview.locations.map(toLocationViewModel),
      recentAssets: inventory.assets.slice(0, 10).map(toAssetCardViewModel),
      checkedOutAssets: checkedOutAssets.slice(0, 10).map(toAssetCardViewModel),
      assetTags: [...(inventory.assetTags ?? [])]
    };
  }
}

function toLocationViewModel(location: LocationSummary): HomeDashboardLocationViewModel {
  return {
    id: location.id,
    title: location.title,
    description: location.description,
    containedAssetCountLabel: location.containedAssetCount.toString(),
    recentAssetLabel: location.recentAssetTitles.join(', '),
    photoLabel: location.hasPhoto ? 'Photo ready' : 'Needs photo',
    photo: location.photo
  };
}

function labelAccessRole(role: AccessRole): string {
  switch (role) {
    case 'owner':
      return 'Owner';
    case 'editor':
      return 'Editor';
    case 'viewer':
      return 'Viewer';
  }
}
