<script lang="ts">
  import type { Inventory, Tenant } from '$lib/domain/inventory';
  import WorkspaceContextSwitcher from './WorkspaceContextSwitcher.svelte';

  let {
    tenants,
    inventories,
    initialTenantId,
    initialInventoryId,
    mobile = false,
    asyncTenantUpdate = false
  }: {
    tenants: Tenant[];
    inventories: Inventory[];
    initialTenantId: string;
    initialInventoryId: string;
    mobile?: boolean;
    asyncTenantUpdate?: boolean;
  } = $props();

  let selectedTenantId = $state(initialTenant());
  let selectedInventoryId = $state(initialInventory());

  function initialTenant(): string {
    return initialTenantId;
  }

  function initialInventory(): string {
    return initialInventoryId;
  }

  function selectTenant(tenantId: string): void {
    if (asyncTenantUpdate) {
      window.setTimeout(() => {
        applyTenant(tenantId);
      }, 10);
      return;
    }
    applyTenant(tenantId);
  }

  function applyTenant(tenantId: string): void {
    selectedTenantId = tenantId;
    selectedInventoryId = inventories.find((inventory) => inventory.tenantId === tenantId)?.id ?? selectedInventoryId;
  }
</script>

<WorkspaceContextSwitcher
  {mobile}
  {tenants}
  {inventories}
  {selectedTenantId}
  {selectedInventoryId}
  onSelectTenant={selectTenant}
  onSelectInventory={(tenantId, inventoryId) => {
    selectedTenantId = tenantId;
    selectedInventoryId = inventoryId;
  }}
/>
