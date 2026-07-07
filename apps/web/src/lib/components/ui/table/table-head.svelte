<script lang="ts">
  import type { HTMLThAttributes } from 'svelte/elements';
  import { cn, type WithElementRef } from '$lib/utils.js';

  let {
    ref = $bindable(null),
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLThAttributes> = $props();
</script>

<th
  bind:this={ref}
  data-slot="table-head"
  class={cn(
    'bg-muted/30 text-muted-foreground min-w-32 px-3 py-2 text-left align-top text-xs font-bold uppercase',
    className
  )}
  {...restProps}
>
  {@render children?.()}
</th>
