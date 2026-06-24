<script lang="ts">
  import { containedAssets, parentTargets, recentlyAddedAssets, topLevelLocations, withTrail } from '$lib/application/workspace';
  import type { AddAssetDraft, Asset, SearchResult, WorkspaceData, WorkspaceMode } from '$lib/domain/inventory';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import AddAssetTray from './AddAssetTray.svelte';
  import AssetDetail from './AssetDetail.svelte';
  import HomeWorkspace from './HomeWorkspace.svelte';
  import LocationView from './LocationView.svelte';
  import MobileNav from './MobileNav.svelte';
  import SearchPanel from './SearchPanel.svelte';
  import SideNav from './SideNav.svelte';
  import TopHeader from './TopHeader.svelte';
  import * as Alert from '$lib/components/ui/alert/index.js';
  import * as Button from '$lib/components/ui/button/index.js';

  let {
    repository,
    initialData,
    onSignOut
  }: {
    repository: InventoryRepository;
    initialData: WorkspaceData;
    onSignOut: () => void;
  } = $props();

  // svelte-ignore state_referenced_locally -- initial route data seeds local workspace state.
  const startingData = initialData;
  let data = $state(startingData);
  let mode = $state<WorkspaceMode>(startingData.context.inventories.length > 0 ? 'home' : 'settings');
  let selectedLocationId = $state<string | null>(null);
  let selectedAssetId = $state<string | null>(null);
  let addOpen = $state(false);
  let busy = $state(false);
  let message = $state('');
  let error = $state('');
  let searchQuery = $state('');
  let searchResults = $state<SearchResult[]>([]);

  let selectedInventory = $derived(data.context.inventories.find((inventory) => inventory.id === data.context.selectedInventoryId) ?? null);
  let selectedLocation = $derived(data.assets.find((asset) => asset.id === selectedLocationId) ?? null);
  let selectedAsset = $derived(selectedAssetId ? data.assets.find((asset) => asset.id === selectedAssetId) ?? null : null);
  let editable = $derived(data.context.capability === 'editor');
  let userLabel = $derived(data.context.principal.email ?? data.context.principal.id);

  async function selectInventory(tenantId: string, inventoryId: string): Promise<void> {
    await run(async () => {
      data = await repository.selectInventory(tenantId, inventoryId);
      mode = 'home';
      selectedLocationId = null;
      selectedAssetId = null;
    });
  }

  async function createStarterInventory(): Promise<void> {
    await run(async () => {
      data = await repository.createTenantWithInventory({ tenantName: 'Home', inventoryName: 'Household' });
      mode = 'home';
      message = 'Created Household.';
    });
  }

  async function createAsset(draft: AddAssetDraft): Promise<void> {
    if (!selectedInventory) {
      error = 'Create an inventory before adding assets.';
      return;
    }
    await run(async () => {
      const asset = await repository.createAsset(data.context.selectedTenantId, selectedInventory.id, draft);
      data = { ...data, assets: [asset, ...data.assets] };
      addOpen = false;
      message = `Saved ${asset.title}.`;
    });
  }

  async function search(): Promise<void> {
    const query = searchQuery.trim();
    if (!query || !data.context.selectedTenantId) {
      searchResults = [];
      return;
    }
    await run(async () => {
      searchResults = await repository.searchAssets(data.context.selectedTenantId, query);
      mode = 'search';
    });
  }

  async function run(task: () => Promise<void>): Promise<void> {
    busy = true;
    error = '';
    message = '';
    try {
      await task();
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Action failed.';
    } finally {
      busy = false;
    }
  }

  function openLocation(asset: Asset): void {
    selectedLocationId = asset.id;
    selectedAssetId = null;
    mode = 'location';
  }

  function openAsset(asset: Asset): void {
    selectedAssetId = asset.id;
    mode = 'asset';
  }

  function openAssetById(assetId: string): void {
    const asset = data.assets.find((candidate) => candidate.id === assetId);
    if (asset) {
      openAsset(asset);
    }
  }
</script>

<div class="product-shell">
  <SideNav
    tenants={data.context.tenants}
    inventories={data.context.inventories}
    selectedTenantId={data.context.selectedTenantId}
    selectedInventoryId={data.context.selectedInventoryId}
    {mode}
    {userLabel}
    onSelectInventory={(tenantId, inventoryId) => { void selectInventory(tenantId, inventoryId); }}
    onModeChange={(nextMode) => { mode = nextMode; }}
    {onSignOut}
  />

  <div class="workspace-column">
    <TopHeader
      inventory={selectedInventory}
      bind:query={searchQuery}
      canEdit={editable}
      onSearch={() => { void search(); }}
      onOpenAdd={() => { addOpen = true; }}
    />

    {#if data.context.inventories.length === 0}
      <section class="workspace-main">
        <div class="empty-state spacious">
          <h1>No inventory yet</h1>
          <p>Create the first inventory for this tenant.</p>
          <Button.Root onclick={() => { void createStarterInventory(); }}>Create Household</Button.Root>
        </div>
      </section>
    {:else if mode === 'location' && selectedLocation}
      <LocationView
        location={selectedLocation}
        assets={containedAssets(data.assets, selectedLocation.id)}
        onBack={() => { mode = 'home'; selectedLocationId = null; }}
        onOpenAsset={openAsset}
      />
    {:else if mode === 'asset' && selectedAsset}
      <AssetDetail
        asset={withTrail(selectedAsset, data.assets)}
        capability={data.context.capability}
        onBack={() => { mode = selectedLocationId ? 'location' : 'home'; selectedAssetId = null; }}
      />
    {:else if mode === 'search'}
      <SearchPanel
        bind:query={searchQuery}
        results={searchResults}
        {busy}
        onSearch={() => { void search(); }}
        onOpenAsset={openAssetById}
      />
    {:else if mode === 'settings'}
      <section class="workspace-main">
        <div class="section-heading">
          <div>
            <h1>Inventory settings</h1>
            <p>Sharing, activity, and administrative details belong here as those APIs are exposed.</p>
          </div>
        </div>
        <div class="empty-state spacious">
          <h2>Settings unavailable</h2>
          <p>This slice keeps settings visible without inventing unsupported controls.</p>
        </div>
      </section>
    {:else}
      <HomeWorkspace
        locations={topLevelLocations(data.assets)}
        recentAssets={recentlyAddedAssets(data.assets)}
        onOpenLocation={openLocation}
        onOpenAsset={openAsset}
        onOpenAdd={() => { addOpen = true; }}
      />
    {/if}
  </div>

  <MobileNav
    {mode}
    canEdit={editable}
    onModeChange={(nextMode) => { mode = nextMode; }}
    onOpenAdd={() => { addOpen = true; }}
  />

  <AddAssetTray
    open={addOpen}
    parentTargets={parentTargets(data.assets)}
    saving={busy}
    onClose={() => { addOpen = false; }}
    onSave={(draft) => { void createAsset(draft); }}
  />

  {#if message}
    <Alert.Root class="toast" variant="default">
      <Alert.Description>{message}</Alert.Description>
    </Alert.Root>
  {/if}
  {#if error}
    <Alert.Root class="toast" variant="destructive">
      <Alert.Description>{error}</Alert.Description>
    </Alert.Root>
  {/if}
</div>
