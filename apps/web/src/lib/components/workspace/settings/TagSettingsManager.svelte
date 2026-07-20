<script lang="ts">
  import { tick } from 'svelte';
  import Plus from '@lucide/svelte/icons/plus';
  import Check from '@lucide/svelte/icons/check';
  import Search from '@lucide/svelte/icons/search';
  import { safeWorkspaceErrorMessage } from '$lib/application/workspaceSafeError';
  import { filterSettingsRecords, isSettingsPermissionDenied, normalizeTagColor, sortSettingsRecords, tagColorAccessibleLabel, utf8ByteLength } from '$lib/application/settingsManagement';
  import { settingsResourceHref } from '$lib/application/settingsManagementNavigation';
  import { hasAccessPermission, type Inventory, type ManagedAssetTag } from '$lib/domain/inventory';
  import type { InventoryTagRepository } from '$lib/ports/inventoryTagRepository';
  import type { SettingsResourceAction } from '$lib/application/workspaceRoute';
  import * as Button from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import WorkspaceTaskSheet from '../action-surface/WorkspaceTaskSheet.svelte';
  import WorkspaceConfirmationDialog from '../action-surface/WorkspaceConfirmationDialog.svelte';
  import SettingsCollectionState from './SettingsCollectionState.svelte';
  import { notifySuccess } from '$lib/components/ui/sonner/notifications';
  import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';

  let { inventory, repository, observer, resourceId = null, action = null, onNavigate, onTagsChange, onPermissionDenied }:
    { inventory: Inventory; repository: InventoryTagRepository; observer: WorkspaceObserver; resourceId?: string | null; action?: SettingsResourceAction; onNavigate: (href: string) => void; onTagsChange: (tags: ManagedAssetTag[]) => void; onPermissionDenied: () => Promise<void> } = $props();

  let tags = $state<ManagedAssetTag[]>([]);
  let loading = $state(true);
  let loadingMore = $state(false);
  let hasMore = $state(false);
  let nextCursor = $state<string | null>(null);
  let error = $state('');
  let appendError = $state('');
  let saving = $state(false);
  let query = $state('');
  let displayName = $state('');
  let color = $state('');
  let formError = $state('');
  let formErrorElement = $state<HTMLElement | null>(null);
  let initializedFor = $state('');
  let discardOpen = $state(false);
  let loadEpoch = 0;
  let lastRequestKey = '';
  let selected = $derived(tags.find((tag) => tag.id === resourceId) ?? null);
  let canManage = $derived(hasAccessPermission(inventory.access, 'edit_asset'));
  let filtered = $derived(filterSettingsRecords(sortSettingsRecords(tags), query));
  let formOpen = $derived(action === 'new' || action === 'edit');
  let archiveOpen = $derived(action === 'archive');
  let collectionHref = $derived(settingsResourceHref({ level: 'inventory', tenantId: inventory.tenantId, inventoryId: inventory.id, collection: 'tags' }));
  let initialName = $derived(selected?.displayName ?? '');
  let initialColor = $derived(selected?.color ?? '');
  let normalizedFormColor = $derived(normalizeTagColor(color));
  let nameBytes = $derived(utf8ByteLength(displayName.trim()));
  let nameByteError = $derived(nameBytes > 80 ? 'Tag name must be 80 UTF-8 bytes or fewer.' : '');
  let formValid = $derived(Boolean(displayName.trim()) && nameBytes <= 80 && normalizedFormColor !== null);
  let dirty = $derived(displayName !== initialName || (normalizedFormColor ?? '') !== initialColor);

  $effect(() => { const requestKey = `${inventory.tenantId}:${inventory.id}`; if (requestKey === lastRequestKey) return; lastRequestKey = requestKey; const epoch = ++loadEpoch; void load(epoch); });
  $effect(() => {
    if (action === 'edit' && !selected) return;
    const key = `${action}:${resourceId ?? ''}:${selected?.updatedAt ?? ''}`;
    if (!formOpen || initializedFor === key) return;
    initializedFor = key;
    displayName = selected?.displayName ?? '';
    color = selected?.color ?? '';
    formError = '';
  });

  async function load(epoch = ++loadEpoch): Promise<void> {
    loading = true; error = '';
    observer.record('workspace.settings_collection_load_started', { resource: 'tag', scope: 'inventory' });
    try {
      const page = await repository.listManagedAssetTags(inventory.tenantId, inventory.id);
      if (epoch !== loadEpoch) return;
      tags = sortSettingsRecords(page.items); hasMore = page.pagination.hasMore; nextCursor = page.pagination.nextCursor; appendError = '';
      observer.record('workspace.settings_collection_loaded', { resource: 'tag', scope: 'inventory', count: tags.length });
    } catch (caught) {
      if (epoch === loadEpoch) { error = safeWorkspaceErrorMessage(caught, 'Tags could not be loaded. Try again.'); observer.record('workspace.settings_collection_load_failed', { resource: 'tag', scope: 'inventory' }); await handlePermissionDenied(caught, 'load'); }
    } finally { if (epoch === loadEpoch) loading = false; }
  }
  async function loadMore(): Promise<void> { if (!hasMore || !nextCursor || loadingMore) return; const epoch = loadEpoch; loadingMore = true; appendError = ''; try { const page = await repository.listManagedAssetTags(inventory.tenantId, inventory.id, nextCursor); if (epoch !== loadEpoch) return; tags = sortSettingsRecords([...tags, ...page.items]); hasMore = page.pagination.hasMore; nextCursor = page.pagination.nextCursor; } catch (caught) { if (epoch === loadEpoch) appendError = safeWorkspaceErrorMessage(caught, 'More tags could not be loaded. Try again.'); } finally { if (epoch === loadEpoch) loadingMore = false; } }

  function route(actionValue: SettingsResourceAction, id?: string): string {
    return settingsResourceHref({ level: 'inventory', tenantId: inventory.tenantId, inventoryId: inventory.id, collection: 'tags', resourceId: id, action: actionValue });
  }

  async function save(): Promise<void> {
    if (saving || !canManage) return;
    const name = displayName.trim();
    const normalizedColor = normalizeTagColor(color);
    if (!name || utf8ByteLength(name) > 80 || normalizedColor === null) {
      formError = !name ? 'Enter a tag name.' : utf8ByteLength(name) > 80 ? 'Tag name must be 80 UTF-8 bytes or fewer.' : 'Enter a six-digit hex color such as #2F80ED.';
      await tick(); formErrorElement?.focus(); return;
    }
    saving = true; formError = '';
    observer.record('workspace.settings_mutation_started', { resource: 'tag', action: action === 'new' ? 'create' : 'update', scope: 'inventory' });
    try {
      const saved = action === 'new'
        ? await repository.createManagedAssetTag(inventory.tenantId, inventory.id, { displayName: name, ...(normalizedColor ? { color: normalizedColor } : {}) })
        : selected
          ? await repository.updateManagedAssetTag(inventory.tenantId, inventory.id, selected.id, { displayName: name, color: normalizedColor ?? '' })
          : null;
      if (!saved) throw new Error('Tag is unavailable.');
      tags = sortSettingsRecords(action === 'new' ? [...tags, saved] : tags.map((tag) => tag.id === saved.id ? saved : tag));
      onTagsChange(tags);
      notifySuccess(action === 'new' ? 'Tag added' : 'Changes saved', { description: `${saved.displayName} is up to date.` });
      observer.record('workspace.settings_mutation_succeeded', { resource: 'tag', action: action === 'new' ? 'create' : 'update', scope: 'inventory' });
      onNavigate(collectionHref);
    } catch (caught) {
      formError = safeWorkspaceErrorMessage(caught, `Tag was not ${action === 'new' ? 'created' : 'saved'}. Try again.`);
      observer.record('workspace.settings_mutation_failed', { resource: 'tag', action: action === 'new' ? 'create' : 'update', scope: 'inventory' });
      await handlePermissionDenied(caught, action === 'new' ? 'create' : 'update');
      await tick(); formErrorElement?.focus();
    } finally { saving = false; }
  }

  async function archive(): Promise<void> {
    if (!selected || saving || !canManage) return;
    saving = true; formError = '';
    observer.record('workspace.settings_mutation_started', { resource: 'tag', action: 'archive', scope: 'inventory' });
    try {
      await repository.archiveManagedAssetTag(inventory.tenantId, inventory.id, selected.id);
      tags = tags.filter((tag) => tag.id !== selected.id);
      onTagsChange(tags);
      observer.record('workspace.settings_mutation_succeeded', { resource: 'tag', action: 'archive', scope: 'inventory' });
      onNavigate(collectionHref);
    } catch (caught) {
      formError = safeWorkspaceErrorMessage(caught, 'Tag was not archived. Try again.');
      observer.record('workspace.settings_mutation_failed', { resource: 'tag', action: 'archive', scope: 'inventory' });
      await handlePermissionDenied(caught, 'archive');
    } finally { saving = false; }
  }

  function requestClose(): void { if (dirty) { discardOpen = true; return; } onNavigate(collectionHref); }
  async function handlePermissionDenied(caught: unknown, actionValue: string): Promise<void> { if (!isSettingsPermissionDenied(caught)) return; observer.record('workspace.settings_permission_denied', { resource: 'tag', action: actionValue, scope: 'inventory' }); await onPermissionDenied(); }
</script>

<section class="settings-resource-page" aria-labelledby="tag-settings-title">
  <header class="settings-resource-header">
    <div><p class="settings-eyebrow">{inventory.name}</p><h1 id="tag-settings-title">Tags</h1><p>Reusable labels for this inventory.</p></div>
    {#if canManage}<Button.Root href={route('new')} onclick={(event) => { event.preventDefault(); onNavigate(route('new')); }}><Plus /> Add Tag</Button.Root>{/if}
  </header>
  {#if loading && tags.length === 0}<SettingsCollectionState kind="loading" title="Loading tags" message="Getting the current tag list." />
  {:else}
    {#if loading}<SettingsCollectionState kind="loading" title="Updating tags" message="Showing the last loaded tags while this list refreshes." />{/if}
    {#if error}<SettingsCollectionState kind="error" title="Tags unavailable" message={error} onRetry={() => { void load(); }} />{/if}
    {#if tags.length > 12}
      <div class="settings-search"><Search aria-hidden="true" /><Label class="visually-hidden" for="tag-settings-search">Search tags</Label><Input id="tag-settings-search" type="search" bind:value={query} placeholder="Search tags" /></div>
    {/if}
    {#if filtered.length === 0}
      <SettingsCollectionState kind="empty" title={query ? 'No matching tags' : 'No active tags'} message={query ? 'Try a different name.' : canManage ? 'Add a tag to organize assets and filters.' : 'No tags are available in this inventory.'} />
    {:else}
      <div class="settings-resource-list" aria-label="Active tags">
        {#each filtered as tag}
          <a class="settings-resource-row settings-tag-resource-row" href={route('edit', tag.id)} aria-label={`${tag.displayName} ${tagColorAccessibleLabel(tag.color)}`} onclick={(event) => { event.preventDefault(); onNavigate(route('edit', tag.id)); }}>
            <span class="settings-tag-row-primary">
              <span class="settings-tag-color-indicator" class:settings-tag-color-empty={!tag.color} style={tag.color ? `--tag-color: ${tag.color}` : undefined} aria-hidden="true"></span>
              <strong>{tag.displayName}</strong>
            </span>
          </a>
        {/each}
      </div>
    {/if}
    {#if appendError}<SettingsCollectionState kind="error" title="Could not load more" message={appendError} onRetry={() => { void loadMore(); }} />{:else if hasMore}<div class="settings-pagination"><Button.Root variant="outline" disabled={loadingMore} onclick={() => { void loadMore(); }}>{loadingMore ? 'Loading…' : 'Load more'}</Button.Root><small>More tags are available.</small></div>{/if}
  {/if}
</section>

<WorkspaceTaskSheet open={formOpen} title={action === 'new' ? 'Add Tag' : selected ? `Edit ${selected.displayName}` : 'Tag unavailable'} description="Tags belong only to this inventory." busy={saving} dismissible={!saving} closeHref={collectionHref} onCloseLink={(event) => { event.preventDefault(); requestClose(); }} onOpenChange={(open) => { if (!open && !saving) requestClose(); }}>
  {#if !canManage}<SettingsCollectionState kind="denied" title="Read only" message="Your draft is preserved, but this account can no longer change tags here." />{/if}
  {#if action === 'edit' && !selected}<SettingsCollectionState kind="error" title="Tag unavailable" message="This tag may have been archived or is no longer available." />
  {:else}
    {#if formError}<p class="settings-form-error" role="alert" tabindex="-1" bind:this={formErrorElement}>{formError}</p>{/if}
    <div class="field-stack"><Label for="settings-tag-name">Display name</Label><Input id="settings-tag-name" bind:value={displayName} maxlength={80} disabled={!canManage} aria-invalid={!displayName.trim() || nameByteError ? 'true' : undefined} aria-describedby="settings-tag-name-help" /><small id="settings-tag-name-help" class:settings-field-error={Boolean(nameByteError)}>{nameByteError || `${nameBytes} of 80 UTF-8 bytes`}</small></div>
    <fieldset class="settings-color-picker" disabled={!canManage}><legend>Color</legend>
      <div class="settings-color-swatches">
        {#each ['#2F80ED', '#6B90AA', '#F5AB4B', '#2E7D32', '#7C3AED', ''] as swatch}
          <Button.Root type="button" variant="outline" class="settings-color-swatch" style={swatch ? `--swatch: ${swatch}` : undefined} aria-label={swatch ? `Use color ${swatch}` : 'Use no color'} aria-pressed={color.toUpperCase() === swatch} onclick={() => { color = swatch; }}>
            <span style={swatch ? `background:${swatch}` : undefined}></span>{#if color.toUpperCase() === swatch}<Check aria-hidden="true" />{/if}
          </Button.Root>
        {/each}
      </div>
      <div class="settings-native-color-row">
        <div class="field-stack"><Label for="settings-tag-native-color">Choose a custom color</Label><Input id="settings-tag-native-color" class="settings-native-color-input" type="color" value={normalizeTagColor(color) || '#2F80ED'} oninput={(event) => { color = event.currentTarget.value.toUpperCase(); }} /></div>
        <Button.Root type="button" variant="outline" disabled={!color} onclick={() => { color = ''; }}>Clear color</Button.Root>
      </div>
      <div class="field-stack"><Label for="settings-tag-color">Hex color (optional)</Label><Input id="settings-tag-color" bind:value={color} placeholder="#2F80ED" aria-describedby="settings-tag-color-help" /><small id="settings-tag-color-help">Leave blank for no color.</small></div>
    </fieldset>
  {/if}
  {#snippet footer()}
    <Button.Root variant="outline" disabled={saving} onclick={requestClose}>Cancel</Button.Root>
    {#if action === 'edit' && selected}<Button.Root variant="destructive" disabled={saving} href={route('archive', selected.id)} onclick={(event) => { event.preventDefault(); onNavigate(route('archive', selected.id)); }}>Archive</Button.Root>{/if}
    <Button.Root disabled={saving || !canManage || !formValid || (action === 'edit' && !dirty)} onclick={() => { void save(); }}>Save</Button.Root>
  {/snippet}
</WorkspaceTaskSheet>

<WorkspaceConfirmationDialog open={discardOpen} title="Discard changes?" description="Your unsaved tag changes will be lost." onOpenChange={(open) => { discardOpen = open; }}>
  {#snippet cancel()}<Button.Root variant="outline" onclick={() => { discardOpen = false; }}>Keep editing</Button.Root>{/snippet}
  {#snippet action()}<Button.Root variant="destructive" onclick={() => { discardOpen = false; onNavigate(collectionHref); }}>Discard changes</Button.Root>{/snippet}
</WorkspaceConfirmationDialog>

<WorkspaceConfirmationDialog open={archiveOpen} title="Archive tag" description={selected ? `${selected.displayName} will no longer be available for new assignments or normal filtering. Existing history remains.` : 'This tag is no longer available.'} busy={saving} onOpenChange={(open) => { if (!open && !saving) onNavigate(collectionHref); }}>
  {#if formError}<p class="settings-form-error" role="alert">{formError}</p>{/if}
  {#snippet cancel()}<Button.Root variant="outline" disabled={saving} onclick={() => onNavigate(collectionHref)}>Cancel</Button.Root>{/snippet}
  {#snippet action()}<Button.Root variant="destructive" disabled={saving || !selected || !canManage} onclick={() => { void archive(); }}>Archive</Button.Root>{/snippet}
</WorkspaceConfirmationDialog>
