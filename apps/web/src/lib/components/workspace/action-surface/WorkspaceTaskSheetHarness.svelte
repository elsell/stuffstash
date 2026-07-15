<script lang="ts">
  import WorkspaceTaskSheet from './WorkspaceTaskSheet.svelte';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';

  let { busy = false, closeHref }: { busy?: boolean; closeHref?: string } = $props();
</script>

<WorkspaceTaskSheet open title="Edit asset" description="Change the asset details." {busy} {closeHref}>
  <Label for="task-name">Name</Label>
  <Input id="task-name" />
  {#snippet footer()}<Button.Root>Save</Button.Root>{/snippet}
</WorkspaceTaskSheet>
