<script lang="ts">
  import { tick } from 'svelte';
  import type { Snippet } from 'svelte';
  import type { HTMLAttributes } from 'svelte/elements';
  import { cn } from '$lib/utils.js';

  type Props = HTMLAttributes<HTMLDivElement> & {
    open: boolean;
    ariaLabel?: string;
    ariaLabelledBy?: string;
    dismissDisabled?: boolean;
    onDismiss: () => void;
    children?: Snippet;
  };

  let {
    open,
    ariaLabel,
    ariaLabelledBy,
    dismissDisabled = false,
    onDismiss,
    children,
    class: className,
    ...restProps
  }: Props = $props();

  let dialogElement = $state<HTMLElement | null>(null);

  $effect(() => {
    if (!open) return;
    void tick().then(() => {
      dialogElement?.focus();
    });
  });

  function dismiss(): void {
    if (dismissDisabled) return;
    onDismiss();
  }

  function handleDialogKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') {
      event.preventDefault();
      dismiss();
      return;
    }
    if (event.key !== 'Tab' || !dialogElement) {
      return;
    }
    const focusable = Array.from(
      dialogElement.querySelectorAll<HTMLElement>(
        'a[href], button:not([disabled]), input:not([disabled]), textarea:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'
      )
    ).filter((element) => !element.hasAttribute('disabled') && element.getAttribute('aria-hidden') !== 'true');
    if (focusable.length === 0) {
      event.preventDefault();
      dialogElement.focus();
      return;
    }
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }
</script>

{#if open}
  <div data-slot="dialog-overlay" class="dialog-overlay" role="presentation">
    <div class="dialog-backdrop" aria-hidden="true" onclick={dismiss}></div>
    <div
      bind:this={dialogElement}
      data-slot="dialog-content"
      class={cn('dialog-content', className)}
      role="dialog"
      aria-modal="true"
      aria-label={ariaLabelledBy ? undefined : ariaLabel}
      aria-labelledby={ariaLabelledBy}
      tabindex="-1"
      onkeydown={handleDialogKeydown}
      {...restProps}
    >
      {@render children?.()}
    </div>
  </div>
{/if}

<style>
  .dialog-overlay {
    align-items: center;
    display: grid;
    inset: 0;
    justify-items: center;
    padding: 1rem;
    position: fixed;
    z-index: 60;
  }

  .dialog-backdrop {
    background: color-mix(in oklab, var(--background) 72%, transparent);
    inset: 0;
    position: absolute;
  }

  .dialog-content {
    max-height: min(90vh, 46rem);
    max-width: min(42rem, 100%);
    overflow: auto;
    position: relative;
    width: 100%;
  }

  .dialog-content:focus-visible {
    outline: 2px solid var(--ring);
    outline-offset: 2px;
  }
</style>
