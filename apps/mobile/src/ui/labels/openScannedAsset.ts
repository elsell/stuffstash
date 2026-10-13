import type { Router } from 'expo-router';

/** Reuse the tab shell and retain Home beneath externally opened details. */
export function openScannedAsset(router: Pick<Router, 'dismissTo'>, assetId: string): void {
  router.dismissTo({ pathname: '/(tabs)/(home)/assets/[assetId]', params: { assetId } }, { withAnchor: true });
}
