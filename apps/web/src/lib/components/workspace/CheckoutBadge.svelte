<script lang="ts">
  import type { CurrentCheckout } from '$lib/domain/inventory';

  let { checkout, compact = false }: { checkout: CurrentCheckout; compact?: boolean } = $props();
</script>

<span class:compact class="checkout-badge" title={`Checked out ${new Date(checkout.checkedOutAt).toLocaleString()}`}>
  Checked out
</span>

<style>
  .checkout-badge {
    display: inline-flex;
    width: fit-content;
    align-items: center;
    gap: 0.25rem;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--color-warning, #a15c00) 26%, transparent);
    background: color-mix(in srgb, var(--color-warning, #a15c00) 10%, var(--color-surface, #fff));
    color: var(--color-warning-foreground, #6b3a00);
    font-size: 0.75rem;
    font-weight: 650;
    line-height: 1;
    padding: 0.28rem 0.48rem;
    white-space: nowrap;
  }

  .checkout-badge.compact {
    font-size: 0.68rem;
    padding: 0.22rem 0.38rem;
  }
</style>
