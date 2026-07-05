<script lang="ts">
  import * as Button from '$lib/components/ui/button/index.js';
  import type { ImportSourceType, Inventory } from '$lib/domain/inventory';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import HomeboxImportPanel from './HomeboxImportPanel.svelte';

  let {
    tenantId,
    inventory,
    repository,
    initialSourceType = 'legacy_homebox',
    onImported
  }: {
    tenantId: string;
    inventory: Inventory;
    repository: InventoryRepository;
    initialSourceType?: ImportSourceType;
    onImported: () => Promise<void>;
  } = $props();

  let sourceType = $state(initialSource());
  let currentInventory = $state(initialInventory());

  function initialSource(): ImportSourceType {
    return initialSourceType;
  }

  function initialInventory(): Inventory {
    return inventory;
  }

  function alternateInventory(): Inventory {
    return {
      ...inventory,
      id: inventory.id === 'inventory-one' ? 'inventory-two' : 'inventory-one',
      name: inventory.id === 'inventory-one' ? 'Garage' : inventory.name
    };
  }
</script>

<Button.Root id="switch-import-source" onclick={() => { sourceType = sourceType === 'legacy_homebox' ? 'legacy_homebox_csv' : 'legacy_homebox'; }}>
  Switch source externally
</Button.Root>
<Button.Root id="switch-import-inventory" onclick={() => { currentInventory = alternateInventory(); }}>
  Switch inventory externally
</Button.Root>

<HomeboxImportPanel
  {tenantId}
  inventory={currentInventory}
  {repository}
  {sourceType}
  onSourceChange={(nextSourceType) => { sourceType = nextSourceType; }}
  {onImported}
/>
