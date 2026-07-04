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
  import { customizationArchiveConfirmation } from '$lib/application/workspaceCustomizationActions';

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

  let confirmation = $derived(
    customizationArchiveConfirmation({
      assetType,
      fieldDefinition,
      busy,
      canArchiveScope
    })
  );
</script>

<section
  bind:this={panelElement}
  class="settings-panel archive-confirmation"
  aria-labelledby="customization-archive-title"
  tabindex="-1"
>
  {#if !confirmation.unavailable}
    <div class="settings-panel-heading">
      <Trash2 aria-hidden="true" />
      <div>
        <h3 id="customization-archive-title">{confirmation.title}</h3>
        <p>{confirmation.targetLabel}</p>
      </div>
    </div>
    <p class="muted-note">{confirmation.description}</p>
    <div class="heading-actions">
      <Button.Root href={fieldsHref} variant="outline" onclick={onClose}>Cancel</Button.Root>
      <Button.Root
        variant="destructive"
        disabled={confirmation.disabled}
        onclick={() => {
          if (assetType) {
            void onArchiveAssetType(assetType);
          } else if (fieldDefinition) {
            void onArchiveFieldDefinition(fieldDefinition);
          }
        }}
      >
        {confirmation.buttonLabel}
      </Button.Root>
    </div>
  {:else}
    <div class="settings-panel-heading">
      <Trash2 aria-hidden="true" />
      <div>
        <h3 id="customization-archive-title">{confirmation.title}</h3>
        <p>{confirmation.targetLabel}</p>
      </div>
    </div>
    <Button.Root href={fieldsHref} variant="outline" onclick={onClose}>{confirmation.buttonLabel}</Button.Root>
  {/if}
</section>
