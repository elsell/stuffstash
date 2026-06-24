<script lang="ts">
  import { containedAssets, moveParentTargets, parentTargets, recentlyAddedAssets, topLevelLocations, withTrail } from '$lib/application/workspace';
  import {
    canCreateAsset,
    canEditAsset,
    canCreateInventory,
    type AddAssetDraft,
    type Asset,
    type AssetLifecycleFilter,
    type SearchResult,
    type UpdateAssetDraft,
    type WorkspaceData,
    type WorkspaceMode
  } from '$lib/domain/inventory';
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
  let loadedAssetDetail = $state<Asset | null>(null);

  let selectedInventory = $derived(data.context.inventories.find((inventory) => inventory.id === data.context.selectedInventoryId) ?? null);
  let selectedTenant = $derived(data.context.tenants.find((tenant) => tenant.id === data.context.selectedTenantId) ?? null);
  let selectedLocation = $derived(data.assets.find((asset) => asset.id === selectedLocationId) ?? null);
  let detailAssets = $derived(
    loadedAssetDetail && !data.assets.some((asset) => asset.id === loadedAssetDetail?.id)
      ? [loadedAssetDetail, ...data.assets]
      : data.assets
  );
  let selectedAsset = $derived(
    selectedAssetId
      ? loadedAssetDetail?.id === selectedAssetId
        ? loadedAssetDetail
        : data.assets.find((asset) => asset.id === selectedAssetId) ?? null
      : null
  );
  let createAssetAllowed = $derived(canCreateAsset(selectedInventory));
  let editAssetAllowed = $derived(canEditAsset(selectedInventory));
  let canCreateStarter = $derived(!data.context.selectedTenantId || canCreateInventory(selectedTenant));
  let userLabel = $derived(data.context.principal.email ?? data.context.principal.id);

  async function selectInventory(tenantId: string, inventoryId: string): Promise<void> {
    await run(async () => {
      data = await repository.selectInventory(tenantId, inventoryId);
      mode = 'home';
      selectedLocationId = null;
      selectedAssetId = null;
      loadedAssetDetail = null;
    });
  }

  async function selectTenant(tenantId: string): Promise<void> {
    await run(async () => {
      data = await repository.selectTenant(tenantId);
      mode = data.context.inventories.length > 0 ? 'home' : 'settings';
      selectedLocationId = null;
      selectedAssetId = null;
      loadedAssetDetail = null;
    });
  }

  async function createStarterInventory(): Promise<void> {
    await run(async () => {
      data = data.context.selectedTenantId
        ? await repository.createInventory(data.context.selectedTenantId, 'Household')
        : await repository.createTenantWithInventory({ tenantName: 'Home', inventoryName: 'Household' });
      mode = 'home';
      message = 'Created Household.';
    });
  }

  async function createAsset(draft: AddAssetDraft): Promise<void> {
    if (!selectedInventory) {
      error = 'Create an inventory before adding assets.';
      return;
    }
    if (!createAssetAllowed) {
      error = 'You do not have permission to add assets in this inventory.';
      return;
    }
    await run(async () => {
      const asset = await repository.createAsset(data.context.selectedTenantId, selectedInventory.id, draft);
      if (data.context.assetLifecycleState === 'active') {
        data = { ...data, assets: [asset, ...data.assets] };
      } else {
        data = await repository.selectAssetLifecycle(asset.tenantId, asset.inventoryId, 'active');
        mode = 'home';
        selectedLocationId = null;
        selectedAssetId = null;
        loadedAssetDetail = null;
      }
      addOpen = false;
      message = `Saved ${asset.title}.`;
    });
  }

  async function updateAsset(draft: UpdateAssetDraft): Promise<void> {
    if (!selectedAsset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    busy = true;
    error = '';
    message = '';
    try {
      const asset = await repository.updateAsset(
        selectedAsset.tenantId,
        selectedAsset.inventoryId,
        selectedAsset.id,
        draft
      );
      replaceWorkspaceAsset(asset);
      loadedAssetDetail = asset;
      message = `Saved ${asset.title}.`;
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Action failed.';
      throw new Error(error);
    } finally {
      busy = false;
    }
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

  async function selectAssetLifecycle(lifecycleState: AssetLifecycleFilter): Promise<void> {
    if (!selectedInventory) {
      return;
    }
    await run(async () => {
      data = await repository.selectAssetLifecycle(data.context.selectedTenantId, selectedInventory.id, lifecycleState);
      mode = 'home';
      selectedLocationId = null;
      selectedAssetId = null;
      loadedAssetDetail = null;
    });
  }

  async function archiveSelectedAsset(): Promise<void> {
    const asset = selectedAsset;
    if (!asset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.archiveAsset(asset.tenantId, asset.inventoryId, asset.id);
      await refreshSelectedAssetLifecycle();
      closeDetailToHome();
      message = `Archived ${asset.title}.`;
    });
  }

  async function restoreSelectedAsset(): Promise<void> {
    const asset = selectedAsset;
    if (!asset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.restoreAsset(asset.tenantId, asset.inventoryId, asset.id);
      await refreshSelectedAssetLifecycle();
      closeDetailToHome();
      message = `Restored ${asset.title}.`;
    });
  }

  async function deleteSelectedAsset(): Promise<void> {
    const asset = selectedAsset;
    if (!asset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.deleteAsset(asset.tenantId, asset.inventoryId, asset.id);
      await refreshSelectedAssetLifecycle();
      closeDetailToHome();
      message = `Deleted ${asset.title}.`;
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
    loadedAssetDetail = null;
    mode = 'location';
  }

  function openAsset(asset: Asset): void {
    void loadAssetDetail(asset.tenantId, asset.inventoryId, asset.id);
  }

  function openAssetById(assetId: string): void {
    const asset =
      data.assets.find((candidate) => candidate.id === assetId) ??
      searchResults.find((result) => result.asset.id === assetId)?.asset;
    if (asset) {
      openAsset(asset);
    }
  }

  async function loadAssetDetail(tenantId: string, inventoryId: string, assetId: string): Promise<void> {
    await run(async () => {
      const asset = await repository.getAsset(tenantId, inventoryId, assetId);
      replaceWorkspaceAsset(asset);
      loadedAssetDetail = asset;
      selectedAssetId = asset.id;
      mode = 'asset';
    });
  }

  function replaceWorkspaceAsset(asset: Asset): void {
    if (asset.tenantId !== data.context.selectedTenantId || asset.inventoryId !== data.context.selectedInventoryId) {
      return;
    }
    if (asset.lifecycleState !== data.context.assetLifecycleState) {
      return;
    }
    const existing = data.assets.some(
      (candidate) =>
        candidate.tenantId === asset.tenantId && candidate.inventoryId === asset.inventoryId && candidate.id === asset.id
    );
    data = {
      ...data,
      assets: existing
        ? data.assets.map((candidate) =>
            candidate.tenantId === asset.tenantId && candidate.inventoryId === asset.inventoryId && candidate.id === asset.id
              ? asset
              : candidate
          )
        : [asset, ...data.assets]
    };
  }

  async function refreshSelectedAssetLifecycle(): Promise<void> {
    if (!selectedInventory) {
      return;
    }
    data = await repository.selectAssetLifecycle(
      data.context.selectedTenantId,
      selectedInventory.id,
      data.context.assetLifecycleState
    );
  }

  function closeDetailToHome(): void {
    mode = 'home';
    selectedLocationId = null;
    selectedAssetId = null;
    loadedAssetDetail = null;
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
    onSelectTenant={(tenantId) => { void selectTenant(tenantId); }}
    onSelectInventory={(tenantId, inventoryId) => { void selectInventory(tenantId, inventoryId); }}
    onModeChange={(nextMode) => { mode = nextMode; }}
    {onSignOut}
  />

  <div class="workspace-column">
    <TopHeader
      tenants={data.context.tenants}
      inventories={data.context.inventories}
      selectedTenantId={data.context.selectedTenantId}
      inventory={selectedInventory}
      bind:query={searchQuery}
      canCreateAsset={createAssetAllowed && data.context.assetLifecycleState === 'active'}
      onSelectTenant={(tenantId) => { void selectTenant(tenantId); }}
      onSelectInventory={(tenantId, inventoryId) => { void selectInventory(tenantId, inventoryId); }}
      onSearch={() => { void search(); }}
      onOpenAdd={() => { addOpen = true; }}
    />

    {#if data.context.inventories.length === 0}
      <section class="workspace-main">
        <div class="empty-state spacious">
          <h1>No inventory yet</h1>
          {#if canCreateStarter}
            <p>{data.context.selectedTenantId ? 'Create the first inventory for this tenant.' : 'Create your first tenant and inventory.'}</p>
            <Button.Root onclick={() => { void createStarterInventory(); }}>Create Household</Button.Root>
          {:else}
            <p>You can view this tenant, but you cannot create inventories in it.</p>
          {/if}
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
        asset={withTrail(selectedAsset, detailAssets)}
        canEdit={editAssetAllowed}
        parentTargets={moveParentTargets(detailAssets, selectedAsset.id)}
        saving={busy}
        onBack={() => { mode = selectedLocationId ? 'location' : 'home'; selectedAssetId = null; }}
        onSave={updateAsset}
        onArchive={archiveSelectedAsset}
        onRestore={restoreSelectedAsset}
        onDelete={deleteSelectedAsset}
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
        lifecycleState={data.context.assetLifecycleState}
        locations={topLevelLocations(data.assets)}
        recentAssets={recentlyAddedAssets(data.assets)}
        archivedAssets={data.assets}
        onOpenLocation={openLocation}
        onOpenAsset={openAsset}
        onOpenAdd={() => { addOpen = true; }}
        onSelectLifecycle={(lifecycleState) => { void selectAssetLifecycle(lifecycleState); }}
      />
    {/if}
  </div>

  <MobileNav
    {mode}
    canCreateAsset={createAssetAllowed && data.context.assetLifecycleState === 'active'}
    onModeChange={(nextMode) => { mode = nextMode; }}
    onOpenAdd={() => { addOpen = true; }}
  />

  <AddAssetTray
    open={addOpen && createAssetAllowed}
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
