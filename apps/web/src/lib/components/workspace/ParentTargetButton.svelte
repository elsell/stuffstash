<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import type { AssetViewModel } from '$lib/domain/inventory';
  import { assetKindLabel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

  let {
    target,
    selected,
    onSelect
  }: {
    target: AssetViewModel;
    selected: boolean;
    onSelect: (id: string) => void;
  } = $props();
</script>

<Button.Root
  type="button"
  variant={selected ? 'secondary' : 'outline'}
  class="parent-target-button"
  aria-pressed={selected}
  onclick={() => onSelect(target.id)}
>
  <span class="parent-target-thumb" aria-hidden="true">
    <AssetThumb asset={target} size="sm" />
  </span>
  <span>
    <strong>{target.title}</strong>
    <small>{assetKindLabel(target.kind)} / {target.containmentTrail}</small>
  </span>
</Button.Root>
