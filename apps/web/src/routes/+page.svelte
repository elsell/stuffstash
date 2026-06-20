<script lang="ts">
  import {
    StuffStashAPIError,
    StuffStashClient,
    type Asset,
    type AssetLifecycleState,
    type Inventory,
    type Principal
  } from '@stuff-stash/api-client';
  import { onMount } from 'svelte';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import InventoryAssetFlow from '$lib/components/InventoryAssetFlow.svelte';
  import SessionPanel from '$lib/components/SessionPanel.svelte';
  import SignInPanel from '$lib/components/SignInPanel.svelte';
  import * as Alert from '$lib/components/ui/alert/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { getStoredSession, signOut, startSignIn, type AuthSession } from '$lib/auth';
  import { loadRuntimeConfig, type RuntimeConfig } from '$lib/runtimeConfig';

  let config: RuntimeConfig | null = null;
  let session: AuthSession | null = null;
  let principal: Principal | null = null;
  let tenantId = '';
  let tenantName = 'Home';
  let inventoryName = 'Main Inventory';
  let selectedInventoryId = '';
  let assetLifecycleState: AssetLifecycleState = 'active';
  let assetKind: 'item' | 'container' | 'location' = 'item';
  let assetTitle = '';
  let assetDescription = '';
  let inventories: Inventory[] = [];
  let assets: Asset[] = [];
  let loading = true;
  let busy = false;
  let message = '';
  let error = '';

  $: selectedInventory = inventories.find((inventory) => inventory.id === selectedInventoryId) ?? inventories[0] ?? null;

  onMount(async () => {
    try {
      config = await loadRuntimeConfig();
      session = getStoredSession();
      if (session) {
        await loadIdentity();
      }
    } catch (caught) {
      error = errorMessage(caught);
    } finally {
      loading = false;
    }
  });

  function client(): StuffStashClient {
    if (!config) {
      throw new Error('Web runtime configuration has not loaded.');
    }
    return new StuffStashClient({
      baseUrl: config.apiBaseUrl,
      tokenProvider: () => getStoredSession()?.idToken ?? null
    });
  }

  async function signIn(): Promise<void> {
    if (!config) {
      return;
    }
    await startSignIn(config);
  }

  async function signOutAndReset(): Promise<void> {
    signOut();
    session = null;
    principal = null;
    inventories = [];
    assets = [];
    tenantId = '';
    selectedInventoryId = '';
    message = 'Signed out.';
  }

  async function loadIdentity(): Promise<void> {
    principal = await client().me();
  }

  async function createTenantAndInventory(): Promise<void> {
    await runTask(async () => {
      const tenant = await client().createTenant(tenantName);
      tenantId = tenant.id;
      const inventory = await client().createInventory(tenant.id, inventoryName);
      inventories = [inventory];
      selectedInventoryId = inventory.id;
      assetLifecycleState = 'active';
      assets = [];
      message = `Created ${inventory.name}.`;
    });
  }

  async function refreshInventories(): Promise<void> {
    if (!tenantId) {
      error = 'Create or enter a tenant before refreshing inventories.';
      return;
    }
    await runTask(async () => {
      inventories = (await client().listInventories(tenantId)).items;
      if (!inventories.some((inventory) => inventory.id === selectedInventoryId)) {
        selectedInventoryId = inventories[0]?.id ?? '';
      }
      await refreshAssets();
      message = 'Inventories refreshed.';
    });
  }

  async function refreshAssets(inventoryId = selectedInventoryId): Promise<void> {
    if (!tenantId || !inventoryId) {
      assets = [];
      return;
    }
    assets = (await client().listAssets(tenantId, inventoryId, 50, undefined, assetLifecycleState)).items;
  }

  async function selectInventory(inventoryId: string): Promise<void> {
    selectedInventoryId = inventoryId;
    await runTask(() => refreshAssets(inventoryId));
  }

  async function selectAssetLifecycle(lifecycleState: AssetLifecycleState): Promise<void> {
    assetLifecycleState = lifecycleState;
    await runTask(refreshAssets);
  }

  async function createAsset(): Promise<void> {
    if (!tenantId || !selectedInventory) {
      error = 'Create an inventory before adding assets.';
      return;
    }
    await runTask(async () => {
      const asset = await client().createAsset(tenantId, selectedInventory.id, {
        kind: assetKind,
        title: assetTitle,
        description: assetDescription
      });
      assets = [asset, ...assets];
      assetTitle = '';
      assetDescription = '';
      message = `Added ${asset.title}.`;
    });
  }

  async function archiveAsset(asset: Asset): Promise<void> {
    if (!tenantId || !selectedInventory) {
      return;
    }
    await runTask(async () => {
      await client().archiveAsset(tenantId, selectedInventory.id, asset.id);
      await refreshAssets();
      message = `Archived ${asset.title}.`;
    });
  }

  async function restoreAsset(asset: Asset): Promise<void> {
    if (!tenantId || !selectedInventory) {
      return;
    }
    await runTask(async () => {
      await client().restoreAsset(tenantId, selectedInventory.id, asset.id);
      await refreshAssets();
      message = `Restored ${asset.title}.`;
    });
  }

  async function deleteAsset(asset: Asset): Promise<void> {
    if (!tenantId || !selectedInventory) {
      return;
    }
    await runTask(async () => {
      await client().deleteAsset(tenantId, selectedInventory.id, asset.id);
      assets = assets.filter((item) => item.id !== asset.id);
      message = `Deleted ${asset.title}.`;
    });
  }

  async function runTask(task: () => Promise<void>): Promise<void> {
    busy = true;
    error = '';
    message = '';
    try {
      await task();
    } catch (caught) {
      error = errorMessage(caught);
    } finally {
      busy = false;
    }
  }

  function errorMessage(caught: unknown): string {
    if (caught instanceof StuffStashAPIError) {
      return caught.message;
    }
    if (caught instanceof Error) {
      return caught.message;
    }
    return 'Something went wrong.';
  }
</script>

<svelte:head>
  <title>Stuff Stash</title>
</svelte:head>

<main class="app-shell">
  <AppHeader {session} onSignOut={signOutAndReset} />

  {#if loading}
    <Card.Root>
      <Card.Content>
        <p class="muted">Loading...</p>
      </Card.Content>
    </Card.Root>
  {:else if !session}
    <SignInPanel onSignIn={signIn} />
  {:else}
    <section class="workspace">
      <SessionPanel
        bind:tenantId
        {principal}
        {busy}
        onLoadIdentity={() => { void runTask(loadIdentity); }}
        onRefreshInventories={() => { void refreshInventories(); }}
      />

      <InventoryAssetFlow
        bind:tenantName
        bind:inventoryName
        bind:assetKind
        bind:assetTitle
        bind:assetDescription
        {inventories}
        {selectedInventory}
        {assetLifecycleState}
        {assets}
        {busy}
        onCreateInventory={() => { void createTenantAndInventory(); }}
        onSelectInventory={(inventoryId) => { void selectInventory(inventoryId); }}
        onSelectAssetLifecycle={(lifecycleState) => { void selectAssetLifecycle(lifecycleState); }}
        onRefreshAssets={() => { void runTask(refreshAssets); }}
        onCreateAsset={() => { void createAsset(); }}
        onArchiveAsset={(asset) => { void archiveAsset(asset); }}
        onRestoreAsset={(asset) => { void restoreAsset(asset); }}
        onDeleteAsset={(asset) => { void deleteAsset(asset); }}
      />
    </section>
  {/if}

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
</main>
