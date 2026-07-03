<script lang="ts" module>
  import type {
    AddAssetSaveResult,
    AddAssetSubmission,
    AssetKind,
    AssetViewModel,
    CustomAssetType,
    CustomFieldDefinition,
    MediaUploadPolicy
  } from '$lib/domain/inventory';

  export type InventoryWorkspaceOverlaysProps = {
    addOpen: boolean;
    createAssetAllowed: boolean;
    addKind: AssetKind;
    addParentAssetId: string | null;
    addCloseHref: string;
    parentTargets: AssetViewModel[];
    mediaPolicy: MediaUploadPolicy;
    customAssetTypes: CustomAssetType[];
    customFieldDefinitions: CustomFieldDefinition[];
    saving: boolean;
    message: string;
    error: string;
    onAddClose: () => void;
    onAddSave: (draft: AddAssetSubmission) => Promise<AddAssetSaveResult>;
  };
</script>

<script lang="ts">
  import * as Alert from '$lib/components/ui/alert/index.js';
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
    saving,
    message,
    error,
    onAddClose,
    onAddSave
  }: InventoryWorkspaceOverlaysProps = $props();
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
  {saving}
  onClose={onAddClose}
  onSave={onAddSave}
/>

{#if message}
  <Alert.Root class="toast" variant="default">
    <Alert.Description>{message}</Alert.Description>
  </Alert.Root>
{/if}
{#if error}
  <Alert.Root class="toast" variant="destructive">
    <Alert.Description>{error}</Alert.Description>
  </Alert.Root>
{/if}
