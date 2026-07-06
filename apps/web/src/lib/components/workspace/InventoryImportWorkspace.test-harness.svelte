<script lang="ts">
  import type { ImportSourceRoute } from '$lib/application/workspaceRoute';
  import type { Inventory } from '$lib/domain/inventory';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import InventoryImportWorkspace from './InventoryImportWorkspace.svelte';

  type ImportJobInventoryRefreshScope = {
    tenantId: string;
    inventoryId: string;
  };

  let {
    tenantId,
    inventory,
    repository,
    initialImportSource = null,
    onImportJobInventoryChanged = async () => {}
  }: {
    tenantId: string;
    inventory: Inventory | null;
    repository: InventoryRepository;
    initialImportSource?: ImportSourceRoute;
    onImportJobInventoryChanged?: (scope: ImportJobInventoryRefreshScope) => Promise<void>;
  } = $props();

  // svelte-ignore state_referenced_locally -- test harness seeds route state once, then local callbacks mutate it.
  let importSource = $state<ImportSourceRoute>(initialImportSource);
</script>

<InventoryImportWorkspace
  {tenantId}
  {inventory}
  {repository}
  {importSource}
  onImportSourceChange={(next) => (importSource = next)}
  {onImportJobInventoryChanged}
/>
