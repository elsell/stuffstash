<script lang="ts" module>
  import type {
    AddAssetSaveResult,
    AddAssetSubmission,
    AssetKind,
    AssetTag,
    CustomAssetType,
    CustomFieldDefinition,
    MediaUploadPolicy,
    ParentTargetViewModel
  } from '$lib/domain/inventory';
  import type { WorkspaceNotification } from '$lib/components/ui/sonner/index.js';

  export type InventoryWorkspaceOverlaysProps = {
    addOpen: boolean;
    createAssetAllowed: boolean;
    addKind: AssetKind;
    addParentAssetId: string | null;
    addCloseHref: string;
    parentTargets: ParentTargetViewModel[];
    mediaPolicy: MediaUploadPolicy;
    customAssetTypes: CustomAssetType[];
    customFieldDefinitions: CustomFieldDefinition[];
    assetTags?: AssetTag[];
    saving: boolean;
    notification: WorkspaceNotification | null;
    error: string;
    onAddClose: () => void;
    onAddSave: (draft: AddAssetSubmission) => Promise<AddAssetSaveResult>;
  };
</script>

<script lang="ts">
  import { notify, notifyError } from '$lib/components/ui/sonner/index.js';
  import AddAssetTray from './AddAssetTray.svelte';

  let {
    addOpen,
    createAssetAllowed,
    addKind,
    addParentAssetId,
    addCloseHref,
    parentTargets,
    mediaPolicy,
    customAssetTypes,
    customFieldDefinitions,
    assetTags = [],
    saving,
    notification,
    error,
    onAddClose,
    onAddSave
  }: InventoryWorkspaceOverlaysProps = $props();

  let lastNotificationKey = '';
  let lastError = '';

  $effect(() => {
    const nextNotificationKey = notification
      ? `${notification.id ?? ''}:${notification.kind}:${notification.title}:${notification.description ?? ''}:${notification.action?.label ?? ''}:${notification.action?.href ?? ''}`
      : '';
    if (notification && nextNotificationKey !== lastNotificationKey) {
      notify(notification);
    }
    lastNotificationKey = nextNotificationKey;
  });

  $effect(() => {
    if (error && error !== lastError) {
      notifyError(error);
    }
    lastError = error;
  });
</script>

<AddAssetTray
  open={addOpen && createAssetAllowed}
  initialKind={addKind}
  initialParentAssetId={addParentAssetId}
  closeHref={addCloseHref}
  {parentTargets}
  {mediaPolicy}
  {customAssetTypes}
  {customFieldDefinitions}
  {assetTags}
  {saving}
  onClose={onAddClose}
  onSave={onAddSave}
/>
