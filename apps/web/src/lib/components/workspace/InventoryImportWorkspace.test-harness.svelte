<script lang="ts">
  import type { ImportDetailTabRoute, ImportSourceRoute } from '$lib/application/workspaceRoute';
  import type { Inventory, Principal } from '$lib/domain/inventory';
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
    initialImportJobId = null,
    initialImportTab = null,
    currentPrincipal = undefined,
    onImportJobSelectionChange = () => {},
    onImportJobTabChange = () => {},
    onImportJobInventoryChanged = async () => {},
    onOpenImportedAssetId = async () => {},
    onOpenInventoryAuditHistory = () => {}
  }: {
    tenantId: string;
    inventory: Inventory | null;
    repository: InventoryRepository;
    initialImportSource?: ImportSourceRoute;
    initialImportJobId?: string | null;
    initialImportTab?: ImportDetailTabRoute | null;
    currentPrincipal?: Principal;
    onImportJobSelectionChange?: (jobId: string | null, tab?: ImportDetailTabRoute | null) => void;
    onImportJobTabChange?: (tab: ImportDetailTabRoute | null) => void;
    onImportJobInventoryChanged?: (scope: ImportJobInventoryRefreshScope) => Promise<void>;
    onOpenImportedAssetId?: (assetId: string) => Promise<void>;
    onOpenInventoryAuditHistory?: () => void;
  } = $props();

  // svelte-ignore state_referenced_locally -- test harness seeds route state once, then local callbacks mutate it.
  let importSource = $state<ImportSourceRoute>(initialImportSource);
  // svelte-ignore state_referenced_locally -- test harness seeds route state once, then local callbacks mutate it.
  let importJobId = $state<string | null>(initialImportJobId);
  // svelte-ignore state_referenced_locally -- test harness seeds route state once, then local callbacks mutate it.
  let importTab = $state<ImportDetailTabRoute | null>(initialImportTab);
</script>

<InventoryImportWorkspace
  {tenantId}
  {inventory}
  {currentPrincipal}
  {repository}
  {importSource}
  {importJobId}
  {importTab}
  onImportSourceChange={(next) => (importSource = next)}
  onImportJobSelectionChange={(jobId, tab) => {
    importJobId = jobId;
    importTab = tab ?? null;
    onImportJobSelectionChange(jobId, tab);
  }}
  onImportJobTabChange={(tab) => {
    importTab = tab;
    onImportJobTabChange(tab);
  }}
  {onImportJobInventoryChanged}
  {onOpenImportedAssetId}
  {onOpenInventoryAuditHistory}
/>
