<script lang="ts">
  import type { Snippet } from 'svelte';
  import { tick } from 'svelte';
  import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';

  let contentElement: HTMLElement | null = $state(null);
  let progressElement: HTMLElement | null = $state(null);
  let wasBusy = $state(false);

  let {
    open,
    title,
    description,
    busy = false,
    onOpenChange = () => {},
    onCloseAutoFocus,
    children,
    cancel,
    action
  }: {
    open: boolean;
    title: string;
    description: string;
    busy?: boolean;
    onOpenChange?: (open: boolean) => void;
    onCloseAutoFocus?: (event: Event) => void;
    children?: Snippet;
    cancel: Snippet;
    action: Snippet;
  } = $props();

  function focusSafeAction(event: Event): void {
    event.preventDefault();
    void tick().then(() => {
      contentElement
        ?.querySelector<HTMLElement>('[data-workspace-confirmation-actions] button:not(:disabled), [data-workspace-confirmation-actions] a[href]')
        ?.focus();
    });
  }

  $effect(() => {
    const nextBusy = busy;
    if (nextBusy && !wasBusy) {
      void tick().then(() => progressElement?.focus());
    }
    wasBusy = nextBusy;
  });

  $effect(() => {
    const element = contentElement;
    if (!open || busy || !element) return;
    void tick().then(() => {
      element
        .querySelector<HTMLElement>('[data-workspace-confirmation-actions] button:not(:disabled), [data-workspace-confirmation-actions] a[href]')
        ?.focus();
    });
  });
</script>

<AlertDialog.Root {open} onOpenChange={(nextOpen) => { if (!busy || nextOpen) onOpenChange(nextOpen); }}>
  <AlertDialog.Content
    bind:ref={contentElement}
    class="workspace-confirmation-dialog [&_[data-slot=button]]:min-h-11 [&_[data-slot=button]]:min-w-11"
    aria-busy={busy}
    onEscapeKeydown={(event) => { if (busy) event.preventDefault(); }}
    onOpenAutoFocus={focusSafeAction}
    {onCloseAutoFocus}
  >
    <AlertDialog.Header>
      <AlertDialog.Title>{title}</AlertDialog.Title>
      <AlertDialog.Description>{description}</AlertDialog.Description>
    </AlertDialog.Header>
    {#if children}<div class="workspace-confirmation-body">{@render children()}</div>{/if}
    {#if busy}
      <p
        bind:this={progressElement}
        class="workspace-surface-progress flex min-h-5 items-center gap-2 text-sm text-muted-foreground outline-none"
        role="status"
        aria-live="polite"
        tabindex="-1"
      >
        <LoaderCircle class="size-4 motion-safe:animate-spin motion-reduce:animate-none" aria-hidden="true" />
        Working…
      </p>
    {/if}
    <AlertDialog.Footer>
      <fieldset
        class="contents"
        data-workspace-confirmation-actions
        disabled={busy}
        aria-disabled={busy}
        inert={busy ? true : undefined}
      >
        {@render cancel()}
        {@render action()}
      </fieldset>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
