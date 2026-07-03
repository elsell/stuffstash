<script lang="ts">
	import Button from './button.svelte';
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
