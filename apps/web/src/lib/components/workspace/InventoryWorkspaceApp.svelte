<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { isAuthenticationRequiredError } from '$lib/application/authenticationRequired';
  import { onMount } from 'svelte';
  import {
    detailAssetList,
    labelAssets,
    parentTargets,
    selectedAssetForDetail
  } from '$lib/application/workspace';
  import { resolveWorkspaceAddRoute } from '$lib/application/workspaceAddRoute';
  import {
    assetDetailBackHref as workspaceAssetDetailBackHref,
    assetDetailBackRoute,
    inventoryHomeNormalizationHref,
    inventoryHomeNormalizationRoute,
    settingsOverviewHref,
    settingsOverviewRoute,
    workspaceAddCloseHref,
    workspaceAddCloseRoute,
    workspaceHomeHref,
    workspaceHomeRoute
  } from '$lib/application/workspaceAppNavigation';
  import {
    applyLoadedWorkspaceAssetDetail,
    loadWorkspaceAssetDetail,
    refreshWorkspaceAssetAttachments
  } from '$lib/application/workspaceAssetDetail';
  import { createAssetWorkflow, replaceWorkspaceAsset } from '$lib/application/workspaceAssetWorkflow';
  import { buildSearchSuggestions, executeWorkspaceSearch } from '$lib/application/workspaceSearch';
  import {
    type AssetRouteAction,
    type ImportSourceRoute,
    type SettingsSection,
    type WorkspaceRouteState
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
    type AssetCheckout,
    type AssetKind,
    type AssetLifecycleFilter,
    type CustomAssetType,
    type CustomFieldDefinition,
    type LocationAsset,
    type SearchCheckoutFilter,
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
  import InventoryWorkspaceChrome from './InventoryWorkspaceChrome.svelte';
  import InventoryWorkspaceOverlays from './InventoryWorkspaceOverlays.svelte';
  import InventoryWorkspaceRouteContent from './InventoryWorkspaceRouteContent.svelte';

  let {
    repository,
    initialData,
    onSignOut,
    onSessionExpired = onSignOut
  }: {
    repository: InventoryRepository & InventoryAccessRepository & InventoryAuditRepository & InventoryCustomizationRepository;
    initialData: WorkspaceData;
    onSignOut: () => void;
    onSessionExpired?: () => void;
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
  let addReturnLocationId = $state<string | null>(null);
  let addReturnAssetId = $state<string | null>(null);
  let assetAction = $state<AssetRouteAction>(null);
  let attachmentId = $state<string | null>(null);
  let attachmentAction = $state<WorkspaceRouteState['attachmentAction']>(null);
  let busy = $state(false);
  let message = $state('');
  let error = $state('');
  let searchQuery = $state('');
  let searchLifecycleState = $state<SearchLifecycleFilter>('active');
  let searchMode = $state<SearchMode>('fuzzy');
  let searchCheckoutState = $state<SearchCheckoutFilter>('any');
  let settingsSection = $state<SettingsSection>('overview');
  let invitationStatus = $state<WorkspaceRouteState['invitationStatus']>('all');
  let accessInvitationAction = $state<WorkspaceRouteState['accessInvitationAction']>(null);
  let accessInvitationId = $state<string | null>(null);
  let auditScope = $state<WorkspaceRouteState['auditScope']>('inventory');
  let customizationAction = $state<WorkspaceRouteState['customizationAction']>(null);
  let customAssetTypeId = $state<string | null>(null);
  let customFieldDefinitionId = $state<string | null>(null);
  let importSource = $state<ImportSourceRoute>(null);
  let searchResults = $state<SearchResult[]>([]);
  let searchSubmitted = $state(false);
  let searchError = $state('');
  let loadedAssetDetail = $state<Asset | null>(null);
  let selectedAssetAttachments = $state<AssetAttachment[]>([]);
  let selectedAssetCheckoutHistory = $state<AssetCheckout[]>([]);
  let assetDetailRequestId = 0;
  let applyingRoute = false;
  let routeUnavailable = $state('');

  let selectedInventory = $derived(data.context.inventories.find((inventory) => inventory.id === data.context.selectedInventoryId) ?? null);
  let selectedTenant = $derived(data.context.tenants.find((tenant) => tenant.id === data.context.selectedTenantId) ?? null);
  let assets = $derived(labelAssets(data.assets, data.context.customAssetTypes));
  let selectedLocation = $derived(
    (assets.find((asset) => asset.id === selectedLocationId && asset.kind === 'location') as LocationAsset | undefined) ??
      (loadedAssetDetail?.id === selectedLocationId && loadedAssetDetail.kind === 'location'
        ? (loadedAssetDetail as LocationAsset)
        : null)
  );
  let detailAssets = $derived(detailAssetList(assets, loadedAssetDetail, data.context.customAssetTypes));
  let selectedAsset = $derived(selectedAssetForDetail(selectedAssetId, assets, loadedAssetDetail, data.context.customAssetTypes));
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
      selectedAssetCheckoutHistory = [];
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
      selectedAssetCheckoutHistory = [];
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
        selectedAssetCheckoutHistory = [];
        attachmentId = null;
        attachmentAction = null;
      }
      if (result.selectedAsset) {
        selectedLocationId = result.selectedAsset.kind === 'location' ? result.selectedAsset.id : null;
        selectedAssetId = result.selectedAsset.kind === 'location' ? null : result.selectedAsset.id;
        loadedAssetDetail = result.selectedAsset;
        selectedAssetAttachments = [];
        selectedAssetCheckoutHistory = [];
        attachmentId = null;
        attachmentAction = null;
      }
      if (result.route) {
        replaceRoute(result.route);
      }
      return result.saveResult;
    } catch (caught) {
      if (handleSessionExpired(caught)) {
        return { saved: false };
      }
      throw caught;
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
      if (handleSessionExpired(caught)) {
        return;
      }
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
        mode: searchMode,
        checkoutState: searchCheckoutState
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
          searchMode,
          searchCheckoutState
        });
      }
      if (!result.query) {
        return;
      }
    } catch (caught) {
      if (handleSessionExpired(caught)) {
        return;
      }
      throw caught;
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
      selectedAssetCheckoutHistory = [];
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

  async function checkoutSelectedAsset(details: string): Promise<void> {
    const asset = selectedAsset;
    if (!asset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.checkoutAsset(asset.tenantId, asset.inventoryId, asset.id, { details: details || undefined });
      const refreshed = await repository.getAsset(asset.tenantId, asset.inventoryId, asset.id);
      selectedAssetCheckoutHistory = await repository.listAssetCheckoutHistory(asset.tenantId, asset.inventoryId, asset.id);
      data = replaceWorkspaceAsset(data, refreshed);
      loadedAssetDetail = refreshed;
      selectedAssetId = refreshed.id;
      message = `Checked out ${refreshed.title}.`;
    });
  }

  async function returnSelectedAsset(details: string): Promise<void> {
    const asset = selectedAsset;
    if (!asset || !selectedInventory) {
      return;
    }
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.returnAsset(asset.tenantId, asset.inventoryId, asset.id, { details: details || undefined });
      const refreshed = await repository.getAsset(asset.tenantId, asset.inventoryId, asset.id);
      selectedAssetCheckoutHistory = await repository.listAssetCheckoutHistory(asset.tenantId, asset.inventoryId, asset.id);
      data = replaceWorkspaceAsset(data, refreshed);
      loadedAssetDetail = refreshed;
      selectedAssetId = refreshed.id;
      message = `Returned ${refreshed.title}.`;
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
      if (handleSessionExpired(caught)) {
        return;
      }
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
    searchCheckoutState = 'any';
  }

  async function applyRoute(route: WorkspaceRouteState): Promise<void> {
    applyingRoute = true;
    try {
      const shouldCanonicalizeAlias = shouldCanonicalizeWorkspaceAlias(route);
      routeUnavailable = '';
      addOpen = false;
      addKind = route.addKind ?? 'item';
      addParentAssetId = null;
      assetAction = route.assetAction;
      attachmentId = route.attachmentId;
      attachmentAction = route.attachmentAction;
      searchQuery = route.searchQuery;
      searchLifecycleState = route.searchLifecycleState;
      searchMode = route.searchMode;
      searchCheckoutState = route.searchCheckoutState;
      settingsSection = route.settingsSection;
      invitationStatus = route.invitationStatus;
      accessInvitationAction = route.accessInvitationAction;
      accessInvitationId = route.accessInvitationId;
      auditScope = route.auditScope;
      customizationAction = route.customizationAction;
      customAssetTypeId = route.customAssetTypeId;
      customFieldDefinitionId = route.customFieldDefinitionId;
      importSource = route.importSource;

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
      const addRoute = resolveWorkspaceAddRoute(route, {
        createAllowed: createAssetAllowed,
        validParentIds: parentTargets(assets).map((target) => target.id),
        selectedTenantId: data.context.selectedTenantId,
        selectedInventoryId: data.context.selectedInventoryId
      });
      addOpen = addRoute.open;
      addKind = addRoute.kind;
      addParentAssetId = addRoute.parentAssetId;
      if (addRoute.open && addRoute.parentAssetId) {
        const parentTarget = parentTargets(assets).find((target) => target.id === addRoute.parentAssetId);
        addReturnLocationId = parentTarget?.kind === 'location' ? parentTarget.id : null;
        addReturnAssetId = null;
      } else if (!addRoute.open) {
        addReturnLocationId = null;
        addReturnAssetId = null;
      }
      if (addRoute.deniedMessage) {
        showUnavailableRoute(addRoute.deniedMessage);
        return;
      }
      if (addRoute.replacementRoute) {
        route = { ...route, addParentAssetId: null };
        replaceRoute(addRoute.replacementRoute);
      }

      if (route.mode === 'locations') {
        invalidateAssetDetailLoad();
        selectedLocationId = null;
        selectedAssetId = null;
        loadedAssetDetail = null;
        selectedAssetAttachments = [];
        selectedAssetCheckoutHistory = [];
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
          selectedAssetCheckoutHistory = [];
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
        selectedAssetCheckoutHistory = [];
        selectedAssetCheckoutHistory = [];
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
        selectedAssetCheckoutHistory = [];
        normalizeSettingsOverviewRoute(route);
        canonicalizeRouteAlias(route, shouldCanonicalizeAlias);
        return;
      }

      closeDetailToHome();
      normalizeInventoryHomeRoute(route);
      canonicalizeRouteAlias(route, shouldCanonicalizeAlias);
    } finally {
      applyingRoute = false;
    }
  }

  function navigateTo(route: Partial<WorkspaceRouteState>): void {
    void applyRoute(pushWorkspaceRoute(route, data.context.selectedTenantId || null, data.context.selectedInventoryId || null));
  }

  function homeRoute(): Partial<WorkspaceRouteState> {
    return workspaceHomeRoute(data.context);
  }

  function homeHref(): string {
    return workspaceHomeHref(data.context);
  }

  function openHome(event: MouseEvent): void {
    if (!shouldHandleWorkspaceLinkClick(event)) {
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
    selectedAssetCheckoutHistory = [];
    selectedAssetCheckoutHistory = [];
    searchResults = [];
    searchSubmitted = false;
  }

  function navigateMode(nextMode: WorkspaceMode): void {
    navigateTo({
      mode: nextMode,
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: nextMode === 'settings' ? settingsSection : 'overview',
      importSource: nextMode === 'import' ? importSource : null
    });
  }

  function openImportSource(source: ImportSourceRoute): void {
    importSource = source;
    navigateTo({
      mode: 'import',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      importSource: source
    });
  }

  async function refreshInventoryAfterImportJob(scope: { tenantId: string; inventoryId: string }): Promise<void> {
    const lifecycleState = data.context.assetLifecycleState;
    if (data.context.selectedTenantId !== scope.tenantId || data.context.selectedInventoryId !== scope.inventoryId) {
      return;
    }
    let refreshed: WorkspaceData;
    try {
      refreshed = await repository.selectAssetLifecycle(scope.tenantId, scope.inventoryId, lifecycleState);
    } catch (caught) {
      if (handleSessionExpired(caught)) {
        return;
      }
      throw caught;
    }
    if (
      data.context.selectedTenantId !== scope.tenantId ||
      data.context.selectedInventoryId !== scope.inventoryId ||
      data.context.assetLifecycleState !== lifecycleState
    ) {
      return;
    }
    data = refreshed;
    invalidateAssetDetailLoad();
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

  function openAdd(kind: AssetKind = 'item', parentAssetId: string | null = null): void {
    addReturnLocationId = parentAssetId && selectedLocationId ? selectedLocationId : mode === 'location' ? selectedLocationId : null;
    addReturnAssetId = !addReturnLocationId && mode === 'asset' ? selectedAssetId : null;
    navigateTo({
      action: 'add',
      addKind: kind,
      addParentAssetId: parentAssetId,
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId
    });
  }

  function closeAdd(): void {
    const closeRoute = workspaceAddCloseRoute(data.context, { mode, selectedLocationId: addReturnLocationId, selectedAssetId: addReturnAssetId });
    addOpen = false;
    replaceRoute(closeRoute);
    if (typeof window !== 'undefined') {
      void applyRoute(currentWorkspaceRoute());
    }
  }

  function addCloseHref(): string {
    return workspaceAddCloseHref(data.context, { mode, selectedLocationId: addReturnLocationId, selectedAssetId: addReturnAssetId });
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

  function normalizeSettingsOverviewRoute(route: WorkspaceRouteState): void {
    if (route.mode !== 'settings' || route.settingsSection !== 'overview' || typeof window === 'undefined') {
      return;
    }
    const canonicalHref = settingsOverviewHref(data.context);
    if (`${window.location.pathname}${window.location.search}` !== canonicalHref) {
      replaceRoute(settingsOverviewRoute(data.context));
    }
  }

  function normalizeInventoryHomeRoute(route: WorkspaceRouteState): void {
    if (
      route.mode !== 'home' ||
      route.action ||
      route.addKind ||
      typeof window === 'undefined' ||
      (!route.inventoryId && window.location.pathname !== '/') ||
      !data.context.selectedTenantId ||
      !data.context.selectedInventoryId
    ) {
      return;
    }
    const canonicalHref = inventoryHomeNormalizationHref(data.context, route);
    if (`${window.location.pathname}${window.location.search}` !== canonicalHref) {
      replaceRoute(inventoryHomeNormalizationRoute(data.context, route));
    }
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
      const detailState = applyLoadedWorkspaceAssetDetail(data, {
        ...result,
        loaded: true,
        asset: result.asset
      });
      data = detailState.data;
      loadedAssetDetail = detailState.loadedAssetDetail;
      selectedAssetId = detailState.selectedAssetId;
      selectedAssetAttachments = detailState.selectedAssetAttachments;
      selectedAssetCheckoutHistory = detailState.selectedAssetCheckoutHistory;
      mode = detailState.mode;
      return true;
    } catch (caught) {
      if (handleSessionExpired(caught)) {
        return false;
      }
      error = caught instanceof Error ? caught.message : 'Unable to load asset.';
      return false;
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

  async function refreshSelectedAttachments(tenantId: string, inventoryId: string, assetId: string): Promise<void> {
    try {
      selectedAssetAttachments = await refreshWorkspaceAssetAttachments(repository, { tenantId, inventoryId, assetId });
    } catch (caught) {
      if (handleSessionExpired(caught)) {
        return;
      }
      throw caught;
    }
  }

  function handleSessionExpired(caught: unknown): boolean {
    if (!isAuthenticationRequiredError(caught)) {
      return false;
    }
    onSessionExpired();
    return true;
  }

  function closeDetailToHome(): void {
    invalidateAssetDetailLoad();
    mode = 'home';
    selectedLocationId = null;
    selectedAssetId = null;
    loadedAssetDetail = null;
    selectedAssetAttachments = [];
    selectedAssetCheckoutHistory = [];
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
    selectedAssetCheckoutHistory = [];
    if (!applyingRoute) {
      replaceRoute(assetDetailBackRoute(data.context, selectedLocationId));
    }
  }

  function assetDetailBackHref(): string {
    return workspaceAssetDetailBackHref(data.context, selectedLocationId);
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

</script>

<InventoryWorkspaceChrome
    tenants={data.context.tenants}
    inventories={data.context.inventories}
    selectedTenantId={data.context.selectedTenantId}
    selectedInventoryId={data.context.selectedInventoryId}
    selectedInventory={selectedInventory}
    {mode}
    {settingsSection}
    {userLabel}
    searchSuggestions={searchSuggestions}
    bind:searchQuery
    canCreateAsset={createAssetAllowed && data.context.assetLifecycleState === 'active'}
    modalOpen={addOpen}
    onSelectTenant={(tenantId) => { void selectTenant(tenantId); }}
    onSelectInventory={(tenantId, inventoryId) => { void selectInventory(tenantId, inventoryId); }}
    onModeChange={navigateMode}
    onSearch={() => { void search(); }}
    onOpenSearchAsset={openSearchAsset}
    onOpenAdd={openAdd}
    {onSignOut}
  >
  <InventoryWorkspaceRouteContent
    workspace={{
      data,
      repository,
      selectedTenant,
      selectedInventory,
      selectedLocation,
      selectedAsset,
      assets,
      detailAssets,
      selectedAssetAttachments,
      selectedAssetCheckoutHistory
    }}
    status={{ busy, canCreateStarter, createAssetAllowed, editAssetAllowed }}
    route={{
      routeUnavailable,
      mode,
      searchResults,
      searchSuggestions,
      searchSubmitted,
      searchError,
      assetAction,
      attachmentId,
      attachmentAction,
      settingsSection,
      invitationStatus,
      accessInvitationAction,
      accessInvitationId,
      auditScope,
      customizationAction,
      customAssetTypeId,
      customFieldDefinitionId,
      importSource
    }}
    hrefs={{ homeHref: homeHref(), assetDetailBackHref: assetDetailBackHref() }}
    bind:searchQuery
    bind:searchLifecycleState
    bind:searchMode
    bind:searchCheckoutState
    handlers={{
      onHome: openHome,
      onCreateStarterInventory: createStarterInventory,
      onOpenLocation: openLocation,
      onEditLocation: openLocationEdit,
      onOpenAsset: openAsset,
      onOpenAdd: openAdd,
      onCloseLocation: closeLocationToLocations,
      onCloseAssetDetail: closeDetailToPrevious,
      onAssetActionOpen: openAssetActionRoute,
      onAssetActionClose: closeAssetActionRoute,
      onAssetSave: updateAsset,
      onAssetArchive: archiveSelectedAsset,
      onAssetRestore: restoreSelectedAsset,
      onAssetDelete: deleteSelectedAsset,
      onAssetCheckout: checkoutSelectedAsset,
      onAssetReturn: returnSelectedAsset,
      onAssetUploadAttachment: uploadSelectedAttachment,
      onAssetArchiveAttachment: archiveSelectedAttachment,
      onAttachmentDeleteOpen: openAttachmentDeleteRoute,
      onAttachmentDeleteClose: closeAttachmentDeleteRoute,
      onAssetDeleteAttachment: deleteSelectedAttachment,
      onSearch: search,
      onOpenSearchAsset: openSearchAsset,
      onImportSourceChange: openImportSource,
      onImportJobInventoryChanged: refreshInventoryAfterImportJob,
      onSettingsSectionChange: openSettingsSection,
      onInvitationStatusChange: openInvitationStatusFilter,
      onAccessInvitationActionOpen: openAccessInvitationAction,
      onAccessInvitationActionClose: closeAccessInvitationAction,
      onAuditScopeChange: openAuditScopeFilter,
      onCustomizationArchiveOpen: openCustomizationArchive,
      onCustomizationArchiveClose: closeCustomizationArchive,
      onCustomizationChange: updateCustomizationContext,
      onSelectLifecycle: selectAssetLifecycle
    }}
  />

</InventoryWorkspaceChrome>

<InventoryWorkspaceOverlays
  {addOpen}
  {createAssetAllowed}
  {addKind}
  {addParentAssetId}
  addCloseHref={addCloseHref()}
  parentTargets={parentTargets(assets)}
  mediaPolicy={data.context.mediaUploadPolicy}
  customAssetTypes={data.context.customAssetTypes}
  customFieldDefinitions={data.context.customFieldDefinitions}
  saving={busy}
  {message}
  {error}
  onAddClose={closeAdd}
  onAddSave={createAsset}
/>
