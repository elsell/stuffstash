<script lang="ts">
  import { t } from '$lib/presentation/localization';
  import type { Component } from 'svelte';
  import * as Button from '$lib/components/ui/button/index.js';

  let {
    label,
    description,
    checked,
    disabled = false,
    icon,
    onToggle
  }: {
    label: string;
    description?: string;
    checked: boolean;
    disabled?: boolean;
    icon?: Component;
    onToggle: () => void;
  } = $props();
</script>

<Button.Root
  type="button"
  variant="outline"
  class="binary-option"
  role="switch"
  aria-checked={checked}
  disabled={disabled}
  data-checked={checked}
  onclick={onToggle}
>
  <span class="binary-option-copy">
    <span class="binary-option-label">
      {#if icon}
        {@const Icon = icon}
        <Icon aria-hidden="true" />
      {/if}
      <strong>{label}</strong>
    </span>
    {#if description}
      <small>{description}</small>
    {/if}
  </span>
  <span class="binary-option-state" aria-hidden="true">
    <span class="binary-option-status">{checked ? t('web.BinaryOption.on') : t('web.BinaryOption.off')}</span>
    <span class="binary-option-track"><span></span></span>
  </span>
</Button.Root>
