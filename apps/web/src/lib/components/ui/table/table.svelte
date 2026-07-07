<script lang="ts">
  import type { HTMLAttributes } from 'svelte/elements';
  import { cn, type WithElementRef } from '$lib/utils.js';

  let {
    ref = $bindable(null),
    class: className,
    'aria-label': ariaLabel,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLDivElement>> = $props();
</script>

<div
  bind:this={ref}
  data-slot="table-wrapper"
  class={cn('border-border min-w-0 overflow-x-auto rounded-lg border', className)}
  {...restProps}
>
  <table data-slot="table" class="w-full border-collapse text-sm" aria-label={ariaLabel}>
    {@render children?.()}
  </table>
</div>
