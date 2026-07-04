<script lang="ts">
	import Button from './button.svelte';

	let {
		onDisabledActivate = () => {}
	}: {
		onDisabledActivate?: () => void;
	} = $props();
</script>

<Button href="/tenants/tenant-one/inventories/inventory-one/add/item" disabled>Add asset</Button>
<Button
	href="/tenants/tenant-one/inventories/inventory-one/assets/asset-one/edit"
	disabled
	aria-disabled={false}
	tabindex={0}
	data-testid="conflicting-disabled-link"
>
	Edit asset
</Button>
<Button
	href="/tenants/tenant-one/inventories/inventory-one/assets/asset-one/archive"
	disabled
	data-testid="disabled-action-link"
	onclick={onDisabledActivate}
	onkeydown={(event) => {
		if (event.key === 'Enter') {
			onDisabledActivate();
		}
	}}
>
	Archive asset
</Button>
