<script lang="ts">
  import AlertCircle from '@lucide/svelte/icons/alert-circle';
  import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
  import type { ImportMessage } from '$lib/domain/inventory';
  import { Badge } from '$lib/components/ui/badge/index.js';

  type Props = {
    messages: ImportMessage[];
    emptyText: string;
    truncated?: boolean;
    truncatedText?: string;
  };

  let { messages, emptyText, truncated = false, truncatedText = 'Showing a sample of import messages.' }: Props = $props();

  let groups = $derived(groupMessages(messages));

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
</script>

<div class="message-list">
  {#each groups as group (group.key)}
    <div class="message-group">
      <div class="message-group-heading">
        <Badge variant={group.severity === 'error' ? 'destructive' : 'secondary'}>{severityLabel(group.severity)}</Badge>
        <div>
          <strong>{group.summary}</strong>
          <span>{group.cause ? `${group.cause} · ${groupCountLabel(group.messages.length)}` : groupCountLabel(group.messages.length)}</span>
        </div>
      </div>
      <div class="message-group-items">
        {#each group.messages as message}
          <div class="message-row">
            <span>{message.sourceName || message.detail || group.summary}</span>
            {#if message.sourceName && message.detail && message.detail !== group.cause}
              <small>{message.detail}</small>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/each}
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

  .message-group {
    border: 1px solid hsl(var(--border));
    border-radius: 8px;
    display: grid;
    gap: 0.55rem;
    padding: 0.75rem;
  }

  .message-group-heading {
    align-items: center;
    display: flex;
    gap: 0.75rem;
  }

  .message-group-heading strong,
  .message-group-heading span,
  .message-row span,
  .message-row small {
    display: block;
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
    align-items: center;
    display: flex;
    gap: 0.75rem;
  }
</style>
