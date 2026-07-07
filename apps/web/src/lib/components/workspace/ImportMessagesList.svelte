<script lang="ts">
  import AlertCircle from '@lucide/svelte/icons/alert-circle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import type { ImportMessage } from '$lib/domain/inventory';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Button from '$lib/components/ui/button/index.js';

  const COLLAPSED_GROUP_LIMIT = 5;
  const EXPANDED_GROUP_LIMIT = 15;
  const COLLAPSED_RECORD_LIMIT = 3;
  const EXPANDED_RECORD_LIMIT = 5;

  type Props = {
    messages: ImportMessage[];
    emptyText: string;
    truncated?: boolean;
    truncatedText?: string;
  };

  let { messages, emptyText, truncated = false, truncatedText = 'Showing a sample of import messages.' }: Props = $props();

  let groups = $derived(groupMessages(messages));
  let expanded = $state(false);
  let visibleGroupLimit = $derived(expanded ? EXPANDED_GROUP_LIMIT : COLLAPSED_GROUP_LIMIT);
  let visibleRecordLimit = $derived(expanded ? EXPANDED_RECORD_LIMIT : COLLAPSED_RECORD_LIMIT);
  let visibleGroups = $derived(groups.slice(0, visibleGroupLimit));
  let hiddenGroupCount = $derived(Math.max(0, groups.length - visibleGroups.length));

  type MessageGroup = {
    key: string;
    severity: ImportMessage['severity'];
    summary: string;
    cause: string;
    messages: ImportMessage[];
  };

  function groupMessages(items: ImportMessage[]): MessageGroup[] {
    const grouped = new Map<string, MessageGroup>();
    for (const message of items) {
      const cause = message.detail || '';
      const key = `${message.severity}:${message.summary}:${cause}`;
      const group = grouped.get(key);
      if (group) {
        group.messages.push(message);
      } else {
        grouped.set(key, {
          key,
          severity: message.severity,
          summary: message.summary,
          cause,
          messages: [message]
        });
      }
    }
    return Array.from(grouped.values());
  }

  function severityLabel(severity: ImportMessage['severity']): string {
    return severity === 'error' ? 'Blocking' : 'Warning';
  }

  function groupCountLabel(count: number): string {
    return count === 1 ? '1 item' : `${count} items`;
  }

  function messageRowLabel(message: ImportMessage, group: MessageGroup): string {
    if (message.sourceName) return message.sourceName;
    if (message.sourceId) return `Source record ${message.sourceId}`;
    return message.detail || group.summary;
  }
</script>

<div class="message-list">
  {#if groups.length > 0}
    <div class="message-list-summary">
      <strong>{groups.length === 1 ? '1 issue group' : `${groups.length} issue groups`}</strong>
      <span>{messages.length === 1 ? '1 affected record' : `${messages.length} affected records`}</span>
    </div>
  {/if}
  {#each visibleGroups as group (group.key)}
    {@const visibleMessages = group.messages.slice(0, visibleRecordLimit)}
    <div class="message-group">
      <div class="message-group-heading">
        <Badge variant={group.severity === 'error' ? 'destructive' : 'secondary'}>{severityLabel(group.severity)}</Badge>
        <div>
          <strong>{group.summary}</strong>
          <span>{group.cause ? `${group.cause} · ${groupCountLabel(group.messages.length)}` : groupCountLabel(group.messages.length)}</span>
        </div>
      </div>
      <div class="message-group-items">
        {#each visibleMessages as message}
          <div class="message-row">
            <span>{messageRowLabel(message, group)}</span>
            {#if message.sourceName && message.detail && message.detail !== group.cause}
              <small>{message.detail}</small>
            {/if}
          </div>
        {/each}
        {#if group.messages.length > visibleMessages.length}
          <small class="message-overflow">{group.messages.length - visibleMessages.length} more in this group</small>
        {/if}
      </div>
    </div>
  {/each}
  {#if hiddenGroupCount > 0}
    <div class="message-overflow-action">
      <span>{hiddenGroupCount} more issue {hiddenGroupCount === 1 ? 'group' : 'groups'} hidden.</span>
      {#if expanded}
        <Button.Root variant="outline" size="sm" onclick={() => (expanded = false)}>Show fewer</Button.Root>
      {:else}
        <Button.Root variant="outline" size="sm" onclick={() => (expanded = true)}>Show more issues</Button.Root>
      {/if}
    </div>
  {:else if expanded && groups.length > COLLAPSED_GROUP_LIMIT}
    <div class="message-overflow-action">
      <span>All issue groups are shown.</span>
      <Button.Root variant="outline" size="sm" onclick={() => (expanded = false)}>Show fewer</Button.Root>
    </div>
  {/if}
  {#if messages.length === 0}
    <div class="quiet-row"><CheckCircle2 size={16} aria-hidden="true" /> {emptyText}</div>
  {/if}
  {#if truncated}
    <div class="quiet-row"><AlertCircle size={16} aria-hidden="true" /> {truncatedText}</div>
  {/if}
</div>

<style>
  .message-list {
    display: grid;
    gap: 0.75rem;
  }

  .message-list-summary,
  .message-overflow-action {
    align-items: center;
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    justify-content: space-between;
  }

  .message-list-summary span,
  .message-overflow,
  .message-overflow-action span {
    color: hsl(var(--muted-foreground));
    font-size: 0.82rem;
  }

  .message-group {
    border-top: 1px solid hsl(var(--border));
    display: grid;
    gap: 0.55rem;
    padding-top: 0.75rem;
  }

  .message-group:first-child {
    border-top: 0;
    padding-top: 0;
  }

  .message-group-heading {
    align-items: flex-start;
    display: flex;
    gap: 0.75rem;
    min-width: 0;
  }

  .message-group-heading > div {
    min-width: 0;
  }

  .message-group-heading strong,
  .message-group-heading span,
  .message-row span,
  .message-row small {
    display: block;
    overflow-wrap: anywhere;
  }

  .message-group-heading span,
  .message-row small {
    color: hsl(var(--muted-foreground));
    font-size: 0.78rem;
  }

  .message-group-items {
    display: grid;
    gap: 0.4rem;
  }

  .message-row,
  .quiet-row {
    align-items: flex-start;
    display: flex;
    gap: 0.75rem;
  }

  .message-overflow {
    display: block;
  }

  @media (max-width: 640px) {
    .message-list-summary,
    .message-overflow-action {
      align-items: flex-start;
      display: grid;
      justify-content: stretch;
    }

    .message-group-heading {
      display: grid;
      gap: 0.45rem;
    }

    .message-row {
      display: block;
    }
  }
</style>
