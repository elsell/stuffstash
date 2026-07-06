<script lang="ts">
	import type { HTMLInputAttributes } from "svelte/elements";
	import { cn, type WithElementRef } from "$lib/utils.js";

	type Props = WithElementRef<Omit<HTMLInputAttributes, "type">> & {
		checked?: boolean;
	};

	let {
		ref = $bindable(null),
		checked = $bindable(false),
		class: className,
		"data-slot": dataSlot = "checkbox",
		...restProps
	}: Props = $props();
</script>

<input
	bind:this={ref}
	bind:checked
	data-slot={dataSlot}
	type="checkbox"
	class={cn(
		"border-input text-primary focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive size-4 shrink-0 rounded border bg-transparent shadow-xs transition-[color,box-shadow] outline-none focus-visible:ring-3 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50",
		className
	)}
	{...restProps}
/>
