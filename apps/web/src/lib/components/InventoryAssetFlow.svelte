<script lang="ts">
  import type { Asset, AssetKind, Inventory } from '@stuff-stash/api-client';

  export let tenantName: string;
  export let inventoryName: string;
  export let assetKind: AssetKind;
  export let assetTitle: string;
  export let assetDescription: string;
  export let inventories: Inventory[];
  export let selectedInventory: Inventory | null;
  export let assets: Asset[];
  export let busy: boolean;
  export let onCreateInventory: () => void;
  export let onRefreshAssets: () => void;
  export let onCreateAsset: () => void;
</script>

<section class="panel flow-panel">
  <div class="section-heading">
    <div>
      <h2>Create inventory</h2>
      <p>Create a tenant and its first inventory, then add an asset.</p>
    </div>
  </div>
  <form class="form-grid" onsubmit={(event) => { event.preventDefault(); onCreateInventory(); }}>
    <label>
      Tenant name
      <input bind:value={tenantName} required maxlength="120" />
    </label>
    <label>
      Inventory name
      <input bind:value={inventoryName} required maxlength="120" />
    </label>
    <button type="submit" disabled={busy}>Create inventory</button>
  </form>

  {#if inventories.length > 0}
    <div class="divider"></div>
    <div class="section-heading compact">
      <div>
        <h2>Assets</h2>
        <p>{selectedInventory?.name ?? 'Inventory'} · {assets.length} active assets</p>
      </div>
      <button class="secondary" type="button" onclick={onRefreshAssets} disabled={busy}>Refresh</button>
    </div>

    <form class="asset-form" onsubmit={(event) => { event.preventDefault(); onCreateAsset(); }}>
      <label>
        Kind
        <select bind:value={assetKind}>
          <option value="item">Item</option>
          <option value="container">Container</option>
          <option value="location">Location</option>
        </select>
      </label>
      <label>
        Title
        <input bind:value={assetTitle} required maxlength="160" placeholder="Cordless drill" />
      </label>
      <label class="wide">
        Description
        <textarea bind:value={assetDescription} placeholder="Optional notes"></textarea>
      </label>
      <button type="submit" disabled={busy || !assetTitle}>Add asset</button>
    </form>

    <div class="asset-list" aria-live="polite">
      {#each assets as asset}
        <article class="asset-row">
          <div>
            <h3>{asset.title}</h3>
            <p>{asset.description || 'No description'}</p>
          </div>
          <span>{asset.kind}</span>
        </article>
      {:else}
        <div class="empty-state">
          <h3>No assets yet</h3>
          <p>Add the first item, container, or location in this inventory.</p>
        </div>
      {/each}
    </div>
  {/if}
</section>
