<script lang="ts">
  import { shouldHandleWorkspaceLinkClick } from '$lib/application/workspaceLinkHandling';
  import { isAuthenticationRequiredError } from '$lib/application/authenticationRequired';
  import { afterNavigate } from '$app/navigation';
  import { onMount, setContext, tick } from 'svelte';
  import { addReturnFocusTarget } from '$lib/application/workspaceAddFocus';
  import { assetThumbnailLoaderContext, type AssetThumbnailLoader } from '$lib/ports/assetThumbnailLoader';
  import {
    detailAssetList,
    labelAssets,
    parentTargets,
    selectedAssetForDetail
  } from '$lib/application/workspace';
  import { accountDisplayLabel } from '$lib/application/workspaceShellNavigation';
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
  import { reconcilePendingAssetTagDrafts } from '$lib/application/workspaceTagDrafts';
  import { buildSearchSuggestions } from '$lib/application/workspaceSearch';
  import { browseFailureMessage } from '$lib/application/workspaceBrowsePresentation';
  import {
    type AssetRouteAction,
    type ImportSourceRoute,
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
    type AssetCheckout,
    type AssetKind,
    type AssetLifecycleFilter,
    type AssetTag,
    type BrowseScope,
    type BrowseSort,
    type BrowseSurface,
    type CustomAssetType,
    type CustomFieldDefinition,
    type LocationAsset,
    type SearchCheckoutFilter,
    type SearchLifecycleFilter,
    type SearchMode,
    type SearchResult,
    type SelectedAttachment,
  type UpdateAssetDraft,
  type UndoableOperationDirection,
    type WorkspaceData,
    type WorkspaceMode
  } from '$lib/domain/inventory';
  import type { InventoryAccessRepository } from '$lib/ports/inventoryAccessRepository';
  import type { InventoryAuditRepository } from '$lib/ports/inventoryAuditRepository';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import type { InventoryRepository } from '$lib/ports/inventoryRepository';
  import type { InventoryBrowseRepository } from '$lib/ports/inventoryBrowseRepository';
  import type { WorkspaceNotification, WorkspaceNotificationAction } from '$lib/components/ui/sonner/index.js';
  import InventoryWorkspaceChrome from './InventoryWorkspaceChrome.svelte';
  import InventoryWorkspaceOverlays from './InventoryWorkspaceOverlays.svelte';
  import InventoryWorkspaceRouteContent from './InventoryWorkspaceRouteContent.svelte';

  let {
    repository,
    initialData,
    onSignOut,
    onSessionExpired = onSignOut
  }: {
    repository: InventoryRepository & InventoryBrowseRepository & InventoryAccessRepository & InventoryAuditRepository & InventoryCustomizationRepository & AssetThumbnailLoader;
    initialData: WorkspaceData;
    onSignOut: () => void;
    onSessionExpired?: () => void;
  } = $props();

  // svelte-ignore state_referenced_locally -- the repository is immutable for the mounted workspace session.
  const workspaceRepository = repository;
  setContext(assetThumbnailLoaderContext, {
    loadAssetThumbnail: workspaceRepository.loadAssetThumbnail.bind(workspaceRepository)
  });

  // svelte-ignore state_referenced_locally -- initial route data seeds local workspace state.
  const startingData = initialData;
  const startingRoute = currentWorkspaceRoute();
  let data = $state(startingData);
  let mode = $state<WorkspaceMode>(startingData.context.inventories.length > 0 ? startingRoute.mode : 'settings');
  let selectedLocationId = $state<string | null>(null);
  let selectedAssetId = $state<string | null>(null);
  let addOpen = $state(false);
  let addKind = $state<AssetKind>('item');
  let addParentAssetId = $state<string | null>(null);
  let addReturnLocationId = $state<string | null>(null);
  let addReturnAssetId = $state<string | null>(null);
  let addReturnFocusElement: HTMLElement | null = null;
  let assetAction = $state<AssetRouteAction>(null);
  let attachmentId = $state<string | null>(null);
  let attachmentAction = $state<WorkspaceRouteState['attachmentAction']>(null);
  let busy = $state(false);
  let notification = $state<WorkspaceNotification | null>(null);
  let error = $state('');
  let searchQuery = $state('');
  let searchLifecycleState = $state<SearchLifecycleFilter>('active');
  let searchMode = $state<SearchMode>('fuzzy');
  let searchCheckoutState = $state<SearchCheckoutFilter>('any');
  let searchTagIds = $state<string[]>([]);
  let browseSurface = $state<BrowseSurface>(startingRoute.browseSurface);
  let browseScope = $state<BrowseScope>(startingRoute.browseScope);
  let browseSort = $state<BrowseSort>(startingRoute.browseSort);
  let settingsSection = $state<SettingsSection>('overview');
  let invitationStatus = $state<WorkspaceRouteState['invitationStatus']>('all');
  let accessInvitationAction = $state<WorkspaceRouteState['accessInvitationAction']>(null);
  let accessInvitationId = $state<string | null>(null);
  let auditScope = $state<WorkspaceRouteState['auditScope']>('inventory');
  let customizationAction = $state<WorkspaceRouteState['customizationAction']>(null);
  let customAssetTypeId = $state<string | null>(null);
  let customFieldDefinitionId = $state<string | null>(null);
  let importSource = $state<ImportSourceRoute>(null);
  let importJobId = $state<string | null>(null);
  let importTab = $state<WorkspaceRouteState['importTab']>(null);
  let searchResults = $state<SearchResult[]>([]);
  let searchSubmitted = $state(false);
  let searchError = $state('');
  let browseAssets = $state<Asset[]>([]);
  let browseNextCursor = $state<string | null>(null);
  let browseHasMore = $state(false);
  let browseLoadingMore = $state(false);
  let browseBusy = $state(false);
  let browseRequestId = 0;
  let browseMapAssets = $state<Asset[]>([]);
  let browseInventoryEmpty = $state(false);
  let browseErrorPhase = $state<'initial' | 'replacement' | 'append' | 'map' | null>(null);
  let loadedAssetDetail = $state<Asset | null>(null);
  let selectedAssetAttachments = $state<AssetAttachment[]>([]);
  let selectedAssetCheckoutHistory = $state<AssetCheckout[]>([]);
  let assetDetailRequestId = 0;
  let applyingRoute = false;
  let activeRouteApplicationKey: string | null = null;
  let queuedRoute: { route: WorkspaceRouteState; href: string } | null = null;
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
  let userLabel = $derived(accountDisplayLabel(data.context.principal));
  let modalSurfaceOpen = $derived(
    addOpen || assetAction !== null || attachmentAction === 'delete' ||
    accessInvitationAction !== null || customizationAction !== null
  );

  onMount(() => {
    void applyRoute(currentWorkspaceRoute());
    const onPopState = () => {
      void applyRoute(currentWorkspaceRoute());
    };
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  });

  afterNavigate(() => {
    void applyRoute(currentWorkspaceRoute());
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
      replaceRoute({ mode: 'home', tenantId: data.context.selectedTenantId, inventoryId: data.context.selectedInventoryId });
      setSuccessNotification('Created Household.', {
        label: 'Open inventory',
        href: workspaceRouteHref(
          { mode: 'home', tenantId: data.context.selectedTenantId, inventoryId: data.context.selectedInventoryId },
          data.context.selectedTenantId,
          data.context.selectedInventoryId
        )
      });
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
    notification = null;
    try {
      const result = await createAssetWorkflow(repository, data, selectedInventory, draft);
      data = result.data;
      if (result.message) {
        setMutationSuccessNotification(
          result.message,
          result.selectedAsset,
          result.selectedAsset ? viewAssetAction(result.selectedAsset) : undefined
        );
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
    notification = null;
    const createdTags = [];
    try {
      const previousParentId = selectedAsset.parentAssetId;
      const reconciledTags = reconcilePendingAssetTagDrafts(data.context.assetTags ?? [], draft.tagIds ?? [], draft.newTags ?? []);
      for (const tag of reconciledTags.newTags) {
        createdTags.push(await repository.createAssetTag(selectedAsset.tenantId, selectedAsset.inventoryId, tag));
      }
      const asset = await repository.updateAsset(
        selectedAsset.tenantId,
        selectedAsset.inventoryId,
        selectedAsset.id,
        {
          ...draft,
          tagIds: [...reconciledTags.tagIds, ...createdTags.map((tag) => tag.id)]
        }
      );
      data = {
        ...replaceWorkspaceAsset(data, asset),
        context: {
          ...data.context,
          assetTags: mergeAssetTags(data.context.assetTags ?? [], createdTags)
        }
      };
      loadedAssetDetail = asset;
      setMutationSuccessNotification(
        `Saved ${asset.title}.`,
        asset,
        asset.parentAssetId !== previousParentId ? parentDestinationAction(asset.parentAssetId) : undefined
      );
    } catch (caught) {
      if (handleSessionExpired(caught)) {
        return;
      }
      if (createdTags.length > 0) {
        data = {
          ...data,
          context: {
            ...data.context,
            assetTags: mergeAssetTags(data.context.assetTags ?? [], createdTags)
          }
        };
      }
      error = caught instanceof Error ? caught.message : 'Action failed.';
      throw new Error(error);
    } finally {
      busy = false;
    }
  }

  async function moveAssetHere(candidate: Asset): Promise<void> {
    const target = selectedAsset ?? selectedLocation;
    if (!target || (target.kind !== 'container' && target.kind !== 'location')) return;
    if (!editAssetAllowed) {
      throw new Error('Move not saved. You do not have permission to move assets in this inventory.');
    }
    busy = true;
    notification = null;
    try {
      const moved = await repository.moveAsset(candidate.tenantId, candidate.inventoryId, candidate.id, target.id);
      data = replaceWorkspaceAsset(data, moved);
      setMutationSuccessNotification(`Moved ${moved.title} into ${target.title}.`, moved, viewAssetAction(moved));
      closeAssetActionRoute();
    } catch (caught) {
      if (handleSessionExpired(caught)) return;
      const reason = caught instanceof Error ? caught.message : 'Choose another asset and try again.';
      throw new Error(`Move not saved. ${candidate.title} stayed where it was. ${reason}`);
    } finally {
      busy = false;
    }
  }

  function mergeAssetTags(existingTags: NonNullable<WorkspaceData['context']['assetTags']>, nextTags: NonNullable<WorkspaceData['context']['assetTags']>): NonNullable<WorkspaceData['context']['assetTags']> {
    if (nextTags.length === 0) {
      return existingTags;
    }
    const byId = new Map(existingTags.map((tag) => [tag.id, tag]));
    for (const tag of nextTags) {
      byId.set(tag.id, tag);
    }
    return Array.from(byId.values()).sort((left, right) => left.displayName.localeCompare(right.displayName));
  }

  async function search(): Promise<void> {
    navigateTo({
      mode: 'browse', tenantId: data.context.selectedTenantId, inventoryId: data.context.selectedInventoryId,
      searchQuery: searchQuery.trim(), searchLifecycleState, searchMode, searchCheckoutState,
      browseSurface: 'list', browseScope, browseSort, browseTagIds: searchTagIds
    });
  }

  async function loadBrowsePage(append = false): Promise<void> {
    const tenantId = data.context.selectedTenantId;
    const inventoryId = data.context.selectedInventoryId;
    if (!tenantId || !inventoryId) return;
    const requestId = ++browseRequestId;
    if (append) browseLoadingMore = true;
    else browseBusy = true;
    searchError = '';
    browseErrorPhase = null;
    try {
      const page = await repository.browseAssets({
        tenantId, inventoryId, query: searchQuery, tagIds: searchTagIds, lifecycleState: searchLifecycleState,
        checkoutState: searchCheckoutState, scope: browseScope, sort: browseSort, mode: searchMode, limit: 20,
        cursor: append ? browseNextCursor ?? undefined : undefined
      });
      const checksDefaultInventoryEmptiness = !append && !searchQuery.trim() && searchTagIds.length === 0 &&
        browseScope === 'all' && searchLifecycleState === 'active' && searchCheckoutState === 'any';
      const inventoryEmpty = checksDefaultInventoryEmptiness && page.assets.length === 0
        ? !(await repository.hasAnyAssets(tenantId, inventoryId))
        : page.assets.length > 0 ? false : browseInventoryEmpty;
      if (requestId !== browseRequestId) return;
      browseAssets = append ? [...browseAssets, ...page.assets] : page.assets;
      searchResults = append ? [...searchResults, ...page.searchResults] : page.searchResults;
      browseNextCursor = page.nextCursor;
      browseHasMore = page.hasMore;
      browseInventoryEmpty = inventoryEmpty;
      searchSubmitted = !!searchQuery.trim() || searchTagIds.length > 0;
    } catch (caught) {
      if (requestId !== browseRequestId || handleSessionExpired(caught)) return;
      const phase = append ? 'append' : browseAssets.length > 0 ? 'replacement' : 'initial';
      searchError = browseFailureMessage(caught, phase);
      browseErrorPhase = phase;
    } finally {
      if (requestId === browseRequestId) {
        browseBusy = false;
        browseLoadingMore = false;
      }
    }
  }

  async function loadBrowseMap(): Promise<void> {
    const tenantId = data.context.selectedTenantId;
    const inventoryId = data.context.selectedInventoryId;
    if (!tenantId || !inventoryId) return;
    const requestId = ++browseRequestId;
    browseBusy = true;
    searchError = '';
    browseErrorPhase = null;
    try {
      const assets = await repository.loadActiveContainmentMap(tenantId, inventoryId);
      if (requestId === browseRequestId) browseMapAssets = assets;
    } catch (caught) {
      if (requestId === browseRequestId && !handleSessionExpired(caught)) {
        searchError = browseFailureMessage(caught, 'map');
        browseErrorPhase = 'map';
      }
    } finally {
      if (requestId === browseRequestId) browseBusy = false;
    }
  }

  async function searchForTag(tag: AssetTag): Promise<void> {
    searchTagIds = searchTagIds.includes(tag.id) ? searchTagIds.filter((tagId) => tagId !== tag.id) : [...searchTagIds, tag.id];
    await search();
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
      const archived = await repository.archiveAsset(asset.tenantId, asset.inventoryId, asset.id);
      await refreshSelectedAssetLifecycle();
      closeDetailToHome();
      setMutationSuccessNotification(`Archived ${asset.title}.`, archived, {
        label: 'View archived',
        href: workspaceRouteHref(
          { mode: 'home', tenantId: asset.tenantId, inventoryId: asset.inventoryId, lifecycleState: 'archived' },
          asset.tenantId,
          asset.inventoryId
        )
      });
    }, { rethrow: true });
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
      const restored = await repository.restoreAsset(asset.tenantId, asset.inventoryId, asset.id);
      await refreshSelectedAssetLifecycle();
      closeDetailToHome();
      setMutationSuccessNotification(`Restored ${asset.title}.`, restored, viewAssetAction(restored));
    }, { rethrow: true });
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
      setSuccessNotification(`Deleted ${asset.title}.`);
    }, { rethrow: true });
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
      const checkout = await repository.checkoutAsset(asset.tenantId, asset.inventoryId, asset.id, { details: details || undefined });
      const refreshed = await repository.getAsset(asset.tenantId, asset.inventoryId, asset.id);
      selectedAssetCheckoutHistory = await repository.listAssetCheckoutHistory(asset.tenantId, asset.inventoryId, asset.id);
      data = replaceWorkspaceAsset(data, refreshed);
      loadedAssetDetail = refreshed;
      selectedAssetId = refreshed.id;
      setMutationSuccessNotification(`Checked out ${refreshed.title}.`, checkout, viewAssetAction(refreshed));
    }, { rethrow: true });
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
      const checkout = await repository.returnAsset(asset.tenantId, asset.inventoryId, asset.id, { details: details || undefined });
      const refreshed = await repository.getAsset(asset.tenantId, asset.inventoryId, asset.id);
      selectedAssetCheckoutHistory = await repository.listAssetCheckoutHistory(asset.tenantId, asset.inventoryId, asset.id);
      data = replaceWorkspaceAsset(data, refreshed);
      loadedAssetDetail = refreshed;
      selectedAssetId = refreshed.id;
      setMutationSuccessNotification(`Returned ${refreshed.title}.`, checkout, viewAssetAction(refreshed));
    }, { rethrow: true });
  }

  async function returnAssetFromHome(asset: Asset): Promise<void> {
    if (!editAssetAllowed || !selectedInventory) {
      error = 'You do not have permission to edit assets in this inventory.';
      return;
    }
    await run(async () => {
      const checkout = await repository.returnAsset(asset.tenantId, asset.inventoryId, asset.id, {});
      const returnedAsset: Asset = { ...asset, currentCheckout: undefined };
      data = replaceWorkspaceAsset(data, returnedAsset);
      setMutationSuccessNotification(`Returned ${returnedAsset.title}.`, checkout, viewAssetAction(returnedAsset));
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
      setSuccessNotification(`Archived ${attachment.fileName}.`, viewAssetByIdAction(attachment.tenantId, attachment.inventoryId, attachment.assetId));
    }, { rethrow: true });
  }

  async function deleteSelectedAttachment(attachment: AssetAttachment): Promise<void> {
    if (!editAssetAllowed) {
      error = 'You do not have permission to edit assets in this inventory.';
      throw new Error(error);
    }
    await run(async () => {
      await repository.deleteAssetAttachment(attachment.tenantId, attachment.inventoryId, attachment.assetId, attachment.id);
      await refreshSelectedAttachments(attachment.tenantId, attachment.inventoryId, attachment.assetId);
      setSuccessNotification(`Deleted ${attachment.fileName}.`, viewAssetByIdAction(attachment.tenantId, attachment.inventoryId, attachment.assetId));
    }, { rethrow: true });
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
    busy = true;
    notification = null;
    try {
      await repository.uploadAssetAttachment(asset.tenantId, asset.inventoryId, asset.id, attachment);
      await refreshSelectedAttachments(asset.tenantId, asset.inventoryId, asset.id);
      setSuccessNotification(`Uploaded ${attachment.name}.`, viewAssetAction(asset));
    } catch (caught) {
      if (handleSessionExpired(caught)) return;
      throw caught instanceof Error ? caught : new Error('Attachment upload failed.');
    } finally {
      busy = false;
    }
  }

  async function run(task: () => Promise<void>, options: { rethrow?: boolean } = {}): Promise<void> {
    busy = true;
    error = '';
    notification = null;
    try {
      await task();
    } catch (caught) {
      if (handleSessionExpired(caught)) {
        return;
      }
      const taskError = caught instanceof Error ? caught.message : 'Action failed.';
      if (options.rethrow) {
        throw caught instanceof Error ? caught : new Error(taskError);
      }
      error = taskError;
    } finally {
      busy = false;
    }
  }

  function setSuccessNotification(title: string, action?: WorkspaceNotificationAction): void {
    notification = { kind: 'success', title, action };
  }

  function setMutationSuccessNotification(
    title: string,
    result: { tenantId: string; inventoryId: string; undoableOperationId?: string } | null | undefined,
    fallbackAction?: WorkspaceNotificationAction
  ): void {
    if (!result?.undoableOperationId) {
      setSuccessNotification(title, fallbackAction);
      return;
    }
    const operationId = result.undoableOperationId;
    notification = {
      id: `asset-operation:${operationId}`,
      kind: 'success',
      title,
      duration: 10_000,
      action: {
        label: 'Undo',
        onClick: () => applyUndoableAssetOperation(result.tenantId, result.inventoryId, operationId, 'undo')
      }
    };
  }

  async function applyUndoableAssetOperation(
    tenantId: string,
    inventoryId: string,
    operationId: string,
    direction: UndoableOperationDirection
  ): Promise<void> {
    if (busy) return;
    busy = true;
    error = '';
    notification = {
      id: `asset-operation:${operationId}`,
      kind: 'info',
      title: direction === 'undo' ? 'Undoing change…' : 'Redoing change…',
      important: true,
      duration: Infinity
    };
    try {
      const asset = await repository.applyAssetOperation(tenantId, inventoryId, operationId, direction);
      await refreshSelectedAssetLifecycle();
      if (selectedAssetId === asset.id || selectedLocationId === asset.id) {
        if (asset.lifecycleState === data.context.assetLifecycleState) {
          loadedAssetDetail = asset;
        } else {
          closeDetailToHome();
        }
      }
      const inverse: UndoableOperationDirection = direction === 'undo' ? 'redo' : 'undo';
      busy = false;
      notification = {
        id: `asset-operation:${operationId}`,
        kind: 'success',
        title: `${direction === 'undo' ? 'Undid' : 'Redid'} change to ${asset.title}.`,
        duration: 10_000,
        action: {
          label: inverse === 'undo' ? 'Undo' : 'Redo',
          onClick: () => applyUndoableAssetOperation(tenantId, inventoryId, operationId, inverse)
        }
      };
    } catch (caught) {
      if (handleSessionExpired(caught)) return;
      try {
        await refreshSelectedAssetLifecycle();
      } catch (refreshError) {
        if (handleSessionExpired(refreshError)) return;
      }
      busy = false;
      notification = {
        id: `asset-operation:${operationId}`,
        kind: 'error',
        title: direction === 'undo' ? 'Couldn’t undo change.' : 'Couldn’t redo change.',
        description: caught instanceof Error ? caught.message : 'The saved operation is no longer available.',
        important: true,
        duration: Infinity
      };
    } finally {
      if (busy) busy = false;
    }
  }

  function viewAssetAction(asset: Asset): WorkspaceNotificationAction {
    return asset.kind === 'location'
      ? {
          label: 'View location',
          href: workspaceRouteHref(
            { mode: 'location', tenantId: asset.tenantId, inventoryId: asset.inventoryId, locationId: asset.id },
            asset.tenantId,
            asset.inventoryId
          )
        }
      : viewAssetByIdAction(asset.tenantId, asset.inventoryId, asset.id);
  }

  function viewAssetByIdAction(tenantId: string, inventoryId: string, assetId: string): WorkspaceNotificationAction {
    return {
      label: 'View asset',
      href: workspaceRouteHref({ mode: 'asset', tenantId, inventoryId, assetId }, tenantId, inventoryId)
    };
  }

  function parentDestinationAction(parentAssetId: string | null): WorkspaceNotificationAction {
    if (!parentAssetId) {
      return { label: 'View home', href: homeHref() };
    }
    const target = parentTargets(assets).find((candidate) => candidate.id === parentAssetId);
    if (!target) {
      return { label: 'View asset', href: workspaceRouteHref({ mode: 'asset', assetId: parentAssetId }, data.context.selectedTenantId, data.context.selectedInventoryId) };
    }
    if (target.kind === 'location') {
      return {
        label: 'View location',
        href: workspaceRouteHref(
          { mode: 'location', tenantId: target.tenantId, inventoryId: target.inventoryId, locationId: target.id },
          target.tenantId,
          target.inventoryId
        )
      };
    }
    return {
      label: 'View parent',
      href: workspaceRouteHref(
        { mode: 'asset', tenantId: target.tenantId, inventoryId: target.inventoryId, assetId: target.id },
        target.tenantId,
        target.inventoryId
      )
    };
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
    searchTagIds = [];
  }

  async function applyRoute(route: WorkspaceRouteState): Promise<void> {
    const routeKey = routeApplicationKey(route);
    if (activeRouteApplicationKey) {
      if (routeKey === activeRouteApplicationKey) {
        queuedRoute = null;
      } else {
        browseRequestId += 1;
        browseBusy = false;
        browseLoadingMore = false;
        queuedRoute = { route, href: currentWorkspaceHref() };
      }
      return;
    }
    activeRouteApplicationKey = routeKey;
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
      searchTagIds = route.browseTagIds;
      browseSurface = route.browseSurface;
      browseScope = route.browseScope;
      browseSort = route.browseSort;
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
      importJobId = route.importJobId;
      importTab = route.importTab;

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

      if (route.mode !== 'search' && route.mode !== 'browse' && route.lifecycleState !== data.context.assetLifecycleState && selectedInventory) {
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
          if (route.assetAction === 'move-here' && !assetRouteActionIsAvailable(route.assetAction, selectedInventory, location)) {
            assetAction = null;
            replaceRoute({
              mode: 'location',
              tenantId: location.tenantId,
              inventoryId: location.inventoryId,
              locationId: location.id
            });
          }
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

      if (route.mode === 'browse' || route.mode === 'search') {
        mode = 'browse';
        selectedLocationId = null;
        selectedAssetId = null;
        loadedAssetDetail = null;
        selectedAssetAttachments = [];
        selectedAssetCheckoutHistory = [];
        selectedAssetCheckoutHistory = [];
        if (route.browseSurface === 'map') await loadBrowseMap();
        else await loadBrowsePage(false);
        canonicalizeRouteAlias(route, shouldCanonicalizeAlias);
        return;
      }

      if (route.mode === 'settings' || route.mode === 'import') {
        mode = route.mode;
        settingsSection = route.settingsSection;
        importSource = route.importSource;
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
      if (activeRouteApplicationKey === routeKey) {
        activeRouteApplicationKey = null;
      }
      applyingRoute = false;
      const nextRoute = queuedRoute;
      queuedRoute = null;
      if (nextRoute) {
        window.history.replaceState({}, '', nextRoute.href);
        void applyRoute(currentWorkspaceRoute());
      }
    }
  }

  function routeApplicationKey(route: WorkspaceRouteState): string {
    return JSON.stringify(route);
  }

  function currentWorkspaceHref(): string {
    return typeof window === 'undefined' ? '/' : `${window.location.pathname}${window.location.search}`;
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
    importJobId = null;
    importTab = null;
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
      importSource: nextMode === 'import' ? importSource : null,
      importJobId: null,
      importTab: null
    });
  }

  function updateBrowseState(next: {
    surface?: BrowseSurface;
    scope?: BrowseScope;
    lifecycleState?: SearchLifecycleFilter;
    checkoutState?: SearchCheckoutFilter;
    sort?: BrowseSort;
    selectedTagIds?: string[];
  }): void {
    navigateTo({
      mode: 'browse',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      browseSurface: next.surface ?? browseSurface,
      browseScope: next.scope ?? browseScope,
      browseSort: next.sort ?? browseSort,
      browseTagIds: next.selectedTagIds ?? searchTagIds,
      searchQuery,
      searchLifecycleState: next.lifecycleState ?? searchLifecycleState,
      searchCheckoutState: next.checkoutState ?? searchCheckoutState,
      searchMode
    });
  }

  function openImportSource(source: ImportSourceRoute): void {
    importSource = source;
    navigateTo({
      mode: 'import',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      importSource: source,
      importJobId: null,
      importTab: null
    });
  }

  function openImportJobRoute(jobId: string | null, tab: WorkspaceRouteState['importTab'] = null): void {
    navigateTo({
      mode: 'import',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      importSource: null,
      importJobId: jobId,
      importTab: jobId ? tab : null
    });
  }

  function openImportJobTabRoute(tab: WorkspaceRouteState['importTab']): void {
    if (mode !== 'import') return;
    if (!importJobId) return;
    navigateTo({
      mode: 'import',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      importSource: null,
      importJobId,
      importTab: tab
    });
  }

  async function openImportedAssetId(assetId: string): Promise<void> {
    await applyRoute(
      pushWorkspaceRoute(
        {
          mode: 'asset',
          tenantId: data.context.selectedTenantId,
          inventoryId: data.context.selectedInventoryId,
          assetId
        },
        data.context.selectedTenantId || null,
        data.context.selectedInventoryId || null
      )
    );
  }

  function openInventoryAuditHistory(): void {
    auditScope = 'inventory';
    navigateTo({
      mode: 'settings',
      tenantId: data.context.selectedTenantId,
      inventoryId: data.context.selectedInventoryId,
      settingsSection: 'activity',
      auditScope: 'inventory'
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
    addReturnFocusElement = typeof document !== 'undefined' && document.activeElement instanceof HTMLElement
      ? document.activeElement
      : null;
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
    const returnFocusElement = addReturnFocusElement;
    addReturnFocusElement = null;
    const closeRoute = workspaceAddCloseRoute(data.context, { mode, selectedLocationId: addReturnLocationId, selectedAssetId: addReturnAssetId });
    addOpen = false;
    replaceRoute(closeRoute);
    if (typeof window !== 'undefined') {
      void (async () => {
        await applyRoute(currentWorkspaceRoute());
        await tick();
        addReturnFocusTarget(returnFocusElement)?.focus();
      })();
    }
  }

  function addCloseHref(): string {
    return workspaceAddCloseHref(data.context, { mode, selectedLocationId: addReturnLocationId, selectedAssetId: addReturnAssetId });
  }

  function openAssetActionRoute(action: Exclude<AssetRouteAction, null>): void {
    const target = selectedAsset ?? selectedLocation;
    if (target) {
      const isLocationEdit = target.kind === 'location' && action === 'edit';
      const isLocationMoveHere = target.kind === 'location' && action === 'move-here';
      navigateTo({
        mode: isLocationMoveHere ? 'location' : 'asset',
        tenantId: target.tenantId,
        inventoryId: target.inventoryId,
        locationId: isLocationEdit || isLocationMoveHere ? target.id : null,
        assetId: isLocationMoveHere ? null : target.id,
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
    const closingAsset = selectedAsset ?? selectedLocation;
    if (closingAsset) {
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
    updateBrowseState({ scope: 'places' });
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
    modalOpen={modalSurfaceOpen}
    onSelectTenant={(tenantId) => { void selectTenant(tenantId); }}
    onSelectInventory={(tenantId, inventoryId) => { void selectInventory(tenantId, inventoryId); }}
    onModeChange={navigateMode}
    onSearch={() => { void search(); }}
    onOpenSearchAsset={openSearchAsset}
    onOpenAdd={openAdd}
    onOpenAccountSettings={() => openSettingsSection('overview')}
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
      browseSurface,
      browseScope,
      browseSort,
      browseTagIds: searchTagIds,
      browseAssets: browseSurface === 'map' ? browseMapAssets : browseAssets,
      browseInventoryEmpty,
      browseHasMore,
      browseLoadingMore,
      browseBusy,
      browseErrorPhase,
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
      importSource,
      importJobId,
      importTab
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
      onOpenLocations: () => updateBrowseState({ scope: 'places' }),
      onBrowseStateChange: updateBrowseState,
      onBrowseLoadMore: () => loadBrowsePage(true),
      onBrowseRetry: () => browseErrorPhase === 'append' ? loadBrowsePage(true) : browseErrorPhase === 'map' ? loadBrowseMap() : loadBrowsePage(false),
      onEditLocation: openLocationEdit,
      onOpenAsset: openAsset,
      onOpenAdd: openAdd,
      onCloseLocation: closeLocationToLocations,
      onCloseAssetDetail: closeDetailToPrevious,
      onAssetActionOpen: openAssetActionRoute,
      onAssetActionClose: closeAssetActionRoute,
      onAssetSave: updateAsset,
      onMoveAssetHere: moveAssetHere,
      onAssetArchive: archiveSelectedAsset,
      onAssetRestore: restoreSelectedAsset,
      onAssetDelete: deleteSelectedAsset,
      onAssetCheckout: checkoutSelectedAsset,
      onAssetReturn: returnSelectedAsset,
      onHomeAssetReturn: returnAssetFromHome,
      onAssetUploadAttachment: uploadSelectedAttachment,
      onAssetArchiveAttachment: archiveSelectedAttachment,
      onAttachmentDeleteOpen: openAttachmentDeleteRoute,
      onAttachmentDeleteClose: closeAttachmentDeleteRoute,
      onAssetDeleteAttachment: deleteSelectedAttachment,
      onAssetTagSearch: searchForTag,
      onSearch: search,
      onOpenSearchAsset: openSearchAsset,
      onImportSourceChange: openImportSource,
      onImportJobSelectionChange: openImportJobRoute,
      onImportJobTabChange: openImportJobTabRoute,
      onImportJobInventoryChanged: refreshInventoryAfterImportJob,
      onOpenImportedAssetId: openImportedAssetId,
      onOpenInventoryAuditHistory: openInventoryAuditHistory,
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
  assetTags={data.context.assetTags ?? []}
  saving={busy}
  {notification}
  {error}
  onAddClose={closeAdd}
  onAddSave={createAsset}
/>
