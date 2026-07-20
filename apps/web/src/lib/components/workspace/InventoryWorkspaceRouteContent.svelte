<script lang="ts" module>
  import type {
	    Asset,
	    AssetAttachment,
	    AssetCheckout,
    AssetKind,
    AssetTag,
    BrowseScope,
    BrowseSort,
    BrowseSurface,
    CustomAssetType,
    CustomFieldDefinition,
    Inventory,
	    LocationAsset,
	    SearchCheckoutFilter,
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
  import type { InventoryTagRepository } from '$lib/ports/inventoryTagRepository';
  import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import type {
    AssetRouteAction,
    ImportSourceRoute,
    SettingsSection,
    WorkspaceRouteState
  } from '$lib/application/workspaceRoute';

  export type RouteContentWorkspace = {
    data: WorkspaceData;
    repository: InventoryRepository & InventoryAccessRepository & InventoryAuditRepository & InventoryCustomizationRepository & InventoryTagRepository;
    observer: WorkspaceObserver;
    selectedTenant: Tenant | null;
    selectedInventory: Inventory | null;
    selectedLocation: LocationAsset | null;
    selectedAsset: Asset | null;
    assets: Asset[];
    detailAssets: Asset[];
    selectedAssetAttachments: AssetAttachment[];
    selectedAssetCheckoutHistory: AssetCheckout[];
  };

  export type RouteContentStatus = {
    busy: boolean;
    canCreateStarter: boolean;
    createAssetAllowed: boolean;
    editAssetAllowed: boolean;
  };

  export type RouteContentRouteState = {
    routeUnavailable: string;
    assetDetailLoading: boolean;
    mode: WorkspaceMode;
    searchResults: SearchResult[];
    searchSuggestions: Asset[];
    searchSubmitted: boolean;
    searchError: string;
    browseSurface: BrowseSurface;
    browseScope: BrowseScope;
    browseSort: BrowseSort;
    browseTagIds: string[];
    browseAssets: Asset[];
    browseInventoryEmpty: boolean;
    browseHasMore: boolean;
    browseLoadingMore: boolean;
    browseBusy: boolean;
    browseErrorPhase: 'initial' | 'replacement' | 'append' | 'map' | null;
    assetAction: AssetRouteAction;
    attachmentId: string | null;
    attachmentAction: WorkspaceRouteState['attachmentAction'];
    settingsSection: SettingsSection;
    settingsLevel: WorkspaceRouteState['settingsLevel'];
    settingsCollection: WorkspaceRouteState['settingsCollection'];
    settingsLifecycle: WorkspaceRouteState['settingsLifecycle'];
    settingsResourceId: WorkspaceRouteState['settingsResourceId'];
    settingsResourceAction: WorkspaceRouteState['settingsResourceAction'];
    invitationStatus: WorkspaceRouteState['invitationStatus'];
    accessInvitationAction: WorkspaceRouteState['accessInvitationAction'];
    accessInvitationId: string | null;
    auditScope: WorkspaceRouteState['auditScope'];
    customizationAction: WorkspaceRouteState['customizationAction'];
    customAssetTypeId: string | null;
    customFieldDefinitionId: string | null;
    importSource: WorkspaceRouteState['importSource'];
    importJobId: WorkspaceRouteState['importJobId'];
    importTab: WorkspaceRouteState['importTab'];
  };

  export type RouteContentHrefs = {
    homeHref: string;
    assetDetailBackHref: string;
  };

  export type RouteContentHandlers = {
    onHome: (event: MouseEvent) => void;
    onOpenLocation: (asset: Asset) => void;
    onOpenLocations?: () => void;
    onBrowseStateChange: (state: {
      surface?: BrowseSurface;
      scope?: BrowseScope;
      lifecycleState?: SearchLifecycleFilter;
      checkoutState?: SearchCheckoutFilter;
      sort?: BrowseSort;
      selectedTagIds?: string[];
    }) => void;
    onBrowseLoadMore: () => Promise<void>;
    onBrowseRetry: () => Promise<void>;
    onEditLocation: (asset: Asset) => void;
    onOpenAsset: (asset: Asset) => Promise<void>;
    onOpenAdd: (kind?: AssetKind, parentAssetId?: string | null, opener?: HTMLElement | null) => void;
    onCloseLocation: () => void;
    onCloseAssetDetail: () => void;
    onAssetActionOpen: (action: Exclude<AssetRouteAction, null>) => void;
    onAssetActionClose: () => void;
    onAssetSave: (draft: UpdateAssetDraft) => Promise<void>;
    onMoveAssetHere: (asset: Asset) => Promise<void>;
    onAssetArchive: () => Promise<void>;
    onAssetRestore: () => Promise<void>;
    onAssetDelete: () => Promise<void>;
    onAssetCheckout: (details: string) => Promise<void>;
    onAssetReturn: (details: string) => Promise<void>;
    onHomeAssetReturn: (asset: Asset) => Promise<void>;
    onAssetUploadAttachment: (attachment: SelectedAttachment) => Promise<void>;
    onAssetArchiveAttachment: (attachment: AssetAttachment) => Promise<void>;
    onAttachmentDeleteOpen: (attachmentId: string) => void;
    onAttachmentDeleteClose: () => void;
    onAssetDeleteAttachment: (attachment: AssetAttachment) => Promise<void>;
    onAssetTagSearch?: (tag: AssetTag) => Promise<void>;
    onSearch: () => Promise<void>;
    onOpenSearchAsset: (asset: Asset) => void;
    onImportSourceChange: (source: ImportSourceRoute) => void;
    onImportJobSelectionChange: (jobId: string | null, tab?: WorkspaceRouteState['importTab']) => void;
    onImportJobTabChange: (tab: WorkspaceRouteState['importTab']) => void;
    onImportJobInventoryChanged: (scope: { tenantId: string; inventoryId: string }) => Promise<void>;
    onOpenImportedAssetId: (assetId: string) => Promise<void>;
    onOpenInventoryAuditHistory: () => void;
    onSettingsSectionChange: (section: SettingsSection) => void;
    onSettingsNavigate: (href: string) => void;
    onSettingsTagsChange: (tags: import('$lib/domain/inventory').ManagedAssetTag[]) => void;
    onSettingsPermissionDenied: () => Promise<void>;
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
    searchCheckoutState: SearchCheckoutFilter;
  };
</script>

<script lang="ts">
  import { moveParentTargets, recentlyChangedAssets, topLevelLocations, withTrail } from '$lib/application/workspace';
  import {
    workspaceNoInventoryPresentation,
    workspaceUnavailableRoutePresentation
  } from '$lib/application/workspaceRouteRecoveryPresentation';
  import * as Button from '$lib/components/ui/button/index.js';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import AssetDetail from './AssetDetail.svelte';
  import BrowsePanel from './BrowsePanel.svelte';
  import HomeWorkspace from './HomeWorkspace.svelte';
  import InventoryImportWorkspace from './InventoryImportWorkspace.svelte';
  import SettingsWorkspace from './settings/SettingsWorkspace.svelte';
  import LocationView from './LocationView.svelte';

  let {
    workspace,
    status,
    route,
    hrefs,
    handlers,
    searchQuery = $bindable(),
    searchLifecycleState = $bindable(),
    searchMode = $bindable(),
    searchCheckoutState = $bindable()
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
    </div>
  </section>
{:else if route.assetDetailLoading}
  <section class="workspace-main" aria-busy="true">
    <div class="empty-state spacious" role="status" aria-live="polite">
      <LoaderCircle class="size-6 motion-safe:animate-spin motion-reduce:animate-none" aria-hidden="true" />
      <h1>Loading asset details</h1>
      <p>Getting the latest details and files.</p>
    </div>
  </section>
{:else if route.mode === 'location' && workspace.selectedLocation}
  <LocationView
    location={workspace.selectedLocation}
    workspaceAssets={workspace.assets}
    canEdit={status.editAssetAllowed}
    canCreateAsset={status.createAssetAllowed}
    saving={status.busy}
    moveHereOpen={route.assetAction === 'move-here'}
    onBack={handlers.onCloseLocation}
    onOpenLocation={handlers.onOpenLocation}
    onEditLocation={handlers.onEditLocation}
    onOpenAsset={handlers.onOpenAsset}
    onOpenAdd={handlers.onOpenAdd}
    onOpenMoveHere={() => handlers.onAssetActionOpen('move-here')}
    onCloseMoveHere={handlers.onAssetActionClose}
    onMoveHere={handlers.onMoveAssetHere}
    onTagSearch={handlers.onAssetTagSearch}
  />
{:else if route.mode === 'asset' && workspace.selectedAsset}
  <AssetDetail
    asset={withTrail(workspace.selectedAsset, workspace.detailAssets)}
    canEdit={status.editAssetAllowed}
    canCreate={status.createAssetAllowed}
    workspaceAssets={workspace.detailAssets}
    parentTargets={moveParentTargets(workspace.detailAssets, workspace.selectedAsset.id)}
    customFieldDefinitions={workspace.data.context.customFieldDefinitions}
    assetTags={workspace.data.context.assetTags ?? []}
    saving={status.busy}
    attachments={workspace.selectedAssetAttachments}
    checkoutHistory={workspace.selectedAssetCheckoutHistory}
    mediaPolicy={workspace.data.context.mediaUploadPolicy}
    action={route.assetAction}
    attachmentId={route.attachmentId}
    attachmentAction={route.attachmentAction}
    backHref={hrefs.assetDetailBackHref}
    onBack={handlers.onCloseAssetDetail}
    onActionOpen={handlers.onAssetActionOpen}
    onActionClose={handlers.onAssetActionClose}
    onOpenAsset={(asset) => asset.kind === 'location' ? handlers.onOpenLocation(asset) : void handlers.onOpenAsset(asset)}
    onOpenAdd={handlers.onOpenAdd}
    onMoveHere={handlers.onMoveAssetHere}
    onSave={handlers.onAssetSave}
    onArchive={handlers.onAssetArchive}
    onRestore={handlers.onAssetRestore}
    onDelete={handlers.onAssetDelete}
    onCheckout={handlers.onAssetCheckout}
    onReturn={handlers.onAssetReturn}
    onUploadAttachment={(attachment: SelectedAttachment) => handlers.onAssetUploadAttachment(attachment)}
    onArchiveAttachment={handlers.onAssetArchiveAttachment}
    onAttachmentDeleteOpen={handlers.onAttachmentDeleteOpen}
    onAttachmentDeleteClose={handlers.onAttachmentDeleteClose}
    onDeleteAttachment={handlers.onAssetDeleteAttachment}
    onTagSearch={handlers.onAssetTagSearch}
  />
{:else if route.mode === 'browse'}
  <BrowsePanel
    tenantId={workspace.data.context.selectedTenantId}
    inventoryId={workspace.data.context.selectedInventoryId}
    inventoryName={workspace.selectedInventory?.name ?? 'Inventory'}
    assets={route.browseAssets}
    placementAssets={workspace.assets}
    inventoryEmpty={route.browseInventoryEmpty}
    bind:query={searchQuery}
    results={route.searchResults}
    suggestions={route.searchSuggestions}
    assetTags={workspace.data.context.assetTags ?? []}
    submitted={route.searchSubmitted}
    error={route.searchError}
    busy={route.browseBusy}
    surface={route.browseSurface}
    scope={route.browseScope}
    lifecycleState={searchLifecycleState}
    {searchMode}
    checkoutState={searchCheckoutState}
    sort={route.browseSort}
    selectedTagIds={route.browseTagIds}
    canCreateAsset={status.createAssetAllowed}
    hasMore={route.browseHasMore}
    loadingMore={route.browseLoadingMore}
    errorPhase={route.browseErrorPhase}
    onStateChange={handlers.onBrowseStateChange}
    onLoadMore={handlers.onBrowseLoadMore}
    onRetry={handlers.onBrowseRetry}
    onSearch={() => { void handlers.onSearch(); }}
    onOpenAsset={handlers.onOpenSearchAsset}
    onOpenAdd={(kind, parentAssetId, opener) => handlers.onOpenAdd(kind, parentAssetId, opener)}
  />
{:else if route.mode === 'import'}
  <InventoryImportWorkspace
    tenantId={workspace.data.context.selectedTenantId}
    inventory={workspace.selectedInventory}
    currentPrincipal={workspace.data.context.principal}
    repository={workspace.repository}
    importSource={route.importSource}
    importJobId={route.importJobId}
    importTab={route.importTab}
    onImportSourceChange={handlers.onImportSourceChange}
    onImportJobSelectionChange={handlers.onImportJobSelectionChange}
    onImportJobTabChange={handlers.onImportJobTabChange}
    onImportJobInventoryChanged={handlers.onImportJobInventoryChanged}
    onOpenImportedAssetId={handlers.onOpenImportedAssetId}
    onOpenInventoryAuditHistory={handlers.onOpenInventoryAuditHistory}
  />
{:else if route.mode === 'settings'}
  <SettingsWorkspace
    principal={workspace.data.context.principal}
    tenant={workspace.selectedTenant}
    inventory={workspace.selectedInventory}
    route={{
      settingsLevel: route.settingsLevel,
      settingsCollection: route.settingsCollection,
      settingsLifecycle: route.settingsLifecycle,
      settingsResourceId: route.settingsResourceId,
      settingsResourceAction: route.settingsResourceAction,
      invitationStatus: route.invitationStatus,
      accessInvitationAction: route.accessInvitationAction,
      accessInvitationId: route.accessInvitationId,
      auditScope: route.auditScope
    }}
    repository={workspace.repository}
    observer={workspace.observer}
    currentAssetTypes={workspace.data.context.customAssetTypes}
    currentFields={workspace.data.context.customFieldDefinitions}
    onNavigate={handlers.onSettingsNavigate}
    onSchemaChange={handlers.onCustomizationChange}
    onTagsChange={handlers.onSettingsTagsChange}
    onPermissionDenied={handlers.onSettingsPermissionDenied}
  />
{:else}
  <HomeWorkspace
    tenantId={workspace.data.context.selectedTenantId}
    inventoryId={workspace.data.context.selectedInventoryId}
    lifecycleState={workspace.data.context.assetLifecycleState}
    locations={topLevelLocations(workspace.assets)}
    recentAssets={recentlyChangedAssets(workspace.assets)}
    archivedAssets={workspace.assets}
    checkedOutAssets={workspace.data.checkedOutAssets.map((entry) => entry.asset)}
    canCreateAsset={status.createAssetAllowed}
    canEditAsset={status.editAssetAllowed}
    onOpenLocation={handlers.onOpenLocation}
    onOpenLocations={handlers.onOpenLocations}
    onOpenAsset={handlers.onOpenAsset}
    onReturnAsset={handlers.onHomeAssetReturn}
    onOpenAdd={(kind = 'location', parentAssetId, opener) => handlers.onOpenAdd(kind, parentAssetId, opener)}
    onSelectLifecycle={(lifecycleState) => { void handlers.onSelectLifecycle(lifecycleState); }}
    onTagSearch={handlers.onAssetTagSearch}
  />
{/if}
