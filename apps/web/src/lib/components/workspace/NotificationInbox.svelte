<script lang="ts">
  import AssetLocationTrail from './AssetLocationTrail.svelte';
  import {formatAssetExpiration} from '$lib/application/expirationPresentation';
  import { onMount } from 'svelte';
  import type { ExpirationNotification } from '$lib/domain/notification';
  import type { NotificationRepository } from '$lib/ports/notificationRepository';
  import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
  import { markAllNotificationsRead } from '$lib/application/notificationInboxBatch';
  import { loadVisibleNotificationPage, openNotification, setNotificationRead } from '$lib/application/notificationInbox';
  import { safeWorkspaceErrorMessage } from '$lib/application/workspaceSafeError';
  import * as Button from '$lib/components/ui/button/index.js';
  import SegmentedControl from './SegmentedControl.svelte';

  let { tenantId, inventoryId, repository, observer, onOpenAsset, onRead = () => {}, onOpenSettings }: {
    tenantId: string; inventoryId: string; repository: NotificationRepository; observer: WorkspaceObserver;
    onOpenAsset: (assetId: string) => void; onRead?: () => void; onOpenSettings?: () => void;
  } = $props();
  let items = $state<ExpirationNotification[]>([]);
  let readIds = $state(new Set<string>());
  let filter = $state('all');
  let loading = $state(true);
  let appendLoading = $state(false);
  let error = $state('');
  let appendError = $state('');
  let openError = $state('');
  let marking = $state(false);
  let markingController: AbortController | undefined;
  let opening = $state<string | null>(null);
  let nextCursor = $state<string | null>(null);
  let hasMore = $state(false);
  let loadController: AbortController | undefined;
  let openController: AbortController | undefined;
  onMount(() => { void load(); return () => { loadController?.abort(); openController?.abort(); markingController?.abort(); }; });

  function dateLabel(item: ExpirationNotification): string { return formatAssetExpiration(item.expiration); }
  async function load(append = false) {
    loadController?.abort();
    const controller = new AbortController(); loadController = controller;
    if (append) { appendLoading = true; appendError = ''; }
    else { loading = true; appendLoading = false; error = ''; appendError = ''; }
    try {
      const page = await loadVisibleNotificationPage(repository, observer, tenantId, inventoryId, { cursor: append ? nextCursor ?? undefined : undefined, unreadOnly: filter === 'unread', signal: controller.signal });
      if (controller.signal.aborted) return;
      items = append ? Array.from(new Map([...items, ...page.items].map((item) => [item.id, item])).values()) : page.items;
      readIds = new Set([...readIds].filter(id => !page.items.some(item => item.id === id)));
      hasMore = page.pagination.hasMore; nextCursor = page.pagination.nextCursor;
    } catch (caught) {
      if (controller.signal.aborted) return;
      const message = safeWorkspaceErrorMessage(caught, 'Notifications could not be loaded. Try again.');
      if (append) appendError = message; else error = message;
    } finally { if (!controller.signal.aborted) { loading = false; appendLoading = false; } }
  }
  async function markAll() {
    if (marking || opening) return;
    loadController?.abort(); appendLoading = false;
    const controller = new AbortController(); markingController = controller;
    marking = true; openError = '';
    try {
      await markAllNotificationsRead(repository, observer, tenantId, inventoryId, controller.signal);
      if (controller.signal.aborted) return;
      onRead();
      await load();
    } catch (caught) {
      if (!controller.signal.aborted) openError = safeWorkspaceErrorMessage(caught, 'Not all notifications could be marked read. Try again.');
    } finally { if (!controller.signal.aborted) marking = false; }
  }
  async function open(item: ExpirationNotification) {
    if (opening || marking) return;
    const controller = new AbortController(); openController = controller;
    opening = item.id; openError = '';
    try {
      const assetId = await openNotification(repository, observer, tenantId, inventoryId, item.id, controller.signal);
      if (controller.signal.aborted) return;
      readIds = new Set([...readIds, item.id]);
      if (filter === 'unread') items = items.filter((value) => value.id !== item.id);
      onRead(); onOpenAsset(assetId);
    } catch (caught) {
      if (!controller.signal.aborted) openError = safeWorkspaceErrorMessage(caught, 'This notification could not be opened. Refresh to check whether it is still available.');
    } finally { if (!controller.signal.aborted) opening = null; }
  }
  async function toggleRead(item: ExpirationNotification) {
    if (opening || marking) return;
    const controller = new AbortController(); markingController = controller;
    marking = true; openError = '';
    try {
      await setNotificationRead(repository, observer, tenantId, inventoryId, item.id, !!(!item.readAt && !readIds.has(item.id)), controller.signal);
      if (controller.signal.aborted) return;
      readIds = new Set([...readIds].filter(id => id !== item.id));
      onRead(); await load();
    } catch (caught) { if (!controller.signal.aborted) openError = safeWorkspaceErrorMessage(caught, 'Could not change read state. Try again.'); }
    finally { if (!controller.signal.aborted) marking = false; }
  }
</script>

<section aria-label="Notification inbox">
  <div class="toolbar">
    <SegmentedControl label="Notification filter" value={filter} options={[{ value: 'all', label: 'All', disabled: !!opening || marking }, { value: 'unread', label: 'Unread', disabled: !!opening || marking }]} onSelect={(value) => { filter = value; void load(); }} />
    <Button.Root variant="ghost" disabled={loading || !!opening || marking} onclick={markAll}>{marking ? 'Marking read…' : 'Mark all read'}</Button.Root>
    <Button.Root variant="ghost" disabled={loading || !!opening || marking} onclick={() => load()}>Refresh</Button.Root>
    {#if onOpenSettings}<Button.Root variant="ghost" onclick={onOpenSettings}>Notification settings</Button.Root>{/if}
  </div>
  {#if openError}<p role="alert">{openError}</p>{/if}
  {#if loading && items.length === 0}<p role="status">Loading notifications…</p>
  {:else}
    {#if error}<p role="alert">{error}</p><Button.Root disabled={marking} onclick={() => load()}>Retry notifications</Button.Root>{/if}
    {#if items.length === 0 && !hasMore && !error}<p>{filter === 'unread' ? 'No unread notifications.' : 'No expiration notifications yet.'}</p>{/if}
    <ul>
      {#each items as item (item.id)}
        <li class:unread={!item.readAt && !readIds.has(item.id)}><Button.Root variant="ghost" class="notification-row" disabled={!!opening || marking} onclick={() => open(item)}>
          <span><strong>{#if !item.readAt && !readIds.has(item.id)}<span class="unread-dot" aria-hidden="true"></span>{/if}{item.title}</strong><span>{item.milestone === 'expired' ? 'Expired' : 'Expires'} {dateLabel(item)}</span></span>
          {#if opening === item.id}<span>Opening…</span>{:else if !item.readAt && !readIds.has(item.id)}<span class="sr-only">Unread</span>{/if}
        </Button.Root>
          <Button.Root variant="ghost" disabled={!!opening || marking} aria-label={`Mark ${item.title} ${item.readAt || readIds.has(item.id) ? 'unread' : 'read'}`} onclick={() => toggleRead(item)}>{item.readAt || readIds.has(item.id) ? 'Mark unread' : 'Mark read'}</Button.Root>
          <AssetLocationTrail segments={item.parentTrail} incomplete={item.parentTrailIncomplete} disabled={!!opening || marking} onOpen={onOpenAsset} />
        </li>
      {/each}
    </ul>
    {#if appendError}<p role="alert">{appendError}</p>{/if}
    {#if hasMore}<Button.Root variant="outline" disabled={appendLoading || !!opening || marking} onclick={() => load(true)}>{appendLoading ? 'Loading more…' : appendError ? 'Retry more notifications' : 'Load more'}</Button.Root>{/if}
  {/if}
</section>

<style>
  section { display: grid; gap: 1rem; }
  .toolbar { display: flex; justify-content: space-between; gap: 1rem; flex-wrap: wrap; }
  ul { list-style: none; padding: 0; margin: 0; }
  li strong { font-weight: 400; }
  li.unread strong { font-weight: 650; }
  li .unread-dot { display: inline-block; width: 0.5rem; height: 0.5rem; margin-inline-end: 0.5rem; border-radius: 50%; background: var(--primary); }
  li { border-bottom: 1px solid var(--border); }
  li :global(.notification-row) { display: flex; justify-content: space-between; width: 100%; height: auto; min-height: 3.5rem; padding: 1rem; text-align: start; white-space: normal; }
  li span span { display: block; font-size: var(--text-metadata-size); color: var(--muted-foreground); }
  [role='alert'] { color: var(--destructive); }
</style>
