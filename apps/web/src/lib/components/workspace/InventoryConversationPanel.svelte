<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import { formatAssetExpiration } from '$lib/application/expirationPresentation';
  import { onDestroy } from 'svelte';
  import MessageCircle from '@lucide/svelte/icons/message-circle';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import * as Button from '$lib/components/ui/button/index.js';
  import WorkspaceTaskSheet from './action-surface/WorkspaceTaskSheet.svelte';
  import { workspaceRouteHref } from '$lib/application/workspaceRoute';
  import { InventoryConversation } from '$lib/application/inventoryConversation';
  import type { InventoryConversationTransport } from '$lib/ports/inventoryConversation';
  let { transport, tenantId, inventoryId, inventoryName, onOpenAsset, onRefresh, onAuthenticationLost }: {
    transport: InventoryConversationTransport; tenantId: string; inventoryId: string; inventoryName: string;
    onOpenAsset: (id: string) => void; onRefresh: () => Promise<boolean>; onAuthenticationLost: () => void;
  } = $props();
  let open = $state(false); let draft = $state(''); let revision = $state(0);
  let opener = $state<HTMLButtonElement | null>(null);
  // svelte-ignore state_referenced_locally -- parent keys the panel by authenticated inventory scope.
  const conversation = new InventoryConversation(transport, { tenantId, inventoryId }, () => { revision++; }, onRefresh, onAuthenticationLost);
  const view = $derived.by(() => { void revision; return conversation.state; });
  onDestroy(() => conversation.dispose());
  function setOpen(value: boolean) { open = value; if (!value) conversation.stop(); }
  function submit() { if (view.busy || view.plan || view.uncertain || !draft.trim()) return; const text = draft; draft = ''; void conversation.send(text); }
  function showAsset(id: string) { setOpen(false); onOpenAsset(id); }
</script>
<Button.Root bind:ref={opener} variant="ghost" size="icon" class="size-11" aria-label={t('web.InventoryConversationPanel.askStuffStash')} onclick={() => { open = true; }}><MessageCircle aria-hidden="true" /></Button.Root>
<WorkspaceTaskSheet {open} title={t('web.InventoryConversationPanel.askStuffStash')} description={inventoryName} onOpenChange={setOpen} initialFocusSelector="textarea" onCloseAutoFocus={(event) => { event.preventDefault(); opener?.focus(); }}>
  <div class="conversation-transcript" aria-label={t('web.InventoryConversationPanel.conversation')}>
    {#if view.messages.length === 0}<p class="text-muted-foreground">{t('web.InventoryConversationPanel.findSomethingAskWhatSStoredHereOrDescribe')}</p>{/if}
    {#each view.messages as message, index (index)}
      <article aria-label={message.role === 'user' ? t('web.InventoryConversationPanel.you') : t('web.InventoryConversationPanel.stuffStash')} class:user-message={message.role === 'user'}>
        <p class="speaker">{message.role === 'user' ? t('web.InventoryConversationPanel.you') : t('web.InventoryConversationPanel.stuffStash')}</p>
        <p class="message-text" dir="auto">{message.text}</p>
        {#if message.assets.length}<ul>{#each message.assets as asset (asset.id)}<li><Button.Root variant="outline" href={workspaceRouteHref({ mode: 'asset', tenantId, inventoryId, assetId: asset.id }, tenantId, inventoryId)} onclick={(event) => { if (!event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey && event.button === 0) { event.preventDefault(); showAsset(asset.id); } }}>{asset.title}</Button.Root></li>{/each}</ul>{/if}
      </article>
    {/each}
  </div>
  <p class="sr-only" aria-live="polite" aria-atomic="true">{view.messages.at(-1)?.role === 'assistant' ? view.messages.at(-1)?.text : ''}</p>
  {#if view.plan}
    <section class="review" aria-label={t('web.InventoryConversationPanel.reviewChanges')}>
      <h3 class="font-semibold">{t('web.InventoryConversationPanel.reviewChanges')}</h3><p dir="auto">{view.plan.summary}</p>
      <ul>{#each view.plan.commands as command, index (index)}<li><p dir="auto">{command.summary}</p>{#each command.changes ?? [] as change}<p class="whitespace-pre-wrap break-words" dir="auto">{change}</p>{/each}{#if command.expiration}<p>{t('web.InventoryConversationPanel.expiresFull', { value: formatAssetExpiration(command.expiration) })}</p>{:else if command.expirationCleared}<p>{t('web.InventoryConversationPanel.removeExpirationDate')}</p>{/if}{#if command.destination}<p class="text-sm text-muted-foreground">{t('web.InventoryConversationPanel.destinationFull', { destination: command.destination })}</p>{/if}</li>{/each}</ul>
      {#each view.plan.risks as risk}<p class="text-sm">{risk}</p>{/each}
      <div class="flex flex-wrap gap-2"><Button.Root disabled={view.busy} onclick={() => conversation.decide(true)}>{t('web.InventoryConversationPanel.approveChanges')}</Button.Root><Button.Root variant="outline" disabled={view.busy} onclick={() => conversation.decide(false)}>{t('web.InventoryConversationPanel.cancelChanges')}</Button.Root></div>
    </section>
  {/if}
  {#if view.error}<p role="alert">{view.error}</p>{/if}
  {#if view.uncertain}<Button.Root variant="outline" disabled={view.busy} onclick={() => { void conversation.recover(); }}>{t('web.InventoryConversationPanel.refreshInventory')}</Button.Root>{/if}
  {#if view.busy}<p role="status">{view.plan ? t('web.InventoryConversationPanel.applyingYourDecision') : t('web.InventoryConversationPanel.working')}</p>{/if}
  {#snippet footer()}
    <form class="composer" onsubmit={(event) => { event.preventDefault(); submit(); }}>
      <Label for="inventory-conversation-message" class="sr-only">{t('web.InventoryConversationPanel.messageStuffStash')}</Label>
      <Textarea class="max-h-40 resize-y" id="inventory-conversation-message" bind:value={draft} rows={2} maxlength={8000} placeholder={t('web.InventoryConversationPanel.messageStuffStash')} disabled={view.busy || !!view.plan || view.uncertain}
        onkeydown={(event) => { if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) { event.preventDefault(); submit(); } }} />
      <div class="flex justify-end gap-2">{#if view.busy}<Button.Root type="button" variant="outline" onclick={() => conversation.stop()}>{t('web.InventoryConversationPanel.stop')}</Button.Root>{/if}<Button.Root type="submit" disabled={view.busy || !!view.plan || view.uncertain || !draft.trim()}>{t('web.InventoryConversationPanel.send')}</Button.Root></div>
    </form>
  {/snippet}
</WorkspaceTaskSheet>
<style>
  .conversation-transcript { display: grid; gap: 1.25rem; min-width: 0; }
  article { min-width: 0; padding: 1rem; border-radius: var(--radius); background: var(--muted); }
  .user-message { margin-left: 1.5rem; background: var(--accent); }
  .speaker { font-size: var(--text-metadata-size); font-weight: 600; margin-bottom: .4rem; }
  .message-text { white-space: pre-wrap; overflow-wrap: anywhere; }
  ul { display: grid; gap: .5rem; margin-top: .75rem; }
  .review { display: grid; gap: .75rem; border: 1px solid var(--border); border-radius: var(--radius); padding: 1rem; }
  .composer { display: grid; gap: .75rem; width: 100%; }
</style>
