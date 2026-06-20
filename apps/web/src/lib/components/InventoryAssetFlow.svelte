<script lang="ts">
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import * as Select from '$lib/components/ui/select/index.js';
  import { Separator } from '$lib/components/ui/separator/index.js';
  import * as Tabs from '$lib/components/ui/tabs/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import type { Asset, AssetKind, AssetLifecycleState, Inventory } from '@stuff-stash/api-client';

  export let tenantName: string;
  export let inventoryName: string;
  export let assetKind: AssetKind;
  export let assetTitle: string;
  export let assetDescription: string;
  export let inventories: Inventory[];
  export let selectedInventory: Inventory | null;
  export let assetLifecycleState: AssetLifecycleState;
  export let assets: Asset[];
  export let busy: boolean;
  export let onCreateInventory: () => void;
  export let onSelectInventory: (inventoryId: string) => void;
  export let onSelectAssetLifecycle: (lifecycleState: AssetLifecycleState) => void;
  export let onRefreshAssets: () => void;
  export let onCreateAsset: () => void;
  export let onArchiveAsset: (asset: Asset) => void;
  export let onRestoreAsset: (asset: Asset) => void;
  export let onDeleteAsset: (asset: Asset) => void;

  const lifecycleOptions: { value: AssetLifecycleState; label: string }[] = [
    { value: 'active', label: 'Active' },
    { value: 'archived', label: 'Archived' }
  ];
</script>

<Card.Root class="flow-panel">
  <Card.Header class="p-0">
    <Card.Title>Create inventory</Card.Title>
    <Card.Description>Create a tenant and its first inventory, then add an asset.</Card.Description>
  </Card.Header>
  <form class="form-grid" onsubmit={(event) => { event.preventDefault(); onCreateInventory(); }}>
    <div class="field-stack">
      <Label for="tenant-name">Tenant name</Label>
      <Input id="tenant-name" bind:value={tenantName} required maxlength={120} />
    </div>
    <div class="field-stack">
      <Label for="inventory-name">Inventory name</Label>
      <Input id="inventory-name" bind:value={inventoryName} required maxlength={120} />
    </div>
    <Button type="submit" disabled={busy}>Create inventory</Button>
  </form>

  {#if inventories.length > 0}
    <Separator />
    <Tabs.Root value={selectedInventory?.id ?? ''} class="inventory-tabs">
      <Tabs.List aria-label="Inventories" class="flex-wrap justify-start">
        {#each inventories as inventory}
          <Tabs.Trigger
            value={inventory.id}
            onclick={() => { onSelectInventory(inventory.id); }}
            disabled={busy}
          >
            {inventory.name}
          </Tabs.Trigger>
        {/each}
      </Tabs.List>
    </Tabs.Root>
    <div class="section-heading compact">
      <div>
        <h2>Assets</h2>
        <p>{selectedInventory?.name ?? 'Inventory'} · {assets.length} {assetLifecycleState} assets</p>
      </div>
      <div class="asset-tools">
        <Tabs.Root value={assetLifecycleState}>
          <Tabs.List aria-label="Asset lifecycle view">
            {#each lifecycleOptions as option}
              <Tabs.Trigger
                value={option.value}
                onclick={() => { onSelectAssetLifecycle(option.value); }}
                disabled={busy}
              >
                {option.label}
              </Tabs.Trigger>
            {/each}
          </Tabs.List>
        </Tabs.Root>
        <Button variant="outline" type="button" onclick={onRefreshAssets} disabled={busy}>Refresh</Button>
      </div>
    </div>

    {#if assetLifecycleState === 'active'}
      <form class="asset-form" onsubmit={(event) => { event.preventDefault(); onCreateAsset(); }}>
        <div class="field-stack">
          <Label for="asset-kind">Kind</Label>
          <Select.Root type="single" bind:value={assetKind}>
            <Select.Trigger id="asset-kind" class="w-full">
              {assetKind}
            </Select.Trigger>
            <Select.Content>
              <Select.Item value="item" label="Item" />
              <Select.Item value="container" label="Container" />
              <Select.Item value="location" label="Location" />
            </Select.Content>
          </Select.Root>
        </div>
        <div class="field-stack">
          <Label for="asset-title">Title</Label>
          <Input id="asset-title" bind:value={assetTitle} required maxlength={160} placeholder="Cordless drill" />
        </div>
        <div class="field-stack wide">
          <Label for="asset-description">Description</Label>
          <Textarea id="asset-description" bind:value={assetDescription} placeholder="Optional notes" />
        </div>
        <Button type="submit" disabled={busy || !assetTitle}>Add asset</Button>
      </form>
    {/if}

    <div class="asset-list" aria-live="polite">
      {#each assets as asset}
        <Card.Root class="asset-row" size="sm">
          <Card.Content class="asset-row-content p-0">
            <div>
              <h3>{asset.title}</h3>
              <p>{asset.description || 'No description'}</p>
            </div>
            <div class="asset-actions">
              <Badge variant="secondary">{asset.kind}</Badge>
              {#if assetLifecycleState === 'active'}
                <Button variant="outline" type="button" onclick={() => { onArchiveAsset(asset); }} disabled={busy}>Archive</Button>
              {:else}
                <Button variant="outline" type="button" onclick={() => { onRestoreAsset(asset); }} disabled={busy}>Restore</Button>
              {/if}
              <Button variant="destructive" type="button" onclick={() => { onDeleteAsset(asset); }} disabled={busy}>Delete</Button>
            </div>
          </Card.Content>
        </Card.Root>
      {:else}
        <div class="empty-state">
          <h3>No assets yet</h3>
          <p>{assetLifecycleState === 'active' ? 'Add the first item, container, or location in this inventory.' : 'Archived assets will appear here.'}</p>
        </div>
      {/each}
    </div>
  {/if}
</Card.Root>
