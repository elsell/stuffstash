<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { tick } from 'svelte';
  import Plus from '@lucide/svelte/icons/plus';
  import Search from '@lucide/svelte/icons/search';
  import { collectSettingsPages, filterSettingsRecords, invalidateSharedSettingsLoads, isSettingsPermissionDenied, mergeCanonicalSettingsRecord, removeCanonicalSettingsRecord, settingsKeyFromName, sortSettingsRecords } from '$lib/application/settingsManagement';
  import { settingsResourceHref } from '$lib/application/settingsManagementNavigation';
  import { safeWorkspaceErrorMessage } from '$lib/application/workspaceSafeError';
  import { hasAccessPermission, type CustomAssetType, type Inventory, type Tenant } from '$lib/domain/inventory';
  import type { InventoryCustomizationRepository } from '$lib/ports/inventoryCustomizationRepository';
  import type { SettingsResourceAction } from '$lib/application/workspaceRoute';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import WorkspaceTaskSheet from '../action-surface/WorkspaceTaskSheet.svelte';
  import WorkspaceConfirmationDialog from '../action-surface/WorkspaceConfirmationDialog.svelte';
  import SettingsCollectionState from './SettingsCollectionState.svelte';
  import { notifySuccess } from '$lib/components/ui/sonner/notifications';
  import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';

  let { level, tenant, inventory = null, repository, observer, canonicalItems, lifecycle = 'active', resourceId = null, action: routeAction = null, onNavigate, onSchemaChange, onPermissionDenied }:
    { level: 'tenant' | 'inventory'; tenant: Tenant; inventory?: Inventory | null; repository: InventoryCustomizationRepository; observer: WorkspaceObserver; canonicalItems: CustomAssetType[]; lifecycle?: 'active' | 'archived'; resourceId?: string | null; action?: SettingsResourceAction; onNavigate: (href: string) => void; onSchemaChange: (types: CustomAssetType[]) => void; onPermissionDenied: () => Promise<void> } = $props();

  let items = $state<CustomAssetType[]>([]); let loading = $state(true); let loadingMore = $state(false); let hasMore = $state(false); let nextCursor = $state<string | null>(null); let error = $state(''); let appendError = $state(''); let saving = $state(false); let query = $state(''); let loadEpoch = 0; let lastRequestKey = '';
  let expirationEnabled = $state(false); let displayName = $state(''); let key = $state(''); let keyManuallyEdited = $state(false); let description = $state(''); let formError = $state(''); let formErrorElement = $state<HTMLElement | null>(null); let initializedFor = $state(''); let discardOpen = $state(false);
  let selected = $derived(items.find((item) => item.id === resourceId) ?? null);
  let canManage = $derived(level === 'tenant' ? hasAccessPermission(tenant.access, 'configure') : hasAccessPermission(inventory?.access, 'configure'));
  let canManageTenant = $derived(hasAccessPermission(tenant.access, 'configure'));
  let ownedSelected = $derived(selected && selected.scope === level ? selected : null);
  let filtered = $derived(filterSettingsRecords(sortSettingsRecords(items), query));
  let inherited = $derived(level === 'inventory' ? filtered.filter((item) => item.scope === 'tenant') : []);
  let local = $derived(level === 'inventory' ? filtered.filter((item) => item.scope === 'inventory') : filtered);
  let formOpen = $derived(routeAction === 'new' || routeAction === 'edit');
  let confirmationOpen = $derived(routeAction === 'archive' || routeAction === 'restore' || routeAction === 'delete');
  let detailOpen = $derived(Boolean(resourceId) && routeAction === null);
  let contextName = $derived(level === 'tenant' ? tenant.name : inventory?.name ?? t('web.AssetTypeSettingsManager.inventoryFallback'));
  let collectionHref = $derived(href());
  let dirty = $derived(expirationEnabled !== (ownedSelected?.expirationEnabled ?? false) || displayName !== (ownedSelected?.displayName ?? '') || key !== (ownedSelected?.key ?? '') || description !== (ownedSelected?.description ?? ''));

  $effect(() => { const requestKey = `${level}:${tenant.id}:${inventory?.id ?? ''}:${lifecycle}`; if (requestKey === lastRequestKey) return; lastRequestKey = requestKey; const epoch = ++loadEpoch; void load(epoch); });
  $effect(() => { if (!formOpen) { initializedFor = ''; return; } if (routeAction === 'edit' && !ownedSelected) return; const marker = `${routeAction}:${resourceId ?? ''}:${ownedSelected?.displayName ?? ''}`; if (!formOpen || initializedFor === marker) return; initializedFor = marker; displayName = ownedSelected?.displayName ?? ''; key = ownedSelected?.key ?? ''; keyManuallyEdited = routeAction !== 'new'; description = ownedSelected?.description ?? ''; expirationEnabled = ownedSelected?.expirationEnabled ?? false; formError = ''; });
  $effect(() => { if (routeAction === 'new' && formOpen && !keyManuallyEdited) key = settingsKeyFromName(displayName); });

  function href(options: { lifecycle?: 'active' | 'archived'; resourceId?: string; action?: SettingsResourceAction } = {}): string {
    return settingsResourceHref({ level, tenantId: tenant.id, inventoryId: inventory?.id, collection: 'asset-types', lifecycle: options.lifecycle ?? lifecycle, resourceId: options.resourceId, action: options.action });
  }
  async function load(epoch = ++loadEpoch): Promise<void> {
    loading = true; error = '';
    observer.record('workspace.settings_collection_load_started', { resource: 'asset_type', scope: level });
    try {
      const page = level === 'tenant' ? await repository.listTenantCustomAssetTypes(tenant.id, undefined, lifecycle) : await repository.listInventoryCustomAssetTypes(tenant.id, inventory!.id, undefined, lifecycle);
      if (epoch !== loadEpoch) return;
      items = sortSettingsRecords(page.items); hasMore = page.pagination.hasMore; nextCursor = page.pagination.nextCursor; appendError = '';
      observer.record('workspace.settings_collection_loaded', { resource: 'asset_type', scope: level, count: items.length });
    } catch (caught) { if (epoch === loadEpoch) { error = safeWorkspaceErrorMessage(caught, t('web.AssetTypeSettingsManager.assetTypesCouldNotBeLoadedTryAgain')); observer.record('workspace.settings_collection_load_failed', { resource: 'asset_type', scope: level }); await handlePermissionDenied(caught, 'load'); } }
    finally { if (epoch === loadEpoch) loading = false; }
  }
  async function loadMore(): Promise<void> { if (!hasMore || !nextCursor || loadingMore) return; const epoch = loadEpoch; loadingMore = true; appendError = ''; try { const page = level === 'tenant' ? await repository.listTenantCustomAssetTypes(tenant.id, nextCursor, lifecycle) : await repository.listInventoryCustomAssetTypes(tenant.id, inventory!.id, nextCursor, lifecycle); if (epoch !== loadEpoch) return; items = sortSettingsRecords([...items, ...page.items]); hasMore = page.pagination.hasMore; nextCursor = page.pagination.nextCursor; } catch (caught) { if (epoch === loadEpoch) appendError = safeWorkspaceErrorMessage(caught, t('web.AssetTypeSettingsManager.moreAssetTypesCouldNotBeLoadedTryAgain')); } finally { if (epoch === loadEpoch) loadingMore = false; } }
  async function save(): Promise<void> {
    if (!canManage || saving) return;
    const name = displayName.trim(); const stableKey = key.trim();
    if (!name || (routeAction === 'new' && !/^[a-z][a-z0-9-]{0,79}$/.test(stableKey))) { formError = !name ? t('web.AssetTypeSettingsManager.enterADisplayName') : t('web.AssetTypeSettingsManager.keyMustStartWithALetterAndUseLowercase'); await tick(); formErrorElement?.focus(); return; }
    saving = true; formError = '';
    observer.record('workspace.settings_mutation_started', { resource: 'asset_type', action: routeAction === 'new' ? 'create' : 'update', scope: level });
    try {
      const saved = routeAction === 'new'
        ? await repository.createCustomAssetType(tenant.id, inventory?.id ?? '', { scope: level, key: stableKey, displayName: name, description: description.trim(), expirationEnabled })
        : ownedSelected ? await repository.updateCustomAssetType(tenant.id, inventory?.id ?? '', ownedSelected.id, level, { displayName: name, description: description.trim(), expirationEnabled }) : null;
      if (!saved) throw new Error('Asset type is unavailable.');
      invalidateSharedSettingsLoads(repository, 'custom-field-supporting-types:');
      items = sortSettingsRecords(routeAction === 'new' ? [...items, saved] : items.map((item) => item.id === saved.id ? saved : item));
      onSchemaChange(mergeCanonicalSettingsRecord(canonicalItems, saved)); notifySuccess(routeAction === 'new' ? t('web.AssetTypeSettingsManager.added') : t('web.AssetTypeSettingsManager.changesSaved'), { description: t('web.AssetTypeSettingsManager.isUpToDate', { displayName: String(saved.displayName) }) }); observer.record('workspace.settings_mutation_succeeded', { resource: 'asset_type', action: routeAction === 'new' ? 'create' : 'update', scope: level }); onNavigate(collectionHref);
    } catch (caught) { formError = safeWorkspaceErrorMessage(caught, t(`web.AssetTypeSettingsManager.${routeAction === 'new' ? 'create' : 'save'}Failed`)); observer.record('workspace.settings_mutation_failed', { resource: 'asset_type', action: routeAction === 'new' ? 'create' : 'update', scope: level }); await handlePermissionDenied(caught, routeAction === 'new' ? 'create' : 'update'); await tick(); formErrorElement?.focus(); }
    finally { saving = false; }
  }
  async function lifecycleAction(): Promise<void> {
    if (!ownedSelected || !canManage || saving || !routeAction) return;
    const selectedRecord = ownedSelected;
    const action = routeAction;
    const destination = collectionHref;
    saving = true; formError = '';
    observer.record('workspace.settings_mutation_started', { resource: 'asset_type', action, scope: level });
    try {
      let changed: CustomAssetType | null = null;
      if (action === 'archive') changed = await repository.archiveCustomAssetType(tenant.id, inventory?.id ?? '', selectedRecord.id, level);
      else if (action === 'restore') changed = await repository.restoreCustomAssetType(tenant.id, inventory?.id ?? '', selectedRecord.id, level);
      else if (action === 'delete') await repository.deleteCustomAssetType(tenant.id, inventory?.id ?? '', selectedRecord.id, level);
      invalidateSharedSettingsLoads(repository, 'custom-field-supporting-types:');
      items = items.filter((item) => item.id !== selectedRecord.id);
      const nextCanonical = action === 'archive' ? removeCanonicalSettingsRecord(canonicalItems, selectedRecord.id) : action === 'restore' && changed ? mergeCanonicalSettingsRecord(canonicalItems, changed) : null;
      observer.record('workspace.settings_mutation_succeeded', { resource: 'asset_type', action, scope: level });
      onNavigate(destination);
      if (nextCanonical) onSchemaChange(nextCanonical);
    } catch (caught) { formError = safeWorkspaceErrorMessage(caught, t(`web.AssetTypeSettingsManager.${action === 'archive' || action === 'restore' || action === 'delete' ? action : 'change'}Failed.${lifecycle}`)); observer.record('workspace.settings_mutation_failed', { resource: 'asset_type', action, scope: level }); await handlePermissionDenied(caught, action); }
    finally { saving = false; }
  }
  function requestClose(): void { if (dirty) { discardOpen = true; return; } onNavigate(collectionHref); }
  async function handlePermissionDenied(caught: unknown, action: string): Promise<void> { if (!isSettingsPermissionDenied(caught)) return; observer.record('workspace.settings_permission_denied', { resource: 'asset_type', action, scope: level }); await onPermissionDenied(); }
  function openRow(item: CustomAssetType): string { return item.scope === level ? href({ resourceId: item.id, ...(item.lifecycleState === 'active' ? { action: 'edit' as const } : {}) }) : href({ resourceId: item.id }); }
  function manageInheritedHref(item: CustomAssetType): string { return settingsResourceHref({ level: 'tenant', tenantId: tenant.id, collection: 'asset-types', resourceId: item.id, ...(item.lifecycleState === 'active' ? { action: 'edit' as const } : {}) }); }
</script>

<section class="settings-resource-page" aria-labelledby="asset-type-settings-title">
  <header class="settings-resource-header"><div><p class="settings-eyebrow">{contextName}</p><h1 id="asset-type-settings-title">{t('web.AssetTypeSettingsManager.assetTypes')}</h1><p>{level === 'tenant' ? t('web.AssetTypeSettingsManager.typesSharedWithEveryInventory') : t('web.AssetTypeSettingsManager.inheritedAndInventoryOnlyClassifications')}</p></div>{#if canManage && lifecycle === 'active'}<Button.Root href={href({ action: 'new' })} onclick={(event) => { event.preventDefault(); onNavigate(href({ action: 'new' })); }}><Plus /> {t('web.AssetTypeSettingsManager.addAssetType')}</Button.Root>{/if}</header>
  <nav class="settings-lifecycle-nav" aria-label={t('web.AssetTypeSettingsManager.assetTypeLifecycle')}><a href={href({ lifecycle: 'active' })} aria-current={lifecycle === 'active' ? 'page' : undefined} onclick={(event) => { event.preventDefault(); onNavigate(href({ lifecycle: 'active' })); }}>{t('web.AssetTypeSettingsManager.active')}</a><a href={href({ lifecycle: 'archived' })} aria-current={lifecycle === 'archived' ? 'page' : undefined} onclick={(event) => { event.preventDefault(); onNavigate(href({ lifecycle: 'archived' })); }}>{t('web.AssetTypeSettingsManager.archived')}</a></nav>
  {#if loading && items.length === 0}<SettingsCollectionState kind="loading" title={t(`web.AssetTypeSettingsManager.loadingAssetTypes.${lifecycle}`)} message={t('web.AssetTypeSettingsManager.gettingTheCurrentSchema')} />
  {:else}
    {#if loading}<SettingsCollectionState kind="loading" title={t('web.AssetTypeSettingsManager.updatingAssetTypes')} message={t('web.AssetTypeSettingsManager.showingTheLastLoadedRowsWhileThisViewRefreshes')} />{/if}
    {#if error}<SettingsCollectionState kind="error" title={t('web.AssetTypeSettingsManager.assetTypesUnavailable')} message={error} onRetry={() => { void load(); }} />{/if}
    {#if items.length > 12}<div class="settings-search"><Search aria-hidden="true" /><Label class="visually-hidden" for="asset-type-search">{t('web.AssetTypeSettingsManager.searchAssetTypes')}</Label><Input id="asset-type-search" type="search" bind:value={query} placeholder={t('web.AssetTypeSettingsManager.searchAssetTypes')} /></div>{/if}
    {#if filtered.length === 0}<SettingsCollectionState kind="empty" title={query ? t('web.AssetTypeSettingsManager.noMatchingAssetTypes') : t(`web.AssetTypeSettingsManager.noAssetTypes.${lifecycle}`)} message={query ? t('web.AssetTypeSettingsManager.tryADifferentNameOrKey') : lifecycle === 'active' && canManage ? t('web.AssetTypeSettingsManager.addAnAssetTypeWhenYouNeedAReusable') : t(`web.AssetTypeSettingsManager.thereAreNoAssetTypesHere.${lifecycle}`)} />
    {:else}
      {#each level === 'inventory' ? [{ title: t('customization.fromHousehold', { household: tenant.name }), rows: inherited }, { title: t('customization.onlyIn', { inventory: String(inventory?.name) }), rows: local }] : [{ title: '', rows: local }] as group}
        {#if group.rows.length}<section class="settings-resource-group" aria-label={group.title || t('web.AssetTypeSettingsManager.assetTypes2', { contextName: String(contextName) })}>{#if group.title}<h2>{group.title}</h2>{/if}<div class="settings-resource-list">{#each group.rows as item}<a class="settings-resource-row" href={openRow(item)} onclick={(event) => { event.preventDefault(); onNavigate(openRow(item)); }}><span><strong>{item.displayName}</strong>{#if item.description}<small>{item.description}</small>{/if}<small>{item.key}</small></span><span class="settings-resource-meta">{#if item.scope === 'tenant' && level === 'inventory'}<Badge variant="outline">{t('web.AssetTypeSettingsManager.inherited')}</Badge>{/if}{#if lifecycle === 'archived'}<Badge variant="secondary">{t('web.AssetTypeSettingsManager.archived')}</Badge>{/if}</span></a>{/each}</div></section>{/if}
      {/each}
    {/if}
    {#if appendError}<SettingsCollectionState kind="error" title={t('web.AssetTypeSettingsManager.couldNotLoadMore')} message={appendError} onRetry={() => { void loadMore(); }} />{:else if hasMore}<div class="settings-pagination"><Button.Root variant="outline" disabled={loadingMore} onclick={() => { void loadMore(); }}>{loadingMore ? t('web.AssetTypeSettingsManager.loading') : t('web.AssetTypeSettingsManager.loadMore')}</Button.Root><small>{t('web.AssetTypeSettingsManager.moreAssetTypesAreAvailable')}</small></div>{/if}
  {/if}
</section>

<WorkspaceTaskSheet open={formOpen} title={routeAction === 'new' ? t('web.AssetTypeSettingsManager.addAssetType') : ownedSelected ? t('web.AssetTypeSettingsManager.edit', { displayName: String(ownedSelected.displayName) }) : t('web.AssetTypeSettingsManager.assetTypeUnavailable')} description={t('web.AssetTypeSettingsManager.managedIn', { contextName: String(contextName) })} busy={saving} closeHref={collectionHref} onCloseLink={(event) => { event.preventDefault(); requestClose(); }} onOpenChange={(open) => { if (!open && !saving) requestClose(); }}>
  {#if !canManage}<SettingsCollectionState kind="denied" title={t('web.AssetTypeSettingsManager.readOnly')} message={t('web.AssetTypeSettingsManager.thisAccountCanViewAssetTypesButCannotChange')} />
  {:else if routeAction === 'edit' && !ownedSelected}<SettingsCollectionState kind="error" title={t('web.AssetTypeSettingsManager.assetTypeUnavailable')} message={t('web.AssetTypeSettingsManager.thisRecordMayBeInheritedArchivedOrNoLonger')} />
  {:else}{#if formError}<p class="settings-form-error" role="alert" tabindex="-1" bind:this={formErrorElement}>{formError}</p>{/if}<div class="field-stack"><Label for="asset-type-name">{t('web.AssetTypeSettingsManager.displayName')}</Label><Input id="asset-type-name" bind:value={displayName} maxlength={120} /></div><div class="field-stack"><Label for="asset-type-key">{t('web.AssetTypeSettingsManager.stableKey')}</Label><Input id="asset-type-key" value={key} readonly={routeAction !== 'new'} aria-describedby="asset-type-key-help" oninput={(event) => { key = event.currentTarget.value; keyManuallyEdited = true; }} /><small id="asset-type-key-help">{t('web.AssetTypeSettingsManager.theKeyCannotChangeAfterCreation')}</small></div><div class="field-stack"><Label for="asset-type-expiration"><Checkbox id="asset-type-expiration" bind:checked={expirationEnabled} disabled={saving} aria-describedby="asset-type-expiration-help" /> {t('web.AssetTypeSettingsManager.trackExpirationDates')}</Label><small id="asset-type-expiration-help">{t('web.AssetTypeSettingsManager.eachAssetCanHaveItsOwnOptionalExpirationDate')}</small></div><div class="field-stack"><Label for="asset-type-description">{t('web.AssetTypeSettingsManager.descriptionOptional')}</Label><Textarea id="asset-type-description" bind:value={description} maxlength={1000} /></div>{/if}
  {#snippet footer()}<Button.Root variant="outline" disabled={saving} onclick={requestClose}>{t('web.AssetTypeSettingsManager.cancel')}</Button.Root>{#if routeAction === 'edit' && ownedSelected}<Button.Root variant="destructive" href={href({ resourceId: ownedSelected.id, action: 'archive' })} onclick={(event) => { event.preventDefault(); onNavigate(href({ resourceId: ownedSelected.id, action: 'archive' })); }}>{t('web.AssetTypeSettingsManager.archive')}</Button.Root>{/if}<Button.Root disabled={saving || !canManage || !dirty || !displayName.trim() || (routeAction === 'new' && !key.trim())} onclick={() => { void save(); }}>{t('web.AssetTypeSettingsManager.save')}</Button.Root>{/snippet}
</WorkspaceTaskSheet>

<WorkspaceConfirmationDialog open={discardOpen} title={t('web.AssetTypeSettingsManager.discardChanges')} description={t('web.AssetTypeSettingsManager.yourUnsavedAssetTypeChangesWillBeLost')} onOpenChange={(open) => { discardOpen = open; }}>
  {#snippet cancel()}<Button.Root variant="outline" onclick={() => { discardOpen = false; }}>{t('web.AssetTypeSettingsManager.keepEditing')}</Button.Root>{/snippet}
  {#snippet action()}<Button.Root variant="destructive" onclick={() => { discardOpen = false; onNavigate(collectionHref); }}>{t('web.AssetTypeSettingsManager.discardChanges2')}</Button.Root>{/snippet}
</WorkspaceConfirmationDialog>

<WorkspaceTaskSheet open={detailOpen} title={selected?.displayName ?? t('web.AssetTypeSettingsManager.assetType')} description={selected?.scope === 'tenant' && level === 'inventory' ? t('web.AssetTypeSettingsManager.inheritedFrom2', { name: String(tenant.name) }) : t(`web.AssetTypeSettingsManager.${selected?.lifecycleState === 'archived' ? 'archivedIn' : 'managedIn'}`, { contextName })} closeHref={collectionHref} onCloseLink={(event) => { event.preventDefault(); onNavigate(collectionHref); }} onOpenChange={(open) => { if (!open) onNavigate(collectionHref); }}>
  {#if !selected}<SettingsCollectionState kind="error" title={t('web.AssetTypeSettingsManager.assetTypeUnavailable')} message={t('web.AssetTypeSettingsManager.thisRecordMayNoLongerExistOrMayNot')} />
  {:else}<dl class="settings-readonly-details"><div><dt>{t('web.AssetTypeSettingsManager.stableKey')}</dt><dd>{selected.key}</dd></div><div><dt>{t('web.AssetTypeSettingsManager.description')}</dt><dd>{selected.description || t('web.AssetTypeSettingsManager.none')}</dd></div><div><dt>{t('web.AssetTypeSettingsManager.expirationTracking')}</dt><dd>{selected.expirationEnabled ? t('web.AssetTypeSettingsManager.on') : t('web.AssetTypeSettingsManager.off')}</dd></div><div><dt>{t('web.AssetTypeSettingsManager.ownership')}</dt><dd>{selected.scope === 'tenant' ? t('web.AssetTypeSettingsManager.inheritedFrom', { name: String(tenant.name) }) : t('web.AssetTypeSettingsManager.onlyIn', { contextName: String(contextName) })}</dd></div></dl>{/if}
  {#snippet footer()}<Button.Root variant="outline" onclick={() => onNavigate(collectionHref)}>{t('web.AssetTypeSettingsManager.done')}</Button.Root>{#if selected?.scope === 'tenant' && level === 'inventory' && canManageTenant}<Button.Root href={manageInheritedHref(selected)} onclick={(event) => { event.preventDefault(); onNavigate(manageInheritedHref(selected)); }}>{t('web.AssetTypeSettingsManager.manageInFull', { name: tenant.name })}</Button.Root>{:else if ownedSelected && ownedSelected.lifecycleState === 'archived' && canManage}<Button.Root href={href({ resourceId: ownedSelected.id, action: 'delete' })} variant="destructive" onclick={(event) => { event.preventDefault(); onNavigate(href({ resourceId: ownedSelected.id, action: 'delete' })); }}>{t('web.AssetTypeSettingsManager.deletePermanently')}</Button.Root><Button.Root href={href({ resourceId: ownedSelected.id, action: 'restore' })} onclick={(event) => { event.preventDefault(); onNavigate(href({ resourceId: ownedSelected.id, action: 'restore' })); }}>{t('web.AssetTypeSettingsManager.restore')}</Button.Root>{/if}{/snippet}
</WorkspaceTaskSheet>

<WorkspaceConfirmationDialog open={confirmationOpen} title={routeAction === 'delete' ? t('web.AssetTypeSettingsManager.deleteAssetTypePermanently') : routeAction === 'restore' ? t('web.AssetTypeSettingsManager.restoreAssetType') : t('web.AssetTypeSettingsManager.archiveAssetType')} description={routeAction === 'delete' ? t('web.AssetTypeSettingsManager.thisCannotBeUndoneDeletionIsBlockedWhileActive') : routeAction === 'restore' ? t('web.AssetTypeSettingsManager.thisTypeWillBecomeAvailableForNewAssignmentsAnd') : t('web.AssetTypeSettingsManager.existingAssetAndFieldReferencesRemainButNewAssignment')} busy={saving} onOpenChange={(open) => { if (!open && !saving) onNavigate(collectionHref); }}>
  {#if formError}<p class="settings-form-error" role="alert">{formError}</p>{/if}
  {#if ownedSelected}<p><strong>{ownedSelected.displayName}</strong></p>{/if}
  {#snippet cancel()}<Button.Root variant="outline" disabled={saving} onclick={() => onNavigate(collectionHref)}>{t('web.AssetTypeSettingsManager.cancel')}</Button.Root>{/snippet}
  {#snippet action()}<Button.Root variant={routeAction === 'delete' || routeAction === 'archive' ? 'destructive' : 'default'} disabled={saving || !ownedSelected || !canManage} onclick={() => { void lifecycleAction(); }}>{routeAction === 'delete' ? t('web.AssetTypeSettingsManager.deletePermanently') : routeAction === 'restore' ? t('web.AssetTypeSettingsManager.restore') : t('web.AssetTypeSettingsManager.archive')}</Button.Root>{/snippet}
</WorkspaceConfirmationDialog>
