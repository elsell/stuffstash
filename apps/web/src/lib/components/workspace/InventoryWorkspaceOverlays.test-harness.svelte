<script lang="ts" module>
  import type { WorkspaceNotification } from '$lib/components/ui/sonner/index.js';
</script>

<script lang="ts">
  import InventoryWorkspaceOverlays from './InventoryWorkspaceOverlays.svelte';

  let notification = $state<WorkspaceNotification | null>(null);
  let error = $state('');

  export function showNotification(next: WorkspaceNotification | null): void {
    notification = next;
  }

  export function showError(next: string): void {
    error = next;
  }
</script>

<InventoryWorkspaceOverlays
  addOpen={false}
  createAssetAllowed={true}
  addKind="item"
  addParentAssetId={null}
  addCloseHref="/tenants/tenant-one/inventories/inventory-one"
  parentTargets={[]}
  mediaPolicy={{ supportedContentTypes: ['image/jpeg', 'image/png', 'image/webp'], maxBytes: 1024 }}
  customAssetTypes={[]}
  customFieldDefinitions={[]}
  saving={false}
  {notification}
  {error}
  onAddClose={() => {}}
  onAddSave={async () => ({ saved: true })}
/>
