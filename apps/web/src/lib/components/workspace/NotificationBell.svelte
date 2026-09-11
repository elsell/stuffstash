<script lang="ts">
  import { onMount } from 'svelte';
  import Bell from '@lucide/svelte/icons/bell';
  import * as Button from '$lib/components/ui/button/index.js';
  import type { NotificationRepository } from '$lib/ports/notificationRepository';
  import type { WorkspaceObserver } from '$lib/observability/workspaceObserver';
  import { countUnreadNotifications } from '$lib/application/notificationInboxBatch';
  import WorkspaceTaskSheet from './action-surface/WorkspaceTaskSheet.svelte';
  import NotificationInbox from './NotificationInbox.svelte';

  let { tenantId, inventoryId, repository, observer, onOpenAsset, onOpenSettings }: {
    tenantId: string; inventoryId: string; repository: NotificationRepository; observer: WorkspaceObserver;
    onOpenAsset: (id: string) => void; onOpenSettings: () => void;
  } = $props();
  const refreshIntervalMs = 30_000;
  let open = $state(false);
  let count = $state<number | null>(null);
  let error = $state(false);
  let registered = false;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let controller: AbortController | undefined;
  let bell: HTMLButtonElement | null = $state(null);
  let navigating = false;
  let disposed = false;
  onMount(() => {
    void refresh();
    const focus = () => { if (!document.hidden) void refresh(); };
    window.addEventListener('focus', focus);
    document.addEventListener('visibilitychange', focus);
    return () => { disposed = true; controller?.abort(); clearTimeout(timer); window.removeEventListener('focus', focus); document.removeEventListener('visibilitychange', focus); };
  });
  async function refresh() {
    if (disposed) return;
    clearTimeout(timer); controller?.abort();
    const request = new AbortController(); controller = request;
    try {
      if (!registered) {
        await repository.initializePreferences(tenantId, inventoryId, Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC', request.signal);
        request.signal.throwIfAborted(); registered = true;
      }
      const loaded = await countUnreadNotifications(repository, observer, tenantId, inventoryId, request.signal);
      request.signal.throwIfAborted(); count = loaded; error = false;
    } catch {
      if (!request.signal.aborted) { count = null; error = true; }
    } finally {
      if (!request.signal.aborted && !disposed) timer = setTimeout(() => { if (!document.hidden) void refresh(); }, refreshIntervalMs);
    }
  }
  function show() { navigating = false; open = true; void refresh(); }
  function navigate(action: () => void) { navigating = true; open = false; action(); }
</script>

<Button.Root bind:ref={bell} variant="ghost" size="icon" class="notification-bell" aria-label={error ? 'Notifications, unread count unavailable' : count === null ? 'Notifications, loading unread count' : `Notifications, ${count} unread`} onclick={show}>
  <Bell aria-hidden="true" />
  {#if count !== null && count > 0}<span class="badge" aria-hidden="true">{count > 99 ? '99+' : count}</span>{/if}
</Button.Root>
<WorkspaceTaskSheet {open} title="Notifications" description="Expiration reminders for this inventory" onOpenChange={(value) => { open = value; }} onCloseAutoFocus={(event) => { event.preventDefault(); if (!navigating) bell?.focus(); }}>
  {#if open}
    {#if error}<p role="alert">The unread count could not be updated.</p><Button.Root variant="outline" onclick={() => refresh()}>Retry unread count</Button.Root>{/if}
    <NotificationInbox {tenantId} {inventoryId} {repository} {observer} onRead={() => { void refresh(); }} onOpenAsset={(id) => navigate(() => onOpenAsset(id))} />
    <Button.Root variant="outline" onclick={() => navigate(onOpenSettings)}>Notification settings</Button.Root>
  {/if}
</WorkspaceTaskSheet>

<style>
  :global(.notification-bell) { position: relative; }
  .badge { position: absolute; top: -0.2rem; right: -0.3rem; border-radius: 999px; padding: 0.1rem 0.3rem; background: var(--primary); color: var(--primary-foreground); font-size: var(--text-caption-size); line-height: 1rem; }
</style>
