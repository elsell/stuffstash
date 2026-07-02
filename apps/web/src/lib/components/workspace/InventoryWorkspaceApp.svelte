<script lang="ts">
  import { onMount } from 'svelte';
  import { containedAssets, moveParentTargets, parentTargets, recentlyAddedAssets, topLevelLocations, withTrail } from '$lib/application/workspace';
  import {
    parseWorkspaceRoute,
    workspaceRouteHref,
    type AssetRouteAction,
    type SettingsSection,
    type WorkspaceRouteState
  } from '$lib/application/workspaceRoute';
  import {
    canCreateAsset,
    canEditAsset,
    canCreateInventory,
    type AddAssetDraft,
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
  let assetAction = $state<AssetRouteAction>(null);
  let busy = $state(false);
  let message = $state('');
  let error = $state('');
  let searchQuery = $state('');
  let searchLifecycleState = $state<SearchLifecycleFilter>('active');
  let searchMode = $state<SearchMode>('fuzzy');
  let settingsSection = $state<SettingsSection>('overview');
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
    void applyRoute(parseWorkspaceRoute(new URL(window.location.href)));
    const onPopState = () => {
      void applyRoute(parseWorkspaceRoute(new URL(window.location.href)));
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
    let createdParent: Asset | null = null;
    try {
      createdParent = draft.parentQuickCreate
        ? await repository.createAsset(data.context.selectedTenantId, selectedInventory.id, {
            kind: draft.parentQuickCreate.kind,
            title: draft.parentQuickCreate.title,
            description: '',
            parentAssetId: draft.parentAssetId,
            customFields: {},
            photos: []
          })
        : null;
      const { parentQuickCreate: _parentQuickCreate, ...assetDraft } = draft;
      const childDraft: AddAssetDraft = {
        ...assetDraft,
        parentAssetId: createdParent?.id ?? draft.parentAssetId
      };
      const asset = await repository.createAsset(data.context.selectedTenantId, selectedInventory.id, childDraft);
      if (data.context.assetLifecycleState === 'active') {
        let uploadFailures = 0;
        const uploaded = [];
        for (const photo of draft.photos) {
          try {
            uploaded.push(await repository.uploadAssetPhoto(asset.tenantId, asset.inventoryId, asset.id, photo));
          } catch {
            uploadFailures += 1;
          }
        }
        const firstPhoto = uploaded[0];
        const savedAsset = firstPhoto
          ? {
              ...asset,
              photo: {
                id: firstPhoto.id,
                assetId: asset.id,
                url: draft.photos[0]?.previewUrl ?? firstPhoto.thumbnailUrl ?? '',
                alt: asset.title
              }
            }
          : asset;
        data = { ...data, assets: createdParent ? [savedAsset, createdParent, ...data.assets] : [savedAsset, ...data.assets] };
        message =
          uploadFailures > 0
            ? `Saved ${asset.title}. ${uploadFailures} photo upload ${uploadFailures === 1 ? 'failed' : 'failed'}.`
            : draft.photos.length > 0
              ? `Saved ${asset.title} with ${draft.photos.length} photo upload.`
              : createdParent
                ? `Saved ${asset.title} in ${createdParent.title}.`
                : `Saved ${asset.title}.`;
        if (uploadFailures > 0) {
          return { saved: false };
        }
      } else {
        data = await repository.selectAssetLifecycle(asset.tenantId, asset.inventoryId, 'active');
        mode = 'home';
        selectedLocationId = null;
        selectedAssetId = null;
        loadedAssetDetail = null;
        selectedAssetAttachments = [];
        message = `Saved ${asset.title}.`;
      }
      addOpen = false;
      replaceRoute({ mode: 'asset', tenantId: asset.tenantId, inventoryId: asset.inventoryId, assetId: asset.id });
      return { saved: true };
    } catch (caught) {
      const failure = caught instanceof Error ? caught.message : 'Action failed.';
      if (createdParent && data.context.assetLifecycleState === 'active' && !data.assets.some((asset) => asset.id === createdParent?.id)) {
        data = { ...data, assets: [createdParent, ...data.assets] };
      }
      error = createdParent
        ? `Created ${createdParent.title}, but could not save ${draft.title}. ${failure}`
        : failure;
      return createdParent ? { saved: false, createdParentId: createdParent.id } : { saved: false };
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
      replaceWorkspaceAsset(asset);
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
    const query = searchQuery.trim();
    if (!query || !data.context.selectedTenantId) {
      searchResults = [];
      searchSubmitted = false;
      searchError = '';
      return;
    }
    busy = true;
    error = '';
    message = '';
    searchError = '';
    searchResults = [];
    searchSubmitted = true;
    try {
      searchResults = await repository.searchAssets({
        tenantId: data.context.selectedTenantId,
        inventoryId: data.context.selectedInventoryId,
        query,
        lifecycleState: searchLifecycleState,
        mode: searchMode
      });
      mode = 'search';
      if (!applyingRoute) {
        replaceRoute({
          mode: 'search',
          tenantId: data.context.selectedTenantId,
          inventoryId: data.context.selectedInventoryId,
          searchQuery: query,
          searchLifecycleState,
          searchMode
        });
      }
    } catch (caught) {
      searchError = caught instanceof Error ? caught.message : 'Search failed.';
      error = searchError;
      mode = 'search';
      if (!applyingRoute) {
        replaceRoute({
          mode: 'search',
          tenantId: data.context.selectedTenantId,
          inventoryId: data.context.selectedInventoryId,
          searchQuery: query,
          searchLifecycleState,
          searchMode
        });
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

  function openAsset(asset: Asset): void {
    navigateTo({ mode: 'asset', tenantId: asset.tenantId, inventoryId: asset.inventoryId, assetId: asset.id });
  }

  function openAssetById(assetId: string): void {
    const asset =
      assets.find((candidate) => candidate.id === assetId) ??
      searchResults.find((result) => result.asset.id === assetId)?.asset;
    if (asset) {
      openAsset(asset);
    }
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
      const shouldCanonicalizeAlias = !!route.inventoryId && !route.tenantId;
      routeUnavailable = '';
      addOpen = route.action === 'add';
      addKind = route.addKind ?? 'item';
      assetAction = route.assetAction;
      searchQuery = route.searchQuery;
      searchLifecycleState = route.searchLifecycleState;
      searchMode = route.searchMode;
      settingsSection = route.settingsSection;

      if (route.tenantId && route.tenantId !== data.context.selectedTenantId) {
        const tenant = data.context.tenants.find((candidate) => candidate.id === route.tenantId);
        if (tenant) {
          await selectTenant(route.tenantId);
        } else {
          showUnavailableRoute('That tenant is not available to this account.');
          return;
        }
      }

      if (route.inventoryId && route.inventoryId !== data.context.selectedInventoryId) {
        const inventory = data.context.inventories.find(
          (candidate) =>
            candidate.id === route.inventoryId &&
            (route.tenantId ? candidate.tenantId === route.tenantId : candidate.tenantId === data.context.selectedTenantId)
        );
        if (inventory) {
          await selectInventory(inventory.tenantId, inventory.id);
        } else {
          showUnavailableRoute('That inventory is not available in the current workspace.');
          return;
        }
      }

      if (route.lifecycleState !== data.context.assetLifecycleState && selectedInventory) {
        await selectAssetLifecycle(route.lifecycleState);
      }

      if (route.mode === 'location' && route.locationId) {
        const location = assets.find((candidate) => candidate.id === route.locationId && candidate.kind === 'location');
        if (location) {
          invalidateAssetDetailLoad();
          selectedLocationId = location.id;
          selectedAssetId = null;
          loadedAssetDetail = null;
          selectedAssetAttachments = [];
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
        if (route.assetAction && (!canEditAsset(selectedInventory) || (route.assetAction !== 'delete' && loadedAssetDetail?.lifecycleState !== 'active'))) {
          assetAction = null;
          replaceRoute({
            mode: 'asset',
            tenantId: data.context.selectedTenantId,
            inventoryId: data.context.selectedInventoryId,
            assetId: route.assetId
          });
          return;
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
    const href = workspaceRouteHref(route, data.context.selectedTenantId || null, data.context.selectedInventoryId || null);
    window.history.pushState({}, '', href);
    void applyRoute(parseWorkspaceRoute(new URL(window.location.href)));
  }

  function showUnavailableRoute(messageText: string): void {
    invalidateAssetDetailLoad();
    routeUnavailable = messageText;
    addOpen = false;
    assetAction = null;
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
      settingsSection: nextMode === 'settings' ? settingsSection : 'overview'
    });
  }

  function openSettingsSection(section: SettingsSection): void {
    settingsSection = section;
    navigateTo({
      mode: 'settings',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: section
    });
  }

  function openAdd(kind: AssetKind = 'item'): void {
    navigateTo({ action: 'add', addKind: kind, tenantId: data.context.selectedTenantId, inventoryId: data.context.selectedInventoryId });
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

  function openAssetActionRoute(action: Exclude<AssetRouteAction, null>): void {
    if (selectedAsset) {
      navigateTo({
        mode: 'asset',
        tenantId: selectedAsset.tenantId,
        inventoryId: selectedAsset.inventoryId,
        assetId: selectedAsset.id,
        assetAction: action,
        action: action === 'edit' ? 'edit' : null
      });
    }
  }

  function closeAssetActionRoute(): void {
    assetAction = null;
    if (selectedAsset) {
      replaceRoute({ mode: 'asset', tenantId: selectedAsset.tenantId, inventoryId: selectedAsset.inventoryId, assetId: selectedAsset.id });
    }
  }

  function replaceRoute(route: Partial<WorkspaceRouteState>): void {
    if (typeof window === 'undefined') {
      return;
    }
    window.history.replaceState(
      {},
      '',
      workspaceRouteHref(route, data.context.selectedTenantId || null, data.context.selectedInventoryId || null)
    );
  }

  function canonicalizeRouteAlias(route: WorkspaceRouteState, shouldCanonicalizeAlias: boolean): void {
    if (!shouldCanonicalizeAlias || !data.context.selectedTenantId || !data.context.selectedInventoryId) {
      return;
    }
    replaceRoute({
      ...route,
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId
    });
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
    try {
      const asset = await repository.getAsset(tenantId, inventoryId, assetId);
      const attachments = await repository.listAssetAttachments(tenantId, inventoryId, assetId);
      if (requestId !== assetDetailRequestId) {
        return false;
      }
      selectedAssetAttachments = attachments;
      replaceWorkspaceAsset(asset);
      loadedAssetDetail = asset;
      selectedAssetId = asset.id;
      mode = 'asset';
      return true;
    } catch (caught) {
      error = caught instanceof Error ? caught.message : 'Asset could not be loaded.';
      return false;
    } finally {
      busy = false;
    }
  }

  function replaceWorkspaceAsset(asset: Asset): void {
    if (asset.tenantId !== data.context.selectedTenantId || asset.inventoryId !== data.context.selectedInventoryId) {
      return;
    }
    if (asset.lifecycleState !== data.context.assetLifecycleState) {
      return;
    }
    const existing = data.assets.some(
      (candidate) =>
        candidate.tenantId === asset.tenantId && candidate.inventoryId === asset.inventoryId && candidate.id === asset.id
    );
    data = {
      ...data,
      assets: existing
        ? data.assets.map((candidate) =>
            candidate.tenantId === asset.tenantId && candidate.inventoryId === asset.inventoryId && candidate.id === asset.id
              ? asset
              : candidate
          )
        : [asset, ...data.assets]
    };
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
    message = 'Import applied.';
    replaceRoute({
      mode: 'home',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      lifecycleState: data.context.assetLifecycleState
    });
  }

  async function refreshSelectedAttachments(tenantId: string, inventoryId: string, assetId: string): Promise<void> {
    selectedAssetAttachments = await repository.listAssetAttachments(tenantId, inventoryId, assetId);
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

  function buildSearchSuggestions(items: Asset[], query: string): Asset[] {
    const normalized = query.trim().toLowerCase();
    if (!normalized) {
      return [];
    }
    return items.filter((asset) => {
      return (
        asset.title.toLowerCase().includes(normalized) ||
        asset.description.toLowerCase().includes(normalized) ||
        asset.customAssetTypeLabel?.toLowerCase().includes(normalized)
      );
    });
  }
</script>

<div class="product-shell">
  <SideNav
    tenants={data.context.tenants}
    inventories={data.context.inventories}
    selectedTenantId={data.context.selectedTenantId}
    selectedInventoryId={data.context.selectedInventoryId}
    {mode}
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
      onOpenSettings={() => navigateMode('settings')}
      onSearch={() => { void search(); }}
      onOpenAsset={openAsset}
      onOpenAdd={openAdd}
    />

    {#if routeUnavailable}
      <section class="workspace-main">
        <div class="empty-state spacious" role="alert">
          <h1>Workspace unavailable</h1>
          <p>{routeUnavailable}</p>
          <Button.Root onclick={() => navigateMode('home')}>Go home</Button.Root>
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
        onBack={closeDetailToHome}
        onOpenLocation={openLocation}
        onEditLocation={openAsset}
        onOpenAsset={openAsset}
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
        onBack={closeDetailToPrevious}
        onActionOpen={openAssetActionRoute}
        onActionClose={closeAssetActionRoute}
        onSave={updateAsset}
        onArchive={archiveSelectedAsset}
        onRestore={restoreSelectedAsset}
        onDelete={deleteSelectedAsset}
        onUploadAttachment={uploadSelectedAttachment}
        onArchiveAttachment={archiveSelectedAttachment}
        onDeleteAttachment={deleteSelectedAttachment}
      />
    {:else if mode === 'search'}
      <SearchPanel
        bind:query={searchQuery}
        bind:lifecycleState={searchLifecycleState}
        bind:searchMode={searchMode}
        results={searchResults}
        submitted={searchSubmitted}
        error={searchError}
        {busy}
        onSearch={() => { void search(); }}
        onOpenAsset={openAssetById}
      />
    {:else if mode === 'import'}
      <HomeboxImportPanel
        tenantId={data.context.selectedTenantId}
        inventory={selectedInventory}
        {repository}
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
        onSectionChange={openSettingsSection}
        onCustomizationChange={updateCustomizationContext}
      />
    {:else}
      <HomeWorkspace
        lifecycleState={data.context.assetLifecycleState}
        locations={topLevelLocations(assets)}
        recentAssets={recentlyAddedAssets(assets)}
        archivedAssets={assets}
        onOpenLocation={openLocation}
        onOpenAsset={openAsset}
        onOpenAdd={() => openAdd('location')}
        onSelectLifecycle={(lifecycleState) => { void selectAssetLifecycle(lifecycleState); }}
      />
    {/if}
  </div>

  <MobileNav
    {mode}
    canCreateAsset={createAssetAllowed && data.context.assetLifecycleState === 'active'}
    onModeChange={navigateMode}
    onOpenAdd={() => openAdd('item')}
  />

  <AddAssetTray
    open={addOpen && createAssetAllowed}
    initialKind={addKind}
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
