import type { AssetTag } from '$lib/domain/inventory';
import { compareNaturalText } from './textCollation';

export function sortAssetTagsByDisplayName(tags: AssetTag[]): AssetTag[] {
  return [...tags].sort((left, right) => compareNaturalText(left.displayName, right.displayName));
}

export function visibleAssetTagOptions(
  tags: AssetTag[],
  expanded: boolean,
  selectedIds: string[] = [],
  limit = 12
): AssetTag[] {
  const sorted = sortAssetTagsByDisplayName(tags);
  if (expanded) {
    return sorted;
  }
  const selected = new Set(selectedIds);
  return sorted.filter((tag, index) => index < limit || selected.has(tag.id));
}
