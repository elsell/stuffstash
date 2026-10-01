import type { AssetKind } from '$lib/domain/inventory';
import { t } from './localization';

export function assetKindLabel(kind: AssetKind): string {
  return t(`asset.kind.${kind}`);
}
