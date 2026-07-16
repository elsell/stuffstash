<script lang="ts">
	import Check from '@lucide/svelte/icons/check';
	import BusyButtonContent from './busy-button-content.svelte';
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
<Button data-testid="ready-button">
	<BusyButtonContent busy={false} icon={Check} label="Confirm connection" busyLabel="Confirming connection" />
</Button>
<Button data-testid="route-close" href="/assets/example" role="button" tabindex={-1}>Close route</Button>
<Button data-testid="outline-button" variant="outline">Cancel</Button>
<Button data-testid="busy-button" disabled>
	<BusyButtonContent busy={true} icon={Check} label="Confirm connection" busyLabel="Confirming connection" />
</Button>
