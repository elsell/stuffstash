<script lang="ts">
  import { t } from '$lib/presentation/localization';
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
  <header><h3 id="case-fixtures-title" tabindex="-1" {...validationAttributes(errors['case-fixtures-title'], 'case-fixtures-title')}>{t('web.CaseFixtures.testInventory')}</h3><p>{t('web.CaseFixtures.theseItemsExistOnlyInsideThisTestCaseYour')}</p></header><ValidationMessage field="case-fixtures-title" message={errors['case-fixtures-title']} />
  {#each value.assets as asset, index (asset.id)}
    {@const hasChildren = value.assets.some(child => child.parentId === asset.id)}
    {@const removalBlocked = fixtureRemovalBlocked(value, asset.id)}
    <fieldset {disabled}>
      <legend>{t('web.CaseFixtures.fixtureFull', { value: index + 1, value2: asset.title ? `: ${asset.title}` : '' })}</legend>
      <Label.Root class="grid gap-2 text-sm">{t('web.CaseFixtures.name')}<Input.Root name={`fixture-title-${asset.id}`} id={`fixture-title-${asset.id}`} {...validationAttributes(errors[`fixture-title-${asset.id}`], `fixture-title-${asset.id}`)} value={asset.title} required oninput={event => update(asset.id, { title: event.currentTarget.value })} /></Label.Root><ValidationMessage field={`fixture-title-${asset.id}`} message={errors[`fixture-title-${asset.id}`]} />
      <WorkflowSelect id={`fixture-kind-${asset.id}`} error={errors[`fixture-kind-${asset.id}`]} label={t('web.CaseFixtures.kind')} value={asset.kind} {disabled}
        options={[...(!hasChildren ? [{ value: 'item', label: t("web.options.CaseFixtures.item") }] : []), { value: 'container', label: t("web.options.CaseFixtures.container") }, { value: 'location', label: t("web.options.CaseFixtures.location") }]}
        onChange={kind => { if (kind === 'container' || kind === 'location' || (kind === 'item' && !hasChildren)) update(asset.id, { kind }); }} />
      {#if hasChildren}<p>{t('web.CaseFixtures.thisFixtureContainsOthersSoItMustRemainA')}</p>{/if}
      <WorkflowSelect id={`fixture-parent-${asset.id}`} error={errors[`fixture-parent-${asset.id}`]} label={t('web.CaseFixtures.inside')} value={asset.parentId} {disabled}
        options={[{ value: '', label: t("web.options.CaseFixtures.noParent") }, ...fixtureParentChoices(value.assets, asset.id).map(parent => ({ value: parent.id, label: parent.title || t("web.options.CaseFixtures.unnamedFixture") }))]}
        onChange={parentId => update(asset.id, { parentId })} />
      <Label.Root class="grid gap-2 text-sm">{t('web.CaseFixtures.description')}<Textarea.Root name={`fixture-description-${asset.id}`} id={`fixture-description-${asset.id}`} {...validationAttributes(errors[`fixture-description-${asset.id}`], `fixture-description-${asset.id}`)} value={asset.description} rows={2} oninput={event => update(asset.id, { description: event.currentTarget.value })} /></Label.Root><ValidationMessage field={`fixture-description-${asset.id}`} message={errors[`fixture-description-${asset.id}`]} />
      <Label.Root class="grid gap-2 text-sm">{t('web.CaseFixtures.tagsOnePerLine')}<Textarea.Root name={`fixture-tags-${asset.id}`} id={`fixture-tags-${asset.id}`} {...validationAttributes(errors[`fixture-tags-${asset.id}`], `fixture-tags-${asset.id}`)} value={asset.tagNames.join('\n')} rows={3}
        oninput={event => update(asset.id, { tagNames: event.currentTarget.value.split('\n') })} /></Label.Root><ValidationMessage field={`fixture-tags-${asset.id}`} message={errors[`fixture-tags-${asset.id}`]} />
      <Button.Root type="button" variant="outline" disabled={disabled || removalBlocked} onclick={() => remove(asset.id)}>{t('web.CaseFixtures.removeFixture')}</Button.Root>
      {#if removalBlocked}<p>{t('web.CaseFixtures.changeReferencesToThisFixtureInContainmentOrExpected')}</p>{/if}
    </fieldset>
  {/each}
  <Button.Root type="button" variant="outline" disabled={disabled || value.assets.length >= 100} onclick={add}>{t('web.CaseFixtures.addTestItem')}</Button.Root>
  {#if value.assets.length >= 100}<p>{t('web.CaseFixtures.aTestCaseSupportsUpTo100Fixtures')}</p>{/if}
</section>
<style>
  .case-fixtures, fieldset { display: grid; gap: 1rem; }
  fieldset { min-width: 0; border: 1px solid var(--border); border-radius: var(--radius); padding: 1rem; }
  legend, h3 { font-weight: 600; } legend { overflow-wrap: anywhere; }
  p { font-size: .9rem; color: var(--muted-foreground); }
</style>
