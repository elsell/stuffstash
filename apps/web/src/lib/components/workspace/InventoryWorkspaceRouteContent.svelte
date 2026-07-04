<script lang="ts" module>
  import type {
    Asset,
    AssetAttachment,
    AssetKind,
    CustomAssetType,
    CustomFieldDefinition,
    Inventory,
    LocationAsset,
    SearchLifecycleFilter,
    SearchMode,
    SearchResult,
    SelectedAttachment,
    Tenant,
    UpdateAssetDraft,
    WorkspaceData,
    WorkspaceMode
  } from '$lib/domain/inventory';
  import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
  import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import type {
    AssetRouteAction,
    SettingsSection,
    WorkspaceRouteState
  } from '$lib/application/workspaceRoute';

  export type RouteContentWorkspace = {
    data: WorkspaceData;
    repository: InventoryRepository & InventoryAccessRepository & InventoryAuditRepository & InventoryCustomizationRepository;
    selectedTenant: Tenant | null;
    selectedInventory: Inventory | null;
    selectedLocation: LocationAsset | null;
    selectedAsset: Asset | null;
    assets: Asset[];
    detailAssets: Asset[];
    selectedAssetAttachments: AssetAttachment[];
  };

  export type RouteContentStatus = {
    busy: boolean;
    canCreateStarter: boolean;
    createAssetAllowed: boolean;
    editAssetAllowed: boolean;
  };

  export type RouteContentRouteState = {
    routeUnavailable: string;
    mode: WorkspaceMode;
    searchResults: SearchResult[];
    searchSuggestions: Asset[];
    searchSubmitted: boolean;
    searchError: string;
    assetAction: AssetRouteAction;
    attachmentId: string | null;
    attachmentAction: WorkspaceRouteState['attachmentAction'];
    settingsSection: SettingsSection;
    invitationStatus: WorkspaceRouteState['invitationStatus'];
    accessInvitationAction: WorkspaceRouteState['accessInvitationAction'];
    accessInvitationId: string | null;
    auditScope: WorkspaceRouteState['auditScope'];
    customizationAction: WorkspaceRouteState['customizationAction'];
    customAssetTypeId: string | null;
    customFieldDefinitionId: string | null;
    importSourceType: WorkspaceRouteState['importSourceType'];
  };

  export type RouteContentHrefs = {
    homeHref: string;
    assetDetailBackHref: string;
  };

  export type RouteContentHandlers = {
    onHome: (event: MouseEvent) => void;
    onCreateStarterInventory: () => Promise<void>;
    onOpenLocation: (asset: Asset) => void;
    onEditLocation: (asset: Asset) => void;
    onOpenAsset: (asset: Asset) => Promise<void>;
    onOpenAdd: (kind?: AssetKind, parentAssetId?: string | null) => void;
    onCloseLocation: () => void;
    onCloseAssetDetail: () => void;
    onAssetActionOpen: (action: Exclude<AssetRouteAction, null>) => void;
    onAssetActionClose: () => void;
    onAssetSave: (draft: UpdateAssetDraft) => Promise<void>;
    onAssetArchive: () => Promise<void>;
    onAssetRestore: () => Promise<void>;
    onAssetDelete: () => Promise<void>;
    onAssetUploadAttachment: (attachment: SelectedAttachment) => Promise<void>;
    onAssetArchiveAttachment: (attachment: AssetAttachment) => Promise<void>;
    onAttachmentDeleteOpen: (attachmentId: string) => void;
    onAttachmentDeleteClose: () => void;
    onAssetDeleteAttachment: (attachment: AssetAttachment) => Promise<void>;
    onSearch: () => Promise<void>;
    onOpenSearchAsset: (asset: Asset) => void;
    onImportSourceChange: (sourceType: WorkspaceRouteState['importSourceType']) => void;
    onImported: () => Promise<void>;
    onSettingsSectionChange: (section: SettingsSection) => void;
    onInvitationStatusChange: (status: WorkspaceRouteState['invitationStatus']) => void;
    onAccessInvitationActionOpen: (action: WorkspaceRouteState['accessInvitationAction'], invitationId: string) => void;
    onAccessInvitationActionClose: () => void;
    onAuditScopeChange: (scope: WorkspaceRouteState['auditScope']) => void;
    onCustomizationArchiveOpen: (action: WorkspaceRouteState['customizationAction'], id: string) => void;
    onCustomizationArchiveClose: () => void;
    onCustomizationChange: (assetTypes: CustomAssetType[], fieldDefinitions: CustomFieldDefinition[]) => void;
    onSelectLifecycle: (lifecycleState: WorkspaceData['context']['assetLifecycleState']) => Promise<void>;
  };

  export type InventoryWorkspaceRouteContentProps = {
    workspace: RouteContentWorkspace;
    status: RouteContentStatus;
    route: RouteContentRouteState;
    hrefs: RouteContentHrefs;
    handlers: RouteContentHandlers;
    searchQuery: string;
    searchLifecycleState: SearchLifecycleFilter;
    searchMode: SearchMode;
  };
</script>

<script lang="ts">
  import { containedAssets, moveParentTargets, recentlyAddedAssets, topLevelLocations, withTrail } from '$lib/application/workspace';
  import {
    workspaceNoInventoryPresentation,
    workspaceUnavailableRoutePresentation
  } from '$lib/application/workspaceRouteRecoveryPresentation';
  import * as Button from '$lib/components/ui/button/index.js';
  import AssetDetail from './AssetDetail.svelte';
  import HomeboxImportPanel from './HomeboxImportPanel.svelte';
  import HomeWorkspace from './HomeWorkspace.svelte';
  import InventorySettings from './InventorySettings.svelte';
  import LocationView from './LocationView.svelte';
  import SearchPanel from './SearchPanel.svelte';

  let {
    workspace,
    status,
    route,
    hrefs,
    handlers,
    searchQuery = $bindable(),
    searchLifecycleState = $bindable(),
    searchMode = $bindable()
  }: InventoryWorkspaceRouteContentProps = $props();

  let routeUnavailablePresentation = $derived(workspaceUnavailableRoutePresentation(route.routeUnavailable));
  let noInventoryPresentation = $derived(
    workspaceNoInventoryPresentation(workspace.data.context.selectedTenantId, status.canCreateStarter)
  );
</script>

{#if route.routeUnavailable}
  <section class="workspace-main">
    <div class="empty-state spacious" role={routeUnavailablePresentation.role}>
      <h1>{routeUnavailablePresentation.title}</h1>
      <p>{routeUnavailablePresentation.message}</p>
      {#if routeUnavailablePresentation.actionLabel}
        <Button.Root href={hrefs.homeHref} onclick={handlers.onHome}>{routeUnavailablePresentation.actionLabel}</Button.Root>
      {/if}
    </div>
  </section>
{:else if workspace.data.context.inventories.length === 0}
  <section class="workspace-main">
    <div class="empty-state spacious">
      <h1>{noInventoryPresentation.title}</h1>
      <p>{noInventoryPresentation.message}</p>
      {#if noInventoryPresentation.actionLabel}
        <Button.Root onclick={() => { void handlers.onCreateStarterInventory(); }}>{noInventoryPresentation.actionLabel}</Button.Root>
      {/if}
    </div>
  </section>
{:else if route.mode === 'location' && workspace.selectedLocation}
  <LocationView
    location={workspace.selectedLocation}
    assets={containedAssets(workspace.assets, workspace.selectedLocation.id)}
    canEdit={status.editAssetAllowed}
    canCreateAsset={status.createAssetAllowed}
    onBack={handlers.onCloseLocation}
    onOpenLocation={handlers.onOpenLocation}
    onEditLocation={handlers.onEditLocation}
    onOpenAsset={handlers.onOpenAsset}
    onOpenAdd={handlers.onOpenAdd}
  />
{:else if route.mode === 'asset' && workspace.selectedAsset}
  <AssetDetail
    asset={withTrail(workspace.selectedAsset, workspace.detailAssets)}
    canEdit={status.editAssetAllowed}
    parentTargets={moveParentTargets(workspace.detailAssets, workspace.selectedAsset.id)}
    customFieldDefinitions={workspace.data.context.customFieldDefinitions}
    saving={status.busy}
    attachments={workspace.selectedAssetAttachments}
    mediaPolicy={workspace.data.context.mediaUploadPolicy}
    action={route.assetAction}
    attachmentId={route.attachmentId}
    attachmentAction={route.attachmentAction}
    backHref={hrefs.assetDetailBackHref}
    onBack={handlers.onCloseAssetDetail}
    onActionOpen={handlers.onAssetActionOpen}
    onActionClose={handlers.onAssetActionClose}
    onSave={handlers.onAssetSave}
    onArchive={handlers.onAssetArchive}
    onRestore={handlers.onAssetRestore}
    onDelete={handlers.onAssetDelete}
    onUploadAttachment={(attachment: SelectedAttachment) => handlers.onAssetUploadAttachment(attachment)}
    onArchiveAttachment={handlers.onAssetArchiveAttachment}
    onAttachmentDeleteOpen={handlers.onAttachmentDeleteOpen}
    onAttachmentDeleteClose={handlers.onAttachmentDeleteClose}
    onDeleteAttachment={handlers.onAssetDeleteAttachment}
  />
{:else if route.mode === 'search'}
  <SearchPanel
    tenantId={workspace.data.context.selectedTenantId}
    inventoryId={workspace.data.context.selectedInventoryId}
    bind:query={searchQuery}
    bind:lifecycleState={searchLifecycleState}
    bind:searchMode={searchMode}
    results={route.searchResults}
    suggestions={route.searchSuggestions}
    submitted={route.searchSubmitted}
    error={route.searchError}
    busy={status.busy}
    onSearch={() => { void handlers.onSearch(); }}
    onOpenAsset={handlers.onOpenSearchAsset}
  />
{:else if route.mode === 'import'}
  <HomeboxImportPanel
    tenantId={workspace.data.context.selectedTenantId}
    inventory={workspace.selectedInventory}
    repository={workspace.repository}
    sourceType={route.importSourceType}
    onSourceChange={handlers.onImportSourceChange}
    onImported={handlers.onImported}
  />
{:else if route.mode === 'settings'}
  <InventorySettings
    tenant={workspace.selectedTenant}
    inventory={workspace.selectedInventory}
    inventoryCount={workspace.data.context.inventories.length}
    accessRepository={workspace.repository}
    auditRepository={workspace.repository}
    customizationRepository={workspace.repository}
    customAssetTypes={workspace.data.context.customAssetTypes}
    customFieldDefinitions={workspace.data.context.customFieldDefinitions}
    section={route.settingsSection}
    invitationStatus={route.invitationStatus}
    accessInvitationAction={route.accessInvitationAction}
    accessInvitationId={route.accessInvitationId}
    auditScope={route.auditScope}
    customizationAction={route.customizationAction}
    customAssetTypeId={route.customAssetTypeId}
    customFieldDefinitionId={route.customFieldDefinitionId}
    onSectionChange={handlers.onSettingsSectionChange}
    onInvitationStatusChange={handlers.onInvitationStatusChange}
    onAccessInvitationActionOpen={handlers.onAccessInvitationActionOpen}
    onAccessInvitationActionClose={handlers.onAccessInvitationActionClose}
    onAuditScopeChange={handlers.onAuditScopeChange}
    onCustomizationArchiveOpen={handlers.onCustomizationArchiveOpen}
    onCustomizationArchiveClose={handlers.onCustomizationArchiveClose}
    onCustomizationChange={handlers.onCustomizationChange}
  />
{:else}
  <HomeWorkspace
    tenantId={workspace.data.context.selectedTenantId}
    inventoryId={workspace.data.context.selectedInventoryId}
    lifecycleState={workspace.data.context.assetLifecycleState}
    browseMode={route.mode === 'locations' ? 'locations' : 'home'}
    locations={topLevelLocations(workspace.assets)}
    recentAssets={recentlyAddedAssets(workspace.assets)}
    archivedAssets={workspace.assets}
    canCreateAsset={status.createAssetAllowed}
    onOpenLocation={handlers.onOpenLocation}
    onOpenAsset={handlers.onOpenAsset}
    onOpenAdd={() => handlers.onOpenAdd('location')}
    onSelectLifecycle={(lifecycleState) => { void handlers.onSelectLifecycle(lifecycleState); }}
  />
{/if}
