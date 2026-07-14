import type { AssetTag } from '$lib/domain/inventory';

export function sortAssetTagsByDisplayName(tags: AssetTag[]): AssetTag[] {
  return [...tags].sort((left, right) =>
    left.displayName.localeCompare(right.displayName, undefined, { sensitivity: 'base' })
  );
}
