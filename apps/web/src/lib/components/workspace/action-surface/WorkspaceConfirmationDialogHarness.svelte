<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import WorkspaceConfirmationDialog from './WorkspaceConfirmationDialog.svelte';
  import * as Button from '$lib/components/ui/button/index.js';

  let {
    busy: initialBusy = false,
    busyOnDelete = false,
    onDelete = () => {}
  }: { busy?: boolean; busyOnDelete?: boolean; onDelete?: () => void } = $props();
  let busy = $state(false);

  $effect(() => {
    if (!busyOnDelete) busy = initialBusy;
  });

  function deleteAction(): void {
    onDelete();
    if (busyOnDelete) busy = true;
  }
</script>

<WorkspaceConfirmationDialog open title={t('web.WorkspaceConfirmationDialogHarness.deleteAsset')} description={t('web.WorkspaceConfirmationDialogHarness.deleteItPermanently')} {busy}>
  {#snippet cancel()}<Button.Root variant="outline">{t('web.WorkspaceConfirmationDialogHarness.cancel')}</Button.Root>{/snippet}
  {#snippet action()}<Button.Root variant="destructive" onclick={deleteAction}>{t('web.WorkspaceConfirmationDialogHarness.delete')}</Button.Root>{/snippet}
</WorkspaceConfirmationDialog>
