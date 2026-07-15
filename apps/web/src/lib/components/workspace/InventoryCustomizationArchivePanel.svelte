<script lang="ts" module>
  import type { CustomAssetType, CustomFieldDefinition } from '$lib/domain/inventory';

  export type InventoryCustomizationArchivePanelProps = {
    assetType: CustomAssetType | null;
    fieldDefinition: CustomFieldDefinition | null;
    busy: boolean;
    error: string;
    fieldsHref: string;
    canArchiveScope: (scope: 'tenant' | 'inventory') => boolean;
    onClose: (event: MouseEvent) => void;
    onDismiss: () => void;
    onCloseAutoFocus: (event: Event) => void;
    onArchiveAssetType: (assetType: CustomAssetType) => Promise<boolean>;
    onArchiveFieldDefinition: (definition: CustomFieldDefinition) => Promise<boolean>;
  };

  export function customizationArchiveFocusTarget(
    trigger: HTMLElement | null,
    fallback: HTMLElement | null
  ): HTMLElement | null {
    return trigger?.isConnected ? trigger : fallback;
  }
</script>

<script lang="ts">
  import { onDestroy } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import { customizationArchiveConfirmation } from '$lib/application/workspaceCustomizationActions';
  import WorkspaceConfirmationDialog from './action-surface/WorkspaceConfirmationDialog.svelte';

  let {
    assetType,
    fieldDefinition,
    busy,
    error,
    fieldsHref,
    canArchiveScope,
    onClose,
    onDismiss,
    onCloseAutoFocus,
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
  let open = $state(true);
  let dismissAfterClose = false;
  let dismissTimer: number | null = null;

  onDestroy(() => {
    if (dismissTimer !== null) window.clearTimeout(dismissTimer);
  });

  function requestDismiss(): void {
    dismissAfterClose = true;
    open = false;
  }

  function handleClose(event: MouseEvent): void {
    onClose(event);
    if (event.defaultPrevented) requestDismiss();
  }

  function handleCloseAutoFocus(event: Event): void {
    onCloseAutoFocus(event);
    if (!dismissAfterClose) return;
    dismissAfterClose = false;
    if (dismissTimer !== null) window.clearTimeout(dismissTimer);
    dismissTimer = window.setTimeout(() => {
      dismissTimer = null;
      onDismiss();
    }, 16);
  }

  async function archive(): Promise<void> {
    const succeeded = assetType
      ? await onArchiveAssetType(assetType)
      : fieldDefinition
        ? await onArchiveFieldDefinition(fieldDefinition)
        : false;
    if (succeeded) requestDismiss();
  }
</script>

<WorkspaceConfirmationDialog
  {open}
  title={confirmation.title}
  description={confirmation.unavailable ? 'The requested archive target is not available.' : confirmation.description}
  {busy}
  onOpenChange={(nextOpen) => { if (!nextOpen) requestDismiss(); }}
  onCloseAutoFocus={handleCloseAutoFocus}
>
  {#snippet children()}
    <p class="muted-note">{confirmation.targetLabel}</p>
    {#if error}<p class="denied-note" role="alert">{error}</p>{/if}
  {/snippet}
  {#snippet cancel()}
    <Button.Root href={fieldsHref} variant="outline" class="min-h-11" disabled={busy} onclick={handleClose} autofocus>
      {confirmation.unavailable ? confirmation.buttonLabel : 'Cancel'}
    </Button.Root>
  {/snippet}
  {#snippet action()}
    {#if !confirmation.unavailable}
      <Button.Root
        variant="destructive"
        class="min-h-11"
        disabled={confirmation.disabled}
        onclick={() => { void archive(); }}
      >
        {confirmation.buttonLabel}
      </Button.Root>
    {/if}
  {/snippet}
</WorkspaceConfirmationDialog>
