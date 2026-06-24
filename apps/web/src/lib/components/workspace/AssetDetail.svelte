<script lang="ts">
  import ArrowLeft from '@lucide/svelte/icons/arrow-left';
  import MoreHorizontal from '@lucide/svelte/icons/more-horizontal';
  import MoveRight from '@lucide/svelte/icons/move-right';
  import Pencil from '@lucide/svelte/icons/pencil';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import type { AssetViewModel, Capability } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

  let {
    asset,
    capability,
    onBack
  }: {
    asset: AssetViewModel;
    capability: Capability;
    onBack: () => void;
  } = $props();
</script>

<section class="workspace-main detail-view" aria-labelledby="asset-title">
  <Button.Root variant="ghost" class="back-button" onclick={onBack}><ArrowLeft /> Back</Button.Root>
  <div class="asset-detail-grid">
    <AssetThumb asset={asset} size="lg" />
    <div class="asset-detail-copy">
      <div class="detail-title-row">
        <div>
          <h1 id="asset-title">{asset.title}</h1>
          <p>{asset.containmentTrail}</p>
        </div>
        <Badge variant={asset.lifecycleState === 'active' ? 'secondary' : 'outline'}>{asset.lifecycleState}</Badge>
      </div>
      <p>{asset.description || 'No description.'}</p>
      <dl class="detail-list">
        <div><dt>Kind</dt><dd>{assetKindLabel(asset.kind)}</dd></div>
        <div><dt>Type</dt><dd>{asset.customAssetTypeLabel ?? 'Base asset'}</dd></div>
        <div><dt>Updated</dt><dd>{asset.updatedAt ? new Date(asset.updatedAt).toLocaleString() : 'Not available'}</dd></div>
      </dl>
      <div class="detail-actions">
        <Button.Root disabled={capability !== 'editor'}><Pencil /> Edit</Button.Root>
        <Button.Root variant="outline" disabled={capability !== 'editor'}><MoveRight /> Move</Button.Root>
        <Button.Root variant="ghost" aria-label="More asset actions"><MoreHorizontal /></Button.Root>
      </div>
      {#if capability !== 'editor'}
        <p class="denied-note">Edit actions require inventory editor access.</p>
      {/if}
    </div>
  </div>
</section>
