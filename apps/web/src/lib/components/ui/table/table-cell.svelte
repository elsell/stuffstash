<script lang="ts">
  import type { HTMLAttributes } from 'svelte/elements';
  import { cn, type WithElementRef } from '$lib/utils.js';

  let {
    ref = $bindable(null),
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLTableCellElement>> = $props();
</script>

<td
  bind:this={ref}
  data-slot="table-cell"
  class={cn('border-border text-muted-foreground min-w-32 border-t px-3 py-2 align-top break-words', className)}
  {...restProps}
>
  {@render children?.()}
</td>
