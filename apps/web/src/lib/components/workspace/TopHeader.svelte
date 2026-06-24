<script lang="ts">
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import type { Inventory } from '$lib/domain/inventory';

  let {
    inventory,
    query = $bindable(''),
    canEdit,
    onSearch,
    onOpenAdd
  }: {
    inventory: Inventory | null;
    query: string;
    canEdit: boolean;
    onSearch: () => void;
    onOpenAdd: () => void;
  } = $props();
</script>

<header class="workspace-header">
  <div class="mobile-context">
    <strong>{inventory?.name ?? 'Stuff Stash'}</strong>
  </div>
  <form class="global-search" onsubmit={(event) => { event.preventDefault(); onSearch(); }}>
    <Search aria-hidden="true" />
    <Input bind:value={query} placeholder="Search this inventory" aria-label="Search this inventory" />
    <Button.Root type="submit" variant="ghost" size="icon-sm" aria-label="Run search"><Search /></Button.Root>
  </form>
  <Button.Root class="header-add" disabled={!canEdit || !inventory} onclick={onOpenAdd}>
    <Plus /> Add
  </Button.Root>
</header>
