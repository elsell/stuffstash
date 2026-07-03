<script lang="ts">
  import { onMount } from 'svelte';
  import { containedAssets, moveParentTargets, parentTargets, recentlyAddedAssets, topLevelLocations, withTrail } from '$lib/application/workspace';
  import { loadWorkspaceAssetDetail, refreshWorkspaceAssetAttachments } from '$lib/application/workspaceAssetDetail';
  import { createAssetWorkflow, replaceWorkspaceAsset } from '$lib/application/workspaceAssetWorkflow';
  import { buildSearchSuggestions, executeWorkspaceSearch } from '$lib/application/workspaceSearch';
  import {
    type AssetRouteAction,
    type SettingsSection,
    type WorkspaceRouteState,
    workspaceRouteHref
  } from '$lib/application/workspaceRoute';
  import {
    assetRouteActionIsAvailable,
    currentWorkspaceRoute,
    findRouteInventory,
    findRouteTenant,
    pushWorkspaceRoute,
    replaceCanonicalWorkspaceAlias,
    replaceWorkspaceRoute,
    shouldCanonicalizeWorkspaceAlias
  } from '$lib/application/workspaceRouteNavigation';
  import {
    canCreateAsset,
    canEditAsset,
    canCreateInventory,
    type AddAssetSaveResult,
    type AddAssetSubmission,
    type Asset,
    type AssetAttachment,
    type AssetKind,
    type AssetLifecycleFilter,
    type CustomAssetType,
    type CustomFieldDefinition,
    type SearchLifecycleFilter,
    type SearchMode,
    type SearchResult,
    type SelectedAttachment,
    type UpdateAssetDraft,
    type WorkspaceData,
    type WorkspaceMode
  } from '$lib/domain/inventory';
  import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
  import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import AddAssetTray from './AddAssetTray.svelte';
  import AssetDetail from './AssetDetail.svelte';
  import HomeboxImportPanel from './HomeboxImportPanel.svelte';
  import HomeWorkspace from './HomeWorkspace.svelte';
  import InventorySettings from './InventorySettings.svelte';
  import LocationView from './LocationView.svelte';
  import MobileNav from './MobileNav.svelte';
  import SearchPanel from './SearchPanel.svelte';
  import SideNav from './SideNav.svelte';
  import TopHeader from './TopHeader.svelte';
  import * as Alert from '$lib/components/ui/alert/index.js';
  import * as Button from '$lib/components/ui/button/index.js';

  let {
    repository,
    initialData,
    onSignOut
  }: {
    repository: InventoryRepository & InventoryAccessRepository & InventoryAuditRepository & InventoryCustomizationRepository;
    initialData: WorkspaceData;
    onSignOut: () => void;
  } = $props();

  // svelte-ignore state_referenced_locally -- initial route data seeds local workspace state.
  const startingData = initialData;
  let data = $state(startingData);
  let mode = $state<WorkspaceMode>(startingData.context.inventories.length > 0 ? 'home' : 'settings');
  let selectedLocationId = $state<string | null>(null);
  let selectedAssetId = $state<string | null>(null);
  let addOpen = $state(false);
  let addKind = $state<AssetKind>('item');
  let addParentAssetId = $state<string | null>(null);
  let assetAction = $state<AssetRouteAction>(null);
  let attachmentId = $state<string | null>(null);
  let attachmentAction = $state<WorkspaceRouteState['attachmentAction']>(null);
  let busy = $state(false);
  let message = $state('');
  let error = $state('');
  let searchQuery = $state('');
  let searchLifecycleState = $state<SearchLifecycleFilter>('active');
  let searchMode = $state<SearchMode>('fuzzy');
  let settingsSection = $state<SettingsSection>('overview');
  let invitationStatus = $state<WorkspaceRouteState['invitationStatus']>('all');
  let accessInvitationAction = $state<WorkspaceRouteState['accessInvitationAction']>(null);
  let accessInvitationId = $state<string | null>(null);
  let auditScope = $state<WorkspaceRouteState['auditScope']>('inventory');
  let customizationAction = $state<WorkspaceRouteState['customizationAction']>(null);
  let customAssetTypeId = $state<string | null>(null);
  let customFieldDefinitionId = $state<string | null>(null);
  let importSourceType = $state<WorkspaceRouteState['importSourceType']>('legacy_homebox');
  let searchResults = $state<SearchResult[]>([]);
  let searchSubmitted = $state(false);
  let searchError = $state('');
  let loadedAssetDetail = $state<Asset | null>(null);
  let selectedAssetAttachments = $state<AssetAttachment[]>([]);
  let assetDetailRequestId = 0;
  let applyingRoute = false;
  let routeUnavailable = $state('');

  let selectedInventory = $derived(data.context.inventories.find((inventory) => inventory.id === data.context.selectedInventoryId) ?? null);
  let selectedTenant = $derived(data.context.tenants.find((tenant) => tenant.id === data.context.selectedTenantId) ?? null);
  let assets = $derived(labelAssets(data.assets, data.context.customAssetTypes));
  let selectedLocation = $derived(assets.find((asset) => asset.id === selectedLocationId) ?? null);
  let detailAssets = $derived(
    loadedAssetDetail && !assets.some((asset) => asset.id === loadedAssetDetail?.id)
      ? [labelAsset(loadedAssetDetail, data.context.customAssetTypes), ...assets]
      : assets
  );
  let selectedAsset = $derived(
    selectedAssetId
      ? loadedAssetDetail?.id === selectedAssetId
        ? labelAsset(loadedAssetDetail, data.context.customAssetTypes)
        : assets.find((asset) => asset.id === selectedAssetId) ?? null
      : null
  );
  let searchSuggestions = $derived(buildSearchSuggestions(assets, searchQuery));
  let createAssetAllowed = $derived(canCreateAsset(selectedInventory));
  let editAssetAllowed = $derived(canEditAsset(selectedInventory));
  let canCreateStarter = $derived(!data.context.selectedTenantId || canCreateInventory(selectedTenant));
  let userLabel = $derived(data.context.principal.email ?? data.context.principal.id);

  onMount(() => {
    void applyRoute(currentWorkspaceRoute());
    const onPopState = () => {
      void applyRoute(currentWorkspaceRoute());
    };
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  });

  async function selectInventory(tenantId: string, inventoryId: string): Promise<void> {
    await run(async () => {
      data = await repository.selectInventory(tenantId, inventoryId);
      routeUnavailable = '';
      invalidateAssetDetailLoad();
      resetSearchState();
      mode = 'home';
      selectedLocationId = null;
      selectedAssetId = null;
      loadedAssetDetail = null;
      selectedAssetAttachments = [];
      attachmentId = null;
      attachmentAction = null;
      if (!applyingRoute) {
        replaceRoute({ mode: 'home', tenantId, inventoryId, lifecycleState: data.context.assetLifecycleState });
      }
    });
  }

  async function selectTenant(tenantId: string): Promise<void> {
    await run(async () => {
      data = await repository.selectTenant(tenantId);
      routeUnavailable = '';
      invalidateAssetDetailLoad();
      resetSearchState();
      mode = data.context.inventories.length > 0 ? 'home' : 'settings';
      selectedLocationId = null;
      selectedAssetId = null;
      loadedAssetDetail = null;
      selectedAssetAttachments = [];
      attachmentId = null;
      attachmentAction = null;
      if (!applyingRoute) {
        replaceRoute({
          mode: 'home',
          tenantId,
          inventoryId: data.context.selectedInventoryId,
          lifecycleState: data.context.assetLifecycleState
        });
      }
    });
  }

  async function createStarterInventory(): Promise<void> {
    await run(async () => {
      data = data.context.selectedTenantId
        ? await repository.createInventory(data.context.selectedTenantId, 'Household')
        : await repository.createTenantWithInventory({ tenantName: 'Home', inventoryName: 'Household' });
      routeUnavailable = '';
      mode = 'home';
      message = 'Created Household.';
      replaceRoute({ mode: 'home', tenantId: data.context.selectedTenantId, inventoryId: data.context.selectedInventoryId });
    });
  }

  async function createAsset(draft: AddAssetSubmission): Promise<AddAssetSaveResult> {
    if (!selectedInventory) {
      error = 'Create an inventory before adding assets.';
      return { saved: false };
    }
    if (!createAssetAllowed) {
      error = 'You do not have permission to add assets in this inventory.';
      return { saved: false };
    }
    busy = true;
    error = '';
    message = '';
    try {
      const result = await createAssetWorkflow(repository, data, selectedInventory, draft);
      data = result.data;
      if (result.message) {
        message = result.message;
      }
      if (result.error) {
        error = result.error;
      }
      if (result.closeAdd) {
        addOpen = false;
      }
      if (result.mode) {
        mode = result.mode;
      }
      if (result.clearDetail) {
        selectedLocationId = null;
        selectedAssetId = null;
        loadedAssetDetail = null;
        selectedAssetAttachments = [];
        attachmentId = null;
        attachmentAction = null;
      }
      if (result.selectedAsset) {
        selectedLocationId = null;
        selectedAssetId = result.selectedAsset.id;
        loadedAssetDetail = result.selectedAsset;
        selectedAssetAttachments = [];
        attachmentId = null;
        attachmentAction = null;
      }
      if (result.route) {
        replaceRoute(result.route);
      }
      return result.saveResult;
    } finally {
      busy = false;
    }
  }

  async function updateAsset(draft: UpdateAssetDraft): Promise<void> {
    if (!selectedAsset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    busy = true;
    error = '';
    message = '';
    try {
      const asset = await repository.updateAsset(
        selectedAsset.tenantId,
        selectedAsset.inventoryId,
        selectedAsset.id,
        draft
      );
      data = replaceWorkspaceAsset(data, asset);
      loadedAssetDetail = asset;
      message = `Saved ${asset.title}.`;
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Action failed.';
      throw new Error(error);
    } finally {
      busy = false;
    }
  }

  async function search(): Promise<void> {
    busy = true;
    error = '';
    message = '';
    searchError = '';
    searchResults = [];
    searchSubmitted = true;
    try {
      const result = await executeWorkspaceSearch({
        repository,
        tenantId: data.context.selectedTenantId,
        inventoryId: data.context.selectedInventoryId,
        query: searchQuery,
        lifecycleState: searchLifecycleState,
        mode: searchMode
      });
      searchQuery = result.query;
      searchResults = result.results;
      searchSubmitted = result.submitted;
      searchError = result.error;
      error = result.error;
      mode = 'search';
      if (!applyingRoute) {
        replaceRoute({
          mode: 'search',
          tenantId: data.context.selectedTenantId,
          inventoryId: data.context.selectedInventoryId,
          searchQuery: result.query,
          searchLifecycleState,
          searchMode
        });
      }
      if (!result.query) {
        return;
      }
    } finally {
      busy = false;
    }
  }

  async function selectAssetLifecycle(lifecycleState: AssetLifecycleFilter): Promise<void> {
    if (!selectedInventory) {
      return;
    }
    await run(async () => {
      data = await repository.selectAssetLifecycle(data.context.selectedTenantId, selectedInventory.id, lifecycleState);
      invalidateAssetDetailLoad();
      mode = 'home';
      selectedLocationId = null;
      selectedAssetId = null;
      loadedAssetDetail = null;
      selectedAssetAttachments = [];
      if (!applyingRoute) {
        replaceRoute({ mode: 'home', tenantId: data.context.selectedTenantId, inventoryId: selectedInventory.id, lifecycleState });
      }
    });
  }

  async function archiveSelectedAsset(): Promise<void> {
    const asset = selectedAsset;
    if (!asset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.archiveAsset(asset.tenantId, asset.inventoryId, asset.id);
      await refreshSelectedAssetLifecycle();
      closeDetailToHome();
      message = `Archived ${asset.title}.`;
    });
  }

  async function restoreSelectedAsset(): Promise<void> {
    const asset = selectedAsset;
    if (!asset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.restoreAsset(asset.tenantId, asset.inventoryId, asset.id);
      await refreshSelectedAssetLifecycle();
      closeDetailToHome();
      message = `Restored ${asset.title}.`;
    });
  }

  async function deleteSelectedAsset(): Promise<void> {
    const asset = selectedAsset;
    if (!asset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.deleteAsset(asset.tenantId, asset.inventoryId, asset.id);
      await refreshSelectedAssetLifecycle();
      closeDetailToHome();
      message = `Deleted ${asset.title}.`;
    });
  }

  async function archiveSelectedAttachment(attachment: AssetAttachment): Promise<void> {
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.archiveAssetAttachment(attachment.tenantId, attachment.inventoryId, attachment.assetId, attachment.id);
      await refreshSelectedAttachments(attachment.tenantId, attachment.inventoryId, attachment.assetId);
      message = `Archived ${attachment.fileName}.`;
    });
  }

  async function deleteSelectedAttachment(attachment: AssetAttachment): Promise<void> {
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.deleteAssetAttachment(attachment.tenantId, attachment.inventoryId, attachment.assetId, attachment.id);
      await refreshSelectedAttachments(attachment.tenantId, attachment.inventoryId, attachment.assetId);
      message = `Deleted ${attachment.fileName}.`;
    });
  }

  async function uploadSelectedAttachment(attachment: SelectedAttachment): Promise<void> {
    const asset = selectedAsset;
    if (!asset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.uploadAssetAttachment(asset.tenantId, asset.inventoryId, asset.id, attachment);
      await refreshSelectedAttachments(asset.tenantId, asset.inventoryId, asset.id);
      message = `Uploaded ${attachment.name}.`;
    });
  }

  async function run(task: () => Promise<void>): Promise<void> {
    busy = true;
    error = '';
    message = '';
    try {
      await task();
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Action failed.';
    } finally {
      busy = false;
    }
  }

  function openLocation(asset: Asset): void {
    navigateTo({ mode: 'location', tenantId: asset.tenantId, inventoryId: asset.inventoryId, locationId: asset.id });
  }

  async function openAsset(asset: Asset): Promise<void> {
    const returnLocationId = mode === 'location' ? selectedLocationId : null;
    await applyRoute(
      pushWorkspaceRoute(
        { mode: 'asset', tenantId: asset.tenantId, inventoryId: asset.inventoryId, assetId: asset.id },
        data.context.selectedTenantId || null,
        data.context.selectedInventoryId || null
      )
    );
    if (returnLocationId && selectedAssetId === asset.id) {
      selectedLocationId = returnLocationId;
    }
  }

  function openLocationEdit(asset: Asset): void {
    navigateTo({
      mode: 'asset',
      tenantId: asset.tenantId,
      inventoryId: asset.inventoryId,
      locationId: asset.id,
      assetId: asset.id,
      action: 'edit',
      assetAction: 'edit'
    });
  }

  function openSearchAsset(asset: Asset): void {
    if (asset.kind === 'location') {
      openLocation(asset);
      return;
    }
    void openAsset(asset);
  }

  function resetSearchState(): void {
    searchQuery = '';
    searchResults = [];
    searchSubmitted = false;
    searchError = '';
    searchLifecycleState = 'active';
    searchMode = 'fuzzy';
  }

  async function applyRoute(route: WorkspaceRouteState): Promise<void> {
    applyingRoute = true;
    try {
      const shouldCanonicalizeAlias = shouldCanonicalizeWorkspaceAlias(route);
      routeUnavailable = '';
      addOpen = route.action === 'add';
      addKind = route.addKind ?? 'item';
      addParentAssetId = null;
      assetAction = route.assetAction;
      attachmentId = route.attachmentId;
      attachmentAction = route.attachmentAction;
      searchQuery = route.searchQuery;
      searchLifecycleState = route.searchLifecycleState;
      searchMode = route.searchMode;
      settingsSection = route.settingsSection;
      invitationStatus = route.invitationStatus;
      accessInvitationAction = route.accessInvitationAction;
      accessInvitationId = route.accessInvitationId;
      auditScope = route.auditScope;
      customizationAction = route.customizationAction;
      customAssetTypeId = route.customAssetTypeId;
      customFieldDefinitionId = route.customFieldDefinitionId;
      importSourceType = route.importSourceType;

      if (route.tenantId && route.tenantId !== data.context.selectedTenantId) {
        const tenantId = findRouteTenant(data, route);
        if (tenantId) {
          await selectTenant(tenantId);
        } else {
          showUnavailableRoute('That tenant is not available to this account.');
          return;
        }
      }

      if (route.inventoryId && route.inventoryId !== data.context.selectedInventoryId) {
        const inventory = findRouteInventory(data, route);
        if (inventory) {
          await selectInventory(inventory.tenantId, inventory.id);
        } else {
          showUnavailableRoute('That inventory is not available in the current workspace.');
          return;
        }
      }

      if (route.mode !== 'search' && route.lifecycleState !== data.context.assetLifecycleState && selectedInventory) {
        await selectAssetLifecycle(route.lifecycleState);
      }
      if (route.action === 'add' && !createAssetAllowed) {
        showUnavailableRoute('You do not have permission to add assets in this inventory.');
        return;
      }
      addParentAssetId = validAddParentId(route.addParentAssetId);
      if (route.action === 'add' && route.addParentAssetId && !addParentAssetId) {
        route = { ...route, addParentAssetId: null };
        replaceRoute({
          action: 'add',
          addKind: route.addKind,
          addParentAssetId: null,
          tenantId: data.context.selectedTenantId,
          inventoryId: data.context.selectedInventoryId
        });
      }

      if (route.mode === 'locations') {
        invalidateAssetDetailLoad();
        selectedLocationId = null;
        selectedAssetId = null;
        loadedAssetDetail = null;
        selectedAssetAttachments = [];
        attachmentId = null;
        attachmentAction = null;
        mode = 'locations';
        canonicalizeRouteAlias(route, shouldCanonicalizeAlias);
        return;
      }

      if (route.mode === 'location' && route.locationId) {
        const location = assets.find((candidate) => candidate.id === route.locationId && candidate.kind === 'location');
        if (location) {
          invalidateAssetDetailLoad();
          selectedLocationId = location.id;
          selectedAssetId = null;
          loadedAssetDetail = null;
          selectedAssetAttachments = [];
          attachmentId = null;
          attachmentAction = null;
          mode = 'location';
          canonicalizeRouteAlias(route, shouldCanonicalizeAlias);
        } else {
          closeDetailToHome();
        }
        return;
      }

      if (route.mode === 'asset' && route.assetId && data.context.selectedInventoryId) {
        const knownAsset = assets.find((candidate) => candidate.id === route.assetId);
        let loaded = false;
        if (knownAsset) {
          loaded = await loadAssetDetail(knownAsset.tenantId, knownAsset.inventoryId, knownAsset.id);
        } else if (data.context.selectedTenantId) {
          loaded = await loadAssetDetail(data.context.selectedTenantId, data.context.selectedInventoryId, route.assetId);
        }
        if (!loaded) {
          showUnavailableRoute('That asset is not available in this inventory.');
          return;
        }
        attachmentId = route.attachmentId;
        attachmentAction = route.attachmentAction;
        if (route.locationId) {
          if (loadedAssetDetail?.kind !== 'location') {
            showUnavailableRoute('That location is not available in this inventory.');
            return;
          }
          selectedLocationId = route.locationId;
        }
        if (!assetRouteActionIsAvailable(route.assetAction, selectedInventory, loadedAssetDetail)) {
          assetAction = null;
          attachmentId = null;
          attachmentAction = null;
          replaceRoute({
            mode: 'asset',
            tenantId: data.context.selectedTenantId,
            inventoryId: data.context.selectedInventoryId,
            assetId: route.assetId
          });
          return;
        }
        if (route.attachmentAction === 'delete') {
          if (!canEditAsset(selectedInventory)) {
            attachmentId = null;
            attachmentAction = null;
            replaceRoute({
              mode: 'asset',
              tenantId: data.context.selectedTenantId,
              inventoryId: data.context.selectedInventoryId,
              assetId: route.assetId
            });
            return;
          }
          const routedAttachment = selectedAssetAttachments.find((attachment) => attachment.id === route.attachmentId);
          if (!routedAttachment || routedAttachment.assetId !== route.assetId) {
            attachmentId = null;
            attachmentAction = null;
            replaceRoute({
              mode: 'asset',
              tenantId: data.context.selectedTenantId,
              inventoryId: data.context.selectedInventoryId,
              assetId: route.assetId
            });
            return;
          }
        }
        canonicalizeRouteAlias(route, shouldCanonicalizeAlias);
        return;
      }

      if (route.mode === 'search') {
        mode = 'search';
        selectedLocationId = null;
        selectedAssetId = null;
        loadedAssetDetail = null;
        selectedAssetAttachments = [];
        if (route.searchQuery.trim()) {
          await search();
        } else {
          searchResults = [];
          searchSubmitted = false;
          searchError = '';
        }
        canonicalizeRouteAlias(route, shouldCanonicalizeAlias);
        return;
      }

      if (route.mode === 'settings' || route.mode === 'import') {
        mode = route.mode;
        settingsSection = route.settingsSection;
        selectedLocationId = null;
        selectedAssetId = null;
        loadedAssetDetail = null;
        selectedAssetAttachments = [];
        canonicalizeRouteAlias(route, shouldCanonicalizeAlias);
        return;
      }

      closeDetailToHome();
      canonicalizeRouteAlias(route, shouldCanonicalizeAlias);
    } finally {
      applyingRoute = false;
    }
  }

  function navigateTo(route: Partial<WorkspaceRouteState>): void {
    void applyRoute(pushWorkspaceRoute(route, data.context.selectedTenantId || null, data.context.selectedInventoryId || null));
  }

  function homeRoute(): Partial<WorkspaceRouteState> {
    return {
      mode: 'home',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      lifecycleState: data.context.assetLifecycleState
    };
  }

  function homeHref(): string {
    return workspaceRouteHref(homeRoute(), data.context.selectedTenantId || null, data.context.selectedInventoryId || null);
  }

  function openHome(event: MouseEvent): void {
    if (!shouldHandleInApp(event)) {
      return;
    }
    event.preventDefault();
    navigateTo(homeRoute());
  }

  function showUnavailableRoute(messageText: string): void {
    invalidateAssetDetailLoad();
    routeUnavailable = messageText;
    addOpen = false;
    assetAction = null;
    attachmentId = null;
    attachmentAction = null;
    mode = 'home';
    selectedLocationId = null;
    selectedAssetId = null;
    loadedAssetDetail = null;
    selectedAssetAttachments = [];
    searchResults = [];
    searchSubmitted = false;
  }

  function navigateMode(nextMode: WorkspaceMode): void {
    navigateTo({
      mode: nextMode,
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: nextMode === 'settings' ? settingsSection : 'overview',
      importSourceType: 'legacy_homebox'
    });
  }

  function openSettingsSection(section: SettingsSection): void {
    settingsSection = section;
    navigateTo({
      mode: 'settings',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: section,
      invitationStatus: section === 'access' ? invitationStatus : 'all',
      auditScope: section === 'activity' ? auditScope : 'inventory'
    });
  }

  function openInvitationStatusFilter(nextInvitationStatus: WorkspaceRouteState['invitationStatus']): void {
    invitationStatus = nextInvitationStatus;
    accessInvitationAction = null;
    accessInvitationId = null;
    navigateTo({
      mode: 'settings',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: 'access',
      invitationStatus: nextInvitationStatus,
      accessInvitationAction: null,
      accessInvitationId: null
    });
  }

  function openAccessInvitationAction(action: WorkspaceRouteState['accessInvitationAction'], invitationId: string): void {
    accessInvitationAction = action;
    accessInvitationId = invitationId;
    navigateTo({
      mode: 'settings',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: 'access',
      invitationStatus,
      accessInvitationAction: action,
      accessInvitationId: invitationId
    });
  }

  function closeAccessInvitationAction(): void {
    accessInvitationAction = null;
    accessInvitationId = null;
    replaceRoute({
      mode: 'settings',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: 'access',
      invitationStatus
    });
  }

  function openAuditScopeFilter(nextAuditScope: WorkspaceRouteState['auditScope']): void {
    auditScope = nextAuditScope;
    navigateTo({
      mode: 'settings',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: 'activity',
      auditScope: nextAuditScope
    });
  }

  function openCustomizationArchive(action: WorkspaceRouteState['customizationAction'], id: string): void {
    customizationAction = action;
    customAssetTypeId = action === 'archive_asset_type' ? id : null;
    customFieldDefinitionId = action === 'archive_field_definition' ? id : null;
    navigateTo({
      mode: 'settings',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: 'fields',
      customizationAction: action,
      customAssetTypeId: action === 'archive_asset_type' ? id : null,
      customFieldDefinitionId: action === 'archive_field_definition' ? id : null
    });
  }

  function closeCustomizationArchive(): void {
    customizationAction = null;
    customAssetTypeId = null;
    customFieldDefinitionId = null;
    replaceRoute({
      mode: 'settings',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: 'fields'
    });
  }

  function openImportSource(nextImportSourceType: WorkspaceRouteState['importSourceType']): void {
    importSourceType = nextImportSourceType;
    navigateTo({
      mode: 'import',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      importSourceType: nextImportSourceType
    });
  }

  function openAdd(kind: AssetKind = 'item', parentAssetId: string | null = null): void {
    navigateTo({
      action: 'add',
      addKind: kind,
      addParentAssetId: parentAssetId,
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId
    });
  }

  function closeAdd(): void {
    addOpen = false;
    replaceRoute({
      mode,
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      lifecycleState: data.context.assetLifecycleState
    });
  }

  function addCloseHref(): string {
    return workspaceRouteHref(
      {
        mode,
        tenantId: data.context.selectedTenantId,
        inventoryId: data.context.selectedInventoryId,
        lifecycleState: data.context.assetLifecycleState
      },
      data.context.selectedTenantId || null,
      data.context.selectedInventoryId || null
    );
  }

  function validAddParentId(parentAssetId: string | null): string | null {
    if (!parentAssetId) {
      return null;
    }
    return parentTargets(assets).some((target) => target.id === parentAssetId) ? parentAssetId : null;
  }

  function openAssetActionRoute(action: Exclude<AssetRouteAction, null>): void {
    if (selectedAsset) {
      const isLocationEdit = selectedAsset.kind === 'location' && action === 'edit';
      navigateTo({
        mode: 'asset',
        tenantId: selectedAsset.tenantId,
        inventoryId: selectedAsset.inventoryId,
        locationId: isLocationEdit ? selectedAsset.id : null,
        assetId: selectedAsset.id,
        assetAction: action,
        action: action === 'edit' ? 'edit' : null
      });
    }
  }

  function openAttachmentDeleteRoute(nextAttachmentId: string): void {
    if (selectedAsset) {
      navigateTo({
        mode: 'asset',
        tenantId: selectedAsset.tenantId,
        inventoryId: selectedAsset.inventoryId,
        assetId: selectedAsset.id,
        attachmentId: nextAttachmentId,
        attachmentAction: 'delete'
      });
    }
  }

  function closeAssetActionRoute(): void {
    assetAction = null;
    if (selectedAsset) {
      const closingAsset = selectedAsset;
      if (closingAsset.kind === 'location') {
        mode = 'location';
        selectedLocationId = closingAsset.id;
        selectedAssetId = null;
        loadedAssetDetail = null;
        selectedAssetAttachments = [];
      }
      replaceRoute(
        closingAsset.kind === 'location'
          ? {
              mode: 'location',
              tenantId: closingAsset.tenantId,
              inventoryId: closingAsset.inventoryId,
              locationId: closingAsset.id
            }
          : { mode: 'asset', tenantId: closingAsset.tenantId, inventoryId: closingAsset.inventoryId, assetId: closingAsset.id }
      );
    }
  }

  function closeAttachmentDeleteRoute(): void {
    attachmentId = null;
    attachmentAction = null;
    if (selectedAsset) {
      replaceRoute(
        selectedAsset.kind === 'location'
          ? {
              mode: 'location',
              tenantId: selectedAsset.tenantId,
              inventoryId: selectedAsset.inventoryId,
              locationId: selectedAsset.id
            }
          : { mode: 'asset', tenantId: selectedAsset.tenantId, inventoryId: selectedAsset.inventoryId, assetId: selectedAsset.id }
      );
    }
  }

  function replaceRoute(route: Partial<WorkspaceRouteState>): void {
    if (typeof window === 'undefined') {
      return;
    }
    replaceWorkspaceRoute(route, data.context.selectedTenantId || null, data.context.selectedInventoryId || null);
  }

  function canonicalizeRouteAlias(route: WorkspaceRouteState, shouldCanonicalizeAlias: boolean): void {
    if (!shouldCanonicalizeAlias) {
      return;
    }
    replaceCanonicalWorkspaceAlias(route, data.context.selectedTenantId || null, data.context.selectedInventoryId || null);
  }

  async function loadAssetDetail(tenantId: string, inventoryId: string, assetId: string): Promise<boolean> {
    const requestId = ++assetDetailRequestId;
    busy = true;
    error = '';
    message = '';
    selectedLocationId = null;
    selectedAssetId = null;
    loadedAssetDetail = null;
    selectedAssetAttachments = [];
    attachmentId = null;
    attachmentAction = null;
    try {
      const result = await loadWorkspaceAssetDetail(repository, tenantId, inventoryId, assetId);
      if (requestId !== assetDetailRequestId) {
        return false;
      }
      if (!result.loaded || !result.asset) {
        error = result.error;
        return false;
      }
      selectedAssetAttachments = result.attachments;
      data = replaceWorkspaceAsset(data, result.asset);
      loadedAssetDetail = result.asset;
      selectedAssetId = result.asset.id;
      mode = 'asset';
      return true;
    } finally {
      busy = false;
    }
  }

  async function refreshSelectedAssetLifecycle(): Promise<void> {
    if (!selectedInventory) {
      return;
    }
    data = await repository.selectAssetLifecycle(
      data.context.selectedTenantId,
      selectedInventory.id,
      data.context.assetLifecycleState
    );
  }

  async function refreshAfterImport(): Promise<void> {
    if (!selectedInventory) {
      return;
    }
    data = await repository.selectAssetLifecycle(data.context.selectedTenantId, selectedInventory.id, 'active');
    invalidateAssetDetailLoad();
    resetSearchState();
    mode = 'home';
    selectedLocationId = null;
    selectedAssetId = null;
    loadedAssetDetail = null;
    selectedAssetAttachments = [];
    attachmentId = null;
    attachmentAction = null;
    message = 'Import applied.';
    replaceRoute({
      mode: 'home',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      lifecycleState: data.context.assetLifecycleState
    });
  }

  async function refreshSelectedAttachments(tenantId: string, inventoryId: string, assetId: string): Promise<void> {
    selectedAssetAttachments = await refreshWorkspaceAssetAttachments(repository, { tenantId, inventoryId, assetId });
  }

  function closeDetailToHome(): void {
    invalidateAssetDetailLoad();
    mode = 'home';
    selectedLocationId = null;
    selectedAssetId = null;
    loadedAssetDetail = null;
    selectedAssetAttachments = [];
    if (!applyingRoute) {
      replaceRoute({
        mode: 'home',
        tenantId: data.context.selectedTenantId,
        inventoryId: data.context.selectedInventoryId,
        lifecycleState: data.context.assetLifecycleState
      });
    }
  }

  function closeLocationToLocations(): void {
    navigateMode('locations');
  }

  function closeDetailToPrevious(): void {
    invalidateAssetDetailLoad();
    mode = selectedLocationId ? 'location' : 'home';
    selectedAssetId = null;
    loadedAssetDetail = null;
    selectedAssetAttachments = [];
    if (!applyingRoute) {
      replaceRoute(
        selectedLocationId
          ? {
              mode: 'location',
              tenantId: data.context.selectedTenantId,
              inventoryId: data.context.selectedInventoryId,
              locationId: selectedLocationId
            }
          : {
              mode: 'home',
              tenantId: data.context.selectedTenantId,
              inventoryId: data.context.selectedInventoryId,
              lifecycleState: data.context.assetLifecycleState
            }
      );
    }
  }

  function assetDetailBackHref(): string {
    return workspaceRouteHref(
      selectedLocationId
        ? {
            mode: 'location',
            tenantId: data.context.selectedTenantId,
            inventoryId: data.context.selectedInventoryId,
            locationId: selectedLocationId
          }
        : {
            mode: 'home',
            tenantId: data.context.selectedTenantId,
            inventoryId: data.context.selectedInventoryId,
            lifecycleState: data.context.assetLifecycleState
          },
      data.context.selectedTenantId || null,
      data.context.selectedInventoryId || null
    );
  }

  function invalidateAssetDetailLoad(): void {
    assetDetailRequestId += 1;
  }

  function updateCustomizationContext(assetTypes: CustomAssetType[], fieldDefinitions: CustomFieldDefinition[]): void {
    data = {
      ...data,
      context: {
        ...data.context,
        customAssetTypes: assetTypes,
        customFieldDefinitions: fieldDefinitions
      }
    };
  }

  function labelAssets(items: Asset[], customAssetTypes: CustomAssetType[]): Asset[] {
    return items.map((asset) => labelAsset(asset, customAssetTypes));
  }

  function labelAsset(asset: Asset, customAssetTypes: CustomAssetType[]): Asset {
    if (asset.customAssetTypeLabel || !asset.customAssetTypeId) {
      return asset;
    }
    return {
      ...asset,
      customAssetTypeLabel: customAssetTypes.find((assetType) => assetType.id === asset.customAssetTypeId)?.displayName
    };
  }

  function shouldHandleInApp(event: MouseEvent): boolean {
    return event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey && !event.defaultPrevented;
  }

</script>

<div class="product-shell">
  <SideNav
    tenants={data.context.tenants}
    inventories={data.context.inventories}
    selectedTenantId={data.context.selectedTenantId}
    selectedInventoryId={data.context.selectedInventoryId}
    {mode}
    {settingsSection}
    {userLabel}
    onSelectTenant={(tenantId) => { void selectTenant(tenantId); }}
    onSelectInventory={(tenantId, inventoryId) => { void selectInventory(tenantId, inventoryId); }}
    onModeChange={navigateMode}
    {onSignOut}
  />

  <div class="workspace-column">
    <TopHeader
      tenants={data.context.tenants}
      inventories={data.context.inventories}
      selectedTenantId={data.context.selectedTenantId}
      inventory={selectedInventory}
      suggestions={searchSuggestions}
      bind:query={searchQuery}
      canCreateAsset={createAssetAllowed && data.context.assetLifecycleState === 'active'}
      onSelectTenant={(tenantId) => { void selectTenant(tenantId); }}
      onSelectInventory={(tenantId, inventoryId) => { void selectInventory(tenantId, inventoryId); }}
      onSearch={() => { void search(); }}
      onOpenAsset={openSearchAsset}
      onOpenAdd={openAdd}
    />

    {#if routeUnavailable}
      <section class="workspace-main">
        <div class="empty-state spacious" role="alert">
          <h1>Workspace unavailable</h1>
          <p>{routeUnavailable}</p>
          <Button.Root href={homeHref()} onclick={openHome}>Go home</Button.Root>
        </div>
      </section>
    {:else if data.context.inventories.length === 0}
      <section class="workspace-main">
        <div class="empty-state spacious">
          <h1>No inventory yet</h1>
          {#if canCreateStarter}
            <p>{data.context.selectedTenantId ? 'Create the first inventory for this tenant.' : 'Create your first tenant and inventory.'}</p>
            <Button.Root onclick={() => { void createStarterInventory(); }}>Create Household</Button.Root>
          {:else}
            <p>You can view this tenant, but you cannot create inventories in it.</p>
          {/if}
        </div>
      </section>
    {:else if mode === 'location' && selectedLocation}
      <LocationView
        location={selectedLocation}
        assets={containedAssets(assets, selectedLocation.id)}
        canEdit={editAssetAllowed}
        canCreateAsset={createAssetAllowed}
        onBack={closeLocationToLocations}
        onOpenLocation={openLocation}
        onEditLocation={openLocationEdit}
        onOpenAsset={openAsset}
        onOpenAdd={openAdd}
      />
    {:else if mode === 'asset' && selectedAsset}
      <AssetDetail
        asset={withTrail(selectedAsset, detailAssets)}
        canEdit={editAssetAllowed}
        parentTargets={moveParentTargets(detailAssets, selectedAsset.id)}
        customFieldDefinitions={data.context.customFieldDefinitions}
        saving={busy}
        attachments={selectedAssetAttachments}
        mediaPolicy={data.context.mediaUploadPolicy}
        action={assetAction}
        {attachmentId}
        {attachmentAction}
        backHref={assetDetailBackHref()}
        onBack={closeDetailToPrevious}
        onActionOpen={openAssetActionRoute}
        onActionClose={closeAssetActionRoute}
        onSave={updateAsset}
        onArchive={archiveSelectedAsset}
        onRestore={restoreSelectedAsset}
        onDelete={deleteSelectedAsset}
        onUploadAttachment={uploadSelectedAttachment}
        onArchiveAttachment={archiveSelectedAttachment}
        onAttachmentDeleteOpen={openAttachmentDeleteRoute}
        onAttachmentDeleteClose={closeAttachmentDeleteRoute}
        onDeleteAttachment={deleteSelectedAttachment}
      />
    {:else if mode === 'search'}
      <SearchPanel
        tenantId={data.context.selectedTenantId}
        inventoryId={data.context.selectedInventoryId}
        bind:query={searchQuery}
        bind:lifecycleState={searchLifecycleState}
        bind:searchMode={searchMode}
        results={searchResults}
        suggestions={searchSuggestions}
        submitted={searchSubmitted}
        error={searchError}
        {busy}
        onSearch={() => { void search(); }}
        onOpenAsset={openSearchAsset}
      />
    {:else if mode === 'import'}
      <HomeboxImportPanel
        tenantId={data.context.selectedTenantId}
        inventory={selectedInventory}
        {repository}
        sourceType={importSourceType}
        onSourceChange={openImportSource}
        onImported={refreshAfterImport}
      />
    {:else if mode === 'settings'}
      <InventorySettings
        tenant={selectedTenant}
        inventory={selectedInventory}
        inventoryCount={data.context.inventories.length}
        accessRepository={repository}
        auditRepository={repository}
        customizationRepository={repository}
        customAssetTypes={data.context.customAssetTypes}
        customFieldDefinitions={data.context.customFieldDefinitions}
        section={settingsSection}
        {invitationStatus}
        {accessInvitationAction}
        {accessInvitationId}
        {auditScope}
        {customizationAction}
        {customAssetTypeId}
        {customFieldDefinitionId}
        onSectionChange={openSettingsSection}
        onInvitationStatusChange={openInvitationStatusFilter}
        onAccessInvitationActionOpen={openAccessInvitationAction}
        onAccessInvitationActionClose={closeAccessInvitationAction}
        onAuditScopeChange={openAuditScopeFilter}
        onCustomizationArchiveOpen={openCustomizationArchive}
        onCustomizationArchiveClose={closeCustomizationArchive}
        onCustomizationChange={updateCustomizationContext}
      />
    {:else}
      <HomeWorkspace
        tenantId={data.context.selectedTenantId}
        inventoryId={data.context.selectedInventoryId}
        lifecycleState={data.context.assetLifecycleState}
        browseMode={mode === 'locations' ? 'locations' : 'home'}
        locations={topLevelLocations(assets)}
        recentAssets={recentlyAddedAssets(assets)}
        archivedAssets={assets}
        canCreateAsset={createAssetAllowed}
        onOpenLocation={openLocation}
        onOpenAsset={openAsset}
        onOpenAdd={() => openAdd('location')}
        onSelectLifecycle={(lifecycleState) => { void selectAssetLifecycle(lifecycleState); }}
      />
    {/if}
  </div>

  <MobileNav
    {mode}
    selectedTenantId={data.context.selectedTenantId}
    selectedInventoryId={data.context.selectedInventoryId}
    {settingsSection}
    canCreateAsset={createAssetAllowed && data.context.assetLifecycleState === 'active'}
    onModeChange={navigateMode}
    onOpenAdd={() => openAdd('item')}
  />

  <AddAssetTray
    open={addOpen && createAssetAllowed}
    initialKind={addKind}
    initialParentAssetId={addParentAssetId}
    closeHref={addCloseHref()}
    parentTargets={parentTargets(assets)}
    mediaPolicy={data.context.mediaUploadPolicy}
    customAssetTypes={data.context.customAssetTypes}
    customFieldDefinitions={data.context.customFieldDefinitions}
    saving={busy}
    onClose={closeAdd}
    onSave={createAsset}
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
</div>
