<script lang="ts">
  let { field, message }: { field: string; message?: string } = $props();
</script>
{#if message}<p id={`${field}-error`} class="text-destructive text-sm">{message}</p>{/if}
