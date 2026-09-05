<script lang="ts">
  import * as Label from '$lib/components/ui/label/index.js';
  import * as Select from '$lib/components/ui/select/index.js';
  import ValidationMessage from './ValidationMessage.svelte';
  import { validationAttributes } from './validationPresentation';
  let { id, label, value, options, disabled = false, error, onChange }: {
    id: string; label: string; value: string; options: { value: string; label: string }[];
    disabled?: boolean; error?: string; onChange: (value: string) => void;
  } = $props();
</script>
<div class="workflow-select">
  <Label.Root for={id}>{label}</Label.Root>
  <Select.Root type="single" {value} {disabled} onValueChange={onChange}>
    <Select.Trigger {id} {...validationAttributes(error, id)} class="w-full">{options.find(option => option.value === value)?.label ?? 'Choose an option'}</Select.Trigger>
    <Select.Content>{#each options as option (option.value)}<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>{/each}</Select.Content>
  </Select.Root>
  <ValidationMessage field={id} message={error} />
</div>
<style>.workflow-select { display: grid; gap: .4rem; font-size: .9rem; font-weight: 500; }</style>
