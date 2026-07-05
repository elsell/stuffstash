<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import { parentTargetMetadataLabel } from '$lib/application/workspaceParentTargets';
  import type { ParentTargetViewModel } from '$lib/domain/inventory';
  import AssetThumb from './AssetThumb.svelte';

  let {
    target,
    selected,
    onSelect
  }: {
    target: ParentTargetViewModel;
    selected: boolean;
    onSelect: (id: string) => void;
  } = $props();

  let metadataLabel = $derived(parentTargetMetadataLabel(target));
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
    <small>{metadataLabel}</small>
  </span>
</Button.Root>
