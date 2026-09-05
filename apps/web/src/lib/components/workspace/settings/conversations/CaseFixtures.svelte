<script lang="ts">
  import * as Label from '$lib/components/ui/label/index.js';
  import type { CaseDefinition, CaseFixtureAsset } from '$lib/domain/conversationCase';
  import { fixtureParentChoices, fixtureRemovalBlocked, nextFixtureId } from '$lib/application/caseFixtureEditing';
  import * as Button from '$lib/components/ui/button/index.js';
  import * as Input from '$lib/components/ui/input/index.js';
  import * as Textarea from '$lib/components/ui/textarea/index.js';
  import WorkflowSelect from './WorkflowSelect.svelte';
  import ValidationMessage from './ValidationMessage.svelte';
  import { validationAttributes } from './validationPresentation';
  let { value, errors = {}, disabled = false, onChange }: { value: CaseDefinition; errors?: Record<string, string>; disabled?: boolean; onChange: (value: CaseDefinition) => void } = $props();
  function update(id: string, patch: Partial<CaseFixtureAsset>) {
    if (disabled) return;
    onChange({ ...value, assets: value.assets.map(asset => asset.id === id ? { ...asset, ...patch } : asset) });
  }
  function add() {
    if (disabled || value.assets.length >= 100) return;
    onChange({ ...value, assets: [...value.assets, { id: nextFixtureId(value.assets), title: '', kind: 'item', parentId: '', description: '', tagNames: [] }] });
  }
  function remove(id: string) {
    if (disabled || fixtureRemovalBlocked(value, id)) return;
    onChange({ ...value, assets: value.assets.filter(asset => asset.id !== id) });
  }
</script>
<section class="case-fixtures" aria-labelledby="case-fixtures-title">
  <header><h3 id="case-fixtures-title" tabindex="-1" {...validationAttributes(errors['case-fixtures-title'], 'case-fixtures-title')}>Test inventory</h3><p>These items exist only inside this test case. Your household inventory is unchanged.</p></header><ValidationMessage field="case-fixtures-title" message={errors['case-fixtures-title']} />
  {#each value.assets as asset, index (asset.id)}
    {@const hasChildren = value.assets.some(child => child.parentId === asset.id)}
    {@const removalBlocked = fixtureRemovalBlocked(value, asset.id)}
    <fieldset {disabled}>
      <legend>Fixture {index + 1}{asset.title ? `: ${asset.title}` : ''}</legend>
      <Label.Root class="grid gap-2 text-sm">Name<Input.Root name={`fixture-title-${asset.id}`} id={`fixture-title-${asset.id}`} {...validationAttributes(errors[`fixture-title-${asset.id}`], `fixture-title-${asset.id}`)} value={asset.title} required oninput={event => update(asset.id, { title: event.currentTarget.value })} /></Label.Root><ValidationMessage field={`fixture-title-${asset.id}`} message={errors[`fixture-title-${asset.id}`]} />
      <WorkflowSelect id={`fixture-kind-${asset.id}`} error={errors[`fixture-kind-${asset.id}`]} label="Kind" value={asset.kind} {disabled}
        options={[...(!hasChildren ? [{ value: 'item', label: 'Item' }] : []), { value: 'container', label: 'Container' }, { value: 'location', label: 'Location' }]}
        onChange={kind => { if (kind === 'container' || kind === 'location' || (kind === 'item' && !hasChildren)) update(asset.id, { kind }); }} />
      {#if hasChildren}<p>This fixture contains others, so it must remain a container or location.</p>{/if}
      <WorkflowSelect id={`fixture-parent-${asset.id}`} error={errors[`fixture-parent-${asset.id}`]} label="Inside" value={asset.parentId} {disabled}
        options={[{ value: '', label: 'No parent' }, ...fixtureParentChoices(value.assets, asset.id).map(parent => ({ value: parent.id, label: parent.title || 'Unnamed fixture' }))]}
        onChange={parentId => update(asset.id, { parentId })} />
      <Label.Root class="grid gap-2 text-sm">Description<Textarea.Root name={`fixture-description-${asset.id}`} id={`fixture-description-${asset.id}`} {...validationAttributes(errors[`fixture-description-${asset.id}`], `fixture-description-${asset.id}`)} value={asset.description} rows={2} oninput={event => update(asset.id, { description: event.currentTarget.value })} /></Label.Root><ValidationMessage field={`fixture-description-${asset.id}`} message={errors[`fixture-description-${asset.id}`]} />
      <Label.Root class="grid gap-2 text-sm">Tags — one per line<Textarea.Root name={`fixture-tags-${asset.id}`} id={`fixture-tags-${asset.id}`} {...validationAttributes(errors[`fixture-tags-${asset.id}`], `fixture-tags-${asset.id}`)} value={asset.tagNames.join('\n')} rows={3}
        oninput={event => update(asset.id, { tagNames: event.currentTarget.value.split('\n') })} /></Label.Root><ValidationMessage field={`fixture-tags-${asset.id}`} message={errors[`fixture-tags-${asset.id}`]} />
      <Button.Root type="button" variant="outline" disabled={disabled || removalBlocked} onclick={() => remove(asset.id)}>Remove fixture</Button.Root>
      {#if removalBlocked}<p>Change references to this fixture in containment or expected results before removing it.</p>{/if}
    </fieldset>
  {/each}
  <Button.Root type="button" variant="outline" disabled={disabled || value.assets.length >= 100} onclick={add}>Add test item</Button.Root>
  {#if value.assets.length >= 100}<p>A test case supports up to 100 fixtures.</p>{/if}
</section>
<style>
  .case-fixtures, fieldset { display: grid; gap: 1rem; }
  fieldset { min-width: 0; border: 1px solid var(--border); border-radius: var(--radius); padding: 1rem; }
  legend, h3 { font-weight: 600; } legend { overflow-wrap: anywhere; }
  p { font-size: .9rem; color: var(--muted-foreground); }
</style>
