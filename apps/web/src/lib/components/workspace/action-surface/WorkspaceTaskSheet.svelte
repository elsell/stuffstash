<script lang="ts">
  import type { Snippet } from 'svelte';
  import * as Sheet from '$lib/components/ui/sheet/index.js';
  import X from '@lucide/svelte/icons/x';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import * as Button from '$lib/components/ui/button/index.js';

  let contentElement: HTMLElement | null = $state(null);

  let {
    open,
    title,
    description,
    busy = false,
    dismissible = true,
    closeHref,
    closeLabel = 'Close',
    initialFocusSelector,
    onCloseLink,
    onCloseAutoFocus,
    onOpenChange = () => {},
    children,
    footer
  }: {
    open: boolean;
    title: string;
    description?: string;
    busy?: boolean;
    dismissible?: boolean;
    closeHref?: string;
    closeLabel?: string;
    initialFocusSelector?: string;
    onCloseLink?: (event: MouseEvent) => void;
    onCloseAutoFocus?: (event: Event) => void;
    onOpenChange?: (open: boolean) => void;
    children: Snippet;
    footer?: Snippet;
  } = $props();

  function focusTaskStart(event?: Event): void {
    event?.preventDefault();
    const fallbackSelector = '.workspace-task-sheet-body input:not([type="hidden"]):not(:disabled), .workspace-task-sheet-body textarea:not(:disabled), .workspace-task-sheet-body select:not(:disabled), .workspace-task-sheet-body button:not(:disabled), .workspace-task-sheet-body a[href]';
    contentElement?.querySelector<HTMLElement>(initialFocusSelector || fallbackSelector)?.focus();
  }

</script>

<Sheet.Root {open} onOpenChange={(nextOpen) => { if ((!busy && dismissible) || nextOpen) onOpenChange(nextOpen); }}>
  <Sheet.Content
    bind:ref={contentElement}
    side="right"
    class="workspace-task-sheet w-full max-w-none gap-0 sm:max-w-xl [&_button]:min-h-11 [&_input]:min-h-11 [&_select]:min-h-11 [&_textarea]:min-h-11 [&_[data-slot=button]]:min-h-11 [&_[data-slot=button]]:min-w-11"
    style="width: 100%;"
    showCloseButton={!closeHref && !busy && dismissible}
    aria-busy={busy}
    onInteractOutside={(event) => { if (busy || !dismissible) event.preventDefault(); }}
    onEscapeKeydown={(event) => { if (busy || !dismissible) event.preventDefault(); }}
    onOpenAutoFocus={focusTaskStart}
    {onCloseAutoFocus}
  >
    <Sheet.Header class="workspace-task-sheet-header relative z-10 shrink-0 border-b bg-popover px-5 py-6 pr-16 text-left text-popover-foreground sm:px-6">
      <Sheet.Title class="text-lg font-semibold">{title}</Sheet.Title>
      {#if description}<Sheet.Description>{description}</Sheet.Description>{/if}
      {#if busy}
        <p class="workspace-surface-progress mt-2 flex min-h-5 items-center gap-2 text-sm text-muted-foreground" role="status" aria-live="polite">
          <LoaderCircle class="size-4 motion-safe:animate-spin motion-reduce:animate-none" aria-hidden="true" />
          Saving changes…
        </p>
      {/if}
    </Sheet.Header>
    {#if closeHref && !busy && dismissible}
      <Sheet.Close>
        {#snippet child({ props })}
          <Button.Root {...props} href={closeHref} variant="ghost" size="icon-sm" class="absolute top-4 right-4 z-20 size-11" aria-label={closeLabel} onclick={onCloseLink}>
            <X />
          </Button.Root>
        {/snippet}
      </Sheet.Close>
    {/if}
    <div
      class="workspace-task-sheet-body grid min-h-0 flex-1 content-start gap-6 overflow-y-auto px-5 py-6 sm:px-6"
      inert={busy ? true : undefined}
      aria-disabled={busy}
    >
      {@render children()}
    </div>
    {#if footer}
      <Sheet.Footer class="workspace-task-sheet-footer relative z-10 shrink-0 border-t bg-popover px-5 py-4 text-popover-foreground sm:flex-row sm:justify-end sm:px-6">
        {@render footer()}
      </Sheet.Footer>
    {/if}
  </Sheet.Content>
</Sheet.Root>
