<script lang="ts">
  import type { HTMLAttributes, HTMLTableAttributes } from 'svelte/elements';
  import { cn, type WithElementRef } from '$lib/utils.js';

  type Props = WithElementRef<HTMLTableAttributes> & {
    wrapperClass?: HTMLAttributes<HTMLDivElement>['class'];
    wrapperRef?: HTMLDivElement | null;
  };

  let {
    ref = $bindable(null),
    wrapperRef = $bindable(null),
    class: className,
    wrapperClass,
    children,
    ...restProps
  }: Props = $props();
</script>

<div
  bind:this={wrapperRef}
  data-slot="table-wrapper"
  class={cn('border-border min-w-0 overflow-x-auto rounded-lg border', wrapperClass)}
>
  <table bind:this={ref} data-slot="table" class={cn('w-full border-collapse text-sm', className)} {...restProps}>
    {@render children?.()}
  </table>
</div>
