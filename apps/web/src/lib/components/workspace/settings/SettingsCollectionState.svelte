<script lang="ts">
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import * as Button from '$lib/components/ui/button/index.js';
  let { kind, title, message, onRetry }: { kind: 'loading' | 'empty' | 'error' | 'denied'; title: string; message: string; onRetry?: () => void } = $props();
</script>

<div class="settings-collection-state" role={kind === 'error' || kind === 'denied' ? 'alert' : kind === 'loading' ? 'status' : undefined} aria-live={kind === 'loading' ? 'polite' : undefined}>
  {#if kind === 'loading'}<LoaderCircle class="motion-safe:animate-spin motion-reduce:animate-none" aria-hidden="true" />{/if}
  <div><h3>{title}</h3><p>{message}</p></div>
  {#if onRetry}<Button.Root variant="outline" onclick={onRetry}>Try again</Button.Root>{/if}
</div>
