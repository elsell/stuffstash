import type { AssetKind, AssetLifecycleState } from '$lib/domain/inventory';
import { t } from './localization';

export function assetKindLabel(kind: AssetKind): string {
  return t(`asset.kind.${kind}`);
}

export function assetLifecycleLabel(state: AssetLifecycleState): string {
  return t(`asset.lifecycle.${state}`);
}
