<script lang="ts">
  import { StuffStashAPIError, StuffStashClient, type Asset, type Inventory, type Principal } from '@stuff-stash/api-client';
  import { onMount } from 'svelte';
  import AppHeader from '$lib/components/AppHeader.svelte';
  import InventoryAssetFlow from '$lib/components/InventoryAssetFlow.svelte';
  import SessionPanel from '$lib/components/SessionPanel.svelte';
  import SignInPanel from '$lib/components/SignInPanel.svelte';
  import { getStoredSession, signOut, startSignIn, type AuthSession } from '$lib/auth';
  import { loadRuntimeConfig, type RuntimeConfig } from '$lib/runtimeConfig';

  let config: RuntimeConfig | null = null;
  let session: AuthSession | null = null;
  let principal: Principal | null = null;
  let tenantId = '';
  let tenantName = 'Home';
  let inventoryName = 'Main Inventory';
  let selectedInventoryId = '';
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
      selectedInventoryId = inventories[0]?.id ?? '';
      await refreshAssets();
      message = 'Inventories refreshed.';
    });
  }

  async function refreshAssets(): Promise<void> {
    if (!tenantId || !selectedInventory) {
      assets = [];
      return;
    }
    assets = (await client().listAssets(tenantId, selectedInventory.id)).items;
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
    <section class="panel">
      <p class="muted">Loading…</p>
    </section>
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
        {assets}
        {busy}
        onCreateInventory={() => { void createTenantAndInventory(); }}
        onRefreshAssets={() => { void runTask(refreshAssets); }}
        onCreateAsset={() => { void createAsset(); }}
      />
    </section>
  {/if}

  {#if message}
    <p class="toast success">{message}</p>
  {/if}
  {#if error}
    <p class="toast danger">{error}</p>
  {/if}
</main>
