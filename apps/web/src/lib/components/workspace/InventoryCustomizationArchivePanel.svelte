<script lang="ts" module>
  import type { CustomAssetType, CustomFieldDefinition } from '$lib/domain/inventory';

  export type InventoryCustomizationArchivePanelProps = {
    assetType: CustomAssetType | null;
    fieldDefinition: CustomFieldDefinition | null;
    busy: boolean;
    fieldsHref: string;
    panelElement: HTMLElement | null;
    canArchiveScope: (scope: 'tenant' | 'inventory') => boolean;
    onClose: (event: MouseEvent) => void;
    onArchiveAssetType: (assetType: CustomAssetType) => Promise<void>;
    onArchiveFieldDefinition: (definition: CustomFieldDefinition) => Promise<void>;
  };
</script>

<script lang="ts">
  import Trash2 from '@lucide/svelte/icons/trash-2';
  import * as Button from '$lib/components/ui/button/index.js';

  let {
    assetType,
    fieldDefinition,
    busy,
    fieldsHref,
    panelElement = $bindable(),
    canArchiveScope,
    onClose,
    onArchiveAssetType,
    onArchiveFieldDefinition
  }: InventoryCustomizationArchivePanelProps = $props();
</script>

<section
  bind:this={panelElement}
  class="settings-panel archive-confirmation"
  aria-labelledby="customization-archive-title"
  tabindex="-1"
>
  {#if assetType}
    <div class="settings-panel-heading">
      <Trash2 aria-hidden="true" />
      <div>
        <h3 id="customization-archive-title">Archive asset type</h3>
        <p>{assetType.displayName}</p>
      </div>
    </div>
    <p class="muted-note">Existing assets keep their data. This type will stop appearing in new asset forms.</p>
    <div class="heading-actions">
      <Button.Root href={fieldsHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root
        variant="destructive"
        disabled={busy || !canArchiveScope(assetType.scope)}
        onclick={() => { void onArchiveAssetType(assetType); }}
      >
        Archive
      </Button.Root>
    </div>
  {:else if fieldDefinition}
    <div class="settings-panel-heading">
      <Trash2 aria-hidden="true" />
      <div>
        <h3 id="customization-archive-title">Archive field definition</h3>
        <p>{fieldDefinition.displayName}</p>
      </div>
    </div>
    <p class="muted-note">Existing assets keep their field values. This field will stop appearing in edit forms.</p>
    <div class="heading-actions">
      <Button.Root href={fieldsHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root
        variant="destructive"
        disabled={busy || !canArchiveScope(fieldDefinition.scope)}
        onclick={() => { void onArchiveFieldDefinition(fieldDefinition); }}
      >
        Archive
      </Button.Root>
    </div>
  {:else}
    <div class="settings-panel-heading">
      <Trash2 aria-hidden="true" />
      <div>
        <h3 id="customization-archive-title">Archive target unavailable</h3>
        <p>This schema item is not available in the current fields list.</p>
      </div>
    </div>
    <Button.Root href={fieldsHref} variant="outline" onclick={onClose}>Back to fields</Button.Root>
  {/if}
</section>
