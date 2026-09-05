<script lang="ts">
  import * as Label from '$lib/components/ui/label/index.js';
  import * as Select from '$lib/components/ui/select/index.js';
  let { id, label, value, options, disabled = false, onChange }: {
    id: string; label: string; value: string; options: { value: string; label: string }[];
    disabled?: boolean; onChange: (value: string) => void;
  } = $props();
</script>
<div class="workflow-select">
  <Label.Root for={id}>{label}</Label.Root>
  <Select.Root type="single" {value} {disabled} onValueChange={onChange}>
    <Select.Trigger {id} class="w-full">{options.find(option => option.value === value)?.label ?? 'Choose a profile'}</Select.Trigger>
    <Select.Content>{#each options as option (option.value)}<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>{/each}</Select.Content>
  </Select.Root>
</div>
<style>.workflow-select { display: grid; gap: .4rem; font-size: .9rem; font-weight: 500; }</style>
