<script lang="ts">
  import { setContext } from 'svelte';
  import type { Asset } from '$lib/domain/inventory';
  import { assetThumbnailLoaderContext, type AssetThumbnailLoader } from '$lib/ports/assetThumbnailLoader';
  import {performanceObserverContext, noopPerformanceObserver, type PerformanceObserver} from '$lib/ports/performanceObserver';
  import AssetThumb from './AssetThumb.svelte';

  let { asset: initialAsset, loader, observer = noopPerformanceObserver }: { asset: Asset; loader: AssetThumbnailLoader; observer?: PerformanceObserver } = $props();
  // svelte-ignore state_referenced_locally -- the initial prop intentionally seeds mutable harness state.
  let asset = $state(initialAsset);
  // svelte-ignore state_referenced_locally -- the test loader is immutable for the harness lifetime.
  setContext(assetThumbnailLoaderContext, loader);
  // svelte-ignore state_referenced_locally -- immutable session observer for this test.
  setContext(performanceObserverContext, observer);

  export function replaceAsset(next: Asset): void {
    asset = next;
  }
</script>

<AssetThumb {asset} />
