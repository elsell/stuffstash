<script lang="ts">
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

<WorkspaceConfirmationDialog open title="Delete asset" description="Delete it permanently?" {busy}>
  {#snippet cancel()}<Button.Root variant="outline">Cancel</Button.Root>{/snippet}
  {#snippet action()}<Button.Root variant="destructive" onclick={deleteAction}>Delete</Button.Root>{/snippet}
</WorkspaceConfirmationDialog>
