import { t } from '$lib/presentation/localization';
import type { Asset, AssetViewModel } from '$lib/domain/inventory';
import { compareNaturalText } from './textCollation';
import { moveParentTargets, withTrail } from './workspace';
import { workspaceRouteHref } from './workspaceRoute';

export interface MoveHereCandidatePage {
  candidates: AssetViewModel[];
  totalCount: number;
  hasMore: boolean;
}

export interface ContainedWorkspaceAsset extends AssetViewModel {
  relativePath: string;
}

export interface ContainableWorkspaceSection {
  key: 'inside' | 'spaces' | 'items';
  heading: string;
  countNoun: 'asset' | 'space' | 'item';
  assets: ContainedWorkspaceAsset[];
  emptyTitle: string;
  emptyMessage: string;
}

export function containableWorkspaceSections(target: Asset, assets: Asset[]): ContainableWorkspaceSection[] {
  if (target.kind !== 'location') {
    const children = containedWorkspaceChildren(target, assets).map((candidate) => ({ ...candidate, relativePath: '' }));
    return [{
      key: 'inside',
      heading: t("contents.insideNamed", { title: target.title }),
      countNoun: 'asset',
      assets: children,
      emptyTitle: t('web.workspaceContainedAssets.nothingInsideYet'),
      emptyMessage: t('web.workspaceContainedAssets.addAnItemOrMoveSomethingIntoThisContainer')
    }];
  }

  const activeAssets = assets.filter(
    (candidate) => candidate.lifecycleState === 'active' && candidate.tenantId === target.tenantId && candidate.inventoryId === target.inventoryId
  );
  const childrenByParent = new Map<string, Asset[]>();
  for (const candidate of activeAssets) {
    if (!candidate.parentAssetId) continue;
    childrenByParent.set(candidate.parentAssetId, [...(childrenByParent.get(candidate.parentAssetId) ?? []), candidate]);
  }
  const directSpaces = (childrenByParent.get(target.id) ?? [])
    .filter((candidate) => candidate.kind !== 'item')
    .map((candidate) => ({ ...withTrail(candidate, assets), relativePath: '' }))
    .sort(compareContainedAsset);
  const items: ContainedWorkspaceAsset[] = [];
  const visited = new Set<string>([target.id]);
  const pending = (childrenByParent.get(target.id) ?? []).map((candidate) => ({ candidate, spacePath: [] as string[] }));
  while (pending.length > 0) {
    const next = pending.shift();
    if (!next || visited.has(next.candidate.id)) continue;
    visited.add(next.candidate.id);
    if (next.candidate.kind === 'item') {
      items.push({ ...withTrail(next.candidate, assets), relativePath: next.spacePath.join(' / ') });
      continue;
    }
    const childPath = [...next.spacePath, next.candidate.title];
    for (const child of childrenByParent.get(next.candidate.id) ?? []) {
      pending.push({ candidate: child, spacePath: childPath });
    }
  }
  items.sort(compareContainedAsset);
  return [
    {
      key: 'spaces',
      heading: t("contents.spacesIn", { title: target.title }),
      countNoun: 'space',
      assets: directSpaces,
      emptyTitle: t('web.workspaceContainedAssets.noNestedSpaces'),
      emptyMessage: t('web.workspaceContainedAssets.containersAndPlacesDirectlyInsideWillAppearHere')
    },
    {
      key: 'items',
      heading: t("contents.itemsIn", { title: target.title }),
      countNoun: 'item',
      assets: items,
      emptyTitle: t('web.workspaceContainedAssets.noItemsHereYet'),
      emptyMessage: t('web.workspaceContainedAssets.addAnItemOrMoveSomethingIntoThisPlace')
    }
  ];
}

export function containedWorkspaceChildren(target: Asset, assets: Asset[]): AssetViewModel[] {
  return assets
    .filter((candidate) =>
      candidate.parentAssetId === target.id &&
      candidate.lifecycleState === 'active' &&
      candidate.tenantId === target.tenantId &&
      candidate.inventoryId === target.inventoryId
    )
    .map((candidate) => withTrail(candidate, assets))
    .sort(compareContainedAsset);
}

export function moveHereCandidatePage(
  target: Asset,
  assets: Asset[],
  query: string,
  limit = 8
): MoveHereCandidatePage {
  const normalizedQuery = query.trim().toLocaleLowerCase();
  const eligible = assets
    .filter((candidate) => candidate.lifecycleState === 'active')
    .filter((candidate) => candidate.id !== target.id && candidate.parentAssetId !== target.id)
    .filter((candidate) => candidate.tenantId === target.tenantId && candidate.inventoryId === target.inventoryId)
    .filter((candidate) => moveParentTargets(assets, candidate.id).some((parent) => parent.id === target.id))
    .map((candidate) => withTrail(candidate, assets))
    .filter((candidate) => !normalizedQuery || candidateSearchScore(candidate, normalizedQuery) !== null)
    .sort((left, right) => {
      const scoreDifference = normalizedQuery
        ? candidateSearchScore(left, normalizedQuery)! - candidateSearchScore(right, normalizedQuery)!
        : 0;
      return scoreDifference || compareContainedAsset(left, right);
    });
  return {
    candidates: eligible.slice(0, Math.max(1, limit)),
    totalCount: eligible.length,
    hasMore: eligible.length > Math.max(1, limit)
  };
}

export function moveHereHref(target: Asset): string {
  return workspaceRouteHref(
    target.kind === 'location' ? {
      mode: 'location',
      tenantId: target.tenantId,
      inventoryId: target.inventoryId,
      locationId: target.id,
      assetAction: 'move-here'
    } : {
      mode: 'asset',
      tenantId: target.tenantId,
      inventoryId: target.inventoryId,
      assetId: target.id,
      assetAction: 'move-here'
    },
    target.tenantId,
    target.inventoryId
  );
}

export function addItemHereHref(target: Asset): string {
  return workspaceRouteHref(
    { action: 'add', addKind: 'item', addParentAssetId: target.id },
    target.tenantId,
    target.inventoryId
  );
}

function compareContainedAsset(left: Asset, right: Asset): number {
  const leftGroup = left.kind === 'item' ? 1 : 0;
  const rightGroup = right.kind === 'item' ? 1 : 0;
  return leftGroup - rightGroup || compareNaturalText(left.title, right.title) || left.id.localeCompare(right.id);
}

function candidateSearchScore(candidate: AssetViewModel, query: string): number | null {
  const title = candidate.title.toLocaleLowerCase();
  if (title === query) return 0;
  if (title.startsWith(query)) return 1;
  if (title.includes(query)) return 2;
  if (candidate.containmentTrail.toLocaleLowerCase().includes(query)) return 3;
  return null;
}
