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

  function initialSource(): ImportSourceType {
    return initialSourceType;
  }
</script>

<Button.Root id="switch-import-source" onclick={() => { sourceType = sourceType === 'legacy_homebox' ? 'legacy_homebox_csv' : 'legacy_homebox'; }}>
  Switch source externally
</Button.Root>

<HomeboxImportPanel
  {tenantId}
  {inventory}
  {repository}
  {sourceType}
  onSourceChange={(nextSourceType) => { sourceType = nextSourceType; }}
  {onImported}
/>
