<script lang="ts">
  import { untrack } from 'svelte';
  import type { Inventory, Tenant } from '$lib/domain/inventory';
  import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
  import InventoryAccessManager from './InventoryAccessManager.svelte';

  let {
    initialTenant,
    initialInventory,
    repository
  }: {
    initialTenant: Tenant | null;
    initialInventory: Inventory | null;
    repository: InventoryAccessRepository;
  } = $props();

  let tenant = $state(untrack(() => initialTenant));
  let inventory = $state(untrack(() => initialInventory));

  export function setContext(nextTenant: Tenant | null, nextInventory: Inventory | null): void {
    tenant = nextTenant;
    inventory = nextInventory;
  }
</script>

<InventoryAccessManager {tenant} {inventory} {repository} />
